package device_manager

import (
	"reflect"
	"testing"

	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
	manifest2 "github.com/furiosa-ai/libfuriosa-kubernetes/pkg/manifest"
	DevicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func NewMockFullDevice(mockDevice device.Device) FuriosaDevice {
	return &fullDevice{
		origin:   mockDevice,
		manifest: manifest2.NewWarboyManifest(mockDevice),
	}
}

func TestDeviceID(t *testing.T) {
	tests := []struct {
		description    string
		mockDevice     device.Device
		expectedResult string
	}{
		{
			description:    "test device id",
			mockDevice:     device.NewMockWarboyDevice(0, 0, "", "", "", "", "", "A76AAD68-6855-40B1-9E86-D080852D1C84"),
			expectedResult: "A76AAD68-6855-40B1-9E86-D080852D1C84",
		},
	}

	for _, tc := range tests {
		fullDev := NewMockFullDevice(tc.mockDevice)
		actualResult, err := fullDev.DeviceID()
		if err != nil {
			t.Errorf("got unexpected error %t", err)
			continue
		}

		if actualResult != tc.expectedResult {
			t.Errorf("expectedResult %s but got %s", tc.expectedResult, actualResult)
			continue
		}
	}
}

func TestPCIBusID(t *testing.T) {
	tests := []struct {
		description    string
		mockDevice     device.Device
		expectedResult string
	}{
		{
			description:    "test pci bus id1",
			mockDevice:     device.NewMockWarboyDevice(0, 0, "0000:51:00.0", "", "", "", "", ""),
			expectedResult: "51",
		},
		{
			description:    "test pci bus id2",
			mockDevice:     device.NewMockWarboyDevice(0, 0, "0011:9e:00.0", "", "", "", "", ""),
			expectedResult: "9e",
		},
	}

	for _, tc := range tests {
		fullDev := NewMockFullDevice(tc.mockDevice)
		actualResult, err := fullDev.PCIBusID()
		if err != nil {
			t.Errorf("got unexpected error %t", err)
			continue
		}

		if actualResult != tc.expectedResult {
			t.Errorf("expectedResult %s but got %s", tc.expectedResult, actualResult)
			continue
		}
	}
}

func TestNUMANode(t *testing.T) {
	tests := []struct {
		description    string
		mockDevice     device.Device
		expectedResult int
		expectError    bool
	}{
		{
			description:    "test numa node 1",
			mockDevice:     device.NewMockWarboyDevice(0, 0, "", "", "", "", "", ""),
			expectedResult: 0,
			expectError:    false,
		},
		{
			description:    "test numa node 2",
			mockDevice:     device.NewMockWarboyDevice(0, 1, "", "", "", "", "", ""),
			expectedResult: 1,
			expectError:    false,
		},
		{
			description:    "test numa node 3",
			mockDevice:     device.NewMockWarboyDevice(0, -1, "", "", "", "", "", ""),
			expectedResult: -1,
			expectError:    true,
		},
	}

	for _, tc := range tests {
		fullDev := NewMockFullDevice(tc.mockDevice)
		actualResult, actualErr := fullDev.NUMANode()
		if actualErr != nil != tc.expectError {
			t.Errorf("unexpected error %t", actualErr)
			continue
		}

		if actualResult != tc.expectedResult {
			t.Errorf("expectedResult %d but got %d", tc.expectedResult, actualResult)
		}
	}
}

// TODO(@bg) add test for IsHealthy API later
func TestDeviceSpecs(t *testing.T) {
	tests := []struct {
		description    string
		mockDevice     device.Device
		expectedResult []*DevicePluginAPIv1Beta1.DeviceSpec
	}{
		{
			description: "test warboy full device",
			mockDevice:  device.NewMockWarboyDevice(0, 0, "", "", "", "", "", ""),
			expectedResult: []*DevicePluginAPIv1Beta1.DeviceSpec{
				{
					ContainerPath: "/dev/npu0_mgmt",
					HostPath:      "/dev/npu0_mgmt",
					Permissions:   "rw",
				},
				{
					ContainerPath: "/dev/npu0",
					HostPath:      "/dev/npu0",
					Permissions:   "rw",
				},
				{
					ContainerPath: "/dev/npu0pe0",
					HostPath:      "/dev/npu0pe0",
					Permissions:   "rw",
				},
				{
					ContainerPath: "/dev/npu0pe1",
					HostPath:      "/dev/npu0pe1",
					Permissions:   "rw",
				},
				{
					ContainerPath: "/dev/npu0pe0-1",
					HostPath:      "/dev/npu0pe0-1",
					Permissions:   "rw",
				},
				{
					ContainerPath: "/dev/npu0ch0",
					HostPath:      "/dev/npu0ch0",
					Permissions:   "rw",
				},
				{
					ContainerPath: "/dev/npu0ch1",
					HostPath:      "/dev/npu0ch1",
					Permissions:   "rw",
				},
				{
					ContainerPath: "/dev/npu0ch2",
					HostPath:      "/dev/npu0ch2",
					Permissions:   "rw",
				},
				{
					ContainerPath: "/dev/npu0ch3",
					HostPath:      "/dev/npu0ch3",
					Permissions:   "rw",
				},
			},
		},
		//TODO(@bg): add testcases for renegade and other npu family later
	}

	for _, tc := range tests {
		fullDev := NewMockFullDevice(tc.mockDevice)
		actualResult := fullDev.DeviceSpecs()
		if !reflect.DeepEqual(actualResult, tc.expectedResult) {
			t.Errorf("expectedResult %v but got %v", tc.expectedResult, actualResult)
		}
	}
}

func TestMounts(t *testing.T) {
	tests := []struct {
		description    string
		mockDevice     device.Device
		expectedResult []*DevicePluginAPIv1Beta1.Mount
	}{
		{
			description: "test warboy mount",
			mockDevice:  device.NewMockWarboyDevice(0, 0, "", "", "", "", "", ""),
			expectedResult: []*DevicePluginAPIv1Beta1.Mount{
				{
					ContainerPath: "/sys/class/npu_mgmt/npu0_mgmt",
					HostPath:      "/sys/class/npu_mgmt/npu0_mgmt",
					ReadOnly:      true,
				},
				{
					ContainerPath: "/sys/class/npu_mgmt/npu0",
					HostPath:      "/sys/class/npu_mgmt/npu0",
					ReadOnly:      true,
				},
				{
					ContainerPath: "/sys/class/npu_mgmt/npu0pe0",
					HostPath:      "/sys/class/npu_mgmt/npu0pe0",
					ReadOnly:      true,
				},
				{
					ContainerPath: "/sys/class/npu_mgmt/npu0pe1",
					HostPath:      "/sys/class/npu_mgmt/npu0pe1",
					ReadOnly:      true,
				},
				{
					ContainerPath: "/sys/class/npu_mgmt/npu0pe0-1",
					HostPath:      "/sys/class/npu_mgmt/npu0pe0-1",
					ReadOnly:      true,
				},
				{
					ContainerPath: "/sys/devices/virtual/npu_mgmt/npu0_mgmt",
					HostPath:      "/sys/devices/virtual/npu_mgmt/npu0_mgmt",
					ReadOnly:      true,
				},
				{
					ContainerPath: "/sys/devices/virtual/npu_mgmt/npu0",
					HostPath:      "/sys/devices/virtual/npu_mgmt/npu0",
					ReadOnly:      true,
				},
				{
					ContainerPath: "/sys/devices/virtual/npu_mgmt/npu0pe0",
					HostPath:      "/sys/devices/virtual/npu_mgmt/npu0pe0",
					ReadOnly:      true,
				},
				{
					ContainerPath: "/sys/devices/virtual/npu_mgmt/npu0pe1",
					HostPath:      "/sys/devices/virtual/npu_mgmt/npu0pe1",
					ReadOnly:      true,
				},
				{
					ContainerPath: "/sys/devices/virtual/npu_mgmt/npu0pe0-1",
					HostPath:      "/sys/devices/virtual/npu_mgmt/npu0pe0-1",
					ReadOnly:      true,
				},
			},
		},
	}

	for _, tc := range tests {
		fullDev := NewMockFullDevice(tc.mockDevice)
		actualResult := fullDev.Mounts()
		if !reflect.DeepEqual(actualResult, tc.expectedResult) {
			t.Errorf("expectedResult %v but got %v", tc.expectedResult, actualResult)
		}
	}
}
