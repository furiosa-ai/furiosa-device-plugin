package server

import (
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
	"strings"
)

type DeviceMap map[string][]device.Device

func BuildDeviceMap() (DeviceMap, error) {
	deviceLister := device.NewDeviceLister()
	devices, err := deviceLister.ListDevices()
	if err != nil {
		return nil, err
	}

	archToDevicesMap := make(DeviceMap)
	for _, d := range devices {
		key := strings.ToLower(string(d.Arch()))
		archToDevicesMap[key] = append(archToDevicesMap[key], d)
	}

	return archToDevicesMap, nil
}
