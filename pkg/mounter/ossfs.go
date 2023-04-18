package mounter

import (
	"os"

	"github.com/fimreal/goutils/ezap"
)

// Implements Mounter
type ossfsMounter struct {
	meta          *FSMeta
	url           string
	bucket        string
	pwFileContent string
}

func newOssfsMounter(cfg *Config) (Mounter, error) {
	return &ossfsMounter{
		meta:          cfg.Meta,
		url:           cfg.Endpoint,
		bucket:        cfg.BucketName,
		pwFileContent: cfg.BucketName + ":" + cfg.AccessKeyID + ":" + cfg.SecretAccessKey,
	}, nil
}

func (ossfs *ossfsMounter) Stage(stageTarget string) error {
	return nil
}

func (ossfs *ossfsMounter) Unstage(stageTarget string) error {
	return nil
}

func (ossfs *ossfsMounter) Mount(source string, target string) error {
	if err := writeossfsPass(ossfs.pwFileContent); err != nil {
		return err
	}
	args := []string{
		ossfs.bucket + ":/" + ossfs.meta.Prefix,
		target,
		"-o", "allow_other",
		"-o", "noxattr",
		"-o", "dbglevel=info",
		"-o", "url=" + ossfs.url,
	}
	args = append(args, ossfs.meta.MountOptions...)
	ezap.Info("cmd: ", OssfsCmd, ", args: ", args, ", target: ", target)
	return fuseMount(target, OssfsCmd, args)
}

func writeossfsPass(pwFileContent string) error {
	pwFileName := "/etc/passwd-ossfs"
	pwFile, err := os.OpenFile(pwFileName, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	_, err = pwFile.WriteString(pwFileContent)
	if err != nil {
		return err
	}
	pwFile.Close()
	return nil
}
