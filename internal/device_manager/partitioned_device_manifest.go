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
	warboyDeviceNodePeRegex = regexp.MustCompile(`^/dev/npu[0-9]+pe\S+$`)
)

func NewPartitionedDeviceManifestWarboy(original manifest.Manifest, partition Partition) (manifest.Manifest, error) {
	deviceNodes, err := generateWarboyPartitionedDeviceNodes(original, partition)
	if err != nil {
		return nil, err
	}

	return &partitionedDeviceManifest{
		arch:        smi.ArchWarboy,
		original:    original,
		partition:   partition,
		deviceNodes: deviceNodes,

		// do not filter any mount paths right now
		// see: https://github.com/furiosa-ai/furiosa-device-plugin/pull/30#discussion_r1819763238
		mounts: original.MountPaths(),
	}, nil
}

// generateWarboyPartitionedDeviceNodes generates (actually filters) Device Nodes by following rules.
//   - /dev/npu{N} will be dropped
//   - /dev/npu{N}pe{partition} will be dropped if {partition} does not match with given partition value
func generateWarboyPartitionedDeviceNodes(original manifest.Manifest, partition Partition) ([]*manifest.DeviceNode, error) {
	var survivedDeviceNodes []*manifest.DeviceNode
	for _, deviceNode := range original.DeviceNodes() {
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
			survivedDeviceNodes = append(survivedDeviceNodes, deviceNode)
		}
	}

	return survivedDeviceNodes, nil
}

var (
	rngdDeviceNodePeRegex = regexp.MustCompile(`^/dev/rngd/npu[0-9]+pe\S+$`)
)

func NewPartitionedDeviceManifestRngd(original manifest.Manifest, partition Partition) (manifest.Manifest, error) {
	deviceNodes, err := filterRngdPartitionedDeviceNodes(original, partition)
	if err != nil {
		return nil, err
	}

	return &partitionedDeviceManifest{
		arch:        smi.ArchRngd,
		original:    original,
		partition:   partition,
		deviceNodes: deviceNodes,
		mounts:      original.MountPaths(), // right now, we don't need to filter any mount paths.
	}, nil
}

// filterRngdPartitionedDeviceNodes filters (actually filters) Device Nodes by following rules.
//   - /dev/rngd/npu{N}pe{partition} will be dropped if {partition} does not match with given partition value
func filterRngdPartitionedDeviceNodes(original manifest.Manifest, partition Partition) ([]*manifest.DeviceNode, error) {
	var survivedDeviceNodes []*manifest.DeviceNode
	for _, deviceNode := range original.DeviceNodes() {
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
