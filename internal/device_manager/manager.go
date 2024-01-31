package device_manager

import (
	"fmt"

	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
)

type DeviceManager interface {
	ResourceName() string
	Devices() []FuriosaDevice
	HealthCheck() error
	//TODO(@bg): add more methods
}

type newDeviceFunc func(originDevice device.Device) FuriosaDevice

var _ DeviceManager = (*deviceManager)(nil)

type deviceManager struct {
	origin         []device.Device
	furiosaDevices []FuriosaDevice
	resourceName   string
	debugMode      bool
}

func (d *deviceManager) HealthCheck() error {
	for _, dev := range d.furiosaDevices {
		healthy, err := dev.IsHealthy()
		if err != nil {
			return err
		}

		if !healthy {
			return fmt.Errorf("device is not healthy")
		}
	}

	return nil
}

func (d *deviceManager) ResourceName() string {
	return d.resourceName
}

func (d *deviceManager) Devices() []FuriosaDevice {
	return d.furiosaDevices
}

func newDeviceFuncResolver(strategy config.ResourceUnitStrategy) (ret newDeviceFunc) {
	// Note: config validation ensure that there is no exception other than listed strategies.
	switch strategy {
	case config.LegacyStrategy, config.GenericStrategy:
		ret = NewFullDevice
	case config.SingleCoreStrategy, config.DualCoreStrategy, config.QuadCoreStrategy:
		ret = NewPartialDevice
	}

	return ret
}

func buildFuriosaDevices(devices []device.Device, newDevFunc newDeviceFunc) []FuriosaDevice {
	var furiosaDevices []FuriosaDevice
	for _, origin := range devices {
		furiosaDevices = append(furiosaDevices, newDevFunc(origin))
	}
	return furiosaDevices
}

func NewDeviceManager(devices []device.Device, strategy config.ResourceUnitStrategy, debugMode bool) (DeviceManager, error) {
	resName, err := buildAndValidateFullResourceEndpointName(devices[0].Arch(), strategy)
	if err != nil {
		return nil, err
	}

	return &deviceManager{
		origin:         devices,
		furiosaDevices: buildFuriosaDevices(devices, newDeviceFuncResolver(strategy)),
		resourceName:   resName,
		debugMode:      debugMode,
	}, nil
}
