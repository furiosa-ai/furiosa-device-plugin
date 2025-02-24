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
			description: "parse non partitioning configuration",
			configPath:  "./tests/non_partitioning.yaml",
			expectedResult: &Config{
				Partitioning: NonePolicyStr,
				DebugMode:    true,
			},
			expectedError: false,
		},
		{
			description: "parse 2core.12gb partitioning configuration",
			configPath:  "./tests/2core.12gb_partitioning.yaml",
			expectedResult: &Config{
				Partitioning: Rngd2Core12GbStr,
				DebugMode:    true,
			},
			expectedError: false,
		},
		{
			description: "parse 4core.24gb partitioning configuration",
			configPath:  "./tests/4core.24gb_partitioning.yaml",
			expectedResult: &Config{
				Partitioning: Rngd4Core24GbStr,
				DebugMode:    true,
			},
			expectedError: false,
		},
		{
			description: "parse uuid list",
			configPath:  "./tests/with_disabled_devices.yaml",
			expectedResult: &Config{
				Partitioning: NonePolicyStr,
				DisabledDeviceUUIDs: map[string][]string{
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
		t.Run(tc.description, func(t *testing.T) {
			actualConf, actualErr := getConfigFromFile(absPath(tc.configPath))
			if tc.expectedError {
				assert.NotNilf(t, actualErr, tc.description)
			} else {
				assert.Nilf(t, actualErr, tc.description)
				assert.Equalf(t, tc.expectedResult, actualConf, tc.description)
			}
		})
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
				Partitioning:        NonePolicyStr,
				DisabledDeviceUUIDs: map[string][]string{"node_a": {"a0", "a1"}},
				DebugMode:           true,
			},
			b: &Config{
				Partitioning:        NonePolicyStr,
				DisabledDeviceUUIDs: map[string][]string{"node_a": {"a0", "a1"}},
				DebugMode:           true,
			},
			expected: true,
		},
		{
			description: "different resource strategy map",
			a: &Config{
				Partitioning:        NonePolicyStr,
				DisabledDeviceUUIDs: map[string][]string{"node_a": {"a0", "a1"}},
				DebugMode:           true,
			},
			b: &Config{
				Partitioning:        Rngd2Core12GbStr,
				DisabledDeviceUUIDs: map[string][]string{"node_a": {"a0", "a1"}},
				DebugMode:           true,
			},
			expected: false,
		},
		{
			description: "different disabled device uuid list map1",
			a: &Config{
				Partitioning:        NonePolicyStr,
				DisabledDeviceUUIDs: map[string][]string{"node_a": {"a0", "a1"}},
				DebugMode:           true,
			},
			b: &Config{
				Partitioning:        NonePolicyStr,
				DisabledDeviceUUIDs: map[string][]string{"node_b": {"a0", "a1"}},
				DebugMode:           true,
			},
			expected: false,
		},
		{
			description: "different disabled device uuid list map2",
			a: &Config{
				Partitioning:        NonePolicyStr,
				DisabledDeviceUUIDs: map[string][]string{"node_a": {"a0", "a1"}},
				DebugMode:           true,
			},
			b: &Config{
				Partitioning:        NonePolicyStr,
				DisabledDeviceUUIDs: map[string][]string{"node_a": {"a1", "a2"}},
				DebugMode:           true,
			},
			expected: false,
		},
		{
			description: "different disabled device uuid list map3",
			a: &Config{
				Partitioning:        NonePolicyStr,
				DisabledDeviceUUIDs: map[string][]string{"node_a": {"a0", "a1"}},
				DebugMode:           true,
			},
			b: &Config{
				Partitioning:        NonePolicyStr,
				DisabledDeviceUUIDs: nil,
				DebugMode:           true,
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			actual := isEqualConfig(tc.a, tc.b)
			assert.Equalf(t, tc.expected, actual, tc.description)
		})
	}
}

func absPath(path string) string {
	ret, _ := filepath.Abs(path)
	return ret
}
