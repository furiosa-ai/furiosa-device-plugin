package device_manager

import (
	"testing"

	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
)

func TestBuildDomainName(t *testing.T) {
	tests := []struct {
		description string
		strategy    config.ResourceUnitStrategy
		expected    string
	}{
		{
			description: "test legacy strategy",
			strategy:    config.LegacyStrategy,
			expected:    "alpha.furiosa.ai",
		},
		{
			description: "test generic strategy",
			strategy:    config.GenericStrategy,
			expected:    "furiosa.ai",
		},
		{
			description: "test single core strategy",
			strategy:    config.SingleCoreStrategy,
			expected:    "furiosa.ai",
		},
		{
			description: "test dual core strategy",
			strategy:    config.DualCoreStrategy,
			expected:    "furiosa.ai",
		},
		{
			description: "test quad core strategy",
			strategy:    config.QuadCoreStrategy,
			expected:    "furiosa.ai",
		},
	}

	for _, tc := range tests {
		actual := buildDomainName(tc.strategy)
		if actual != tc.expected {
			t.Errorf("expectedResult %s but got %s", tc.expected, actual)
		}
	}
}

func TestBuildFullEndpoint(t *testing.T) {
	tests := []struct {
		description    string
		arch           device.Arch
		strategy       config.ResourceUnitStrategy
		expectedResult string
		expectError    bool
	}{
		{
			description:    "test warboy legacy strategy",
			arch:           device.ArchWarboy,
			strategy:       config.LegacyStrategy,
			expectedResult: "npu",
			expectError:    false,
		},
		{
			description:    "test warboy generic strategy",
			arch:           device.ArchWarboy,
			strategy:       config.GenericStrategy,
			expectedResult: "warboy",
			expectError:    false,
		},
		{
			description:    "test warboy single core strategy",
			arch:           device.ArchWarboy,
			strategy:       config.SingleCoreStrategy,
			expectedResult: "warboy-1core.8gb",
			expectError:    false,
		},
		{
			description:    "test warboy dual core strategy",
			arch:           device.ArchWarboy,
			strategy:       config.DualCoreStrategy,
			expectedResult: "warboy-2core.16gb",
			expectError:    false,
		},
		{
			description:    "test warboy quad core strategy",
			arch:           device.ArchWarboy,
			strategy:       config.QuadCoreStrategy,
			expectedResult: "",
			expectError:    true,
		},
		{
			description:    "test rngd legacy strategy",
			arch:           device.ArchRngd,
			strategy:       config.LegacyStrategy,
			expectedResult: "npu",
			expectError:    false,
		},
		{
			description:    "test rngd generic strategy",
			arch:           device.ArchRngd,
			strategy:       config.GenericStrategy,
			expectedResult: "rngd",
			expectError:    false,
		},
		{
			description:    "test rngd single core strategy",
			arch:           device.ArchRngd,
			strategy:       config.SingleCoreStrategy,
			expectedResult: "rngd-1core.6gb",
			expectError:    false,
		},
		{
			description:    "test rngd dual core strategy",
			arch:           device.ArchRngd,
			strategy:       config.DualCoreStrategy,
			expectedResult: "rngd-2core.12gb",
			expectError:    false,
		},
		{
			description:    "test rngd quad core strategy",
			arch:           device.ArchRngd,
			strategy:       config.QuadCoreStrategy,
			expectedResult: "rngd-4core.24gb",
			expectError:    false,
		},
	}

	for _, tc := range tests {
		actualResult, actualErr := buildFullEndpoint(tc.arch, tc.strategy)
		if actualErr != nil != tc.expectError {
			t.Errorf("unexpected error %t", actualErr)
			continue
		}

		if actualResult != tc.expectedResult {
			t.Errorf("expectedResult %s but got %s", tc.expectedResult, actualResult)
		}
	}
}

func TestBuildAndValidateFullResourceEndpointName(t *testing.T) {
	tests := []struct {
		description    string
		arch           device.Arch
		strategy       config.ResourceUnitStrategy
		expectedResult string
		expectError    bool
	}{
		{
			description:    "test warboy legacy strategy",
			arch:           device.ArchWarboy,
			strategy:       config.LegacyStrategy,
			expectedResult: "alpha.furiosa.ai/npu",
			expectError:    false,
		},
		{
			description:    "test warboy generic strategy",
			arch:           device.ArchWarboy,
			strategy:       config.GenericStrategy,
			expectedResult: "furiosa.ai/warboy",
			expectError:    false,
		},
		{
			description:    "test warboy single core strategy",
			arch:           device.ArchWarboy,
			strategy:       config.SingleCoreStrategy,
			expectedResult: "furiosa.ai/warboy-1core.8gb",
			expectError:    false,
		},
		{
			description:    "test warboy dual core strategy",
			arch:           device.ArchWarboy,
			strategy:       config.DualCoreStrategy,
			expectedResult: "furiosa.ai/warboy-2core.16gb",
			expectError:    false,
		},
		{
			description:    "test warboy quad core strategy",
			arch:           device.ArchWarboy,
			strategy:       config.QuadCoreStrategy,
			expectedResult: "",
			expectError:    true,
		},
		{
			description:    "test rngd legacy strategy",
			arch:           device.ArchRngd,
			strategy:       config.LegacyStrategy,
			expectedResult: "alpha.furiosa.ai/npu",
			expectError:    false,
		},
		{
			description:    "test rngd generic strategy",
			arch:           device.ArchRngd,
			strategy:       config.GenericStrategy,
			expectedResult: "furiosa.ai/rngd",
			expectError:    false,
		},
		{
			description:    "test rngd single core strategy",
			arch:           device.ArchRngd,
			strategy:       config.SingleCoreStrategy,
			expectedResult: "furiosa.ai/rngd-1core.6gb",
			expectError:    false,
		},
		{
			description:    "test rngd dual core strategy",
			arch:           device.ArchRngd,
			strategy:       config.DualCoreStrategy,
			expectedResult: "furiosa.ai/rngd-2core.12gb",
			expectError:    false,
		},
		{
			description:    "test rngd quad core strategy",
			arch:           device.ArchRngd,
			strategy:       config.QuadCoreStrategy,
			expectedResult: "furiosa.ai/rngd-4core.24gb",
			expectError:    false,
		},
	}

	for _, tc := range tests {
		actualResult, actualErr := buildAndValidateFullResourceEndpointName(tc.arch, tc.strategy)
		if actualErr != nil != tc.expectError {
			t.Errorf("unexpected error %t", actualErr)
			continue
		}

		if actualResult != tc.expectedResult {
			t.Errorf("expectedResult %s but got %s", tc.expectedResult, actualResult)
		}
	}
}
