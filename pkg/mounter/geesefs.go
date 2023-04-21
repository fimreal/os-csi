package mounter

import (
	"os"

	"github.com/fimreal/goutils/ezap"
)

// Implements Mounter
type geesefsMounter struct {
	meta            *FSMeta
	endpoint        string
	bucket          string
	region          string
	accessKeyID     string
	secretAccessKey string
	// pwFileContent string
}

func newGeeseFSMounter(cfg *Config) (Mounter, error) {
	return &geesefsMounter{
		meta:            cfg.Meta,
		endpoint:        cfg.Endpoint,
		bucket:          cfg.BucketName,
		region:          cfg.Region,
		accessKeyID:     cfg.AccessKeyID,
		secretAccessKey: cfg.SecretAccessKey,
		// pwFileContent: cfg.BucketName + ":" + cfg.AccessKeyID + ":" + cfg.SecretAccessKey,
	}, nil
}

func (*geesefsMounter) Stage(stageTarget string) error {
	return nil
}
func (*geesefsMounter) Unstage(stageTarget string) error {
	return nil
}

func (geesefs *geesefsMounter) Mount(source string, target string) error {
	os.Setenv("AWS_ACCESS_KEY_ID", geesefs.accessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", geesefs.secretAccessKey)

	args := []string{
		"--endpoint", geesefs.endpoint,
		"-o", "allow_other",
		"--log-file", "/dev/stderr",
	}
	if geesefs.region != "" {
		args = append(args, "--region", geesefs.region)
	}
	args = append(args,
		geesefs.bucket+":/"+geesefs.meta.Prefix,
		target,
	)
	args = append(args, geesefs.meta.MountOptions...)
	ezap.Info("cmd: ", GeeseFsCmd, ", args: ", args, ", target: ", target)
	return fuseMount(target, GeeseFsCmd, args)
}
