package device_manager

import (
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/cdi_spec_gen"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/furiosa_device"
	devicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"tags.cncf.io/container-device-interface/specs-go"
)

const (
	groupName   = "legacy"
	readOnlyOpt = "ro"
)

func buildDeviceSpecToContainerAllocateResponse(devices ...furiosa_device.FuriosaDevice) (*devicePluginAPIv1Beta1.ContainerAllocateResponse, error) {
	cdiSpec, err := cdi_spec_gen.NewSpec(cdi_spec_gen.WithGroupDevice(groupName, devices...))
	if err != nil {
		return nil, err
	}

	raw := cdiSpec.Raw()
	resp := &devicePluginAPIv1Beta1.ContainerAllocateResponse{}
	resp.Devices = transformDeviceNodes(raw.Devices)
	resp.Mounts = transformMounts(raw.Devices)
	return resp, nil
}

func transformDeviceNodes(devices []specs.Device) []*devicePluginAPIv1Beta1.DeviceSpec {
	var out []*devicePluginAPIv1Beta1.DeviceSpec
	for _, device := range devices {
		for _, node := range device.ContainerEdits.DeviceNodes {
			out = append(out, &devicePluginAPIv1Beta1.DeviceSpec{
				HostPath:      node.HostPath,
				ContainerPath: node.Path,
				Permissions:   node.Permissions,
			})
		}
	}

	return out
}

func transformMounts(devices []specs.Device) []*devicePluginAPIv1Beta1.Mount {
	var out []*devicePluginAPIv1Beta1.Mount
	for _, device := range devices {
		for _, mount := range device.ContainerEdits.Mounts {
			out = append(out, &devicePluginAPIv1Beta1.Mount{
				HostPath:      mount.HostPath,
				ContainerPath: mount.ContainerPath,
				ReadOnly:      optsToReadOnly(mount.Options),
			})
		}
	}

	return out
}

func optsToReadOnly(opts []string) bool {
	for _, opt := range opts {
		if opt == readOnlyOpt {
			return true
		}
	}
	return false
}
