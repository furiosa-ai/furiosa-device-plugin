package device_manager

import (
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/stretchr/testify/assert"

	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
)

func TestBuildFullEndpoint(t *testing.T) {
	tests := []struct {
		description    string
		arch           smi.Arch
		strategy       config.ResourceUnitStrategy
		expectedResult string
		expectError    bool
	}{
		{
			description:    "test warboy generic strategy",
			arch:           smi.ArchWarboy,
			strategy:       config.GenericStrategy,
			expectedResult: "warboy",
			expectError:    false,
		},
		{
			description:    "test warboy single core strategy",
			arch:           smi.ArchWarboy,
			strategy:       config.SingleCoreStrategy,
			expectedResult: "warboy-1core.8gb",
			expectError:    false,
		},
		{
			description:    "test warboy dual core strategy",
			arch:           smi.ArchWarboy,
			strategy:       config.DualCoreStrategy,
			expectedResult: "warboy-2core.16gb",
			expectError:    false,
		},
		{
			description:    "test warboy quad core strategy",
			arch:           smi.ArchWarboy,
			strategy:       config.QuadCoreStrategy,
			expectedResult: "",
			expectError:    true,
		},
		{
			description:    "test rngd generic strategy",
			arch:           smi.ArchRngd,
			strategy:       config.GenericStrategy,
			expectedResult: "rngd",
			expectError:    false,
		},
		{
			description:    "test rngd single core strategy",
			arch:           smi.ArchRngd,
			strategy:       config.SingleCoreStrategy,
			expectedResult: "rngd-1core.6gb",
			expectError:    false,
		},
		{
			description:    "test rngd dual core strategy",
			arch:           smi.ArchRngd,
			strategy:       config.DualCoreStrategy,
			expectedResult: "rngd-2core.12gb",
			expectError:    false,
		},
		{
			description:    "test rngd quad core strategy",
			arch:           smi.ArchRngd,
			strategy:       config.QuadCoreStrategy,
			expectedResult: "rngd-4core.24gb",
			expectError:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			actualResult, actualErr := buildFullEndpoint(tc.arch, tc.strategy)
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
		strategy       config.ResourceUnitStrategy
		expectedResult string
		expectError    bool
	}{
		{
			description:    "test warboy generic strategy",
			arch:           smi.ArchWarboy,
			strategy:       config.GenericStrategy,
			expectedResult: "furiosa.ai/warboy",
			expectError:    false,
		},
		{
			description:    "test warboy single core strategy",
			arch:           smi.ArchWarboy,
			strategy:       config.SingleCoreStrategy,
			expectedResult: "furiosa.ai/warboy-1core.8gb",
			expectError:    false,
		},
		{
			description:    "test warboy dual core strategy",
			arch:           smi.ArchWarboy,
			strategy:       config.DualCoreStrategy,
			expectedResult: "furiosa.ai/warboy-2core.16gb",
			expectError:    false,
		},
		{
			description:    "test warboy quad core strategy",
			arch:           smi.ArchWarboy,
			strategy:       config.QuadCoreStrategy,
			expectedResult: "",
			expectError:    true,
		},
		{
			description:    "test rngd generic strategy",
			arch:           smi.ArchRngd,
			strategy:       config.GenericStrategy,
			expectedResult: "furiosa.ai/rngd",
			expectError:    false,
		},
		{
			description:    "test rngd single core strategy",
			arch:           smi.ArchRngd,
			strategy:       config.SingleCoreStrategy,
			expectedResult: "furiosa.ai/rngd-1core.6gb",
			expectError:    false,
		},
		{
			description:    "test rngd dual core strategy",
			arch:           smi.ArchRngd,
			strategy:       config.DualCoreStrategy,
			expectedResult: "furiosa.ai/rngd-2core.12gb",
			expectError:    false,
		},
		{
			description:    "test rngd quad core strategy",
			arch:           smi.ArchRngd,
			strategy:       config.QuadCoreStrategy,
			expectedResult: "furiosa.ai/rngd-4core.24gb",
			expectError:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			actualResult, actualErr := buildAndValidateFullResourceEndpointName(tc.arch, tc.strategy)
			if tc.expectError {
				assert.Error(t, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tc.expectedResult, actualResult)
		})
	}
}
