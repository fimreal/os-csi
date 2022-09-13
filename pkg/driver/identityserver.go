package driver

import csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"

type identityServer struct {
	*csicommon.DefaultIdentityServer
}

// newIdentityServer create identity server
func newIdentityServer(d *csicommon.CSIDriver) *identityServer {
	return &identityServer{
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d),
	}
}
