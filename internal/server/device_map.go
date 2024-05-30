package server

import (
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/smi"
)

type DeviceMap map[smi.Arch][]smi.Device

func BuildDeviceMap() (DeviceMap, error) {
	devices, err := smi.GetDevices()
	if err != nil {
		return nil, err
	}

	archToDevicesMap := make(DeviceMap)
	for _, d := range devices {
		info, err := d.DeviceInfo()
		if err != nil {
			return nil, err
		}

		key := info.Arch()
		archToDevicesMap[key] = append(archToDevicesMap[key], d)
	}

	return archToDevicesMap, nil
}
