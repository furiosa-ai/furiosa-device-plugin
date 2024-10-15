package device_manager

import (
	"fmt"

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
