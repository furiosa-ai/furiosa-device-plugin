package device_manager

import (
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/npu_allocator"
	"github.com/stretchr/testify/assert"
	devicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	totalCoresOfRNGD = 8
)

func TestFinalIndexGeneration_RNGD_PartitionedDevice(t *testing.T) {
	rngdMockDevices := smi.GetStaticMockDevices(smi.ArchRngd)

	tests := []struct {
		description                  string
		strategy                     config.ResourceUnitStrategy
		expectedIndexes              []int
		expectedIndexToDeviceUUIDMap map[int]string // key: index, value: uuid
	}{
		{
			description: "Single Core Strategy",
			strategy:    config.SingleCoreStrategy,
			expectedIndexes: func() []int {
				indexes := make([]int, 64)
				for i := range indexes {
					indexes[i] = i
				}

				return indexes
			}(),
			expectedIndexToDeviceUUIDMap: func() map[int]string {
				mapping := make(map[int]string)
				for i := 0; i < 64; i++ {
					deviceInfo, _ := rngdMockDevices[i/8].DeviceInfo()
					mapping[i] = deviceInfo.UUID()
				}

				return mapping
			}(),
		},
		{
			description: "Dual Core Strategy",
			strategy:    config.DualCoreStrategy,
			expectedIndexes: func() []int {
				indexes := make([]int, 32)
				for i := range indexes {
					indexes[i] = i
				}

				return indexes
			}(),
			expectedIndexToDeviceUUIDMap: func() map[int]string {
				mapping := make(map[int]string)
				for i := 0; i < 32; i++ {
					deviceInfo, _ := rngdMockDevices[i/4].DeviceInfo()
					mapping[i] = deviceInfo.UUID()
				}

				return mapping
			}(),
		},
		{
			description: "Quad Core Strategy",
			strategy:    config.QuadCoreStrategy,
			expectedIndexes: func() []int {
				indexes := make([]int, 16)
				for i := range indexes {
					indexes[i] = i
				}

				return indexes
			}(),
			expectedIndexToDeviceUUIDMap: func() map[int]string {
				mapping := make(map[int]string)
				for i := 0; i < 16; i++ {
					deviceInfo, _ := rngdMockDevices[i/2].DeviceInfo()
					mapping[i] = deviceInfo.UUID()
				}

				return mapping
			}(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			deviceMgr, _ := NewDeviceManager(smi.ArchRngd, rngdMockDevices, tc.strategy, nil, false)

			furiosaDeviceMap := deviceMgr.(*deviceManager).furiosaDevices
			furiosaDevices := make([]FuriosaDevice, 0, len(furiosaDeviceMap))
			for _, device := range furiosaDeviceMap {
				furiosaDevices = append(furiosaDevices, device)
			}

			slices.SortFunc(furiosaDevices, func(dev1, dev2 FuriosaDevice) int {
				return dev1.Index() - dev2.Index()
			})

			finalIndexes := make([]int, 0, len(furiosaDevices))
			for _, device := range furiosaDevices {
				finalIndexes = append(finalIndexes, device.Index())
			}

			assert.Equal(t, tc.expectedIndexes, finalIndexes)

			finalIndexToDeviceUUIDMap := make(map[int]string)
			for _, furiosaDevice := range furiosaDevices {
				finalIndexToDeviceUUIDMap[furiosaDevice.Index()] = furiosaDevice.(*partitionedDevice).uuid
			}

			assert.Equal(t, tc.expectedIndexToDeviceUUIDMap, finalIndexToDeviceUUIDMap)
		})
	}
}

