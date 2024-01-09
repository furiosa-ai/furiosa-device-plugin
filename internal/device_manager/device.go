package device_manager

type DevicePolicy string

type DeviceInfo interface {
	DeviceID() string
	PCIBusID() string
	NUMANode() int
	IsHealthy() (bool, error)
	IsFullDevice() bool
	AllowedPolicy() []DevicePolicy
}

type Manifest interface {
	EnvVars() map[string]string
	Annotations() map[string]string
	DeviceNodes() []DeviceNode
	MountPaths() []Mount
	//TODO(@bg) add method and struct for CDI later
}

type FuriosaDevice interface {
	DeviceInfo
	Manifest
}

// Mount is subset of oci-runtime Mount spec
type Mount struct {
	ContainerPath string
	HostPath      string
	ReadyOnly     bool
}

// DeviceNode is subset struct of oci-runtime DeviceNode spec
// TODO(@bg) add more fields for cdi spec if needed, for example: FileMode, UID, GID, Type, Major, Minor
type DeviceNode struct {
	ContainerPath string
	HostPath      string
	// Cgroups permissions of the device, candidates are one or more of
	// * r - allows container to read from the specified device.
	// * w - allows container to write to the specified device.
	// * m - allows container to create device files that do not yet exist.
	Permissions string
}
