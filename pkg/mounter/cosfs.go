package mounter

import (
	"fmt"
	"os"

	"github.com/fimreal/os-csi/pkg/cos"
)

// Implements Mounter
type cosfsMounter struct {
	meta          *cos.FSMeta
	url           string
	bucket        string
	pwFileContent string
}

const (
	cosfsCmd = "cosfs"
)

func newCosfsMounter(meta *cos.FSMeta, cfg *cos.Config) (Mounter, error) {
	return &cosfsMounter{
		meta:          meta,
		url:           cfg.Endpoint,
		bucket:        meta.BucketName,
		pwFileContent: meta.BucketName + ":" + cfg.AccessKeyID + ":" + cfg.SecretAccessKey,
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
		// "-d",
		"-o", "allow_other",
		// "-o", "del_cache",
		"-o", "noxattr",
		"-o", "dbglevel=info",
		"-o", fmt.Sprintf("url=%s", cosfs.url),
	}
	args = append(args, cosfs.meta.MountOptions...)
	fmt.Println("cmd: ", cosfsCmd, "\nargs: ", args, "\ntarget: ", target)
	return fuseMount(target, cosfsCmd, args)
}

func writecosfsPass(pwFileContent string) error {
	// pwFileName := "/etc/passwd-cosfs"
	pwFileName := fmt.Sprintf("%s/.passwd-cosfs", os.Getenv("HOME"))
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
