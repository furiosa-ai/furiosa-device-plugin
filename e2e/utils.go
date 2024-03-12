package e2e

import (
	"encoding/json"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
)

type CoreRange struct {
	Type  string `json:"type"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

type DeviceFile struct {
	Path        string    `json:"path"`
	Filename    string    `json:"filename"`
	DeviceIndex int       `json:"device_index"`
	CoreRange   CoreRange `json:"core_range"`
	Mode        string    `json:"mode"`
}

type Device struct {
	Arch     string       `json:"arch"`
	Dev      string       `json:"dev"`
	UUID     string       `json:"uuid"`
	Cores    []int        `json:"cores"`
	DevFiles []DeviceFile `json:"device_files"`
}

type Devices struct {
	Devices []Device `json:"devices"`
}

func MarshalDevices(devices []device.Device) ([]byte, error) {
	payload := Devices{}

	for _, dev := range devices {
		uuid, err := dev.DeviceUUID()
		if err != nil {
			return nil, err
		}

		var convertedDevFiles []DeviceFile
		for _, file := range dev.DevFiles() {
			converted := DeviceFile{
				Path:        file.Path(),
				Filename:    file.Filename(),
				DeviceIndex: int(file.DeviceIndex()),
				CoreRange: CoreRange{
					Type:  string(file.CoreRange().Type()),
					Start: int(file.CoreRange().Start()),
					End:   int(file.CoreRange().End()),
				},
				Mode: string(file.Mode()),
			}

			convertedDevFiles = append(convertedDevFiles, converted)
		}

		var convertedCores []int
		for _, core := range dev.Cores() {
			convertedCores = append(convertedCores, int(core))
		}

		deviceMarshal := Device{
			Arch:     string(dev.Arch()),
			Dev:      dev.Name(),
			UUID:     uuid,
			Cores:    convertedCores,
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
