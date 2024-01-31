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
			description:    "test renegade legacy strategy",
			arch:           device.ArchRenegade,
			strategy:       config.LegacyStrategy,
			expectedResult: "npu",
			expectError:    false,
		},
		{
			description:    "test renegade generic strategy",
			arch:           device.ArchRenegade,
			strategy:       config.GenericStrategy,
			expectedResult: "renegade",
			expectError:    false,
		},
		{
			description:    "test renegade single core strategy",
			arch:           device.ArchRenegade,
			strategy:       config.SingleCoreStrategy,
			expectedResult: "renegade-1core.6gb",
			expectError:    false,
		},
		{
			description:    "test renegade dual core strategy",
			arch:           device.ArchRenegade,
			strategy:       config.DualCoreStrategy,
			expectedResult: "renegade-2core.12gb",
			expectError:    false,
		},
		{
			description:    "test renegade quad core strategy",
			arch:           device.ArchRenegade,
			strategy:       config.QuadCoreStrategy,
			expectedResult: "renegade-4core.24gb",
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
			description:    "test renegade legacy strategy",
			arch:           device.ArchRenegade,
			strategy:       config.LegacyStrategy,
			expectedResult: "alpha.furiosa.ai/npu",
			expectError:    false,
		},
		{
			description:    "test renegade generic strategy",
			arch:           device.ArchRenegade,
			strategy:       config.GenericStrategy,
			expectedResult: "furiosa.ai/renegade",
			expectError:    false,
		},
		{
			description:    "test renegade single core strategy",
			arch:           device.ArchRenegade,
			strategy:       config.SingleCoreStrategy,
			expectedResult: "furiosa.ai/renegade-1core.6gb",
			expectError:    false,
		},
		{
			description:    "test renegade dual core strategy",
			arch:           device.ArchRenegade,
			strategy:       config.DualCoreStrategy,
			expectedResult: "furiosa.ai/renegade-2core.12gb",
			expectError:    false,
		},
		{
			description:    "test renegade quad core strategy",
			arch:           device.ArchRenegade,
			strategy:       config.QuadCoreStrategy,
			expectedResult: "furiosa.ai/renegade-4core.24gb",
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
