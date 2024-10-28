package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

const (
	GlobalConfigMountPath = "/etc/config/config.yaml"
)

const (
	legacyStrategyStr  = "legacy"
	genericStrategyStr = "generic"
	singleCoreStr      = "single-core"
	dualCoreStr        = "dual-core"
	quadCoreStr        = "quad-core"
)

type ResourceUnitStrategy string

const (
	LegacyStrategy     ResourceUnitStrategy = legacyStrategyStr
	GenericStrategy    ResourceUnitStrategy = genericStrategyStr
	SingleCoreStrategy ResourceUnitStrategy = singleCoreStr
	DualCoreStrategy   ResourceUnitStrategy = dualCoreStr
	QuadCoreStrategy   ResourceUnitStrategy = quadCoreStr
)

type AllocationMode string

const (
	ScoreBased AllocationMode = "ScoreBased"
	BinPacking AllocationMode = "BinPacking"
)

// Config holds the configuration for running this device plugin.
type Config struct {
	ResourceStrategy          ResourceUnitStrategy `yaml:"resourceStrategy"`
	AllocationMode            AllocationMode       `yaml:"allocationMode"`
	DisabledDeviceUUIDListMap map[string][]string  `yaml:"disabledDeviceUUIDListMap"`
	DebugMode                 bool                 `yaml:"debugMode"`
}

func getDefaultConfig() *Config {
	return &Config{
		ResourceStrategy:          GenericStrategy,
		AllocationMode:            ScoreBased,
		DisabledDeviceUUIDListMap: nil,
		DebugMode:                 false,
	}
}

type ConfigChangeEvent struct {
	IsError  bool
	Filename string
	Detail   string
}

func GetConfigWithWatcher(configPath string, confUpdateChan chan *ConfigChangeEvent) (*Config, error) {
	conf, err := getConfigFromFile(configPath)
	if err != nil {
		return nil, err
	}
	err = startWatchingConfigChange(confUpdateChan, configPath, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func getConfigFromFile(configPath string) (*Config, error) {
	err, config := validateConfigYaml(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config file: %w", err)
	}

	if config.ResourceStrategy == "" {
		config.ResourceStrategy = GenericStrategy
	}

	if config.AllocationMode == "" {
		config.AllocationMode = ScoreBased
	}

	return config, nil
}

func validateConfigYaml(configFilePath string) (error, *Config) {
	configYaml := getDefaultConfig()
	file, err := os.Open(configFilePath)
	if err != nil {
		return err, nil
	}

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	err = decoder.Decode(configYaml)
	if err != nil {
		return err, nil
	}
	return nil, configYaml
}

func startWatchingConfigChange(confUpdateChan chan *ConfigChangeEvent, filePath string, prevConf *Config) error {
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
				// Since k8s configmap is mounted as a symlink, we need to detect the symlink update via `any remove event` in the same directory.
				maybeUpdated := event.Has(fsnotify.Remove) || (event.Has(targetOp) && event.Name == filePath)
				if !maybeUpdated {
					continue
				}
				newConf, err := getConfigFromFile(filePath)
				if err != nil {
					log.Err(err).Msgf("failed to read updated config file: %s", filePath)
					confUpdateChan <- &ConfigChangeEvent{IsError: true, Filename: filePath, Detail: err.Error()}
					continue
				}
				if !isEqualConfig(newConf, prevConf) {
					confUpdateChan <- &ConfigChangeEvent{IsError: false, Filename: filePath, Detail: "config is updated"}
				} else {
					log.Info().Msg("config file has been updated but no config change is detected")
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

func isEqualConfig(c1, c2 *Config) bool {
	return reflect.DeepEqual(c1, c2)
}
