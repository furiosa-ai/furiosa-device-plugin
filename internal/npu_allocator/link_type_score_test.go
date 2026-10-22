package npu_allocator

import (
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestLinkTypeScore(t *testing.T) {
	tests := []struct {
		linkType api.LinkType
		expected uint
	}{
		{api.LinkTypeNoC, 70},
		{api.LinkTypeBridge, 30},
		{api.LinkTypeCPU, 20},
		{api.LinkTypeInterconnect, 10},
		{api.LinkTypeUnknown, 0},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, linkTypeScore(tc.linkType), "link type %v", tc.linkType)
	}
}
