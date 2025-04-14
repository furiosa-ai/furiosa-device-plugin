package device_manager

import (
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/furiosa_device"
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/stretchr/testify/assert"
)

func TestBuildFullEndpoint(t *testing.T) {
	tests := []struct {
		description    string
		arch           smi.Arch
		policy         furiosa_device.PartitioningPolicy
		expectedResult string
		expectError    bool
	}{
		{
			description:    "test warboy non partitioning policy",
			arch:           smi.ArchWarboy,
			policy:         furiosa_device.NonePolicy,
			expectedResult: "warboy",
			expectError:    false,
		},
		{
			description:    "test rngd non partitioning policy",
			arch:           smi.ArchRngd,
			policy:         furiosa_device.NonePolicy,
			expectedResult: "rngd",
			expectError:    false,
		},
		{
			description:    "test rngd 2core.12gb partitioning policy",
			arch:           smi.ArchRngd,
			policy:         furiosa_device.DualCorePolicy,
			expectedResult: "rngd-2core.12gb",
			expectError:    false,
		},
		{
			description:    "test rngd 4core.24gb partitioning policy",
			arch:           smi.ArchRngd,
			policy:         furiosa_device.QuadCorePolicy,
			expectedResult: "rngd-4core.24gb",
			expectError:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			actualResult, actualErr := buildFullEndpoint(tc.arch, tc.policy)
			if tc.expectError {
				assert.Error(t, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tc.expectedResult, actualResult)
		})
	}
}

func TestBuildAndValidateFullResourceEndpointName(t *testing.T) {
	tests := []struct {
		description    string
		arch           smi.Arch
		policy         furiosa_device.PartitioningPolicy
		expectedResult string
		expectError    bool
	}{
		{
			description:    "test warboy non partitioning policy",
			arch:           smi.ArchWarboy,
			policy:         furiosa_device.NonePolicy,
			expectedResult: "furiosa.ai/warboy",
			expectError:    false,
		},
		{
			description:    "test rngd non partitioning policy",
			arch:           smi.ArchRngd,
			policy:         furiosa_device.NonePolicy,
			expectedResult: "furiosa.ai/rngd",
			expectError:    false,
		},
		{
			description:    "test rngd 2core.12gb partitioning policy",
			arch:           smi.ArchRngd,
			policy:         furiosa_device.DualCorePolicy,
			expectedResult: "furiosa.ai/rngd-2core.12gb",
			expectError:    false,
		},
		{
			description:    "test rngd 4core.24gb partitioning policy",
			arch:           smi.ArchRngd,
			policy:         furiosa_device.QuadCorePolicy,
			expectedResult: "furiosa.ai/rngd-4core.24gb",
			expectError:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			actualResult, actualErr := buildAndValidateFullResourceEndpointName(tc.arch, tc.policy)
			if tc.expectError {
				assert.Error(t, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tc.expectedResult, actualResult)
		})
	}
}
