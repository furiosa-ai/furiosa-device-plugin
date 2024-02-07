package device_manager

import (
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/npu_allocator"
	DevicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	readOnlyOpt = "ro"
)

type DeviceInfo interface {
	DeviceID() string
	PCIBusID() string
	NUMANode() int
	IsHealthy() (bool, error)
	IsFullDevice() bool
}

type Manifest interface {
	EnvVars() map[string]string
	Annotations() map[string]string
	DeviceSpecs() []*DevicePluginAPIv1Beta1.DeviceSpec
	Mounts() []*DevicePluginAPIv1Beta1.Mount
	CDIDevices() []*DevicePluginAPIv1Beta1.CDIDevice
}

type FuriosaDevice interface {
	DeviceInfo
	Manifest
	npu_allocator.Device
}
