package device_manager

import (
	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
)

type DeviceManager interface {
	ResourceName() string
	Devices() []FuriosaDevice
	HealthCheck() error
	//TODO(@bg): add more methods
}

var _ DeviceManager = (*deviceManager)(nil)

type deviceManager struct {
	strategy  config.ResourceUnitStrategy
	origin    []device.Device
	debugMode bool
}

func (d deviceManager) HealthCheck() error {
	//TODO(@bg) examine all devices and return error if happened
	return nil
}

func (d deviceManager) ResourceName() string {
	// TODO(@bg): resource name should be determined by configured policy
	panic("implement me")
}

func (d deviceManager) Devices() []FuriosaDevice {
	//TODO implement me
	panic("implement me")
}

func NewDeviceManager(devices []device.Device, strategy config.ResourceUnitStrategy, debugMode bool) DeviceManager {
	//TODO(@bg): resource name should be validated with NameIsDNSSubdomain(...)
	//TODO(@bg): create furiosa devices based on the given policy
	return &deviceManager{
		strategy:  strategy,
		origin:    devices,
		debugMode: debugMode,
	}
}
