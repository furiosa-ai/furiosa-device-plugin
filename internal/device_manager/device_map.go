package device_manager

import (
	"cmp"
	"slices"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

type DeviceMap map[api.Arch][]api.Device

// BuildDeviceMap enumerates the devices and groups them by arch. Each per-arch
// slice is sorted (vendor ID, device ID, dev-node index); sameDevices relies
// on this order.
func BuildDeviceMap() (DeviceMap, error) {
	devices, err := api.ListDevices()
	if err != nil {
		return nil, err
	}

	// Grouping appends in iteration order, so sorting once here leaves every
	// per-arch slice sorted.
	sortDevices(devices)

	archToDevicesMap := make(DeviceMap)
	for i := range devices {
		key := devices[i].Arch()
		archToDevicesMap[key] = append(archToDevicesMap[key], devices[i])
	}

	return archToDevicesMap, nil
}

// sortDevices sorts devices in place by vendor ID, device ID and dev-node
// index, the same order api.ListDevices uses.
func sortDevices(devices []api.Device) {
	slices.SortFunc(devices, func(a, b api.Device) int {
		return cmp.Or(
			cmp.Compare(a.VendorID(), b.VendorID()),
			cmp.Compare(a.DeviceID(), b.DeviceID()),
			cmp.Compare(a.DevNodeIndex(), b.DevNodeIndex()),
		)
	})
}
