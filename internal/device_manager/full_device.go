package device_manager

import (
	"errors"

	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/manifest"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/npu_allocator"
	devicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

var _ FuriosaDevice = (*fullDevice)(nil)

type fullDevice struct {
	origin    device.Device
	manifest  manifest.Manifest
	deviceID  string
	pciBusID  string
	numaNode  int
	isBlocked bool
}

func parseDeviceInfo(originDevice device.Device) (deviceID, pciBusID string, numaNode int, err error) {
	deviceID, err = originDevice.DeviceUUID()
	if err != nil {
		return "", "", 0, err
	}

	busname, err := originDevice.Busname()
	if err != nil {
		return "", "", 0, err
	}

	pciBusID, err = parseBusIDfromBDF(busname)
	if err != nil {
		return "", "", 0, err
	}

	unsignedNumaNode, err := originDevice.NumaNode()
	if err != nil {
		if errors.Is(err, device.UnexpectedValue) {
			return deviceID, pciBusID, -1, nil
		} else {
			return "", "", 0, err
		}
	} else {
		numaNode = int(unsignedNumaNode)
	}

	return deviceID, pciBusID, numaNode, err
}

func NewFullDevice(originDevice device.Device, isBlocked bool) (FuriosaDevice, error) {
	deviceID, pciBusID, numaNode, err := parseDeviceInfo(originDevice)
	if err != nil {
		return nil, err
	}

	var newFullDeviceManifest manifest.Manifest
	switch originDevice.Arch() {
	case device.ArchWarboy:
		newFullDeviceManifest = manifest.NewWarboyManifest(originDevice)
	case device.ArchRenegade:
		//FIXME(@bg): create right manifest using device arch once manifest is ready for the renegade
	}

	return &fullDevice{
		origin:    originDevice,
		manifest:  newFullDeviceManifest,
		deviceID:  deviceID,
		pciBusID:  pciBusID,
		numaNode:  int(numaNode),
		isBlocked: isBlocked,
	}, nil
}

func (f *fullDevice) DeviceID() string {
	return f.deviceID
}

func (f *fullDevice) PCIBusID() string {
	return f.pciBusID
}

func (f *fullDevice) NUMANode() int {
	return f.numaNode
}

func (f *fullDevice) IsHealthy() (bool, error) {
	//TODO(@bg): use more sophisticated way
	if f.isBlocked {
		return false, nil
	}
	liveness, err := f.origin.Alive()
	if err != nil {
		return liveness, err
	}
	return liveness, nil
}

func (f *fullDevice) IsFullDevice() bool {
	return true
}

func (f *fullDevice) EnvVars() map[string]string {
	return f.manifest.EnvVars()
}

func (f *fullDevice) Annotations() map[string]string {
	return f.manifest.Annotations()
}

func buildDeviceSpec(node *manifest.DeviceNode) *devicePluginAPIv1Beta1.DeviceSpec {
	return &devicePluginAPIv1Beta1.DeviceSpec{
		ContainerPath: node.ContainerPath,
		HostPath:      node.HostPath,
		Permissions:   node.Permissions,
	}
}

func (f *fullDevice) DeviceSpecs() []*devicePluginAPIv1Beta1.DeviceSpec {
	var deviceSpecs []*devicePluginAPIv1Beta1.DeviceSpec

	for _, deviceNode := range f.manifest.DeviceNodes() {
		deviceSpecs = append(deviceSpecs, buildDeviceSpec(deviceNode))
	}

	return deviceSpecs
}

func (f *fullDevice) Mounts() []*devicePluginAPIv1Beta1.Mount {
	var mounts []*devicePluginAPIv1Beta1.Mount

	for _, mount := range f.manifest.MountPaths() {
		var readOnly = false
		// NOTE(@bg): available options are "nodev", "bind", "noexec" and file permission("ro", "rw", ...).
		// However, device-plugin only consume file permission.
		for _, opt := range mount.Options {
			if opt == readOnlyOpt {
				readOnly = true
				break
			}
		}

		mounts = append(mounts, &devicePluginAPIv1Beta1.Mount{
			ContainerPath: mount.ContainerPath,
			HostPath:      mount.HostPath,
			ReadOnly:      readOnly,
		})
	}

	return mounts
}

func (f *fullDevice) CDIDevices() []*devicePluginAPIv1Beta1.CDIDevice {
	//TODO(@bg): CDI will be supported once libfuriosa-kubernetes is ready for CDI and DRA.
	return nil
}

func (f *fullDevice) ID() string {
	return f.DeviceID()
}

func (f *fullDevice) TopologyHintKey() string {
	return f.PCIBusID()
}

func (f *fullDevice) Equal(target npu_allocator.Device) bool {
	converted, isFullDevice := target.(*fullDevice)
	if !isFullDevice {
		return false
	}

	if f.DeviceID() != converted.DeviceID() {
		return false
	}

	return true
}
