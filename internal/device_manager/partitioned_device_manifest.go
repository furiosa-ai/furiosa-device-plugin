package device_manager

import (
	"fmt"
	"regexp"

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
	deviceNodes, err := filterPartitionedDeviceNodes(original, partition)
	if err != nil {
		return nil, err
	}

	return &partitionedDeviceManifest{
		arch:        arch,
		original:    original,
		partition:   partition,
		deviceNodes: deviceNodes,

		// do not filter any mount paths right now as right now, we don't need to filter any mount paths.
		// see: https://github.com/furiosa-ai/furiosa-device-plugin/pull/30#discussion_r1819763238
		mounts: original.MountPaths(),
	}, nil
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

const (
	deviceIdExp  = "device_id"
	startCoreExp = "start_core"
	endCoreExp   = "end_core"

	regexpPattern = `^\S+npu(?P<` + deviceIdExp + `>\d+)((?:pe)(?P<` + startCoreExp + `>\d+)(-(?P<` + endCoreExp + `>\d+))?)?$`
)

var (
	deviceNodePeRegex = regexp.MustCompile(regexpPattern)
	deviceNodeSubExps = deviceNodePeRegex.SubexpNames()
)

// filterPartitionedDeviceNodes filters (actually filters) Device Nodes by following rules.
//   - npu{N}pe{partition} will be dropped if {partition} does not match with given partition value
func filterPartitionedDeviceNodes(original manifest.Manifest, partition Partition) ([]*manifest.DeviceNode, error) {
	var survivedDeviceNodes []*manifest.DeviceNode
	for _, deviceNode := range original.DeviceNodes() {
		path := deviceNode.ContainerPath
		matches := deviceNodePeRegex.FindStringSubmatch(path)
		namedMatches := map[string]string{}
		for i, match := range matches {
			subExp := deviceNodeSubExps[i]
			if subExp == "" {
				continue
			}

			namedMatches[subExp] = match
		}

		if len(namedMatches) > 0 {
			deviceId := namedMatches[deviceIdExp]
			if deviceId == "" {
				continue
			}

			startCore := namedMatches[startCoreExp]
			endCore := namedMatches[endCoreExp]

			var partitionPostfix string
			if endCore == "" {
				partitionPostfix = startCore
			} else {
				partitionPostfix = fmt.Sprintf("%s-%s", startCore, endCore)
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
