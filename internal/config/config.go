package config

import (
	"os"
	"path/filepath"
	"reflect"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
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
	DeviceDenyList      []string                              `yaml:"deviceDenyList"`
	DebugMode           bool                                  `yaml:"debugMode"`
}

func GetMergedConfigWithWatcher(confUpdateChan chan *fsnotify.Event, localConfigPath string) (*Config, error) {
	var err error
	var localConf map[string]interface{}

	globalConf, err := readInConfigAsMap(globalConfigMountPath)
	if err != nil {
		return nil, err
	}

	//check whether local config exist
	if !ensureLocalConfigExist(localConfigPath) {
		localConf = make(map[string]interface{})
	} else {
		localConf, err = readInConfigAsMap(localConfigPath)
		if err != nil {
			return nil, err
		}
	}
	mergeMaps(globalConf, localConf)
	config, err := convertToConfig(globalConf)
	if err != nil {
		return nil, err
	}
	err = validateConfig(config)
	if err != nil {
		return nil, err
	}

	startFileWatch(confUpdateChan, globalConfigMountPath)
	startFileWatch(confUpdateChan, localConfigPath)

	return config, nil
}

func readInConfigAsMap(configFilePath string) (map[string]interface{}, error) {
	contents, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	err = yaml.Unmarshal(contents, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func convertToConfig(confAsMap map[string]interface{}) (*Config, error) {
	conf := getDefaultConfig()

	v := viper.New()
	for key, val := range confAsMap {
		v.Set(key, val)
	}
	err := v.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func validateConfig(conf *Config) error {
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

	return validate.Struct(conf)
}

func startFileWatch(confUpdateChan chan *fsnotify.Event, filePath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = watcher.Add(filepath.Dir(filePath))
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write && event.Name == filePath {
					confUpdateChan <- &event
				}
			}
		}
	}()
	return nil
}

func mergeMaps(dst, src map[string]interface{}) {
	for k, v := range src {
		if v == nil {
			continue
		}
		if reflect.TypeOf(v).Kind() == reflect.Map {
			// if dst[k] does not exist, or is not a map, override it with a new map
			_, hasKey := dst[k]
			if !hasKey || reflect.TypeOf(dst[k]).Kind() != reflect.Map || dst[k] == nil {
				dst[k] = make(map[string]interface{})
			}
			mergeMaps(dst[k].(map[string]interface{}), v.(map[string]interface{}))
		} else {
			dst[k] = v
		}
	}
}

func getDefaultConfig() *Config {
	return &Config{
		ResourceStrategyMap: nil,
		DebugMode:           false,
	}
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
