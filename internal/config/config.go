package config

import (
	"path/filepath"
	"reflect"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

const (
	configType            = "yaml"
	globalConfigMountPath = "/etc/config/global_config.yaml"
	legacyStrategyStr     = "legacy"
	genericStrategyStr    = "generic"
	singlePeIsolationStr  = "single-pe-isolation"
	twoPeFusionStr        = "2pe-fusion"
	fourPeFusionStr       = "4pe-fusion"
)

type DeviceStrategy string

const (
	LegacyStrategy    DeviceStrategy = legacyStrategyStr
	GenericStrategy   DeviceStrategy = genericStrategyStr
	SinglePeIsolation DeviceStrategy = singlePeIsolationStr
	TwoPeFusion       DeviceStrategy = twoPeFusionStr
	FourPeFusion      DeviceStrategy = fourPeFusionStr
)

type DeviceStrategyConfig struct {
	Strategy string `yaml:"strategy" validate:"required"`
}

type Config struct {
	DeviceStrategy DeviceStrategyConfig `yaml:"deviceStrategy" validate:"required"`
	DebugMode      bool                 `yaml:"debugMode"`
}

func (c *Config) IsDebugMode() bool {
	return c.DebugMode
}

func (c *Config) GetDeviceStrategy() DeviceStrategy {
	var strategy DeviceStrategy = ""
	//Note(@bg): struct validation guarantees that the value is one of following.
	switch c.DeviceStrategy.Strategy {
	case legacyStrategyStr:
		strategy = LegacyStrategy
	case genericStrategyStr:
		strategy = GenericStrategy
	case singlePeIsolationStr:
		strategy = SinglePeIsolation
	case twoPeFusionStr:
		strategy = TwoPeFusion
	case fourPeFusionStr:
		strategy = FourPeFusion
	}
	return strategy
}

func GetMergedConfigWithWatcher(confUpdateChan chan *fsnotify.Event, localConfigPath string) (*Config, error) {
	globalConf, globalErr := getValidatedConfigAndWatch(confUpdateChan, globalConfigMountPath)
	if globalErr != nil {
		return nil, globalErr
	}

	if localConfigPath == "" {
		return globalConf, nil
	}

	localConf, localErr := getValidatedConfigAndWatch(confUpdateChan, localConfigPath)
	if localErr != nil {
		return nil, localErr
	}

	mergeConfig(globalConf, localConf)

	return globalConf, nil
}

// mergeConfig merge global config and local config.
// the fields of the global config will be updated with the fields of the local config.
func mergeConfig(global interface{}, local interface{}) {
	glbVal := reflect.ValueOf(global).Elem()
	localVal := reflect.ValueOf(local).Elem()

	for i := 0; i < glbVal.NumField(); i++ {
		glbField := glbVal.Field(i)
		localField := localVal.Field(i)

		// recursively merge inner struct
		if glbField.Kind() == reflect.Struct {
			mergeConfig(glbField.Addr().Interface(), localField.Addr().Interface())
			continue
		}

		if !localField.IsZero() {
			glbField.Set(localField)
		}
	}
}

func getValidatedConfigAndWatch(confUpdateChan chan *fsnotify.Event, configFilePath string) (*Config, error) {
	v, conf, err := getConfigFromFile(filepath.Dir(configFilePath), filepath.Base(configFilePath))
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	validate.RegisterStructValidation(func(sl validator.StructLevel) {
		deviceStrategy := sl.Current().Interface().(DeviceStrategyConfig)
		switch deviceStrategy.Strategy {
		case legacyStrategyStr:
			return
		case genericStrategyStr:
			return
		case singlePeIsolationStr:
			return
		case twoPeFusionStr:
			return
		case fourPeFusionStr:
			return
		default:
			sl.ReportError(deviceStrategy.Strategy, "Strategy", "DeviceStrategy", "required", "")
		}

	}, DeviceStrategyConfig{})

	err = validate.Struct(conf)
	if err != nil {
		return nil, err
	}

	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		confUpdateChan <- &in
	})

	return conf, nil
}

func getDefaultConfig() *Config {
	return &Config{
		DebugMode: false,
	}
}

func getConfigFromFile(filepath string, filename string) (*viper.Viper, *Config, error) {
	v := viper.New()
	v.SetConfigType(configType)

	v.AddConfigPath(filepath)
	v.SetConfigName(filename)

	err := v.ReadInConfig()
	if err != nil {
		return nil, nil, err
	}

	conf := getDefaultConfig()
	err = v.Unmarshal(conf)
	if err != nil {
		return nil, nil, err
	}

	return v, conf, nil
}
