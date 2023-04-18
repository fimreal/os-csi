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
	"github.com/fimreal/os-csi/pkg/osclient/s3"
)

type Client interface {
	BucketExists() (bool, error)
	CreatePrefix(prefix string) error
	RemovePrefix(prefix string) error
	GetConfig() *mounter.Config
}

func New(cfg *mounter.Config) (Client, error) {

	mounterType := cfg.Mounter

	switch mounterType {
	case mounter.CosfsCmd:
		return cos.NewClient(cfg)
	case mounter.OssfsCmd:
		return oss.NewClient(cfg)
	case mounter.GeeseFsCmd, mounter.S3FsCmd:
		return s3.NewClient(cfg)
	default:
		return nil, errors.New("not supported mounterType: " + mounterType)
	}
}
