package device_manager

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/bradfitz/iter"
	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
)

func MockDeviceSlices() []device.Device {
	mockDevices := []device.Device{}
	for i := range iter.N(8) {
		mockDevices = append(mockDevices,
			device.NewMockWarboyDevice(uint8(i), 0, fmt.Sprintf("0000:%d:00.0", i), "", "", "", "", strconv.Itoa(i)))
	}
	return mockDevices
}

func TestBuildFuriosaDevices(t *testing.T) {
	tests := []struct {
		description      string
		devices          []device.Device
		strategy         config.ResourceUnitStrategy
		expectFullDevice bool
	}{
		{
			description:      "test legacy strategy",
			strategy:         config.LegacyStrategy,
			devices:          MockDeviceSlices(),
			expectFullDevice: true,
		},
		{
			description:      "test generic strategy",
			strategy:         config.GenericStrategy,
			devices:          MockDeviceSlices(),
			expectFullDevice: true,
		},
		{
			description:      "test single core strategy",
			strategy:         config.SingleCoreStrategy,
			devices:          MockDeviceSlices(),
			expectFullDevice: false,
		},
		{
			description:      "test dual core strategy",
			strategy:         config.DualCoreStrategy,
			devices:          MockDeviceSlices(),
			expectFullDevice: false,
		},
		{
			description:      "test quad core strategy",
			strategy:         config.QuadCoreStrategy,
			devices:          MockDeviceSlices(),
			expectFullDevice: false,
		},
	}

	for _, tc := range tests {
		actualDevices, err := buildFuriosaDevices(tc.devices, newDeviceFuncResolver(tc.strategy))
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}
		for _, actualDevice := range actualDevices {
			if tc.expectFullDevice {
				if _, ok := actualDevice.(*fullDevice); !ok {
					t.Errorf("expect full device but type assertion failed")
				}
			} else {
				if _, ok := actualDevice.(*partialDevice); !ok {
					t.Errorf("expect partial device but type assertion failed")
				}
			}
		}

	}
}
