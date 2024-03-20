package config

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

func TestValidateConfigYaml(t *testing.T) {
	tests := []struct {
		description   string
		configPath    string
		expectedError bool
	}{
		{
			description:   "test schema of legacy strategy",
			configPath:    "./tests/global_legacy_strategy.yaml",
			expectedError: false,
		},
		{
			description:   "test schema of generic strategy",
			configPath:    "./tests/global_generic_strategy.yaml",
			expectedError: false,
		},
		{
			description:   "test schema of full override",
			configPath:    "./tests/override_full.yaml",
			expectedError: false,
		},
		{
			description:   "test wrong schema",
			configPath:    "./tests/wrong_format.yaml",
			expectedError: true,
		},
	}

	for _, tc := range tests {
		err := validateConfigYaml(absPath(tc.configPath))
		if tc.expectedError {
			assert.NotNilf(t, err, tc.description)
		} else {
			assert.Nilf(t, err, tc.description)
		}
	}
}

func TestReadInConfigAsMap(t *testing.T) {
	testFile := absPath("./tests/override_full.yaml")
	confMap, err := readInConfigAsMap(testFile)
	assert.Nil(t, err)

	globalMap := confMap.Global
	assert.Equal(t, map[string]interface{}{"warboy": "generic", "rngd": "generic"}, globalMap["resourceStrategyMap"])
	assert.Equal(t, false, globalMap["debugMode"])

	localMap := confMap.Overrides
	assert.Contains(t, localMap, "this")
	assert.Contains(t, localMap, "other")

	thisMap := localMap["this"]
	assert.Equal(t, map[string]interface{}{"warboy": "single-core", "rngd": "single-core"}, thisMap["resourceStrategyMap"])
	assert.Equal(t, []interface{}{"b0"}, thisMap["disabledDeviceUUIDList"])
	assert.Equal(t, true, thisMap["debugMode"])
}

func TestConvertToConfig(t *testing.T) {
	tests := []struct {
		description    string
		input          map[string]interface{}
		expectedResult *Config
		expectedError  bool
	}{
		{
			description: "convert full map",
			input: map[string]interface{}{
				"resourceStrategyMap":    map[string]string{"warboy": "generic", "rngd": "generic"},
				"disabledDeviceUUIDList": []string{"a0", "a1"},
				"debugMode":              true,
			},
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: GenericStrategy,
					rngd:   GenericStrategy,
				},
				DisabledDeviceUUIDList: []string{"a0", "a1"},
				DebugMode:              true,
			},
			expectedError: false,
		},
		{
			description: "convert partial map",
			input: map[string]interface{}{
				"resourceStrategyMap": map[string]string{"warboy": "generic", "rngd": "generic"},
			},
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: GenericStrategy,
					rngd:   GenericStrategy,
				},
				DisabledDeviceUUIDList: nil,
				DebugMode:              false,
			},
			expectedError: false,
		},
	}

	for _, tc := range tests {
		actual, err := convertToConfig(tc.input)
		if tc.expectedError {
			assert.NotNilf(t, err, tc.description)
		} else {
			assert.Nilf(t, err, tc.description)
			assert.Equalf(t, tc.expectedResult, actual, tc.description)
		}
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		description   string
		config        *Config
		expectedError bool
	}{
		{
			description: "valid config",
			config: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: GenericStrategy,
					rngd:   GenericStrategy,
				},
			},
			expectedError: false,
		},
		{
			description: "invalid config (warboy with quad core strategy)",
			config: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: QuadCoreStrategy,
					rngd:   GenericStrategy,
				},
			},
			expectedError: true,
		},
		{
			description: "invalid config (invalid strategy)",
			config: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: "foo",
					rngd:   "bar",
				},
			},
			expectedError: true,
		},
		{
			description: "invalid config (unknown resource kind)",
			config: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: GenericStrategy,
					rngd:   GenericStrategy,
					"Foo":  GenericStrategy,
				},
			},
			expectedError: true,
		},
		{
			description:   "invalid config (missing required field)",
			config:        &Config{},
			expectedError: true,
		},
	}

	for _, tc := range tests {
		err := validateConfig(tc.config)
		if tc.expectedError {
			assert.NotNilf(t, err, tc.description)
		} else {
			assert.Nilf(t, err, tc.description)
		}
	}
}

