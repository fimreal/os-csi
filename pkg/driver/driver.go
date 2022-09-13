package driver

import (
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/fimreal/goutils/ezap"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
)

type driver struct {
	driver   *csicommon.CSIDriver
	endpoint string

	// ids *identityServer
	// ns  *nodeServer
	// cs  *controllerServer
}

var (
	driverVersion = "v0.1.0"
	driverName    = "csi.cosfs"
)

// New initializes the driver
func NewDriver(endpoint string, nodeID string) (*driver, error) {
	csiDriver := csicommon.NewCSIDriver(driverName, driverVersion, nodeID)
	if csiDriver == nil {
		ezap.Fatal("Failed to initialize CSI Driver.")
	}

	d := &driver{
		endpoint: endpoint,
		driver:   csiDriver,
	}
	return d, nil
}

func (d *driver) Start() {
	ezap.Infof("Driver: %v ", driverName)
	ezap.Infof("Version: %v ", driverVersion)
	// Initialize default library driver

	d.driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME})
	d.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER})

	s := csicommon.NewNonBlockingGRPCServer()
	s.Start(d.endpoint, newIdentityServer(d.driver), newControllerServer(d.driver), newNodeServer(d.driver))
	s.Wait()
}
