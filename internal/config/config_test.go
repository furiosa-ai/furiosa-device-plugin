package config

import (
	"path/filepath"
	"reflect"
	"testing"
)

func getConfigFromFile(configPath string) (*Config, error) {
	configPath = absPath(configPath)
	conf, err := readInConfigAsMap(configPath)
	if err != nil {
		return nil, err
	}
	return convertToConfig(conf)
}

func TestGetConfigFromSingleFile(t *testing.T) {
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
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy:   LegacyStrategy,
					Renegade: LegacyStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description: "parse generic configuration",
			configPath:  "./tests/generic_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy:   GenericStrategy,
					Renegade: GenericStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description: "parse single-core configuration",
			configPath:  "./tests/single_core_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy:   SingleCoreStrategy,
					Renegade: SingleCoreStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description: "parse dual-core configuration",
			configPath:  "./tests/dual_core_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy:   DualCoreStrategy,
					Renegade: DualCoreStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description: "parse quad-core configuration",
			configPath:  "./tests/quad_core_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Renegade: QuadCoreStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description: "parse mixed strategy configuration",
			configPath:  "./tests/mixed_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy:   LegacyStrategy,
					Renegade: QuadCoreStrategy,
				},
				DebugMode: true,
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

	for _, tc := range tests {
		actualConf, actualErr := getConfigFromFile(tc.configPath)

		if actualErr != nil != tc.expectedError {
			t.Errorf("got unexpected error %t", actualErr)
			continue
		}

		if !reflect.DeepEqual(actualConf, tc.expectedResult) {
			t.Errorf("expected %v but got %v", tc.expectedResult, actualConf)
		}
	}
}

func validateConfigFromFile(configFilePath string) (*Config, error) {
	conf, err := getConfigFromFile(configFilePath)
	// getConfigFromFile should be validated already
	if err != nil {
		panic(err)
	}
	err = validateConfig(conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		description    string
		configFilePath string
		expectedResult *Config
		expectedError  bool
	}{
		{
			description:    "test legacy ResourceUnitStrategy",
			configFilePath: absPath("./tests/legacy_strategy.yaml"),
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy:   LegacyStrategy,
					Renegade: LegacyStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description:    "test generic ResourceUnitStrategy",
			configFilePath: absPath("./tests/generic_strategy.yaml"),
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy:   GenericStrategy,
					Renegade: GenericStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description:    "test single core ResourceUnitStrategy",
			configFilePath: absPath("./tests/single_core_strategy.yaml"),
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy:   SingleCoreStrategy,
					Renegade: SingleCoreStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description:    "test dual core ResourceUnitStrategy",
			configFilePath: absPath("./tests/dual_core_strategy.yaml"),
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy:   DualCoreStrategy,
					Renegade: DualCoreStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description:    "test quad core ResourceUnitStrategy",
			configFilePath: absPath("./tests/quad_core_strategy.yaml"),
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Renegade: QuadCoreStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description:    "test mixed ResourceUnitStrategy",
			configFilePath: absPath("./tests/mixed_strategy.yaml"),
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy:   LegacyStrategy,
					Renegade: QuadCoreStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description:    "test validation error with wrong ResourceUnitStrategy",
			configFilePath: absPath("./tests/wrong_strategy.yaml"),
			expectedResult: nil,
			expectedError:  true,
		},
		{
			description:    "test validation error with wrong resource kind",
			configFilePath: absPath("./tests/wrong_kind.yaml"),
			expectedResult: nil,
			expectedError:  true,
		},
		{
			description:    "test validation error with missing required field",
			configFilePath: absPath("./tests/missing_required.yaml"),
			expectedResult: nil,
			expectedError:  true,
		},
	}
	for _, tc := range tests {
		actualConf, actualErr := validateConfigFromFile(tc.configFilePath)
		if actualErr != nil != tc.expectedError {
			t.Errorf("got unexpected error %t", actualErr)
			continue
		}

		if !reflect.DeepEqual(actualConf, tc.expectedResult) {
			t.Errorf("expected %v but got %v", tc.expectedResult, actualConf)
		}
	}
}

