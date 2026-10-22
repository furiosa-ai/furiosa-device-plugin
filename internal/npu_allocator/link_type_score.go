package npu_allocator

import "github.com/furiosa-ai/furiosa-smi-go/pkg/api"

// linkTypeScore maps api.LinkType to the additive topology score used by the allocator.
// The weights keep the ranking of the former libfuriosa_smi link type values.
func linkTypeScore(linkType api.LinkType) uint {
	switch linkType {
	case api.LinkTypeNoC:
		return 70
	case api.LinkTypeBridge:
		return 30
	case api.LinkTypeCPU:
		return 20
	case api.LinkTypeInterconnect:
		return 10
	default:
		return 0
	}
}
