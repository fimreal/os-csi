package driver

import (
	"context"
	"fmt"
	"path"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/fimreal/goutils/ezap"
	"github.com/fimreal/os-csi/pkg/mounter"
	"github.com/fimreal/os-csi/pkg/osclient"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
}

func newControllerServer(d *csicommon.CSIDriver) *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
	}
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	params := req.Parameters
	capacityBytes := req.CapacityRange.RequiredBytes
	bucketName := params[mounter.BucketKey]
	volumeID := sanitizeVolumeID(req.Name)

	// Check arguments
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "volume name is empty")
	}
	if len(req.VolumeCapabilities) <= 0 {
		return nil, status.Error(codes.InvalidArgument, "volume has no capabilities")
	}
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		ezap.Errorf("invalid create volume req: %v", req)
		return nil, err
	}

	ezap.Infof("Got a request to create volume %s", volumeID)

	// 使用 sdk 构造连接
	secret := req.Secrets
	cfg := &mounter.Config{
		AccessKeyID:     secret["accessKeyID"],
		SecretAccessKey: secret["secretAccessKey"],
		Endpoint:        secret["endpoint"],
		BucketName:      bucketName,
		Mounter:         secret["mounter"],
		// Meta: &mounter.FSMeta{
		// 	Prefix:        volumeID,
		// 	MountOptions:  []string{},
		// 	CapacityBytes: capacityBytes,
		// }
	}
	client, err := osclient.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize osclient: %s", err)
	}

	// 检查 bucket 是否存在
	exists, err := client.BucketExists()
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket %s exists: %v", bucketName, err)
	} else if !exists {
		return nil, fmt.Errorf("not find bucket %s", bucketName)
	}

	// 尝试以 pvc 名称创建目录
	if err = client.CreatePrefix(volumeID); err != nil {
		return nil, fmt.Errorf("failed to create prefix dir %s: %v", volumeID, err)
	}

	ezap.Infof("create volume %s", volumeID)
	params["capacity"] = fmt.Sprintf("%v", capacityBytes)

	response := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      path.Join(bucketName, volumeID),
			CapacityBytes: capacityBytes,
			VolumeContext: params,
		}}

	ezap.Infof("Success create Volume: %s, Size: %d", volumeID, capacityBytes)
	return response, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	volumeID := req.VolumeId
	bucketName, prefix := VolumeIDToBucketPrefix(volumeID)

	// Check arguments
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		ezap.Infof("Invalid delete volume req: %v", req)
		return nil, err
	}
	ezap.Infof("Deleting volume %s", prefix)

	// 使用 sdk 构造连接
	secret := req.Secrets
	cfg := &mounter.Config{
		AccessKeyID:     secret["accessKeyID"],
		SecretAccessKey: secret["secretAccessKey"],
		Endpoint:        secret["endpoint"],
		BucketName:      bucketName,
		Mounter:         secret["mounter"],
	}
	client, err := osclient.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize osclient: %s", err)
	}

	if err := client.RemovePrefix(prefix); err != nil {
		return nil, fmt.Errorf("unable to remove prefix dir: %w", err)
	}
	ezap.Infof("Prefix dir %s removed", prefix)

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	volumeID := req.VolumeId
	bucketName, _ := VolumeIDToBucketPrefix(volumeID)

	// Check arguments
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "volumeID missing in request")
	}
	if len(req.VolumeCapabilities) <= 0 {
		return nil, status.Error(codes.InvalidArgument, "volume has no capabilities")
	}

	// 使用 sdk 构造连接
	secret := req.Secrets
	cfg := &mounter.Config{
		AccessKeyID:     secret["accessKeyID"],
		SecretAccessKey: secret["secretAccessKey"],
		Endpoint:        secret["endpoint"],
		BucketName:      bucketName,
		Mounter:         secret["mounter"],
	}
	client, err := osclient.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize osclient: %s", err)
	}

	exists, err := client.BucketExists()
	if err != nil {
		return nil, err
	}

	if !exists {
		// return an error if the bucket of the requested volume does not exist
		return nil, status.Error(codes.NotFound, "bucket of volume with id "+volumeID+" does not exist")
	}

	supportedAccessMode := &csi.VolumeCapability_AccessMode{
		Mode: csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER,
	}

	for _, capability := range req.VolumeCapabilities {
		if capability.GetAccessMode().GetMode() != supportedAccessMode.GetMode() {
			return &csi.ValidateVolumeCapabilitiesResponse{Message: "Only single node writer is supported"}, nil
		}
	}

	return &csi.ValidateVolumeCapabilitiesResponse{
		Confirmed: &csi.ValidateVolumeCapabilitiesResponse_Confirmed{
			VolumeCapabilities: []*csi.VolumeCapability{
				{
					AccessMode: supportedAccessMode,
				},
			},
		},
	}, nil
}

func (cs *controllerServer) ControllerGetVolume(_ context.Context, _ *csi.ControllerGetVolumeRequest) (*csi.ControllerGetVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented ControllerGetVolume")
}

func (cs *controllerServer) ControllerExpandVolume(_ context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ControllerExpandVolume is not implemented")
}

func (cs *controllerServer) ControllerPublishVolume(_ context.Context, _ *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented ControllerPublishVolume")
}

func (cs *controllerServer) ControllerUnpublishVolume(_ context.Context, _ *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented ControllerUnpublishVolume")
}

func (cs *controllerServer) ListVolumes(_ context.Context, _ *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented ListVolumes")
}

func (cs *controllerServer) GetCapacity(_ context.Context, _ *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Unimplemented GetCapacity")
}
