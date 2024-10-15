package device_manager

import (
	"fmt"

	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/manifest"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/npu_allocator"
	devicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

var _ FuriosaDevice = (*partitionedDevice)(nil)

// e.g. If UUID is a3e78042-9cc7-4344-9541-d2d3ffd28106 and Partition is 0-1,
// DeviceID should be "a3e78042-9cc7-4344-9541-d2d3ffd28106_cores_0-1".
const deviceIdDelimiter = "_cores_"

type Partition string

var (
	WarboySingleCorePartitions = []Partition{"0", "1"}
	WarboyDualCorePartitions   = []Partition{"0-1"}

	RngdSingleCorePartitions = []Partition{"0", "1", "2", "3", "4", "5", "6", "7"}
	RngdDualCorePartitions   = []Partition{"0-1", "2-3", "4-5", "6-7"}
	RngdQuadCorePartitions   = []Partition{"0-3", "4-7"}
)

// generatePartitionsByArchAndStrategy generates N partitions based on architecture and strategy.
// If specific strategy is not applicable to given architecture, it returns error.
func generatePartitionsByArchAndStrategy(arch smi.Arch, strategy config.ResourceUnitStrategy) ([]Partition, error) {
	switch strategy {
	case config.SingleCoreStrategy:
		switch arch {
		case smi.ArchWarboy: // warboy: 0, 1
			return WarboySingleCorePartitions, nil

		case smi.ArchRngd: // rngd: 0, 1, 2, 3, 4, 5, 6, 7
			return RngdSingleCorePartitions, nil
		}

	case config.DualCoreStrategy:
		switch arch {
		case smi.ArchWarboy: // warboy: 0-1
			return WarboyDualCorePartitions, nil

		case smi.ArchRngd: // rngd: 0-1, 2-3, 4-5, 6-7
			return RngdDualCorePartitions, nil
		}

	case config.QuadCoreStrategy:
		switch arch {
		case smi.ArchWarboy: // Warboy only supports SingleCore and DualCore strategy.
			return nil, fmt.Errorf("warboy only supports single-core and dual-core strategies")

		case smi.ArchRngd: // rngd: 0-3, 4-7
			return RngdQuadCorePartitions, nil
		}
	}

	// should not reach here!
	return nil, fmt.Errorf("unsupported strategy %s for architecture %s", strategy, arch.ToString())
}

type partitionedDevice struct {
	origin     smi.Device
	manifest   manifest.Manifest
	uuid       string
	partition  Partition // e.g. "0", "1", "0-1", "2-3", "0-3", ...
	strategy   config.ResourceUnitStrategy
	pciBusID   string
	numaNode   int
	isDisabled bool
}

// NewPartitionedDevices returns list of FuriosaDevice based on given config.ResourceUnitStrategy.
func NewPartitionedDevices(originDevice smi.Device, strategy config.ResourceUnitStrategy, isDisabled bool) ([]FuriosaDevice, error) {
	arch, uuid, pciBusID, numaNode, err := parseOriginDeviceInfo(originDevice)
	if err != nil {
		return nil, err
	}

	var originalManifest manifest.Manifest
	var manifestErr error

	// This block checks architecture and gets manifest of it.
	// If architecture is invalid, it returns error.
	switch arch {
	case smi.ArchWarboy:
		originalManifest, manifestErr = manifest.NewWarboyManifest(originDevice)

	case smi.ArchRngd:
		originalManifest, manifestErr = manifest.NewRngdManifest(originDevice)

	default:
		return nil, fmt.Errorf("unsupported architecture: %s", arch.ToString())
	}

	if manifestErr != nil {
		return nil, manifestErr
	}

	partitionedDevices := make([]FuriosaDevice, 0)
	partitions, err := generatePartitionsByArchAndStrategy(arch, strategy)
	if err != nil {
		return nil, err
	}

	for _, partition := range partitions {
		partitionedManifest, err := NewPartitionedDeviceManifest(arch, originalManifest, partition)
		if err != nil {
			return nil, err
		}

		partitionedDevices = append(partitionedDevices, &partitionedDevice{
			origin:     originDevice,
			manifest:   partitionedManifest,
			uuid:       uuid,
			partition:  partition,
			strategy:   strategy,
			pciBusID:   pciBusID,
			numaNode:   int(numaNode),
			isDisabled: isDisabled,
		})
	}

	return partitionedDevices, nil
}

func (p *partitionedDevice) DeviceID() string {
	// e.g. If UUID is a3e78042-9cc7-4344-9541-d2d3ffd28106 and Partition is 0-1,
	// DeviceID should be "a3e78042-9cc7-4344-9541-d2d3ffd28106_cores_0-1".
	return fmt.Sprintf("%s%s%s", p.uuid, deviceIdDelimiter, p.partition)
}

func (p *partitionedDevice) PCIBusID() string {
	return p.pciBusID
}

func (p *partitionedDevice) NUMANode() int {
	return p.numaNode
}

func (p *partitionedDevice) IsHealthy() (bool, error) {
	if p.isDisabled {
		return false, nil
	}

	liveness, err := p.origin.Liveness()
	if err != nil {
		return liveness, err
	}

	return liveness, nil
}

func (p *partitionedDevice) IsExclusiveDevice() bool {
	return false
}

func (p *partitionedDevice) EnvVars() map[string]string {
	return p.manifest.EnvVars()
}

func (p *partitionedDevice) Annotations() map[string]string {
	return p.manifest.Annotations()
}

func (p *partitionedDevice) DeviceSpecs() []*devicePluginAPIv1Beta1.DeviceSpec {
	deviceSpecs := make([]*devicePluginAPIv1Beta1.DeviceSpec, 0)
	for _, deviceNode := range p.manifest.DeviceNodes() {
		deviceSpecs = append(deviceSpecs, &devicePluginAPIv1Beta1.DeviceSpec{
			ContainerPath: deviceNode.ContainerPath,
			HostPath:      deviceNode.HostPath,
			Permissions:   deviceNode.Permissions,
		})
	}

	return deviceSpecs
}

func (p *partitionedDevice) Mounts() []*devicePluginAPIv1Beta1.Mount {
	mounts := make([]*devicePluginAPIv1Beta1.Mount, 0)
	for _, mount := range p.manifest.MountPaths() {
		readOnly := false
		for _, opt := range mount.Options {
			if opt == readOnlyOpt {
				readOnly = true
				break
			}
		}

		mounts = append(mounts, &devicePluginAPIv1Beta1.Mount{
			ContainerPath: mount.ContainerPath,
			HostPath:      mount.HostPath,
			ReadOnly:      readOnly,
		})
	}

	return mounts
}

func (p *partitionedDevice) CDIDevices() []*devicePluginAPIv1Beta1.CDIDevice {
	return nil
}

func (p *partitionedDevice) GetID() string {
	// TODO: implement it when bin-packing allocator is ready.
	return ""
}

func (p *partitionedDevice) GetTopologyHintKey() npu_allocator.TopologyHintKey {
	// TODO: implement it when bin-packing allocator is ready.
	return ""
}

func (p *partitionedDevice) Equal(target npu_allocator.Device) bool {
	// TODO: implement it when bin-packing allocator is ready.
	return false
}
