// goosefs lite
// for tencent cos
package mounter

import (
	"fmt"
	"os"

	"github.com/fimreal/os-csi/pkg/cos"
)

// Implements Mounter
type goosefsMounter struct {
	meta          *cos.FSMeta
	url           string
	bucket        string
	pwFileContent string
}

const (
	goosefsCmd = "goosefs-lite"
)

func newGoosefsMounter(meta *cos.FSMeta, cfg *cos.Config) (Mounter, error) {
	return &goosefsMounter{
		meta:          meta,
		url:           cfg.Endpoint,
		bucket:        meta.BucketName,
		pwFileContent: meta.BucketName + ":" + cfg.AccessKeyID + ":" + cfg.SecretAccessKey,
	}, nil
}

func (goosefs *goosefsMounter) Stage(stageTarget string) error {
	return nil
}

func (goosefs *goosefsMounter) Unstage(stageTarget string) error {
	return nil
}

func (goosefs *goosefsMounter) Mount(source string, target string) error {
	if err := writegoosefsPass(goosefs.pwFileContent); err != nil {
		return err
	}
	args := []string{
		"-o", "dbglevel=info",
		"-o", fmt.Sprintf("url=%s", goosefs.url),
		"cosn://" + goosefs.bucket + ":/" + goosefs.meta.Prefix,
		target,
	}
	args = append(args, goosefs.meta.MountOptions...)
	fmt.Println("cmd: ", goosefsCmd, "\nargs: ", args, "\ntarget: ", target)
	return fuseMount(target, goosefsCmd, args)
}

func writegoosefsPass(pwFileContent string) error {
	pwFileName := fmt.Sprintf("%s/.passwd-goosefs", os.Getenv("HOME"))
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
