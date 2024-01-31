package device_manager

import (
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/manifest"
	DevicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

var _ FuriosaDevice = (*fullDevice)(nil)

type fullDevice struct {
	origin   device.Device
	manifest manifest.Manifest
}

func NewFullDevice(originDevice device.Device) FuriosaDevice {
	var newFullDeviceManifest manifest.Manifest
	switch originDevice.Arch() {
	case device.ArchWarboy:
		newFullDeviceManifest = manifest.NewWarboyManifest(originDevice)
	case device.ArchRenegade:
		//FIXME(@bg): create right manifest using device arch once manifest is ready for the renegade
	}
	return &fullDevice{
		origin:   originDevice,
		manifest: newFullDeviceManifest,
	}
}

func (f *fullDevice) DeviceID() (string, error) {
	uuid, err := f.origin.DeviceUUID()
	if err != nil {
		return "", err
	}

	return uuid, nil
}

func (f *fullDevice) PCIBusID() (string, error) {
	busname, err := f.origin.Busname()
	if err != nil {
		return "", err
	}

	busId, err := parseBusIDfromBDF(busname)
	if err != nil {
		return "", err
	}

	return busId, nil
}

func (f *fullDevice) NUMANode() (int, error) {
	numaNode, err := f.origin.NumaNode()
	if err != nil {
		return -1, err
	}
	return int(numaNode), nil
}

func (f *fullDevice) IsHealthy() (bool, error) {
	//TODO(@bg): use more sophisticated way
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

func buildDeviceSpec(node *manifest.DeviceNode) *DevicePluginAPIv1Beta1.DeviceSpec {
	return &DevicePluginAPIv1Beta1.DeviceSpec{
		ContainerPath: node.ContainerPath,
		HostPath:      node.HostPath,
		Permissions:   node.Permissions,
	}
}

func (f *fullDevice) DeviceSpecs() []*DevicePluginAPIv1Beta1.DeviceSpec {
	var deviceSpecs []*DevicePluginAPIv1Beta1.DeviceSpec

	for _, deviceNode := range f.manifest.DeviceNodes() {
		deviceSpecs = append(deviceSpecs, buildDeviceSpec(deviceNode))
	}

	return deviceSpecs
}

func (f *fullDevice) Mounts() []*DevicePluginAPIv1Beta1.Mount {
	var mounts []*DevicePluginAPIv1Beta1.Mount

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

		mounts = append(mounts, &DevicePluginAPIv1Beta1.Mount{
			ContainerPath: mount.ContainerPath,
			HostPath:      mount.HostPath,
			ReadOnly:      readOnly,
		})
	}

	return mounts
}

func (f *fullDevice) CDIDevices() []*DevicePluginAPIv1Beta1.CDIDevice {
	//TODO(@bg): CDI will be supported once libfuriosa-kubernetes is ready for CDI and DRA.
	return nil
}
