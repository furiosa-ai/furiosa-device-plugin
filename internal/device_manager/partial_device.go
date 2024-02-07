package device_manager

import (
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/npu_allocator"
	DevicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

var _ FuriosaDevice = (*partialDevice)(nil)

type partialDevice struct {
}

func NewPartialDevice(originDevice device.Device) (FuriosaDevice, error) {
	return &partialDevice{}, nil
}

func (p partialDevice) DeviceID() string {
	//TODO implement me
	return ""
}

func (p partialDevice) PCIBusID() string {
	//TODO implement me
	return ""
}

func (p partialDevice) NUMANode() int {
	//TODO implement me
	return 0
}

func (p partialDevice) IsHealthy() (bool, error) {
	//TODO implement me
	return false, nil
}

func (p partialDevice) IsFullDevice() bool {
	return false
}

func (p partialDevice) EnvVars() map[string]string {
	//TODO implement me
	return nil
}

func (p partialDevice) Annotations() map[string]string {
	//TODO implement me
	return nil
}

func (p partialDevice) DeviceSpecs() []*DevicePluginAPIv1Beta1.DeviceSpec {
	//TODO implement me
	return nil
}

func (p partialDevice) Mounts() []*DevicePluginAPIv1Beta1.Mount {
	//TODO implement me
	return nil
}

func (p partialDevice) CDIDevices() []*DevicePluginAPIv1Beta1.CDIDevice {
	//TODO implement me
	return nil
}

func (p partialDevice) ID() string {
	//TODO implement me
	panic("implement me")
}

func (p partialDevice) TopologyHintKey() string {
	//TODO implement me
	panic("implement me")
}

func (p partialDevice) Equal(target npu_allocator.Device) bool {
	//TODO implement me
	panic("implement me")
}