func TestDeviceIDs_RNGD_PartitionedDevice(t *testing.T) {
	rngdMockDevice := smi.GetStaticMockDevices(smi.ArchRngd)[0]
	rngdMockDeviceUUID := "A76AAD68-6855-40B1-9E86-D080852D1C80"

	tests := []struct {
		description     string
		mockDevice      smi.Device
		strategy        config.ResourceUnitStrategy
		expectedResults []string
	}{
		{
			description: "should return a list of RNGD Device ID for single core strategy",
			mockDevice:  rngdMockDevice,
			strategy:    config.SingleCoreStrategy,
			expectedResults: []string{
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "0"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "1"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "2"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "3"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "4"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "5"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "6"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "7"),
			},
		},
		{
			description: "should return a list of RNGD Device ID for dual core strategy",
			mockDevice:  rngdMockDevice,
			strategy:    config.DualCoreStrategy,
			expectedResults: []string{
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "0-1"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "2-3"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "4-5"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "6-7"),
			},
		},
		{
			description: "should return a list of RNGD Device ID for quad core strategy",
			mockDevice:  rngdMockDevice,
			strategy:    config.QuadCoreStrategy,
			expectedResults: []string{
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "0-3"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "4-7"),
			},
		},
	}

	for _, tc := range tests {
		numOfCoresPerPartition := tc.strategy.CoreSize()
		partitionedDevices, err := NewPartitionedDevices(0, tc.mockDevice, numOfCoresPerPartition, totalCoresOfRNGD/numOfCoresPerPartition, false)
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}

		if len(partitionedDevices) != len(tc.expectedResults) {
			t.Errorf("length of expectedResults and partitioned devices are not equal for strategy %s: expected: %d, got: %d", tc.strategy, len(tc.expectedResults), len(partitionedDevices))
			continue
		}

		for i, device := range partitionedDevices {
			expectedDeviceId := tc.expectedResults[i]
			actualDeviceId := device.DeviceID()
			if expectedDeviceId != actualDeviceId {
				t.Errorf("expected Device ID %s, got %s", expectedDeviceId, actualDeviceId)
				continue
			}
		}
	}
}

func TestPCIBusIDs_RNGD_PartitionedDevice(t *testing.T) {
	rngdMockDevice0 := smi.GetStaticMockDevices(smi.ArchRngd)[0]
	rngdMockDevice0PciBusId := "27"

	rngdMockDevice1 := smi.GetStaticMockDevices(smi.ArchRngd)[1]
	rngdMockDevice1PciBusId := "2a"

	tests := []struct {
		description    string
		mockDevice     smi.Device
		strategy       config.ResourceUnitStrategy
		expectedResult string
	}{
		{
			description:    "returned devices must have same PCI Bus IDs - RNGD 0",
			mockDevice:     rngdMockDevice0,
			strategy:       config.SingleCoreStrategy,
			expectedResult: rngdMockDevice0PciBusId,
		},
		{
			description:    "returned devices must have same PCI Bus IDs - RNGD 1",
			mockDevice:     rngdMockDevice1,
			strategy:       config.SingleCoreStrategy,
			expectedResult: rngdMockDevice1PciBusId,
		},
	}

	for _, tc := range tests {
		numOfCoresPerPartition := tc.strategy.CoreSize()
		partitionedDevices, err := NewPartitionedDevices(0, tc.mockDevice, numOfCoresPerPartition, totalCoresOfRNGD/numOfCoresPerPartition, false)
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}

		expectedPCIBusID := tc.expectedResult
		for _, device := range partitionedDevices {
			actualPCIBusID := device.PCIBusID()
			if expectedPCIBusID != actualPCIBusID {
				t.Errorf("expected PCIBusID %s, got %s", expectedPCIBusID, actualPCIBusID)
				continue
			}
		}
	}
}

func TestNUMANode_RNGD_PartitionedDevice(t *testing.T) {
	rngdMockDevice0 := smi.GetStaticMockDevices(smi.ArchRngd)[0]
	rngdMockDevice0NUMANode := 0

	rngdMockDevice1 := smi.GetStaticMockDevices(smi.ArchRngd)[4]
	rngdMockDevice1NUMANode := 1

	tests := []struct {
		description    string
		mockDevice     smi.Device
		strategy       config.ResourceUnitStrategy
		expectedResult int
	}{
		{
			description:    "returned devices must have same NUMA node - RNGD 0",
			mockDevice:     rngdMockDevice0,
			strategy:       config.SingleCoreStrategy,
			expectedResult: rngdMockDevice0NUMANode,
		},
		{
			description:    "returned devices must have same NUMA node - RNGD 1",
			mockDevice:     rngdMockDevice1,
			strategy:       config.SingleCoreStrategy,
			expectedResult: rngdMockDevice1NUMANode,
		},
	}

	for _, tc := range tests {
		numOfCoresPerPartition := tc.strategy.CoreSize()
		partitionedDevices, err := NewPartitionedDevices(0, tc.mockDevice, numOfCoresPerPartition, totalCoresOfRNGD/numOfCoresPerPartition, true)
		if err != nil {
			t.Errorf("unexpected error %t", err)
			continue
		}

		expectedNUMANode := tc.expectedResult
		for _, device := range partitionedDevices {
			actualNUMANode := device.NUMANode()
			if expectedNUMANode != actualNUMANode {
				t.Errorf("expected NUMA node %d, got %d", expectedNUMANode, actualNUMANode)
				continue
			}
		}
	}
}

