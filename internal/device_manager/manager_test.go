package device_manager

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/bradfitz/iter"
	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/npu_allocator"
	"github.com/google/uuid"

	devicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func MockDeviceUUIDSeed(n int) (ret []string) {
	cache := map[string]string{}
	for range iter.N(n) {
		var newSeed string
		for {
			newUUID, _ := uuid.NewRandom()
			if _, exist := cache[newUUID.String()]; !exist {
				newSeed = newUUID.String()
				break
			}
		}
		ret = append(ret, newSeed)
	}
	return ret
}

func MockDeviceBusnameSeed(n int) (ret []string) {
	for seed := range iter.N(n) {
		ret = append(ret, fmt.Sprintf("0000:%s:00.0", strconv.FormatInt(int64(seed), 16)))
	}
	return ret
}

func MockDeviceSlices(n int, uuidSeed []string, busnameSeed []string) (mockDevices []device.Device) {
	if len(uuidSeed) == 0 {
		uuidSeed = MockDeviceUUIDSeed(n)
	}

	if len(busnameSeed) == 0 {
		busnameSeed = MockDeviceBusnameSeed(n)
	}

	for i := range iter.N(n) {
		mockDevices = append(mockDevices,
			device.NewMockWarboyDevice(uint8(i), 0, busnameSeed[i], "", "", "", "", uuidSeed[i]))
	}
	return mockDevices
}

func MockFuriosaDevices(mockDevices []device.Device) (ret map[string]FuriosaDevice) {
	if len(mockDevices) == 0 {
		mockDevices = MockDeviceSlices(8, nil, nil)
	}
	ret = make(map[string]FuriosaDevice, len(mockDevices))
	for _, mockDevice := range mockDevices {
		key, _ := mockDevice.DeviceUUID()
		mockFuriosaDevice, _ := NewMockFullDevice(mockDevice, false)
		ret[key] = mockFuriosaDevice
	}

	return ret
}

