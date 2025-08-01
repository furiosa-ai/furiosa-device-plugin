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
	NonePolicyStr         = "none"
	Rngd2Core12GbStr      = "2core.12gb"
	Rngd4Core24GbStr      = "4core.24gb"
)

// Config holds the configuration for running this device plugin.
type Config struct {
	Partitioning        string              `yaml:"partitioning"`
	DebugMode           bool                `yaml:"debugMode"`
	DisabledDeviceUUIDs map[string][]string `yaml:"disabledDeviceUUIDs"`
}

func getDefaultConfig() *Config {
	return &Config{
		Partitioning:        NonePolicyStr,
		DebugMode:           false,
		DisabledDeviceUUIDs: nil,
	}
}

type ConfigChangeEvent struct {
	IsError  bool
	Filename string
	Detail   string
}

// LoadConfigOrGetDefault attempts to load config from GlobalConfigMountPath if the file exists,
// otherwise it returns the default config.
func LoadConfigOrGetDefault() (*Config, chan *ConfigChangeEvent, error) {
	confUpdateChan := make(chan *ConfigChangeEvent, 1)
	if _, statErr := os.Stat(GlobalConfigMountPath); statErr == nil {
		conf, err := getConfigWithWatcher(GlobalConfigMountPath, confUpdateChan)
		return conf, confUpdateChan, err
	}

	return getDefaultConfig(), confUpdateChan, nil
}

func getConfigWithWatcher(configPath string, confUpdateChan chan *ConfigChangeEvent) (*Config, error) {
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

func validatePartitioningPolicy(policy string) error {
	switch policy {
	case NonePolicyStr, Rngd2Core12GbStr, Rngd4Core24GbStr:
		return nil
	}

	return fmt.Errorf("invalid partitioning policy: %s", policy)
}

func getConfigFromFile(configPath string) (*Config, error) {
	config, err := validateConfigYaml(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to validate config file: %w", err)
	}

	err = validatePartitioningPolicy(config.Partitioning)
	if err != nil {
		return nil, fmt.Errorf("failed to validate partitioning policy: %w", err)
	}

	return config, nil
}

func validateConfigYaml(configFilePath string) (*Config, error) {
	configYaml := getDefaultConfig()
	file, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)
	err = decoder.Decode(configYaml)
	if err != nil {
		return nil, err
	}
	return configYaml, nil
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
