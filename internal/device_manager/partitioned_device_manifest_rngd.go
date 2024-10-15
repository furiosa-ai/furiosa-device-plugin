package device_manager

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/manifest"
)

var (
	rngdDeviceNodePeRegex, _ = regexp.Compile(`^/dev/rngd/npu[0-9]+pe\S+$`)
)

func NewPartitionedDeviceManifestRngd(original manifest.Manifest, partition Partition) (manifest.Manifest, error) {
	deviceNodes, err := generateRngdPartitionedDeviceNodes(original, partition)
	if err != nil {
		return nil, err
	}

	mounts, err := generateRngdPartitionedMounts(original, partition)
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

// generateRngdPartitionedDeviceNodes generates (actually filters) Device Nodes by following rules.
//   - /dev/rngd/npu{N}pe{partition} will be dropped if {partition} does not match with given partition value
func generateRngdPartitionedDeviceNodes(original manifest.Manifest, partition Partition) ([]*manifest.DeviceNode, error) {
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

			if Partition(partitionPostfix) == partition {
				survivedDeviceNodes = append(survivedDeviceNodes, deviceNode)
			}
		} else {
			survivedDeviceNodes = append(survivedDeviceNodes, deviceNode)
		}
	}

	return survivedDeviceNodes, nil
}

// generateWarboyPartitionedMounts generates Mounts
func generateRngdPartitionedMounts(original manifest.Manifest, _ Partition) ([]*manifest.Mount, error) {
	// right now, we don't need to filter any mount paths.
	return original.MountPaths(), nil
}
