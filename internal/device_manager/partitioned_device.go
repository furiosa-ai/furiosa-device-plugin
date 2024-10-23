package device_manager

import (
	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/npu_allocator"
	devicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

var _ FuriosaDevice = (*partitionedDevice)(nil)

type partitionedDevice struct{}

func NewPartitionedDevices(_ int, _ smi.Device, _ config.ResourceUnitStrategy, _ bool) ([]FuriosaDevice, error) {
	//TODO implement me
	return nil, nil
}

func (p partitionedDevice) DeviceID() string {
	//TODO implement me
	return ""
}

func (p partitionedDevice) PCIBusID() string {
	//TODO implement me
	return ""
}

func (p partitionedDevice) NUMANode() int {
	//TODO implement me
	return 0
}

func (p partitionedDevice) IsHealthy() (bool, error) {
	//TODO implement me
	return false, nil
}

func (p partitionedDevice) IsExclusiveDevice() bool {
	return false
}

func (p partitionedDevice) EnvVars() map[string]string {
	//TODO implement me
	return nil
}

func (p partitionedDevice) Annotations() map[string]string {
	//TODO implement me
	return nil
}

func (p partitionedDevice) DeviceSpecs() []*devicePluginAPIv1Beta1.DeviceSpec {
	//TODO implement me
	return nil
}

func (p partitionedDevice) Mounts() []*devicePluginAPIv1Beta1.Mount {
	//TODO implement me
	return nil
}

func (p partitionedDevice) CDIDevices() []*devicePluginAPIv1Beta1.CDIDevice {
	//TODO implement me
	return nil
}

func (p partitionedDevice) Index() int {
	//TODO implement me
	panic("implement me")
}

func (p partitionedDevice) ID() string {
	//TODO implement me
	panic("implement me")
}

func (p partitionedDevice) TopologyHintKey() npu_allocator.TopologyHintKey {
	//TODO implement me
	panic("implement me")
}

func (p partitionedDevice) Equal(target npu_allocator.Device) bool {
	//TODO implement me
	panic("implement me")
}
