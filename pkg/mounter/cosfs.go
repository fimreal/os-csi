package mounter

import (
	"os"

	"github.com/fimreal/goutils/ezap"
)

const (
	CosfsCmd         = "cosfs"
	CosfsMounterType = "cosfs"
)

// Implements Mounter
type cosfsMounter struct {
	meta          *FSMeta
	url           string
	bucket        string
	pwFileContent string
}

func newCosfsMounter(cfg *Config) (Mounter, error) {
	return &cosfsMounter{
		meta:          cfg.Meta,
		url:           cfg.Endpoint,
		bucket:        cfg.BucketName,
		pwFileContent: cfg.BucketName + ":" + cfg.AccessKeyID + ":" + cfg.SecretAccessKey,
	}, nil
}

func (cosfs *cosfsMounter) Stage(stageTarget string) error {
	return nil
}

func (cosfs *cosfsMounter) Unstage(stageTarget string) error {
	return nil
}

func (cosfs *cosfsMounter) Mount(source string, target string) error {
	if err := writecosfsPass(cosfs.pwFileContent); err != nil {
		return err
	}
	args := []string{
		cosfs.bucket + ":/" + cosfs.meta.Prefix,
		target,
		"-o", "allow_other",
		"-o", "noxattr",
		"-o", "dbglevel=info",
		"-o", "url=" + cosfs.url,
	}
	args = append(args, cosfs.meta.MountOptions...)
	ezap.Info("cmd: ", CosfsCmd, "\nargs: ", args, "\ntarget: ", target)
	return fuseMount(target, CosfsCmd, args)
}

func writecosfsPass(pwFileContent string) error {
	pwFileName := "/etc/passwd-cosfs"
	// pwFileName := fmt.Sprintf("%s/.passwd-cosfs", os.Getenv("HOME"))
	pwFile, err := os.OpenFile(pwFileName, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer pwFile.Close()

	_, err = pwFile.WriteString(pwFileContent)
	if err != nil {
		return err
	}
	return nil
}
