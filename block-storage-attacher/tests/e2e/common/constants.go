package common

// busyboxTestImage is the list of images used in common test. These images should be prepulled
// before a tests starts, so that the tests won't fail due image pulling flakes.
const (
	BusyboxTestImage     = "redis"
	PluginName           = "ibm.io/ibmc-blockattacher"
	MountPath            = "/host/etc"
	NamespaceName        = "block-volume-attacher-e2e-namespace"
	ProvisionerPodName   = ""
	DeploymentName       = "block-attacher"
	ClaimPrefix          = "armada-portworx-"
	PluginImage          = "armada-master/armada-block-volume-attacher:latest"
	VolumeMountPath      = "/host"
	DeploymentGOFilePATH = "src/github.ibm.com/alchemy-containers/armada-storage-e2e/deploy/kube-config/deployment.yaml"
)
