package driver

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/fimreal/os-csi/pkg/mounter"
	"k8s.io/utils/mount"
)

func getMeta(prefix string, context map[string]string) *mounter.FSMeta {
	mountOptions := make([]string, 0)
	mountOptStr := context[mounter.OptionsKey]
	if mountOptStr != "" {
		re, _ := regexp.Compile(`([^\s"]+|"([^"\\]+|\\")*")+`)
		re2, _ := regexp.Compile(`"([^"\\]+|\\")*"`)
		re3, _ := regexp.Compile(`\\(.)`)
		for _, opt := range re.FindAll([]byte(mountOptStr), -1) {
			// Unquote options
			opt = re2.ReplaceAllFunc(opt, func(q []byte) []byte {
				return re3.ReplaceAll(q[1:len(q)-1], []byte("$1"))
			})
			mountOptions = append(mountOptions, string(opt))
		}
	}
	capacity, _ := strconv.ParseInt(context["capacity"], 10, 64)
	return &mounter.FSMeta{
		// BucketName: bucketName,
		Prefix:     prefix,
		// Mounter:       context[mounter.TypeKey],
		MountOptions:  mountOptions,
		CapacityBytes: capacity,
	}
}

// volumeIDToBucketPrefix returns the bucket name and prefix based on the volumeID.
// Prefix is empty if volumeID does not have a slash in the name.
func VolumeIDToBucketPrefix(volumeID string) (string, string) {
	// if the volumeID has a slash in it, this volume is
	// stored under a certain prefix within the bucket.
	splitVolumeID := strings.SplitN(volumeID, "/", 2)
	if len(splitVolumeID) > 1 {
		return splitVolumeID[0], splitVolumeID[1]
	}

	return volumeID, ""
}

func checkMount(targetPath string) (bool, error) {
	notMnt, err := mount.New("").IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(targetPath, 0750); err != nil {
				return false, err
			}
			notMnt = true
		} else {
			return false, err
		}
	}
	return notMnt, nil
}

func sanitizeVolumeID(volumeID string) string {
	volumeID = strings.ToLower(volumeID)
	if len(volumeID) > 63 {
		h := sha1.New()
		io.WriteString(h, volumeID)
		volumeID = hex.EncodeToString(h.Sum(nil))
	}
	return volumeID
}

// func GenConfig(req *csi.CreateVolumeRequest) *mounter.Config {
// 	params := req.Parameters
// 	capacityBytes := req.CapacityRange.RequiredBytes
// 	bucketName := params[mounter.BucketKey]
// 	volumeID := sanitizeVolumeID(req.Name)
// 	secret := req.Secrets

// return &mounter.Config{
// 	AccessKeyID:     secret["accessKeyID"],
// 	SecretAccessKey: secret["secretAccessKey"],
// 	Endpoint:        secret["endpoint"],
// 	BucketName:      bucketName,
// 	Mounter:         secret["mounter"],
// 	Meta: &mounter.FSMeta{
// 		BucketName:    bucketName,
// 		Prefix:        volumeID,
// 		MountOptions:  []string{},
// 		CapacityBytes: capacityBytes,
// 	},
// 	}
// }