func TestDeviceSpecs_RNGD_PartitionedDevice(t *testing.T) {
	rngdMockDevice := smi.GetStaticMockDevices(smi.ArchRngd)[0]

	rngdExpectedResultCandidatesForSingleCoreStrategy := func() [][]*devicePluginAPIv1Beta1.DeviceSpec {
		candidates := make([][]*devicePluginAPIv1Beta1.DeviceSpec, 0, 8)
		for i := 0; i < 8; i++ {
			candidates = append(candidates, []*devicePluginAPIv1Beta1.DeviceSpec{
				{
					ContainerPath: "/dev/rngd/npu0mgmt",
					HostPath:      "/dev/rngd/npu0mgmt",
					Permissions:   "rw",
				},
				{
					ContainerPath: fmt.Sprintf("/dev/rngd/npu0pe%d", i),
					HostPath:      fmt.Sprintf("/dev/rngd/npu0pe%d", i),
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
			})
		}

		return candidates
	}()

	rngdExpectedResultCandidatesForDualCoreStrategy := func() [][]*devicePluginAPIv1Beta1.DeviceSpec {
		candidates := make([][]*devicePluginAPIv1Beta1.DeviceSpec, 0, 8)
		for i := 0; i < 8; i += 2 {
			candidates = append(candidates, []*devicePluginAPIv1Beta1.DeviceSpec{
				{
					ContainerPath: "/dev/rngd/npu0mgmt",
					HostPath:      "/dev/rngd/npu0mgmt",
					Permissions:   "rw",
				},
				{
					ContainerPath: fmt.Sprintf("/dev/rngd/npu0pe%d-%d", i, i+1),
					HostPath:      fmt.Sprintf("/dev/rngd/npu0pe%d-%d", i, i+1),
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
			})
		}

		return candidates
	}()

	rngdExpectedResultCandidatesForQuadCoreStrategy := func() [][]*devicePluginAPIv1Beta1.DeviceSpec {
		candidates := make([][]*devicePluginAPIv1Beta1.DeviceSpec, 0, 8)
		for i := 0; i < 8; i += 4 {
			candidates = append(candidates, []*devicePluginAPIv1Beta1.DeviceSpec{
				{
					ContainerPath: "/dev/rngd/npu0mgmt",
					HostPath:      "/dev/rngd/npu0mgmt",
					Permissions:   "rw",
				},
				{
					ContainerPath: fmt.Sprintf("/dev/rngd/npu0pe%d-%d", i, i+3),
					HostPath:      fmt.Sprintf("/dev/rngd/npu0pe%d-%d", i, i+3),
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
			})
		}

		return candidates
	}()

	tests := []struct {
		description              string
		mockDevice               smi.Device
		strategy                 config.ResourceUnitStrategy
		expectedResultCandidates [][]*devicePluginAPIv1Beta1.DeviceSpec
	}{
		{
			description:              "[SingleCoreStrategy] each RNGD mock device must contains all DeviceSpecs",
			mockDevice:               rngdMockDevice,
			strategy:                 config.SingleCoreStrategy,
			expectedResultCandidates: rngdExpectedResultCandidatesForSingleCoreStrategy,
		},
		{
			description:              "[DualCoreStrategy] each RNGD mock device must contains all DeviceSpecs",
			mockDevice:               rngdMockDevice,
			strategy:                 config.DualCoreStrategy,
			expectedResultCandidates: rngdExpectedResultCandidatesForDualCoreStrategy,
		},
		{
			description:              "[QuadCoreStrategy] each RNGD mock device must contains all DeviceSpecs",
			mockDevice:               rngdMockDevice,
			strategy:                 config.QuadCoreStrategy,
			expectedResultCandidates: rngdExpectedResultCandidatesForQuadCoreStrategy,
		},
	}

	for _, tc := range tests {
		numOfCoresPerPartition := tc.strategy.CoreSize()
		partitionedDevices, err := NewPartitionedDevices(0, tc.mockDevice, numOfCoresPerPartition, totalCoresOfRNGD/numOfCoresPerPartition, false)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.description, err)
			continue
		}

		if len(partitionedDevices) != len(tc.expectedResultCandidates) {
			t.Errorf("%s: expected %d partitioned devices, got %d", tc.description, len(tc.expectedResultCandidates), len(partitionedDevices))
		}

		for i, device := range partitionedDevices {
			actualResult := device.DeviceSpecs()
			if !reflect.DeepEqual(actualResult, tc.expectedResultCandidates[i]) {
				t.Errorf("%s: expected %v, got %v", tc.description, tc.expectedResultCandidates[i], actualResult)
			}
		}
	}
}

func TestIsHealthy_RNGD_PartitionedDevice(t *testing.T) {
	tests := []struct {
		description     string
		mockDevice      smi.Device
		strategy        config.ResourceUnitStrategy
		isDisabled      bool
		expectedResults bool
	}{
		{
			description:     "Enabled device must be healthy - RNGD",
			mockDevice:      smi.GetStaticMockDevices(smi.ArchRngd)[0],
			strategy:        config.SingleCoreStrategy,
			isDisabled:      true,
			expectedResults: false,
		},
		{
			description:     "Disabled device must be unhealthy - RNGD",
			mockDevice:      smi.GetStaticMockDevices(smi.ArchRngd)[0],
			strategy:        config.SingleCoreStrategy,
			isDisabled:      true,
			expectedResults: false,
		},
	}

	for _, tc := range tests {
		numOfCoresPerPartition := tc.strategy.CoreSize()
		partitionedDevices, err := NewPartitionedDevices(0, tc.mockDevice, numOfCoresPerPartition, totalCoresOfRNGD/numOfCoresPerPartition, tc.isDisabled)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.description, err)
			continue
		}

		for _, device := range partitionedDevices {
			actualResult, err := device.IsHealthy()
			if err != nil {
				t.Errorf("%s: unexpected error: %v", tc.description, err)
				continue
			}

			if actualResult != tc.expectedResults {
				t.Errorf("expectedResults %v but got %v", tc.expectedResults, actualResult)
			}
		}
	}
}

