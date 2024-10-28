package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigFromFile(t *testing.T) {
	tests := []struct {
		description    string
		configPath     string
		expectedResult *Config
		expectedError  bool
	}{
		{
			description: "parse legacy configuration",
			configPath:  "./tests/legacy_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategy: LegacyStrategy,
				AllocationMode:   ScoreBased,
				DebugMode:        true,
			},
			expectedError: false,
		},
		{
			description: "parse generic configuration",
			configPath:  "./tests/generic_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategy: GenericStrategy,
				AllocationMode:   ScoreBased,
				DebugMode:        true,
			},
			expectedError: false,
		},
		{
			description: "parse single-core configuration",
			configPath:  "./tests/single_core_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategy: SingleCoreStrategy,
				AllocationMode:   ScoreBased,
				DebugMode:        true,
			},
			expectedError: false,
		},
		{
			description: "parse dual-core configuration",
			configPath:  "./tests/dual_core_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategy: DualCoreStrategy,
				AllocationMode:   ScoreBased,
				DebugMode:        true,
			},
			expectedError: false,
		},
		{
			description: "parse quad-core configuration",
			configPath:  "./tests/quad_core_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategy: QuadCoreStrategy,
				AllocationMode:   ScoreBased,
				DebugMode:        true,
			},
			expectedError: false,
		},
		{
			description: "parse uuid list",
			configPath:  "./tests/with_disabled_devices.yaml",
			expectedResult: &Config{
				ResourceStrategy: GenericStrategy,
				AllocationMode:   ScoreBased,
				DisabledDeviceUUIDListMap: map[string][]string{
					"node_a": {"uuid1", "uuid2"},
					"node_b": {"uuid1", "uuid2"},
				},
			},
			expectedError: false,
		},
		{
			description:    "try empty config",
			configPath:     "./tests/empty.yaml",
			expectedResult: nil,
			expectedError:  true,
		},
		{
			description:    "try wrong format1",
			configPath:     "./tests/wrong_resource_strategy.yaml",
			expectedResult: nil,
			expectedError:  true,
		},
		{
			description:    "try wrong format2",
			configPath:     "./tests/wrong_disabled_devices.yaml",
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tc := range tests {
		actualConf, actualErr := getConfigFromFile(absPath(tc.configPath))
		if tc.expectedError {
			assert.NotNilf(t, actualErr, tc.description)
		} else {
			assert.Nilf(t, actualErr, tc.description)
			assert.Equalf(t, tc.expectedResult, actualConf, tc.description)
		}
	}
}

func TestIsEqualConfig(t *testing.T) {
	tests := []struct {
		description string
		a           *Config
		b           *Config
		expected    bool
	}{
		{
			description: "equal configs",
			a: &Config{
				ResourceStrategy:          GenericStrategy,
				DisabledDeviceUUIDListMap: map[string][]string{"node_a": {"a0", "a1"}},
				DebugMode:                 true,
			},
			b: &Config{
				ResourceStrategy:          GenericStrategy,
				DisabledDeviceUUIDListMap: map[string][]string{"node_a": {"a0", "a1"}},
				DebugMode:                 true,
			},
			expected: true,
		},
		{
			description: "different resource strategy map",
			a: &Config{
				ResourceStrategy:          GenericStrategy,
				DisabledDeviceUUIDListMap: map[string][]string{"node_a": {"a0", "a1"}},
				DebugMode:                 true,
			},
			b: &Config{
				ResourceStrategy:          SingleCoreStrategy,
				DisabledDeviceUUIDListMap: map[string][]string{"node_a": {"a0", "a1"}},
				DebugMode:                 true,
			},
			expected: false,
		},
		{
			description: "different disabled device uuid list map1",
			a: &Config{
				ResourceStrategy:          GenericStrategy,
				DisabledDeviceUUIDListMap: map[string][]string{"node_a": {"a0", "a1"}},
				DebugMode:                 true,
			},
			b: &Config{
				ResourceStrategy:          GenericStrategy,
				DisabledDeviceUUIDListMap: map[string][]string{"node_b": {"a0", "a1"}},
				DebugMode:                 true,
			},
			expected: false,
		},
		{
			description: "different disabled device uuid list map2",
			a: &Config{
				ResourceStrategy:          GenericStrategy,
				DisabledDeviceUUIDListMap: map[string][]string{"node_a": {"a0", "a1"}},
				DebugMode:                 true,
			},
			b: &Config{
				ResourceStrategy:          GenericStrategy,
				DisabledDeviceUUIDListMap: map[string][]string{"node_a": {"a1", "a2"}},
				DebugMode:                 true,
			},
			expected: false,
		},
		{
			description: "different disabled device uuid list map3",
			a: &Config{
				ResourceStrategy:          GenericStrategy,
				DisabledDeviceUUIDListMap: map[string][]string{"node_a": {"a0", "a1"}},
				DebugMode:                 true,
			},
			b: &Config{
				ResourceStrategy:          GenericStrategy,
				DisabledDeviceUUIDListMap: nil,
				DebugMode:                 true,
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		actual := isEqualConfig(tc.a, tc.b)
		assert.Equalf(t, tc.expected, actual, tc.description)
	}
}

func absPath(path string) string {
	ret, _ := filepath.Abs(path)
	return ret
}
