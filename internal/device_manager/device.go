package device_manager

import (
	DevicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
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
	DeviceNodes() []*DevicePluginAPIv1Beta1.DeviceSpec
	MountPaths() []*DevicePluginAPIv1Beta1.Mount
	CDIDevices() []*DevicePluginAPIv1Beta1.CDIDevice
}

type FuriosaDevice interface {
	DeviceInfo
	Manifest
}
