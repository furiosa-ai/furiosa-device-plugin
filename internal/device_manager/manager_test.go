package device_manager

import (
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/furiosa_device"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/npu_allocator"
	"github.com/stretchr/testify/assert"

	devicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func MockFuriosaDevices(mockDevices []smi.Device) (ret map[string]furiosa_device.FuriosaDevice) {
	if len(mockDevices) == 0 {
		mockDevices = smi.GetStaticMockDevices(smi.ArchRngd)
	}

	ret = make(map[string]furiosa_device.FuriosaDevice, len(mockDevices))

	mockFuriosaDevices, _ := furiosa_device.NewFuriosaDevices(mockDevices, nil, furiosa_device.NonePolicy)
	for _, mockFuriosaDevice := range mockFuriosaDevices {
		ret[mockFuriosaDevice.DeviceID()] = mockFuriosaDevice
	}

	return ret
}

func TestFetchByID(t *testing.T) {
	mockDevices := smi.GetStaticMockDevices(smi.ArchRngd)
	var seedUUID []string

	for i, mockDevice := range mockDevices {
		if i == 2 {
			break
		}

		info, _ := mockDevice.DeviceInfo()
		seedUUID = append(seedUUID, info.UUID())
	}

	mockFuriosaDevices := MockFuriosaDevices(mockDevices)
	actual, err := fetchByID(mockFuriosaDevices, seedUUID)
	assert.NoError(t, err)

	var actualIDs []string
	for _, furiosaDevice := range actual {
		actualIDs = append(actualIDs, furiosaDevice.DeviceID())
	}

	assert.Equal(t, seedUUID, actualIDs)
}

func TestFetchDevicesByID(t *testing.T) {
	mockDevices := smi.GetStaticMockDevices(smi.ArchRngd)
	var seedUUID []string

	for _, mockDevice := range mockDevices {
		info, _ := mockDevice.DeviceInfo()
		seedUUID = append(seedUUID, info.UUID())
	}

	mockFuriosaDevices := MockFuriosaDevices(mockDevices)
	actual, err := fetchDevicesByID(mockFuriosaDevices, seedUUID)
	assert.NoError(t, err)

	var actualIDs []string
	for _, d := range actual {
		actualIDs = append(actualIDs, d.ID())
	}

	assert.Equal(t, seedUUID, actualIDs)
}

// staticMockTopologyHintProvider build hint matrix for optimized 2socket server
// which has two pcie switches per socket and two devices per switch.
func staticMockTopologyHintProvider() npu_allocator.TopologyHintProvider {
	hints := map[npu_allocator.TopologyHintKey]map[npu_allocator.TopologyHintKey]uint{
		"27": {"27": 70, "2a": 30, "51": 20, "57": 20, "9e": 10, "a4": 10, "c7": 10, "ca": 10},
		"2a": {"2a": 70, "51": 20, "57": 20, "9e": 10, "a4": 10, "c7": 10, "ca": 10},
		"51": {"51": 70, "57": 30, "9e": 10, "a4": 10, "c7": 10, "ca": 10},
		"57": {"3": 70, "9e": 10, "a4": 10, "c7": 10, "ca": 10},
		"9e": {"9e": 70, "a4": 30, "c7": 20, "ca": 20},
		"a4": {"a4": 70, "c7": 20, "ca": 20},
		"c7": {"c7": 70, "ca": 30},
		"ca": {"ca": 70},
	}
	return func(device1, device2 npu_allocator.Device) uint {
		topologyHintKey1 := device1.TopologyHintKey()
		topologyHintKey2 := device2.TopologyHintKey()

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

func prefix(prefix string, origin []string) []string {
	var ret []string

	for _, ele := range origin {
		ret = append(ret, prefix+ele)
	}

	return ret
}

func TestGetContainerPreferredAllocationResponseWithScoreBasedOptimalNpuAllocator(t *testing.T) {
	tests := []struct {
		description    string
		available      []string
		required       []string
		request        int
		expectedResult *devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse
		expectError    bool
	}{
		// start with socket 0
		{
			description: "request one device from socket 0 of 2 sockets",
			available:   []string{"0", "1", "2", "3"},
			required:    nil,
			request:     1,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0"},
			},
			expectError: false,
		},
		{
			description: "request one pre-allocated device from socket 0 of 2 sockets",
			available:   []string{"0", "1", "2", "3"},
			required:    []string{"3"},
			request:     1,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"3"},
			},
			expectError: false,
		},
		{
			description: "request two devices from socket 0 of 2 sockets",
			available:   []string{"0", "1", "2", "3"},
			required:    nil,
			request:     2,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1"},
			},
			expectError: false,
		},
		{
			description: "request two pre-allocated devices from socket 0 of 2 sockets",
			available:   []string{"0", "1", "2", "3"},
			required:    []string{"2", "3"},
			request:     2,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"2", "3"},
			},
			expectError: false,
		},
		{
			description: "request two devices(one is pre-allocated) from socket 0 of 2 sockets",
			available:   []string{"0", "1", "2", "3"},
			required:    []string{"2"},
			request:     2,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"2", "3"},
			},
			expectError: false,
		},
		{
			description: "request three devices from socket 0 of 2 sockets",
			available:   []string{"0", "1", "2", "3"},
			required:    nil,
			request:     3,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "2"},
			},
			expectError: false,
		},
		{
			description: "request three devices(one is pre-allocated) from socket 0 of 2 sockets",
			available:   []string{"0", "1", "2", "3"},
			required:    []string{"3"},
			request:     3,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "3"},
			},
			expectError: false,
		},
		{
			description: "request four devices from socket 0 of 2 sockets",
			available:   []string{"0", "1", "2", "3"},
			required:    nil,
			request:     4,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "2", "3"},
			},
			expectError: false,
		},
		{
			description: "request four devices(two are pre-allocated) from socket 0 of 2 sockets",
			available:   []string{"0", "1", "2", "3"},
			required:    []string{"2", "3"},
			request:     4,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "2", "3"},
			},
			expectError: false,
		},
		// NOTE(@bg): skip pre-allocated cases for socket 1
		{
			description: "request one device from socket 1 of 2 sockets",
			available:   []string{"4", "5", "6", "7"},
			required:    nil,
			request:     1,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"4"},
			},
			expectError: false,
		},
		{
			description: "request two devices from socket 1 of 2 sockets",
			available:   []string{"4", "5", "6", "7"},
			required:    nil,
			request:     2,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"4", "5"},
			},
			expectError: false,
		},
		{
			description: "request four devices from socket 1 of 2 sockets",
			available:   []string{"4", "5", "6", "7"},
			required:    nil,
			request:     3,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"4", "5", "6"},
			},
			expectError: false,
		},
		{
			description: "request four devices from socket 1 of 2 sockets",
			available:   []string{"4", "5", "6", "7"},
			required:    nil,
			request:     4,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"4", "5", "6", "7"},
			},
			expectError: false,
		},
		// add cases for requesting devices across sockets
		{
			description: "request five devices across 2 sockets",
			available:   []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			required:    nil,
			request:     5,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "2", "3", "4"},
			},
			expectError: false,
		},
		{
			description: "request six devices across 2 sockets",
			available:   []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			required:    nil,
			request:     6,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "2", "3", "4", "5"},
			},
			expectError: false,
		},
		{
			description: "request seven devices across 2 sockets",
			available:   []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			required:    nil,
			request:     7,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "2", "3", "4", "5", "6"},
			},
			expectError: false,
		},
		{
			description: "request eight devices across 2 sockets",
			available:   []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			required:    nil,
			request:     8,
			expectedResult: &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
				DeviceIDs: []string{"0", "1", "2", "3", "4", "5", "6", "7"},
			},
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			mockDevices := smi.GetStaticMockDevices(smi.ArchRngd)
			mockFuriosaDevices := MockFuriosaDevices(mockDevices)
			allocator, _ := npu_allocator.NewMockScoreBasedOptimalNpuAllocator(staticMockTopologyHintProvider())
			mockDeviceManager := &deviceManager{
				origin:         mockDevices,
				furiosaDevices: mockFuriosaDevices,
				resourceName:   "furiosa.ai/npu",
				debugMode:      false,
				allocator:      allocator,
			}

			completeAvailable := prefix("A76AAD68-6855-40B1-9E86-D080852D1C8", tc.available)
			completeRequired := prefix("A76AAD68-6855-40B1-9E86-D080852D1C8", tc.required)
			actualResult, actualError := mockDeviceManager.GetContainerPreferredAllocationResponse(completeAvailable, completeRequired, tc.request)
			if actualError != nil != tc.expectError {
				t.Errorf("unexpected error %t", actualError)
			}

			completeExpectedResult := prefix("A76AAD68-6855-40B1-9E86-D080852D1C8", tc.expectedResult.DeviceIDs)

			assert.Equal(t, completeExpectedResult, actualResult.DeviceIDs)
		})
	}
}

