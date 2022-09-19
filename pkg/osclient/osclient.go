/*
为实现 controllerserver（创建，删除 volume 目录）

osclient 使用对应服务 sdk 连接对象存储服务器

黑箱不同用法，cos 和 oss 类似，需要提前指定 bucket 名字创建连接
*/
package osclient

import (
	"errors"

	"github.com/fimreal/os-csi/pkg/mounter"
	"github.com/fimreal/os-csi/pkg/osclient/cos"
	"github.com/fimreal/os-csi/pkg/osclient/oss"
)

type Client interface {
	BucketExists() (bool, error)
	// CreateBucket(bucketName string) error
	CreatePrefix(prefix string) error
	RemovePrefix(prefix string) error
	GetConfig() *mounter.Config
}

func NewFromSecret(secret map[string]string) (Client, error) {

	mounterType := secret["mounter"]
	cfg := &mounter.Config{
		AccessKeyID:     secret["accessKeyID"],
		SecretAccessKey: secret["secretAccessKey"],
		Endpoint:        secret["endpoint"],
		BucketName:      secret["bucketName"],
		Mounter:         mounterType,
	}

	switch mounterType {
	case mounter.CosfsMounterType:
		return cos.NewClient(cfg)
	case mounter.OssfsMounterType:
		return oss.NewClient(cfg)
	default:
		return nil, errors.New("not supported mounterType: " + mounterType)
	}
}

func New(cfg *mounter.Config) (Client, error) {

	mounterType := cfg.Mounter

	switch mounterType {
	case mounter.CosfsMounterType:
		return cos.NewClient(cfg)
	case mounter.OssfsMounterType:
		return oss.NewClient(cfg)
	default:
		return nil, errors.New("not supported mounterType: " + mounterType)
	}
}
