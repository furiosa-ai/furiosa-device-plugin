package device_manager

import (
	"testing"

	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/furiosa_device"
	"github.com/stretchr/testify/assert"
)

func TestTransformPartitioningConfig(t *testing.T) {
	tests := []struct {
		description string
		policy      string
		expected    furiosa_device.PartitioningPolicy
		expectError bool
	}{
		{
			description: "none policy",
			policy:      config.NonePolicyStr,
			expected:    furiosa_device.NonePolicy,
			expectError: false,
		},
		{
			description: "2core.12gb",
			policy:      config.Rngd2Core12GbStr,
			expected:    furiosa_device.DualCorePolicy,
			expectError: false,
		},
		{
			description: "4core.24gb",
			policy:      config.Rngd4Core24GbStr,
			expected:    furiosa_device.QuadCorePolicy,
			expectError: false,
		},
		{
			description: "invalid",
			policy:      "invalid",
			expected:    furiosa_device.NonePolicy,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			result, err := transformPartitioningConfig(tc.policy)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestValidatePartitioningConfig(t *testing.T) {
	tests := []struct {
		description string
		deviceMap   DeviceMap
		policy      furiosa_device.PartitioningPolicy
		expectError bool
	}{
		{
			description: "non policy with warboy",
			deviceMap:   DeviceMap{smi.ArchWarboy: nil},
			policy:      furiosa_device.NonePolicy,
			expectError: false,
		},
		{
			description: "dual core policy with warboy",
			deviceMap:   DeviceMap{smi.ArchWarboy: nil},
			policy:      furiosa_device.DualCorePolicy,
			expectError: true,
		},
		{
			description: "quad core policy with warboy",
			deviceMap:   DeviceMap{smi.ArchWarboy: nil},
			policy:      furiosa_device.QuadCorePolicy,
			expectError: true,
		},
		{
			description: "non policy with rngd",
			deviceMap:   DeviceMap{smi.ArchRngd: nil},
			policy:      furiosa_device.NonePolicy,
			expectError: false,
		},
		{
			description: "dual core policy with rngd",
			deviceMap:   DeviceMap{smi.ArchRngd: nil},
			policy:      furiosa_device.DualCorePolicy,
			expectError: false,
		},
		{
			description: "quad core policy with rngd",
			deviceMap:   DeviceMap{smi.ArchRngd: nil},
			policy:      furiosa_device.QuadCorePolicy,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			err := validatePartitioningConfig(tc.deviceMap, tc.policy)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
