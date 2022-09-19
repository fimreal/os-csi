package oss

import (
	"bytes"
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/fimreal/os-csi/pkg/mounter"
)

type Client struct {
	Config *mounter.Config
	oss    *oss.Client
	// ctx    context.Context
}

func NewClient(cfg *mounter.Config) (*Client, error) {

	ossClient, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.SecretAccessKey, oss.Timeout(10, 120))
	return &Client{
		Config: cfg,
		oss:    ossClient,
	}, err
}

func (c *Client) BucketExists() (bool, error) {
	return c.oss.IsBucketExist(c.Config.BucketName)
}

func (c *Client) CreatePrefix(prefix string) error {
	if prefix == "" {
		return nil
	}
	b, err := c.oss.Bucket(c.Config.BucketName)
	if err != nil {
		return err
	}
	return b.PutObject(prefix+"/", bytes.NewReader([]byte("")))
}

func (c *Client) RemovePrefix(prefix string) error {
	b, err := c.oss.Bucket(c.Config.BucketName)
	if err != nil {
		return err
	}
	marker := oss.Marker("")
	optPrefix := oss.Prefix(prefix)
	count := 0
	for {
		lor, err := b.ListObjects(marker, optPrefix)
		if err != nil {
			return err
		}

		objects := []string{}
		for _, object := range lor.Objects {
			objects = append(objects, object.Key)
		}

		// 删除目录及目录下所有文件
		delRes, err := b.DeleteObjects(objects, oss.DeleteObjectsQuiet(true))
		if err != nil {
			return err
		}

		if len(delRes.DeletedObjects) > 0 {
			return fmt.Errorf("delete object failure, %v", delRes.DeletedObjects)
		}
		count += len(objects)
		optPrefix = oss.Prefix(lor.Prefix)
		marker = oss.Marker(lor.NextMarker)
		if !lor.IsTruncated {
			break
		}
	}
	return nil
}

func (c *Client) GetConfig() *mounter.Config {
	return c.Config
}