func TestMergeConfig(t *testing.T) {
	tests := []struct {
		description      string
		globalConfigPath string
		localConfigPath  string
		expectedConfig   *Config
	}{
		{
			description:      "merge same configs",
			globalConfigPath: absPath("./tests/generic_strategy.yaml"),
			localConfigPath:  absPath("./tests/generic_strategy.yaml"),
			expectedConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy:   "generic",
					Renegade: "generic",
				},
				DisabledDevices: nil,
				DebugMode:       true,
			},
		},
		{
			description:      "merge device ResourceUnitStrategy",
			globalConfigPath: absPath("./tests/legacy_strategy.yaml"),
			localConfigPath:  absPath("./tests/generic_strategy.yaml"),
			expectedConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy:   "generic",
					Renegade: "generic",
				},
				DisabledDevices: nil,
				DebugMode:       true,
			},
		},
		{
			description:      "merge with nothing",
			globalConfigPath: absPath("./tests/generic_strategy.yaml"),
			localConfigPath:  absPath("./tests/empty.yaml"),
			expectedConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Warboy:   "generic",
					Renegade: "generic",
				},
				DisabledDevices: nil,
				DebugMode:       true,
			},
		},
		{
			description:      "merge with zero value",
			globalConfigPath: absPath("./tests/generic_strategy.yaml"),
			localConfigPath:  absPath("./tests/zero.yaml"),
			expectedConfig: &Config{
				ResourceStrategyMap: nil,
				DisabledDevices:     []string{},
				DebugMode:           false,
			},
		},
	}

	testMergeConfig := func(globalConfigPath string, localConfigPath string) *Config {
		g, err := readInConfigAsMap(globalConfigPath)
		if err != nil {
			panic(err)
		}
		l, err := readInConfigAsMap(localConfigPath)
		if err != nil {
			panic(err)
		}
		mergeMaps(g, l)
		conf, err := convertToConfig(g)
		if err != nil {
			panic(err)
		}
		return conf
	}

	for _, tc := range tests {
		actualConfig := testMergeConfig(tc.globalConfigPath, tc.localConfigPath)
		if !reflect.DeepEqual(actualConfig, tc.expectedConfig) {
			t.Errorf("%s: expected %v but got %v", tc.description, tc.expectedConfig, actualConfig)
		}
	}
}

func TestIsDebugMode(t *testing.T) {
	tests := []struct {
		description       string
		config            *Config
		expectedDebugMode bool
	}{
		{
			description: "test true",
			config: &Config{
				DebugMode: true,
			},
			expectedDebugMode: true,
		},
		{
			description: "test false",
			config: &Config{
				DebugMode: false,
			},
			expectedDebugMode: false,
		},
	}

	for _, tc := range tests {
		if tc.config.DebugMode != tc.expectedDebugMode {
			t.Errorf("expected %v but got %v", tc.expectedDebugMode, tc.config.DebugMode)
		}
	}
}

func absPath(path string) string {
	ret, _ := filepath.Abs(path)
	return ret
}

// func TestResourceUnitStrategyConfig(t *testing.T) {
// 	tests := []struct {
// 		description            string
// 		config                 *Config
// 		expectedDeviceStrategy ResourceUnitStrategy
// 	}{
// 		{
// 			description: "test legacy ResourceUnitStrategy",
// 			config: &Config{
// 				ResourceUnitStrategyConfig: ResourceStrategy{ResourceUnitStrategy: legacyStrategyStr},
// 			},
// 			expectedDeviceStrategy: LegacyStrategy,
// 		},
// 		{
// 			description: "test generic ResourceUnitStrategy",
// 			config: &Config{
// 				ResourceUnitStrategyConfig: ResourceStrategy{ResourceUnitStrategy: genericStrategyStr},
// 			},
// 			expectedDeviceStrategy: GenericStrategy,
// 		},
// 		{
// 			description: "test single core ResourceUnitStrategy",
// 			config: &Config{
// 				ResourceUnitStrategyConfig: ResourceStrategy{ResourceUnitStrategy: singleCoreStr},
// 			},
// 			expectedDeviceStrategy: SingleCoreStrategy,
// 		},
// 		{
// 			description: "test dual core ResourceUnitStrategy",
// 			config: &Config{
// 				ResourceUnitStrategyConfig: ResourceStrategy{ResourceUnitStrategy: dualCoreStr},
// 			},
// 			expectedDeviceStrategy: DualCoreStrategy,
// 		},
// 		{
// 			description: "test quad core ResourceUnitStrategy",
// 			config: &Config{
// 				ResourceUnitStrategyConfig: ResourceStrategy{ResourceUnitStrategy: quadCoreStr},
// 			},
// 			expectedDeviceStrategy: QuadCoreStrategy,
// 		},
// 	}
//
// 	for _, tc := range tests {
// 		if tc.config.GetResourceUnitStrategyConfig() != tc.expectedDeviceStrategy {
// 			t.Errorf("expected %v but got %v", tc.expectedDeviceStrategy, tc.config.GetResourceUnitStrategyConfig())
// 		}
// 	}
// }