func TestGetMergedConfigFromFile(t *testing.T) {
	tests := []struct {
		description    string
		configPath     string
		expectedResult *Config
		expectedError  bool
	}{
		{
			description: "parse legacy configuration",
			configPath:  "./tests/global_legacy_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: LegacyStrategy,
					rngd:   LegacyStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description: "parse generic configuration",
			configPath:  "./tests/global_generic_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: GenericStrategy,
					rngd:   GenericStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description: "parse single-core configuration",
			configPath:  "./tests/global_single_core_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: SingleCoreStrategy,
					rngd:   SingleCoreStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description: "parse dual-core configuration",
			configPath:  "./tests/global_dual_core_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: DualCoreStrategy,
					rngd:   DualCoreStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description: "parse quad-core configuration",
			configPath:  "./tests/global_quad_core_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					rngd: QuadCoreStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description: "parse mixed strategy configuration",
			configPath:  "./tests/global_mixed_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: LegacyStrategy,
					rngd:   QuadCoreStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description: "parse and override with full config",
			configPath:  "./tests/override_full.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: SingleCoreStrategy,
					rngd:   SingleCoreStrategy,
				},
				DisabledDeviceUUIDList: []string{"b0"},
				DebugMode:              true,
			},
			expectedError: false,
		},
		{
			description: "parse and override with partial config",
			configPath:  "./tests/override_partial.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: GenericStrategy,
					rngd:   SingleCoreStrategy,
				},
				DisabledDeviceUUIDList: nil,
				DebugMode:              true,
			},
			expectedError: false,
		},
		{
			description: "parse and override with empty config",
			configPath:  "./tests/override_empty.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: GenericStrategy,
					rngd:   GenericStrategy,
				},
				DisabledDeviceUUIDList: nil,
				DebugMode:              false,
			},
			expectedError: false,
		},
		{
			description: "parse and override with zero values (merge maps, override otherwise)",
			configPath:  "./tests/override_zero.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: GenericStrategy,
					rngd:   GenericStrategy,
				},
				DisabledDeviceUUIDList: []string{},
				DebugMode:              false,
			},
			expectedError: false,
		},
		{
			description:    "try wrong format",
			configPath:     "./tests/wrong_format.yaml",
			expectedResult: nil,
			expectedError:  true,
		},
	}

	mockNodeNameGetter := newMockNodenameGetter("this")
	for _, tc := range tests {
		actualConf, actualErr := getMergedConfigFromFile(absPath(tc.configPath), mockNodeNameGetter)
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
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: GenericStrategy,
					rngd:   GenericStrategy,
				},
				DisabledDeviceUUIDList: []string{"a0", "a1"},
				DebugMode:              true,
			},
			b: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: GenericStrategy,
					rngd:   GenericStrategy,
				},
				DisabledDeviceUUIDList: []string{"a0", "a1"},
				DebugMode:              true,
			},
			expected: true,
		},
		{
			description: "different resource strategy map",
			a: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: GenericStrategy,
					rngd:   GenericStrategy,
				},
				DisabledDeviceUUIDList: []string{"a0", "a1"},
				DebugMode:              true,
			},
			b: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy: SingleCoreStrategy,
					rngd:   SingleCoreStrategy,
				},
				DisabledDeviceUUIDList: []string{"a0", "a1"},
				DebugMode:              true,
			},
			expected: false,
		},
		{
			description: "different disabled device uuid list",
			a: &Config{
				ResourceStrategyMap:    map[ResourceKind]ResourceUnitStrategy{},
				DisabledDeviceUUIDList: []string{"a0", "a1"},
				DebugMode:              true,
			},
			b: &Config{
				ResourceStrategyMap:    map[ResourceKind]ResourceUnitStrategy{},
				DisabledDeviceUUIDList: []string{"a0", "a2"},
				DebugMode:              true,
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