func TestMounts_RNGD_PartitionedDevice(t *testing.T) {
	rngdMockDevice := smi.GetStaticMockDevices(smi.ArchRngd)[0]
	rngdMockDeviceMounts := []*devicePluginAPIv1Beta1.Mount{
		{
			ContainerPath: "/sys/class/rngd_mgmt/rngd!npu0mgmt",
			HostPath:      "/sys/class/rngd_mgmt/rngd!npu0mgmt",
			ReadOnly:      true,
		},
		{
			ContainerPath: "/sys/devices/virtual/rngd_mgmt/rngd!npu0mgmt",
			HostPath:      "/sys/devices/virtual/rngd_mgmt/rngd!npu0mgmt",
			ReadOnly:      true,
		},
	}

	tests := []struct {
		description     string
		mockDevice      smi.Device
		strategy        config.ResourceUnitStrategy
		expectedResults []*devicePluginAPIv1Beta1.Mount
	}{
		{
			description:     "each RNGD mock device must contains all Mounts",
			mockDevice:      rngdMockDevice,
			strategy:        config.SingleCoreStrategy,
			expectedResults: rngdMockDeviceMounts,
		},
	}

	for _, tc := range tests {
		numOfCoresPerPartition := tc.strategy.CoreSize()
		partitionedDevices, err := NewPartitionedDevices(0, tc.mockDevice, numOfCoresPerPartition, totalCoresOfRNGD/numOfCoresPerPartition, false)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.description, err)
			continue
		}

		for _, device := range partitionedDevices {
			actualResults := device.Mounts()
			if !reflect.DeepEqual(actualResults, tc.expectedResults) {
				t.Errorf("expectedResults %v but got %v", tc.expectedResults, actualResults)
			}
		}
	}
}

func TestID_RNGD_PartitionedDevice(t *testing.T) {
	rngdMockDevice := smi.GetStaticMockDevices(smi.ArchRngd)[0]
	rngdMockDeviceUUID := "A76AAD68-6855-40B1-9E86-D080852D1C80"

	tests := []struct {
		description     string
		mockDevice      smi.Device
		strategy        config.ResourceUnitStrategy
		expectedResults []string
	}{
		{
			description: "should return a list of RNGD Device ID for single core strategy",
			mockDevice:  rngdMockDevice,
			strategy:    config.SingleCoreStrategy,
			expectedResults: []string{
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "0"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "1"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "2"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "3"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "4"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "5"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "6"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "7"),
			},
		},
		{
			description: "should return a list of RNGD Device ID for dual core strategy",
			mockDevice:  rngdMockDevice,
			strategy:    config.DualCoreStrategy,
			expectedResults: []string{
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "0-1"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "2-3"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "4-5"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "6-7"),
			},
		},
		{
			description: "should return a list of RNGD Device ID for quad core strategy",
			mockDevice:  rngdMockDevice,
			strategy:    config.QuadCoreStrategy,
			expectedResults: []string{
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "0-3"),
				fmt.Sprintf("%s%s%s", rngdMockDeviceUUID, deviceIdDelimiter, "4-7"),
			},
		},
	}

	for _, tc := range tests {
		numOfCoresPerPartition := tc.strategy.CoreSize()
		partitionedDevices, err := NewPartitionedDevices(0, tc.mockDevice, numOfCoresPerPartition, totalCoresOfRNGD/numOfCoresPerPartition, false)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.description, err)
			continue
		}

		if len(partitionedDevices) != len(tc.expectedResults) {
			t.Errorf("length of expectedResults and partitioned devices are not equal for strategy %s: expected: %d, got: %d", tc.strategy, len(tc.expectedResults), len(partitionedDevices))
			continue
		}

		for i, device := range partitionedDevices {
			expectedId := tc.expectedResults[i]
			actualId := device.ID()
			if expectedId != actualId {
				t.Errorf("expected ID %s, got %s", expectedId, actualId)
				continue
			}
		}
	}
}