// TODO(@bg): add test cases for CDI
func TestGetContainerAllocateResponseForWarboy(t *testing.T) {
	tests := []struct {
		description    string
		deviceIDs      []string
		expectedResult *devicePluginAPIv1Beta1.ContainerAllocateResponse
		expectError    bool
	}{
		{
			description: "allocate one device",
			deviceIDs:   []string{"0"},
			expectedResult: &devicePluginAPIv1Beta1.ContainerAllocateResponse{
				Envs:   nil,
				Mounts: nil,
				Devices: []*devicePluginAPIv1Beta1.DeviceSpec{
					{
						ContainerPath: "/dev/rngd/npu0mgmt",
						HostPath:      "/dev/rngd/npu0mgmt",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0pe0",
						HostPath:      "/dev/rngd/npu0pe0",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0pe1",
						HostPath:      "/dev/rngd/npu0pe1",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0pe0-1",
						HostPath:      "/dev/rngd/npu0pe0-1",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0pe2",
						HostPath:      "/dev/rngd/npu0pe2",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0pe3",
						HostPath:      "/dev/rngd/npu0pe3",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0pe2-3",
						HostPath:      "/dev/rngd/npu0pe2-3",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0pe0-3",
						HostPath:      "/dev/rngd/npu0pe0-3",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0pe4",
						HostPath:      "/dev/rngd/npu0pe4",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0pe5",
						HostPath:      "/dev/rngd/npu0pe5",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0pe4-5",
						HostPath:      "/dev/rngd/npu0pe4-5",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0pe6",
						HostPath:      "/dev/rngd/npu0pe6",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0pe7",
						HostPath:      "/dev/rngd/npu0pe7",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0pe6-7",
						HostPath:      "/dev/rngd/npu0pe6-7",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0pe4-7",
						HostPath:      "/dev/rngd/npu0pe4-7",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch0",
						HostPath:      "/dev/rngd/npu0ch0",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch1",
						HostPath:      "/dev/rngd/npu0ch1",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch2",
						HostPath:      "/dev/rngd/npu0ch2",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch3",
						HostPath:      "/dev/rngd/npu0ch3",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch4",
						HostPath:      "/dev/rngd/npu0ch4",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch5",
						HostPath:      "/dev/rngd/npu0ch5",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch6",
						HostPath:      "/dev/rngd/npu0ch6",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch7",
						HostPath:      "/dev/rngd/npu0ch7",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch0r",
						HostPath:      "/dev/rngd/npu0ch0r",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch1r",
						HostPath:      "/dev/rngd/npu0ch1r",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch2r",
						HostPath:      "/dev/rngd/npu0ch2r",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch3r",
						HostPath:      "/dev/rngd/npu0ch3r",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch4r",
						HostPath:      "/dev/rngd/npu0ch4r",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch5r",
						HostPath:      "/dev/rngd/npu0ch5r",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch6r",
						HostPath:      "/dev/rngd/npu0ch6r",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0ch7r",
						HostPath:      "/dev/rngd/npu0ch7r",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0dmar",
						HostPath:      "/dev/rngd/npu0dmar",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0bar0",
						HostPath:      "/dev/rngd/npu0bar0",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0bar2",
						HostPath:      "/dev/rngd/npu0bar2",
						Permissions:   "rw",
					},
					{
						ContainerPath: "/dev/rngd/npu0bar4",
						HostPath:      "/dev/rngd/npu0bar4",
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
		t.Run(tc.description, func(t *testing.T) {
			mockDevices := smi.GetStaticMockDevices(smi.ArchRngd)
			mockFuriosaDevices := MockFuriosaDevices(mockDevices)
			mockDeviceManager := &deviceManager{
				origin:         mockDevices,
				furiosaDevices: mockFuriosaDevices,
				resourceName:   "furiosa.ai/npu",
				debugMode:      false,
				allocator:      nil,
			}

			actualResult, actualError := mockDeviceManager.GetContainerAllocateResponse(prefix("A76AAD68-6855-40B1-9E86-D080852D1C8", tc.deviceIDs))
			if tc.expectError {
				assert.Error(t, actualError)
			} else {
				assert.NoError(t, actualError)
			}

			assert.Equal(t, tc.expectedResult, actualResult)
		})
	}
}
