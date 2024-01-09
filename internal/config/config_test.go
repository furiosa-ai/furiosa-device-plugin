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
				DeviceStrategy: DeviceStrategyConfig{Strategy: "legacy"},
				DebugMode:      true,
			},
			expectedError: false,
		},
		{
			description:    "parse generic configuration",
			configPath:     "./tests/",
			configFilename: "generic_strategy.yaml",
			expectedResult: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "generic"},
				DebugMode:      true,
			},
			expectedError: false,
		},
		{
			description:    "parse single-pe-isolation configuration",
			configPath:     "./tests/",
			configFilename: "single_pe_isolation_strategy.yaml",
			expectedResult: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "single-pe-isolation"},
				DebugMode:      true,
			},
			expectedError: false,
		},
		{
			description:    "parse 2pe-fusion configuration",
			configPath:     "./tests/",
			configFilename: "two_pe_fusion_strategy.yaml",
			expectedResult: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "2pe-fusion"},
				DebugMode:      true,
			},
			expectedError: false,
		},
		{
			description:    "parse 4pe-fusion configuration",
			configPath:     "./tests/",
			configFilename: "four_pe_fusion_strategy.yaml",
			expectedResult: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "4pe-fusion"},
				DebugMode:      true,
			},
			expectedError: false,
		},
		{
			description:    "try wrong configuration",
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
			description:    "test legacy Strategy",
			configFilePath: abs("./tests/legacy_strategy.yaml"),
			expectedResult: &Config{
				DeviceStrategy: DeviceStrategyConfig{
					Strategy: "legacy",
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description:    "test generic Strategy",
			configFilePath: abs("./tests/generic_strategy.yaml"),
			expectedResult: &Config{
				DeviceStrategy: DeviceStrategyConfig{
					Strategy: "generic",
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description:    "test single pe isolation Strategy",
			configFilePath: abs("./tests/single_pe_isolation_strategy.yaml"),
			expectedResult: &Config{
				DeviceStrategy: DeviceStrategyConfig{
					Strategy: "single-pe-isolation",
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description:    "two pe fusion Strategy",
			configFilePath: abs("./tests/two_pe_fusion_strategy.yaml"),
			expectedResult: &Config{
				DeviceStrategy: DeviceStrategyConfig{
					Strategy: "2pe-fusion",
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description:    "four pe fusion Strategy",
			configFilePath: abs("./tests/four_pe_fusion_strategy.yaml"),
			expectedResult: &Config{
				DeviceStrategy: DeviceStrategyConfig{
					Strategy: "4pe-fusion",
				},
				DebugMode: true,
			},
			expectedError: false,
		},
		{
			description:    "test validation error with wrong Strategy",
			configFilePath: abs("./tests/wrong_strategy.yaml"),
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
				DeviceStrategy: DeviceStrategyConfig{Strategy: "generic"},
				DebugMode:      false,
			},
			localConfig: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "generic"},
				DebugMode:      false,
			},
			expectedConfig: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "generic"},
				DebugMode:      false,
			},
		},
		{
			description: "merge device Strategy",
			globalConfig: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "legacy"},
				DebugMode:      false,
			},
			localConfig: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "generic"},
				DebugMode:      false,
			},
			expectedConfig: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "generic"},
				DebugMode:      false,
			},
		},
		{
			description: "merge debug mode",
			globalConfig: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "generic"},
				DebugMode:      false,
			},
			localConfig: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "generic"},
				DebugMode:      true,
			},
			expectedConfig: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "generic"},
				DebugMode:      true,
			},
		},
		{
			description: "merge Strategy and debug mode",
			globalConfig: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "legacy"},
				DebugMode:      false,
			},
			localConfig: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "generic"},
				DebugMode:      true,
			},
			expectedConfig: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: "generic"},
				DebugMode:      true,
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

func TestGetDeviceStrategy(t *testing.T) {
	tests := []struct {
		description            string
		config                 *Config
		expectedDeviceStrategy DeviceStrategy
	}{
		{
			description: "test legacy Strategy",
			config: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: legacyStrategyStr},
			},
			expectedDeviceStrategy: LegacyStrategy,
		},
		{
			description: "test generic Strategy",
			config: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: genericStrategyStr},
			},
			expectedDeviceStrategy: GenericStrategy,
		},
		{
			description: "test single pe isolation Strategy",
			config: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: singlePeIsolationStr},
			},
			expectedDeviceStrategy: SinglePeIsolation,
		},
		{
			description: "test two pe fusion Strategy",
			config: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: twoPeFusionStr},
			},
			expectedDeviceStrategy: TwoPeFusion,
		},
		{
			description: "test four pe fusion Strategy",
			config: &Config{
				DeviceStrategy: DeviceStrategyConfig{Strategy: fourPeFusionStr},
			},
			expectedDeviceStrategy: FourPeFusion,
		},
	}

	for _, tc := range tests {
		if tc.config.GetDeviceStrategy() != tc.expectedDeviceStrategy {
			t.Errorf("expected %v but got %v", tc.expectedDeviceStrategy, tc.config.GetDeviceStrategy())
		}
	}
}
