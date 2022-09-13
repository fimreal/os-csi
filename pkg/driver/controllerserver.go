package driver

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/fimreal/goutils/ezap"
	"github.com/fimreal/tencent-cos-csi-driver/pkg/cos"
	"github.com/fimreal/tencent-cos-csi-driver/pkg/mounter"
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
	params := req.GetParameters()
	capacityBytes := req.GetCapacityRange().GetRequiredBytes()
	volumeID := sanitizeVolumeID(req.GetName())
	bucketName := volumeID
	prefix := ""

	// check if bucket name is overridden
	if params[mounter.BucketKey] != "" {
		bucketName = params[mounter.BucketKey]
		prefix = volumeID
		volumeID = path.Join(bucketName, prefix)
	}

	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		ezap.Errorf("invalid create volume req: %v", req)
		return nil, err
	}

	// Check arguments
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Name missing in request")
	}
	if req.GetVolumeCapabilities() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities missing in request")
	}

	ezap.Infof("Got a request to create volume %s", volumeID)

	client, err := cos.NewClientFromSecret(req.GetSecrets())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cos client: %s", err)
	}

	exists, err := client.BucketExists()
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket %s exists: %v", bucketName, err)
	} else if !exists {
		return nil, fmt.Errorf("failed to find bucket %s", bucketName)
	}

	if err = client.CreatePrefix(prefix); err != nil {
		return nil, fmt.Errorf("failed to create prefix dir %s: %v", prefix, err)
	}

	ezap.Infof("create volume %s", volumeID)
	params["capacity"] = fmt.Sprintf("%v", capacityBytes)

	response := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			VolumeId:      volumeID,
			CapacityBytes: capacityBytes,
			VolumeContext: params,
		}}

	ezap.Infof("Success create Volume: %s, Size: %d", volumeID, capacityBytes)
	return response, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	_, prefix := volumeIDToBucketPrefix(volumeID)

	// Check arguments
	if len(volumeID) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		ezap.Infof("Invalid delete volume req: %v", req)
		return nil, err
	}
	ezap.Infof("Deleting volume %s", volumeID)

	client, err := cos.NewClientFromSecret(req.GetSecrets())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %s", err)
	}

	if err := client.RemovePrefix(prefix); err != nil {
		return nil, fmt.Errorf("unable to remove prefix dir: %w", err)
	}
	ezap.Infof("Prefix dir %s removed", prefix)

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) ValidateVolumeCapabilities(ctx context.Context, req *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {
	// Check arguments
	if len(req.GetVolumeId()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}
	if req.GetVolumeCapabilities() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capabilities missing in request")
	}

	client, err := cos.NewClientFromSecret(req.GetSecrets())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize client: %s", err)
	}
	exists, err := client.BucketExists()
	if err != nil {
		return nil, err
	}

	if !exists {
		// return an error if the bucket of the requested volume does not exist
		return nil, status.Error(codes.NotFound, fmt.Sprintf("bucket of volume with id %s does not exist", req.GetVolumeId()))
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

func sanitizeVolumeID(volumeID string) string {
	volumeID = strings.ToLower(volumeID)
	if len(volumeID) > 63 {
		h := sha1.New()
		io.WriteString(h, volumeID)
		volumeID = hex.EncodeToString(h.Sum(nil))
	}
	return volumeID
}