func TestTopologyHintKey_RNGD_PartitionedDevice(t *testing.T) {
	rngdMockDevice0 := smi.GetStaticMockDevices(smi.ArchRngd)[0]
	rngdMockDevice0PciBusId := "27"

	rngdMockDevice1 := smi.GetStaticMockDevices(smi.ArchRngd)[1]
	rngdMockDevice1PciBusId := "2a"

	tests := []struct {
		description    string
		mockDevice     smi.Device
		strategy       config.ResourceUnitStrategy
		expectedResult npu_allocator.TopologyHintKey
	}{
		{
			description:    "returned devices must have same TopologyHintKeys - RNGD 0",
			mockDevice:     rngdMockDevice0,
			strategy:       config.SingleCoreStrategy,
			expectedResult: npu_allocator.TopologyHintKey(rngdMockDevice0PciBusId),
		},
		{
			description:    "returned devices must have same TopologyHintKeys - RNGD 1",
			mockDevice:     rngdMockDevice1,
			strategy:       config.SingleCoreStrategy,
			expectedResult: npu_allocator.TopologyHintKey(rngdMockDevice1PciBusId),
		},
	}

	for _, tc := range tests {
		numOfCoresPerPartition := tc.strategy.CoreSize()
		partitionedDevices, err := NewPartitionedDevices(0, tc.mockDevice, numOfCoresPerPartition, totalCoresOfRNGD/numOfCoresPerPartition, false)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.description, err)
			continue
		}

		for _, device := range partitionedDevices {
			actualResult := device.TopologyHintKey()
			if actualResult != tc.expectedResult {
				t.Errorf("expectedResults %s, got %s", tc.expectedResult, actualResult)
				continue
			}
		}
	}
}

func TestEqual_RNGD_PartitionedDevice(t *testing.T) {
	tests := []struct {
		description      string
		mockSourceDevice smi.Device
		mockTargetDevice smi.Device
		strategy         config.ResourceUnitStrategy
		expected         bool
	}{
		{
			description:      "expect source and target are identical",
			mockSourceDevice: smi.GetStaticMockDevices(smi.ArchRngd)[0],
			mockTargetDevice: smi.GetStaticMockDevices(smi.ArchRngd)[0],
			strategy:         config.SingleCoreStrategy,
			expected:         true,
		},
		{
			description:      "expect source and target are not identical",
			mockSourceDevice: smi.GetStaticMockDevices(smi.ArchRngd)[0],
			mockTargetDevice: smi.GetStaticMockDevices(smi.ArchRngd)[1],
			strategy:         config.SingleCoreStrategy,
			expected:         false,
		},
	}

	for _, tc := range tests {
		numOfCoresPerPartition := tc.strategy.CoreSize()

		sourcePartitionedDevices, err := NewPartitionedDevices(0, tc.mockSourceDevice, numOfCoresPerPartition, totalCoresOfRNGD/numOfCoresPerPartition, false)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.description, err)
			continue
		}

		targetPartitionedDevices, err := NewPartitionedDevices(0, tc.mockTargetDevice, numOfCoresPerPartition, totalCoresOfRNGD/numOfCoresPerPartition, false)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.description, err)
			continue
		}

		if len(sourcePartitionedDevices) != len(targetPartitionedDevices) {
			t.Errorf("length of sourcePartitionedDevices and targetPartitionedDevices is different!")
			continue
		}

		for i := range sourcePartitionedDevices {
			sourceDevice := sourcePartitionedDevices[i]
			targetDevice := targetPartitionedDevices[i]

			actualResult := sourceDevice.Equal(targetDevice)
			if actualResult != tc.expected {
				t.Errorf("expected: %v, got: %v", tc.expected, actualResult)
				continue
			}
		}
	}
}
