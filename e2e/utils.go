package e2e

import (
	"encoding/json"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/smi"
)

type DeviceFile struct {
	Path  string   `json:"path"`
	Cores []uint32 `json:"cores"`
}

type Device struct {
	Arch     string       `json:"arch"`
	Dev      string       `json:"dev"`
	UUID     string       `json:"uuid"`
	CoreNum  uint32       `json:"core_num"`
	DevFiles []DeviceFile `json:"device_files"`
}

type Devices struct {
	Devices []Device `json:"devices"`
}

func MarshalDevices(devices []smi.Device) ([]byte, error) {
	payload := Devices{}

	for _, dev := range devices {
		var convertedDevFiles []DeviceFile
		files, err := dev.DeviceFiles()
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			converted := DeviceFile{
				Path:  file.Path(),
				Cores: file.Cores(),
			}

			convertedDevFiles = append(convertedDevFiles, converted)
		}

		info, err := dev.DeviceInfo()
		if err != nil {
			return nil, err
		}

		deviceMarshal := Device{
			Arch:     info.Arch().ToString(),
			Dev:      info.Name(),
			UUID:     info.UUID(),
			CoreNum:  info.CoreNum(),
			DevFiles: convertedDevFiles,
		}

		payload.Devices = append(payload.Devices, deviceMarshal)
	}

	return json.Marshal(payload)
}

func UnmarshalDevices(payload []byte) (*Devices, error) {
	devices := &Devices{}
	err := json.Unmarshal(payload, devices)
	if err != nil {
		return nil, err
	}

	return devices, nil
}
