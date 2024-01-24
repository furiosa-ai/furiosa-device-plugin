package server

import "github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"

type DeviceMap map[device.Arch][]device.Device

func BuildDeviceMap() (DeviceMap, error) {
	deviceLister := device.NewDeviceLister()
	devices, err := deviceLister.ListDevices()
	if err != nil {
		return nil, err
	}

	archToDevicesMap := make(DeviceMap)
	for _, d := range devices {
		archToDevicesMap[d.Arch()] = append(archToDevicesMap[d.Arch()], d)
	}

	return archToDevicesMap, nil
}
