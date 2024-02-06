package device_manager

import (
	"fmt"

	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
)

type DeviceManager interface {
	ResourceName() string
	Devices() []string
	HealthCheck() error
}

type newDeviceFunc func(originDevice device.Device) (FuriosaDevice, error)

var _ DeviceManager = (*deviceManager)(nil)

type deviceManager struct {
	origin         []device.Device
	furiosaDevices map[string]FuriosaDevice
	resourceName   string
	debugMode      bool
}

func (d *deviceManager) Devices() (ret []string) {
	for id := range d.furiosaDevices {
		ret = append(ret, id)
	}

	return ret
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

func buildFuriosaDevices(devices []device.Device, newDevFunc newDeviceFunc) (map[string]FuriosaDevice, error) {
	furiosaDevices := map[string]FuriosaDevice{}
	for _, origin := range devices {
		furiosaDevice, err := newDevFunc(origin)
		if err != nil {
			return nil, err
		}
		furiosaDevices[furiosaDevice.DeviceID()] = furiosaDevice

	}
	return furiosaDevices, nil
}

func NewDeviceManager(devices []device.Device, strategy config.ResourceUnitStrategy, debugMode bool) (DeviceManager, error) {
	resName, err := buildAndValidateFullResourceEndpointName(devices[0].Arch(), strategy)
	if err != nil {
		return nil, err
	}

	furiosaDevices, err := buildFuriosaDevices(devices, newDeviceFuncResolver(strategy))
	if err != nil {
		return nil, err
	}
	return &deviceManager{
		origin:         devices,
		furiosaDevices: furiosaDevices,
		resourceName:   resName,
		debugMode:      debugMode,
	}, nil
}
