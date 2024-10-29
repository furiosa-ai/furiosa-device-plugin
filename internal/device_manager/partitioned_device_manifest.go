package device_manager

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/manifest"
)

type partitionedDeviceManifest struct {
	arch        smi.Arch
	original    manifest.Manifest
	partition   Partition
	deviceNodes []*manifest.DeviceNode
	mounts      []*manifest.Mount
}

func NewPartitionedDeviceManifest(arch smi.Arch, original manifest.Manifest, partition Partition) (manifest.Manifest, error) {
	switch arch {
	case smi.ArchWarboy:
		return NewPartitionedDeviceManifestWarboy(original, partition)

	case smi.ArchRngd:
		return NewPartitionedDeviceManifestRngd(original, partition)

	default:
		return nil, fmt.Errorf("unsupported architecture: %s", arch.ToString())
	}
}

func (p *partitionedDeviceManifest) EnvVars() map[string]string {
	return p.original.EnvVars()
}

func (p *partitionedDeviceManifest) Annotations() map[string]string {
	return p.original.Annotations()
}

func (p *partitionedDeviceManifest) DeviceNodes() []*manifest.DeviceNode {
	return p.deviceNodes
}

func (p *partitionedDeviceManifest) MountPaths() []*manifest.Mount {
	return p.mounts
}

var (
	warboyDeviceNodeWholeDeviceRegex = regexp.MustCompile(`^/dev/npu[0-9]+$`)
	warboyDeviceNodePeRegex          = regexp.MustCompile(`^/dev/npu[0-9]+pe\S+$`)

	warboyMountWholeDeviceRegex = regexp.MustCompile(`^(/sys/class|/sys/devices/virtual)/npu_mgmt/npu[0-9]+$`)
	warboyMountPeRegex          = regexp.MustCompile(`^(/sys/class|/sys/devices/virtual)/npu_mgmt/npu[0-9]+pe\S+$`)
)

func NewPartitionedDeviceManifestWarboy(original manifest.Manifest, partition Partition) (manifest.Manifest, error) {
	deviceNodes, err := generateWarboyPartitionedDeviceNodes(original, partition)
	if err != nil {
		return nil, err
	}

	mounts, err := generateWarboyPartitionedMounts(original, partition)
	if err != nil {
		return nil, err
	}

	return &partitionedDeviceManifest{
		arch:        smi.ArchWarboy,
		original:    original,
		partition:   partition,
		deviceNodes: deviceNodes,
		mounts:      mounts,
	}, nil
}

// generateWarboyPartitionedDeviceNodes generates (actually filters) Device Nodes by following rules.
//   - /dev/npu{N} will be dropped
//   - /dev/npu{N}pe{partition} will be dropped if {partition} does not match with given partition value
func generateWarboyPartitionedDeviceNodes(original manifest.Manifest, partition Partition) ([]*manifest.DeviceNode, error) {
	originalDeviceNodes := original.DeviceNodes()
	survivedDeviceNodes := make([]*manifest.DeviceNode, 0, len(originalDeviceNodes))
	for _, deviceNode := range originalDeviceNodes {
		path := deviceNode.ContainerPath
		if warboyDeviceNodePeRegex.MatchString(path) {
			// /dev/npu{N}pe{partition} will be dropped if {partition} does not match with given partition value
			elements := strings.Split(path, "/")
			target := elements[len(elements)-1]

			var devNum int
			var partitionPostfix string
			_, err := fmt.Sscanf(target, "npu%dpe%s", &devNum, &partitionPostfix)
			if err != nil {
				return nil, err
			}

			if partitionPostfix == partition.String() {
				survivedDeviceNodes = append(survivedDeviceNodes, deviceNode)
			}
		} else {
			if warboyDeviceNodeWholeDeviceRegex.MatchString(path) {
				// /dev/npu{N} will be dropped
				continue
			}

			survivedDeviceNodes = append(survivedDeviceNodes, deviceNode)
		}
	}

	return survivedDeviceNodes, nil
}

// generateWarboyPartitionedMounts generates (actually filters) Mounts by following rules.
//   - npu{N} will be dropped
//   - npu{N}pe{partition} will be dropped if {partition} does not match with given partition value
func generateWarboyPartitionedMounts(original manifest.Manifest, partition Partition) ([]*manifest.Mount, error) {
	originalMounts := original.MountPaths()
	survivedMounts := make([]*manifest.Mount, 0, len(originalMounts))
	for _, mount := range originalMounts {
		path := mount.ContainerPath
		if warboyMountPeRegex.MatchString(path) {
			// npu{N}pe{partition} will be dropped if {partition} does not match with given partition value
			elements := strings.Split(path, "/")
			target := elements[len(elements)-1]

			var devNum int
			var partitionPostfix string
			_, err := fmt.Sscanf(target, "npu%dpe%s", &devNum, &partitionPostfix)
			if err != nil {
				return nil, err
			}

			if partitionPostfix == partition.String() {
				survivedMounts = append(survivedMounts, mount)
			}
		} else {
			if warboyMountWholeDeviceRegex.MatchString(path) {
				// npu{N} will be dropped
				continue
			}

			survivedMounts = append(survivedMounts, mount)
		}
	}

	return survivedMounts, nil
}

var (
	rngdDeviceNodePeRegex, _ = regexp.Compile(`^/dev/rngd/npu[0-9]+pe\S+$`)
)

func NewPartitionedDeviceManifestRngd(original manifest.Manifest, partition Partition) (manifest.Manifest, error) {
	deviceNodes, err := filterRngdPartitionedDeviceNodes(original, partition)
	if err != nil {
		return nil, err
	}

	mounts, err := filterRngdPartitionedMounts(original, partition)
	if err != nil {
		return nil, err
	}

	return &partitionedDeviceManifest{
		arch:        smi.ArchRngd,
		original:    original,
		partition:   partition,
		deviceNodes: deviceNodes,
		mounts:      mounts,
	}, nil
}

// filterRngdPartitionedDeviceNodes filters (actually filters) Device Nodes by following rules.
//   - /dev/rngd/npu{N}pe{partition} will be dropped if {partition} does not match with given partition value
func filterRngdPartitionedDeviceNodes(original manifest.Manifest, partition Partition) ([]*manifest.DeviceNode, error) {
	originalDeviceNodes := original.DeviceNodes()
	survivedDeviceNodes := make([]*manifest.DeviceNode, 0, len(originalDeviceNodes))
	for _, deviceNode := range originalDeviceNodes {
		path := deviceNode.ContainerPath
		if rngdDeviceNodePeRegex.MatchString(path) {
			// /dev/rngd/npu{N}pe{partition} will be dropped if {partition} does not match with given partition value
			elements := strings.Split(path, "/")
			target := elements[len(elements)-1]

			var devNum int
			var partitionPostfix string
			_, err := fmt.Sscanf(target, "npu%dpe%s", &devNum, &partitionPostfix)
			if err != nil {
				return nil, err
			}

			if partitionPostfix == partition.String() {
				survivedDeviceNodes = append(survivedDeviceNodes, deviceNode)
			}
		} else {
			survivedDeviceNodes = append(survivedDeviceNodes, deviceNode)
		}
	}

	return survivedDeviceNodes, nil
}

// filterRngdPartitionedMounts filters Mounts
func filterRngdPartitionedMounts(original manifest.Manifest, _ Partition) ([]*manifest.Mount, error) {
	// right now, we don't need to filter any mount paths.
	return original.MountPaths(), nil
}