func TestBuildFuriosaDevices(t *testing.T) {
	tests := []struct {
		description      string
		strategy         config.ResourceUnitStrategy
		expectFullDevice bool
	}{
		{
			description:      "test legacy strategy",
			strategy:         config.LegacyStrategy,
			expectFullDevice: true,
		},
		{
			description:      "test generic strategy",
			strategy:         config.GenericStrategy,
			expectFullDevice: true,
		},
		{
			description:      "test single core strategy",
			strategy:         config.SingleCoreStrategy,
			expectFullDevice: false,
		},
		{
			description:      "test dual core strategy",
			strategy:         config.DualCoreStrategy,
			expectFullDevice: false,
		},
		{
			description:      "test quad core strategy",
			strategy:         config.QuadCoreStrategy,
			expectFullDevice: false,
		},
	}

	for _, tc := range tests {
		devices := MockDeviceSlices(8, nil, nil)
		actualDevices, err := buildFuriosaDevices(devices, nil, newDeviceFuncResolver(tc.strategy))
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

func TestFetchByID(t *testing.T) {
	seedUUID := MockDeviceUUIDSeed(8)
	mockDevices := MockDeviceSlices(8, seedUUID, nil)
	mockFuriosaDevices := MockFuriosaDevices(mockDevices)
	actual, err := fetchByID(mockFuriosaDevices, seedUUID)
	if err != nil {
		t.Errorf("failed with error %t", err)
		return
	}

	var actualIDs []string
	for _, furiosaDevice := range actual {
		actualIDs = append(actualIDs, furiosaDevice.DeviceID())
	}

	if !reflect.DeepEqual(actualIDs, seedUUID) {
		t.Errorf("expectedResult %v but got %v", seedUUID, actualIDs)
	}
}

func TestFetchDevicesByID(t *testing.T) {
	seedUUID := MockDeviceUUIDSeed(8)
	mockDevices := MockDeviceSlices(8, seedUUID, nil)
	mockFuriosaDevices := MockFuriosaDevices(mockDevices)
	actual, err := fetchDevicesByID(mockFuriosaDevices, seedUUID)
	if err != nil {
		t.Errorf("failed with error %t", err)
		return
	}

	var actualIDs []string
	for _, ele := range actual {
		if furiosaDevice, ok := ele.(FuriosaDevice); !ok {
			t.Errorf("type assertion failed")
			return
		} else {
			actualIDs = append(actualIDs, furiosaDevice.DeviceID())
		}
	}

	if !reflect.DeepEqual(actualIDs, seedUUID) {
		t.Errorf("expectedResult %v but got %v", seedUUID, actualIDs)
	}
}

// staticMockTopologyHintProvider build hint matrix for optimized 2socket server
// which has two pcie switches per socket and two devices per switch.
func staticMockTopologyHintProvider() npu_allocator.TopologyHintProvider {
	hints := map[string]map[string]uint{
		"0": {"0": 70, "1": 30, "2": 20, "3": 20, "4": 10, "5": 10, "6": 10, "7": 10},
		"1": {"1": 70, "2": 20, "3": 20, "4": 10, "5": 10, "6": 10, "7": 10},
		"2": {"2": 70, "3": 30, "4": 10, "5": 10, "6": 10, "7": 10},
		"3": {"3": 70, "4": 10, "5": 10, "6": 10, "7": 10},
		"4": {"4": 70, "5": 30, "6": 20, "7": 20},
		"5": {"5": 70, "6": 20, "7": 20},
		"6": {"6": 70, "7": 30},
		"7": {"7": 70},
	}
	return func(topologyHintKey1, topologyHintKey2 string) uint {
		if topologyHintKey1 > topologyHintKey2 {
			topologyHintKey1, topologyHintKey2 = topologyHintKey2, topologyHintKey1
		}

		if hint, ok := hints[topologyHintKey1][topologyHintKey2]; ok {
			return hint
		}

		return 0
	}
}

// TODO(@bg): Add test for bin packing allocator once it is ready.

func TestGetContainerPreferredAllocationResponseWithScoreBasedOptimalNpuAllocator(t *testing.T) {
	tests := []struct {
		description    string
		deviceSeedUUID []string
		available      []string
		required       []string
		request        int
		expectedResult *devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse
		expectError    bool
	}{
		// start with socket 0
		{
			description:    "request one device from socket 0 of 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"0", "1", "2", "3"},
			required:       nil,
			request:        1,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0"},
			},
			expectError: false,
		},
		{
			description:    "request one pre-allocated device from socket 0 of 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"0", "1", "2", "3"},
			required:       []string{"3"},
			request:        1,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"3"},
			},
			expectError: false,
		},
		{
			description:    "request two devices from socket 0 of 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"0", "1", "2", "3"},
			required:       nil,
			request:        2,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1"},
			},
			expectError: false,
		},
		{
			description:    "request two pre-allocated devices from socket 0 of 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"0", "1", "2", "3"},
			required:       []string{"2", "3"},
			request:        2,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"2", "3"},
			},
			expectError: false,
		},
		{
			description:    "request two devices(one is pre-allocated) from socket 0 of 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"0", "1", "2", "3"},
			required:       []string{"2"},
			request:        2,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"2", "3"},
			},
			expectError: false,
		},
		{
			description:    "request three devices from socket 0 of 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"0", "1", "2", "3"},
			required:       nil,
			request:        3,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "2"},
			},
			expectError: false,
		},
		{
			description:    "request three devices(one is pre-allocated) from socket 0 of 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"0", "1", "2", "3"},
			required:       []string{"3"},
			request:        3,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "3"},
			},
			expectError: false,
		},
		{
			description:    "request four devices from socket 0 of 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"0", "1", "2", "3"},
			required:       nil,
			request:        4,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "2", "3"},
			},
			expectError: false,
		},
		{
			description:    "request four devices(two are pre-allocated) from socket 0 of 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"0", "1", "2", "3"},
			required:       []string{"2", "3"},
			request:        4,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "2", "3"},
			},
			expectError: false,
		},
		// NOTE(@bg): skip pre-allocated cases for socket 1
		{
			description:    "request one device from socket 1 of 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"4", "5", "6", "7"},
			required:       nil,
			request:        1,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"4"},
			},
			expectError: false,
		},
		{
			description:    "request two devices from socket 1 of 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"4", "5", "6", "7"},
			required:       nil,
			request:        2,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"4", "5"},
			},
			expectError: false,
		},
		{
			description:    "request four devices from socket 1 of 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"4", "5", "6", "7"},
			required:       nil,
			request:        3,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"4", "5", "6"},
			},
			expectError: false,
		},
		{
			description:    "request four devices from socket 1 of 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"4", "5", "6", "7"},
			required:       nil,
			request:        4,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"4", "5", "6", "7"},
			},
			expectError: false,
		},
		// add cases for requesting devices across sockets
		{
			description:    "request five devices across 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			required:       nil,
			request:        5,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "2", "3", "4"},
			},
			expectError: false,
		},
		{
			description:    "request six devices across 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			required:       nil,
			request:        6,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "2", "3", "4", "5"},
			},
			expectError: false,
		},
		{
			description:    "request seven devices across 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			required:       nil,
			request:        7,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "2", "3", "4", "5", "6"},
			},
			expectError: false,
		},
		{
			description:    "request eight devices across 2 sockets",
			deviceSeedUUID: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			available:      []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			required:       nil,
			request:        8,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		mockDevices := MockDeviceSlices(len(tc.deviceSeedUUID), tc.deviceSeedUUID, nil)
		mockFuriosaDevices := MockFuriosaDevices(mockDevices)
		allocator, _ := npu_allocator.NewMockScoreBasedOptimalNpuAllocator(staticMockTopologyHintProvider())
		mockDeviceManager := &deviceManager{
			origin:         mockDevices,
			furiosaDevices: mockFuriosaDevices,
			resourceName:   "furiosa.ai/npu",
			debugMode:      false,
			allocator:      allocator,
		}

		actualResult, actualError := mockDeviceManager.GetContainerPreferredAllocationResponse(tc.available, tc.required, tc.request)
		if actualError != nil != tc.expectError {
			t.Errorf("unexpected error %t", actualError)
		}

		if !reflect.DeepEqual(actualResult, tc.expectedResult) {
			t.Errorf("expectedResult %v but got %v", tc.expectedResult, actualResult)
		}
	}
}

