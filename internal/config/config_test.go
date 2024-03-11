package config

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetConfigFromFile(t *testing.T) {
	tests := []struct {
		description    string
		configPath     string
		configFilename string
		expectedResult *Config
		expectedError  bool
	}{
		{
			description:    "parse legacy configuration",
			configPath:     "./tests/",
			configFilename: "legacy_strategy.yaml",
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
			description:    "parse generic configuration",
			configPath:     "./tests/",
			configFilename: "generic_strategy.yaml",
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
			description:    "parse single-core configuration",
			configPath:     "./tests/",
			configFilename: "single_core_strategy.yaml",
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
			description:    "parse dual-core configuration",
			configPath:     "./tests/",
			configFilename: "dual_core_strategy.yaml",
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
			description:    "parse quad-core configuration",
			configPath:     "./tests/",
			configFilename: "quad_core_strategy.yaml",
			expectedResult: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{
					Renegade: QuadCoreStrategy,
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description:    "parse mixed strategy configuration",
			configPath:     "./tests/",
			configFilename: "mixed_strategy.yaml",
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
			configPath:     "./tests/",
			configFilename: "wrong_format.yaml",
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tc := range tests {
		_, actualConf, actualErr := getConfigFromFile(tc.configPath, tc.configFilename)

		if actualErr != nil != tc.expectedError {
			t.Errorf("got unexpected error %t", actualErr)
			continue
		}

		if !reflect.DeepEqual(actualConf, tc.expectedResult) {
			t.Errorf("expected %v but got %v", tc.expectedResult, actualConf)
		}
	}
}

func abs(path string) string {
	ret, _ := filepath.Abs(path)
	return ret
}

func TestGetValidatedConfigAndWatch(t *testing.T) {
	tests := []struct {
		description    string
		configFilePath string
		expectedResult *Config
		expectedError  bool
	}{
		{
			description:    "test legacy ResourceUnitStrategy",
			configFilePath: abs("./tests/legacy_strategy.yaml"),
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
			configFilePath: abs("./tests/generic_strategy.yaml"),
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
			configFilePath: abs("./tests/single_core_strategy.yaml"),
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
			configFilePath: abs("./tests/dual_core_strategy.yaml"),
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
			configFilePath: abs("./tests/quad_core_strategy.yaml"),
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
			configFilePath: abs("./tests/mixed_strategy.yaml"),
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
			configFilePath: abs("./tests/wrong_strategy.yaml"),
			expectedResult: nil,
			expectedError:  true,
		},
		{
			description:    "test validation error with wrong resource kind",
			configFilePath: abs("./tests/wrong_kind.yaml"),
			expectedResult: nil,
			expectedError:  true,
		},
		{
			description:    "test validation error with missing required field",
			configFilePath: abs("./tests/missing_required.yaml"),
			expectedResult: nil,
			expectedError:  true,
		},
	}
	for _, tc := range tests {
		actualConf, actualErr := getValidatedConfigAndWatch(nil, tc.configFilePath)
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
		description    string
		globalConfig   *Config
		localConfig    *Config
		expectedConfig *Config
	}{
		{
			description: "merge same configs",
			globalConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{Warboy: "generic"},
				DebugMode:           false,
			},
			localConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{Warboy: "generic"},
				DebugMode:           false,
			},
			expectedConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{Warboy: "generic"},
				DebugMode:           false,
			},
		},
		{
			description: "merge device ResourceUnitStrategy",
			globalConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{Warboy: "legacy"},
				DebugMode:           false,
			},
			localConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{Warboy: "generic"},
				DebugMode:           false,
			},
			expectedConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{Warboy: "generic"},
				DebugMode:           false,
			},
		},
		{
			description: "merge debug mode",
			globalConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{Warboy: "generic"},
				DebugMode:           false,
			},
			localConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{Warboy: "generic"},
				DebugMode:           true,
			},
			expectedConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{Warboy: "generic"},
				DebugMode:           true,
			},
		},
		{
			description: "merge ResourceUnitStrategy and debug mode",
			globalConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{Warboy: "legacy"},
				DebugMode:           false,
			},
			localConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{Warboy: "generic"},
				DebugMode:           true,
			},
			expectedConfig: &Config{
				ResourceStrategyMap: map[ResourceKind]ResourceUnitStrategy{Warboy: "generic"},
				DebugMode:           true,
			},
		},
		{
			description: "merge with zero value",
			globalConfig: &Config{
				DebugMode: true,
			},
			localConfig: &Config{
				DebugMode: false,
			},
			expectedConfig: &Config{
				DebugMode: false,
			},
		},
	}

	for _, tc := range tests {
		mergeConfig(tc.globalConfig, tc.localConfig)
		if !reflect.DeepEqual(tc.globalConfig, tc.expectedConfig) {
			t.Errorf("expected %v but got %v", tc.expectedConfig, tc.globalConfig)
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
		if tc.config.IsDebugMode() != tc.expectedDebugMode {
			t.Errorf("expected %v but got %v", tc.expectedDebugMode, tc.config.IsDebugMode())
		}
	}
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
