package driver

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/fimreal/goutils/ezap"
	"github.com/fimreal/tencent-cos-csi-driver/pkg/cos"
	"github.com/fimreal/tencent-cos-csi-driver/pkg/mounter"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/utils/mount"
)

type nodeServer struct {
	*csicommon.DefaultNodeServer
}

func newNodeServer(d *csicommon.CSIDriver) *nodeServer {
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d),
	}
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()
	stagingTargetPath := req.GetStagingTargetPath()
	bucketName, prefix := volumeIDToBucketPrefix(volumeID)

	// Check arguments
	if req.GetVolumeCapability() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(stagingTargetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Staging Target path missing in request")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	notMnt, err := checkMount(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	// TODO: Implement readOnly & mountFlags
	readOnly := req.GetReadonly()
	mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()
	attrib := req.GetVolumeContext()

	ezap.Infof("target %v\nreadonly %v\nvolumeId %v\nattributes %v\nmountflags %v\n",
		targetPath, readOnly, volumeID, attrib, mountFlags)

	cos, err := cos.NewClientFromSecret(req.GetSecrets())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %s", err)
	}

	meta := getMeta(bucketName, prefix, req.VolumeContext)
	mounter, err := mounter.New(meta, cos.Config)
	if err != nil {
		return nil, err
	}
	if err := mounter.Mount(stagingTargetPath, targetPath); err != nil {
		return nil, err
	}

	ezap.Infof("volume %s successfully mounted to %s", volumeID, targetPath)

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	targetPath := req.GetTargetPath()

	// Check arguments
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(targetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	if err := mounter.FuseUnmount(targetPath); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	ezap.Infof("volume %s has been unmounted.", volumeID)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	stagingTargetPath := req.GetStagingTargetPath()
	bucketName, prefix := volumeIDToBucketPrefix(volumeID)

	// Check arguments
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	if len(stagingTargetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	if req.VolumeCapability == nil {
		return nil, status.Error(codes.InvalidArgument, "NodeStageVolume Volume Capability must be provided")
	}

	notMnt, err := checkMount(stagingTargetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if !notMnt {
		return &csi.NodeStageVolumeResponse{}, nil
	}
	client, err := cos.NewClientFromSecret(req.GetSecrets())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %s", err)
	}

	meta := getMeta(bucketName, prefix, req.VolumeContext)
	mounter, err := mounter.New(meta, client.Config)
	if err != nil {
		return nil, err
	}
	if err := mounter.Stage(stagingTargetPath); err != nil {
		return nil, err
	}

	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	stagingTargetPath := req.GetStagingTargetPath()

	// Check arguments
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if len(stagingTargetPath) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	return &csi.NodeUnstageVolumeResponse{}, nil
}

// NodeGetCapabilities returns the supported capabilities of the node server
func (ns *nodeServer) NodeGetCapabilities(ctx context.Context, req *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {
	// currently there is a single NodeServer capability according to the spec
	nscap := &csi.NodeServiceCapability{
		Type: &csi.NodeServiceCapability_Rpc{
			Rpc: &csi.NodeServiceCapability_RPC{
				Type: csi.NodeServiceCapability_RPC_STAGE_UNSTAGE_VOLUME,
			},
		},
	}

	return &csi.NodeGetCapabilitiesResponse{
		Capabilities: []*csi.NodeServiceCapability{
			nscap,
		},
	}, nil
}

func (ns *nodeServer) NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (*csi.NodeExpandVolumeResponse, error) {
	return &csi.NodeExpandVolumeResponse{}, status.Error(codes.Unimplemented, "NodeExpandVolume is not implemented")
}

func getMeta(bucketName, prefix string, context map[string]string) *cos.FSMeta {
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
	return &cos.FSMeta{
		// BucketName:    bucketName,
		Prefix:        prefix,
		Mounter:       context[mounter.TypeKey],
		MountOptions:  mountOptions,
		CapacityBytes: capacity,
	}
}

// volumeIDToBucketPrefix returns the bucket name and prefix based on the volumeID.
// Prefix is empty if volumeID does not have a slash in the name.
func volumeIDToBucketPrefix(volumeID string) (string, string) {
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
