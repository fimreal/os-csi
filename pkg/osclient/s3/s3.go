package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/fimreal/goutils/ezap"
	"github.com/fimreal/os-csi/pkg/mounter"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	Config *mounter.Config
	minio  *minio.Client
	ctx    context.Context
}

func NewClient(cfg *mounter.Config) (*Client, error) {
	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	ssl := u.Scheme == "https"

	minioClient, err := minio.New(u.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.Region),
		Secure: ssl,
	})
	return &Client{
		Config: cfg,
		minio:  minioClient,
		ctx:    context.Background(),
	}, err
}

func (c *Client) BucketExists() (bool, error) {
	return c.minio.BucketExists(c.ctx, c.Config.BucketName)
}

func (c *Client) CreatePrefix(prefix string) error {
	if prefix == "" {
		return nil
	}
	_, err := c.minio.PutObject(c.ctx, c.Config.BucketName, prefix+"/", bytes.NewReader([]byte("")), 0, minio.PutObjectOptions{})
	return err
}

// func (c *Client) CreateBucket(bucketName string) error {
// 	return c.minio.MakeBucket(c.ctx, bucketName, minio.MakeBucketOptions{Region: c.Config.Region})
// }

func (c *Client) GetConfig() *mounter.Config {
	return c.Config
}

func (c *Client) RemovePrefix(prefix string) (err error) {

	if err = c.removeObjects(c.Config.BucketName, prefix); err == nil {
		return c.minio.RemoveObject(c.ctx, c.Config.BucketName, prefix, minio.RemoveObjectOptions{})
	}

	ezap.Warnf("removeObjects failed with: %s, will try removeObjectsOneByOne", err)

	if err = c.removeObjectsOneByOne(c.Config.BucketName, prefix); err == nil {
		return c.minio.RemoveObject(c.ctx, c.Config.BucketName, prefix, minio.RemoveObjectOptions{})
	}

	return err
}

// https://github.com/minio/minio-go/blob/master/examples/s3/removeobjects.go
func (c *Client) removeObjects(bucketname, prefix string) error {
	objectsCh := make(chan minio.ObjectInfo)
	var listErr error

	go func() {
		defer close(objectsCh)

		for object := range c.minio.ListObjects(
			c.ctx,
			bucketname,
			minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
			if object.Err != nil {
				listErr = object.Err
				return
			}
			objectsCh <- object
		}
	}()

	if listErr != nil {
		ezap.Error("Error listing objects", listErr)
		return listErr
	}

	select {
	default:
		errorCh := c.minio.RemoveObjects(c.ctx, bucketname, objectsCh, minio.RemoveObjectsOptions{GovernanceBypass: true})
		haveErrWhenRemoveObjects := false
		for e := range errorCh {
			ezap.Errorf("Failed to remove object %s, error: %s", e.ObjectName, e.Err)
			haveErrWhenRemoveObjects = true
		}
		if haveErrWhenRemoveObjects {
			return errors.New("Failed to remove all objects of bucket " + bucketname)
		}
	}

	return nil
}

// will delete files one by one without file lock
func (client *Client) removeObjectsOneByOne(bucketName, prefix string) error {
	parallelism := 16
	objectsCh := make(chan minio.ObjectInfo, 1)
	guardCh := make(chan int, parallelism)
	var listErr error
	totalObjects := 0
	removeErrors := 0

	go func() {
		defer close(objectsCh)

		for object := range client.minio.ListObjects(client.ctx, bucketName,
			minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
			if object.Err != nil {
				listErr = object.Err
				return
			}
			totalObjects++
			objectsCh <- object
		}
	}()

	if listErr != nil {
		ezap.Error("Error listing objects", listErr)
		return listErr
	}

	for object := range objectsCh {
		guardCh <- 1
		go func() {
			err := client.minio.RemoveObject(client.ctx, bucketName, object.Key,
				minio.RemoveObjectOptions{VersionID: object.VersionID})
			if err != nil {
				ezap.Errorf("Failed to remove object %s, error: %s", object.Key, err)
				removeErrors++
			}
			<-guardCh
		}()
	}
	for i := 0; i < parallelism; i++ {
		guardCh <- 1
	}
	for i := 0; i < parallelism; i++ {
		<-guardCh
	}

	if removeErrors > 0 {
		return fmt.Errorf("failed to remove %v objects out of total %v of path %s", removeErrors, totalObjects, bucketName)
	}

	return nil
}
