package device_manager

import (
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/smi"
	"reflect"
	"testing"

	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/manifest"
	devicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func NewMockExclusiveDevice(mockDevice smi.Device, isDisabled bool) (FuriosaDevice, error) {
	_, deviceID, pciBusID, numaNode, err := parseDeviceInfo(mockDevice)
	if err != nil {
		return nil, err
	}

	mockManifest, err := manifest.NewWarboyManifest(mockDevice)
	if err != nil {
		return nil, err
	}

	return &exclusiveDevice{
		origin:     mockDevice,
		manifest:   mockManifest,
		deviceID:   deviceID,
		pciBusID:   pciBusID,
		numaNode:   int(numaNode),
		isDisabled: isDisabled,
	}, nil
}

func TestDeviceID(t *testing.T) {
	tests := []struct {
		description    string
		mockDevice     smi.Device
		expectedResult string
	}{
		{
			description:    "test device id",
			mockDevice:     smi.GetStaticMockWarboyDevice(0),
			expectedResult: "A76AAD68-6855-40B1-9E86-D080852D1C80",
		},
	}

	for _, tc := range tests {
		exclusiveDev, err := NewMockExclusiveDevice(tc.mockDevice, false)
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}
		actualResult := exclusiveDev.DeviceID()
		if actualResult != tc.expectedResult {
			t.Errorf("expectedResult %s but got %s", tc.expectedResult, actualResult)
			continue
		}
	}
}

func TestPCIBusID(t *testing.T) {
	tests := []struct {
		description    string
		mockDevice     smi.Device
		expectedResult string
	}{
		{
			description:    "test pci bus id1",
			mockDevice:     smi.GetStaticMockWarboyDevice(0),
			expectedResult: "27",
		},
		{
			description:    "test pci bus id2",
			mockDevice:     smi.GetStaticMockWarboyDevice(1),
			expectedResult: "2a",
		},
	}

	for _, tc := range tests {
		exclusiveDev, err := NewMockExclusiveDevice(tc.mockDevice, false)
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}

		actualResult := exclusiveDev.PCIBusID()
		if actualResult != tc.expectedResult {
			t.Errorf("expectedResult %s but got %s", tc.expectedResult, actualResult)
			continue
		}
	}
}

func TestNUMANode(t *testing.T) {
	tests := []struct {
		description    string
		mockDevice     smi.Device
		expectedResult int
		expectError    bool
	}{
		{
			description:    "test numa node 1",
			mockDevice:     smi.GetStaticMockWarboyDevice(0),
			expectedResult: 0,
			expectError:    false,
		},
		{
			description:    "test numa node 2",
			mockDevice:     smi.GetStaticMockWarboyDevice(4),
			expectedResult: 1,
			expectError:    false,
		},
	}

	for _, tc := range tests {
		exclusiveDev, err := NewMockExclusiveDevice(tc.mockDevice, false)
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}

		actualResult := exclusiveDev.NUMANode()
		if actualResult != tc.expectedResult {
			t.Errorf("expectedResult %d but got %d", tc.expectedResult, actualResult)
		}
	}
}