// TODO(@bg): add test cases for CDI
// TODO(@bg): add test cases for renegade
func TestGetContainerAllocateResponseForWarboy(t *testing.T) {
	tests := []struct {
		description    string
		deviceSeedUUID []string
		deviceIDs      []string
		expectedResult *devicePluginAPIv1Beta1.ContainerAllocateResponse
		expectError    bool
	}{
		{
			description:    "allocate one device",
			deviceSeedUUID: []string{"0"},
			deviceIDs:      []string{"0"},
			expectedResult: &devicePluginAPIv1Beta1.ContainerAllocateResponse{
				Envs: nil,
				Mounts: []*devicePluginAPIv1Beta1.Mount{
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
				Devices: []*devicePluginAPIv1Beta1.DeviceSpec{
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
				Annotations: nil,
				CDIDevices:  nil,
			},
			expectError: false,
		},
		{
			description:    "allocate two devices",
			deviceSeedUUID: []string{"0", "1"},
			deviceIDs:      []string{"0", "1"},
			expectedResult: &devicePluginAPIv1Beta1.ContainerAllocateResponse{
				Envs: nil,
				Mounts: []*devicePluginAPIv1Beta1.Mount{
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
					{
						ContainerPath: "/sys/class/npu_mgmt/npu1_mgmt",
						HostPath:      "/sys/class/npu_mgmt/npu1_mgmt",
						ReadOnly:      true,
					},
					{
						ContainerPath: "/sys/class/npu_mgmt/npu1",
						HostPath:      "/sys/class/npu_mgmt/npu1",
						ReadOnly:      true,
					},
					{
						ContainerPath: "/sys/class/npu_mgmt/npu1pe0",
						HostPath:      "/sys/class/npu_mgmt/npu1pe0",
						ReadOnly:      true,
					},
					{
						ContainerPath: "/sys/class/npu_mgmt/npu1pe1",
						HostPath:      "/sys/class/npu_mgmt/npu1pe1",
						ReadOnly:      true,
					},
					{
						ContainerPath: "/sys/class/npu_mgmt/npu1pe0-1",
						HostPath:      "/sys/class/npu_mgmt/npu1pe0-1",
						ReadOnly:      true,
					},
					{
						ContainerPath: "/sys/devices/virtual/npu_mgmt/npu1_mgmt",
						HostPath:      "/sys/devices/virtual/npu_mgmt/npu1_mgmt",
						ReadOnly:      true,
					},
					{
						ContainerPath: "/sys/devices/virtual/npu_mgmt/npu1",
						HostPath:      "/sys/devices/virtual/npu_mgmt/npu1",
						ReadOnly:      true,
					},
					{
						ContainerPath: "/sys/devices/virtual/npu_mgmt/npu1pe0",
						HostPath:      "/sys/devices/virtual/npu_mgmt/npu1pe0",
						ReadOnly:      true,
					},
					{
						ContainerPath: "/sys/devices/virtual/npu_mgmt/npu1pe1",
						HostPath:      "/sys/devices/virtual/npu_mgmt/npu1pe1",
						ReadOnly:      true,
					},
					{
						ContainerPath: "/sys/devices/virtual/npu_mgmt/npu1pe0-1",
						HostPath:      "/sys/devices/virtual/npu_mgmt/npu1pe0-1",
						ReadOnly:      true,
					},
				},
				Devices: []*devicePluginAPIv1Beta1.DeviceSpec{
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
					{
						ContainerPath: "/dev/npu1_mgmt",
						HostPath:      "/dev/npu1_mgmt",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/npu1",
						HostPath:      "/dev/npu1",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/npu1pe0",
						HostPath:      "/dev/npu1pe0",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/npu1pe1",
						HostPath:      "/dev/npu1pe1",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/npu1pe0-1",
						HostPath:      "/dev/npu1pe0-1",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/npu1ch0",
						HostPath:      "/dev/npu1ch0",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/npu1ch1",
						HostPath:      "/dev/npu1ch1",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/npu1ch2",
						HostPath:      "/dev/npu1ch2",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/npu1ch3",
						HostPath:      "/dev/npu1ch3",
						Permissions:   "rw",
					},
				},
				Annotations: nil,
				CDIDevices:  nil,
			},
			expectError: false,
		},
		// skip other cases
	}

	for _, tc := range tests {
		mockDevices := MockDeviceSlices(len(tc.deviceSeedUUID), tc.deviceSeedUUID, nil)
		mockFuriosaDevices := MockFuriosaDevices(mockDevices)
		mockDeviceManager := &deviceManager{
			origin:         mockDevices,
			furiosaDevices: mockFuriosaDevices,
			resourceName:   "furiosa.ai/npu",
			debugMode:      false,
			allocator:      nil,
		}

		actualResult, actualError := mockDeviceManager.GetContainerAllocateResponse(tc.deviceIDs)
		if actualError != nil != tc.expectError {
			t.Errorf("unexpected error %t", actualError)
		}

		if !reflect.DeepEqual(actualResult, tc.expectedResult) {
			t.Errorf("expectedResult %v but got %v", tc.expectedResult, actualResult)
		}
	}
}
