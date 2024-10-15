package device_manager

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/manifest"
)

var (
	warboyDeviceNodeWholeDeviceRegex, _ = regexp.Compile(`^/dev/npu[0-9]+$`)
	warboyDeviceNodePeRegex, _          = regexp.Compile(`^/dev/npu[0-9]+pe\S+$`)

	warboyMountWholeDeviceRegex, _ = regexp.Compile(`^(/sys/class|/sys/devices/virtual)/npu_mgmt/npu[0-9]+$`)
	warboyMountPeRegex, _          = regexp.Compile(`^(/sys/class|/sys/devices/virtual)/npu_mgmt/npu[0-9]+pe\S+$`)
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

			if Partition(partitionPostfix) == partition {
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

			if Partition(partitionPostfix) == partition {
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