func TestDeviceSpecs(t *testing.T) {
	tests := []struct {
		description    string
		mockDevice     smi.Device
		expectedResult []*devicePluginAPIv1Beta1.DeviceSpec
	}{
		{
			description: "test warboy exclusive device",
			mockDevice:  smi.GetStaticMockWarboyDevice(0),
			expectedResult: []*devicePluginAPIv1Beta1.DeviceSpec{
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
		//TODO(@bg): add testcases for rngd and other npu family later
	}

	for _, tc := range tests {
		exclusiveDev, err := NewMockExclusiveDevice(tc.mockDevice, false)
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}

		actualResult := exclusiveDev.DeviceSpecs()
		if !reflect.DeepEqual(actualResult, tc.expectedResult) {
			t.Errorf("expectedResult %v but got %v", tc.expectedResult, actualResult)
		}
	}
}

// This function tests the IsHealthy API only in terms of the deny list.
func TestIsHealthy(t *testing.T) {
	tests := []struct {
		description    string
		mockDevice     smi.Device
		isDisabled     bool
		expectedResult bool
	}{
		{
			description:    "test healthy device",
			mockDevice:     smi.GetStaticMockWarboyDevice(0),
			isDisabled:     false,
			expectedResult: true,
		},
		{
			description:    "test unhealthy device",
			mockDevice:     smi.GetStaticMockWarboyDevice(0),
			isDisabled:     true,
			expectedResult: false,
		},
	}

	for _, tc := range tests {
		exclusiveDev, err := NewMockExclusiveDevice(tc.mockDevice, tc.isDisabled)
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}

		actualResult, err := exclusiveDev.IsHealthy()
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}

		if actualResult != tc.expectedResult {
			t.Errorf("expectedResult %t but got %t", tc.expectedResult, actualResult)
		}
	}
}

func TestMounts(t *testing.T) {
	tests := []struct {
		description    string
		mockDevice     smi.Device
		expectedResult []*devicePluginAPIv1Beta1.Mount
	}{
		{
			description: "test warboy mount",
			mockDevice:  smi.GetStaticMockWarboyDevice(0),
			expectedResult: []*devicePluginAPIv1Beta1.Mount{
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
		exclusiveDev, err := NewMockExclusiveDevice(tc.mockDevice, false)
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}

		actualResult := exclusiveDev.Mounts()
		if !reflect.DeepEqual(actualResult, tc.expectedResult) {
			t.Errorf("expectedResult %v but got %v", tc.expectedResult, actualResult)
		}
	}
}

func TestID(t *testing.T) {
	tests := []struct {
		description    string
		mockDevice     smi.Device
		expectedResult string
	}{
		{
			description:    "test id",
			mockDevice:     smi.GetStaticMockWarboyDevice(0),
			expectedResult: "A76AAD68-6855-40B1-9E86-D080852D1C80",
		},
	}

	for _, tc := range tests {
		exclusiveDev, err := NewMockExclusiveDevice(tc.mockDevice, false)
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}
		actualResult := exclusiveDev.ID()
		if actualResult != tc.expectedResult {
			t.Errorf("expectedResult %s but got %s", tc.expectedResult, actualResult)
			continue
		}
	}
}

func TestTopologyHintKey(t *testing.T) {
	tests := []struct {
		description    string
		mockDevice     smi.Device
		expectedResult string
	}{
		{
			description:    "test topology hint",
			mockDevice:     smi.GetStaticMockWarboyDevice(0),
			expectedResult: "27",
		},
	}

	for _, tc := range tests {
		exclusiveDev, err := NewMockExclusiveDevice(tc.mockDevice, false)
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}

		actualResult := exclusiveDev.TopologyHintKey()
		if actualResult != tc.expectedResult {
			t.Errorf("expectedResult %s but got %s", tc.expectedResult, actualResult)
			continue
		}
	}
}

func TestEqual(t *testing.T) {
	tests := []struct {
		description      string
		mockSourceDevice smi.Device
		mockTargetDevice smi.Device
		expected         bool
	}{
		{
			description:      "expect source and target are identical",
			mockSourceDevice: smi.GetStaticMockWarboyDevice(0),
			mockTargetDevice: smi.GetStaticMockWarboyDevice(0),
			expected:         true,
		},
		{
			description:      "expect source and target are not identical",
			mockSourceDevice: smi.GetStaticMockWarboyDevice(0),
			mockTargetDevice: smi.GetStaticMockWarboyDevice(1),
			expected:         false,
		},
	}
	for _, tc := range tests {
		source, err := NewMockExclusiveDevice(tc.mockSourceDevice, false)
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}

		target, err := NewMockExclusiveDevice(tc.mockTargetDevice, false)
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}

		actual := source.Equal(target)
		if actual != tc.expected {
			t.Errorf("expectedResult %v but got %v", tc.expected, actual)
			continue
		}
	}
}
