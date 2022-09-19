package cos

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/fimreal/goutils/ezap"
	"github.com/fimreal/os-csi/pkg/mounter"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/tencentyun/cos-go-sdk-v5/debug"
)

type Client struct {
	Config *mounter.Config
	cos    *cos.Client
	ctx    context.Context
}

func NewClient(cfg *mounter.Config) (*Client, error) {

	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}
	u.Host = cfg.BucketName + "." + u.Host

	cosClient := cos.NewClient(
		&cos.BaseURL{BucketURL: u},
		&http.Client{
			Transport: &cos.AuthorizationTransport{
				SecretID:  cfg.AccessKeyID,
				SecretKey: cfg.SecretAccessKey,
				Transport: &debug.DebugRequestTransport{
					RequestHeader:  true,
					RequestBody:    true,
					ResponseHeader: true,
					ResponseBody:   true,
				},
			},
		})

	return &Client{
		Config: cfg,
		cos:    cosClient,
		ctx:    context.Background(),
	}, nil
}

func (c *Client) BucketExists() (bool, error) {
	return c.cos.Bucket.IsExist(c.ctx)
}

func (c *Client) CreatePrefix(prefix string) error {
	if prefix == "" {
		return nil
	}
	_, err := c.cos.Object.Put(c.ctx, prefix+"/", strings.NewReader(""), nil)
	return err
}

func (c *Client) RemovePrefix(prefix string) error {
	var err error
	// 尝试直接删除文件夹
	if _, err = c.cos.Object.Delete(c.ctx, prefix+"/"); err == nil {
		return nil
	}
	ezap.Warnf("removeObjects failed with: %s, will try removeObjectsOneByOne", err)

	// 批量删除文件夹及下面文件
	var marker string
	opt := &cos.BucketGetOptions{
		Prefix:  prefix + "/",
		MaxKeys: 1000,
	}
	isTruncated := true
	for isTruncated {
		opt.Marker = marker
		v, _, err := c.cos.Bucket.Get(c.ctx, opt)
		if err != nil {
			return err
		}
		for _, content := range v.Contents {
			_, err = c.cos.Object.Delete(c.ctx, content.Key)
			if err != nil {
				return err
			}
		}
		isTruncated = v.IsTruncated
		marker = v.NextMarker
	}

	return err
}

func (c *Client) GetConfig() *mounter.Config {
	return c.Config
}
