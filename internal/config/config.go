package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	GlobalConfigMountPath = "/etc/config/global_config.yaml"
)

const (
	configType         = "yaml"
	legacyStrategyStr  = "legacy"
	genericStrategyStr = "generic"
	singleCoreStr      = "single-core"
	dualCoreStr        = "dual-core"
	quadCoreStr        = "quad-core"
	warboyStr          = "warboy"
	renegadeStr        = "renegade"
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

// Config holds the configuration for running this device plugin.
type Config struct {
	ResourceStrategyMap    map[ResourceKind]ResourceUnitStrategy `yaml:"resourceStrategyMap" validate:"required"`
	DisabledDeviceUUIDList []string                              `yaml:"disabledDeviceUUIDList"`
	DebugMode              bool                                  `yaml:"debugMode"`
}

// ConfigYaml is the schema of the config file. This struct is used only for validation purpose.
type ConfigYaml struct {
	Global Config `yaml:"global" validate:"required"`
	// Overrides is a map of nodename to config.
	Overrides map[string]Config `yaml:"overrides"`
}

// ConfigYamlMap is used to read in the config file as a map.
// The map is used to merge the global and overrided config.
type ConfigYamlMap struct {
	Global    map[string]interface{}            `yaml:"global"`
	Overrides map[string]map[string]interface{} `yaml:"overrides"`
}

type ConfigChangeEvent struct {
	IsError  bool
	Filename string
	Detail   string
}

func GetMergedConfigWithWatcher(configPath string, nodeNameGetter NodeNameGetter, confUpdateChan chan *ConfigChangeEvent) (*Config, error) {
	err := validateConfigYaml(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config file: %w", err)
	}

	confYamlMap, err := readInConfigAsMap(configPath)
	if err != nil {
		return nil, err
	}

	globalConf := confYamlMap.Global
	nodeName := nodeNameGetter.GetNodename()
	if nodeName == "" {
		log.Warn().Msg("NODE_NAME env is not set, using global config only")
	} else {
		localConf := confYamlMap.Overrides[nodeName]
		mergeMaps(globalConf, localConf)
	}

	config, err := convertToConfig(globalConf)
	if err != nil {
		return nil, err
	}
	err = validateConfig(config)
	if err != nil {
		return nil, err
	}

	err = startFileWatch(confUpdateChan, configPath, true)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func validateConfigYaml(configFilePath string) error {
	file, err := os.Open(configFilePath)
	if err != nil {
		return err
	}
	configYaml := ConfigYaml{}
	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	err = decoder.Decode(&configYaml)
	if err != nil {
		return err
	}
	return nil
}

func readInConfigAsMap(configFilePath string) (*ConfigYamlMap, error) {
	contents, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	result := &ConfigYamlMap{}
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

func startFileWatch(confUpdateChan chan *ConfigChangeEvent, filePath string, detectSymlinkRemove bool) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = watcher.Add(filepath.Dir(filePath))
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Error().Msg("watcher.Events channel is closed, exiting watcher loop for file: " + filePath)
					return
				}
				targetOp := fsnotify.Create | fsnotify.Write | fsnotify.Remove | fsnotify.Rename
				// Since k8s configmap is mounted as a symlink, we need to detect the symlink update via Remove event
				if event.Has(fsnotify.Remove) && detectSymlinkRemove {
					log.Info().Msg("detected symlink update in the global config path")
					confUpdateChan <- &ConfigChangeEvent{IsError: false, Filename: filePath, Detail: "symlink updated"}
				} else if event.Name == filePath && event.Has(targetOp) {
					confUpdateChan <- &ConfigChangeEvent{IsError: false, Filename: filePath, Detail: event.String()}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					log.Error().Msgf("watcher.Error channel is closed, exiting watcher loop for file: %s", filePath)
					return
				}
				log.Err(err).Msgf("failed to watch config file: %s", filePath)
				confUpdateChan <- &ConfigChangeEvent{IsError: true, Filename: filePath, Detail: err.Error()}
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
		ResourceStrategyMap:    nil,
		DisabledDeviceUUIDList: nil,
		DebugMode:              false,
	}
}
