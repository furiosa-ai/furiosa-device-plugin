package device_manager

import (
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/rs/zerolog"
)

type DeviceMap map[smi.Arch][]smi.Device

func BuildDeviceMap(logger zerolog.Logger) (DeviceMap, error) {
	err := smi.Init()
	if err != nil {
		return nil, err
	}

	devices, err := smi.ListDevices()
	if err != nil {
		return nil, err
	}

	archToDevicesMap := make(DeviceMap)
	for _, d := range devices {
		info, err := d.DeviceInfo()
		if err != nil {
			logger.Err(err).Msg("failed to get device info for some device")
			continue
		}

		key := info.Arch()
		archToDevicesMap[key] = append(archToDevicesMap[key], d)
	}

	return archToDevicesMap, nil
}
