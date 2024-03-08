package config

import (
	"os"
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
	singleCoreStr         = "single-core"
	dualCoreStr           = "dual-core"
	quadCoreStr           = "quad-core"
	warboyStr             = "warboy"
	renegadeStr           = "renegade"
)

type ResourceUnitStrategy string

const (
	LegacyStrategy     ResourceUnitStrategy = legacyStrategyStr
	GenericStrategy    ResourceUnitStrategy = genericStrategyStr
	SingleCoreStrategy ResourceUnitStrategy = singleCoreStr
	DualCoreStrategy   ResourceUnitStrategy = dualCoreStr
	QuadCoreStrategy   ResourceUnitStrategy = quadCoreStr
)

type ResourceKind string

const (
	Warboy   ResourceKind = warboyStr
	Renegade ResourceKind = renegadeStr
)

type Config struct {
	ResourceStrategyMap map[ResourceKind]ResourceUnitStrategy `yaml:"resourceStrategyMap" validate:"required"`
	DebugMode           bool                                  `yaml:"debugMode"`
}

func (c *Config) IsDebugMode() bool {
	return c.DebugMode
}

func (c *Config) GetResourceStrategyMap() map[ResourceKind]ResourceUnitStrategy {
	return c.ResourceStrategyMap
}

func ensureLocalConfigExist(localConfigPath string) bool {
	if localConfigPath == "" {
		return false
	}

	if info, err := os.Stat(localConfigPath); err != nil || info.IsDir() {
		return false
	}

	return true
}

func GetMergedConfigWithWatcher(confUpdateChan chan *fsnotify.Event, localConfigPath string) (*Config, error) {
	globalConf, globalErr := getValidatedConfigAndWatch(confUpdateChan, globalConfigMountPath)
	if globalErr != nil {
		return nil, globalErr
	}

	//check whether local config exist
	if !ensureLocalConfigExist(localConfigPath) {
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
		conf := sl.Current().Interface().(Config)

		for key, strategy := range conf.ResourceStrategyMap {
			switch key {
			case Warboy:
				switch strategy {
				case LegacyStrategy:
				case GenericStrategy:
				case SingleCoreStrategy:
				case DualCoreStrategy:
				default:
					// Unknown or unsupported strategy(quad core)
					sl.ReportError(conf.ResourceStrategyMap, "ResourceStrategyMap", "resourceStrategyMap", "required", "")
				}
			case Renegade:
				switch strategy {
				case LegacyStrategy:
				case GenericStrategy:
				case SingleCoreStrategy:
				case DualCoreStrategy:
				case QuadCoreStrategy:
				default:
					// Unknown strategy
					sl.ReportError(conf.ResourceStrategyMap, "ResourceStrategyMap", "resourceStrategyMap", "required", "")
				}
			default:
				//Unknown resource kind.
				sl.ReportError(conf.ResourceStrategyMap, "ResourceStrategyMap", "resourceStrategyMap", "required", "")
			}
		}
	}, Config{})

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
		ResourceStrategyMap: nil,
		DebugMode:           false,
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
