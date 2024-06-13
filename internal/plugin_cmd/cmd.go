package plugin_cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/furiosa-device-plugin/internal/device_manager"
	"github.com/furiosa-ai/furiosa-device-plugin/internal/server"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	devicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	cmdUse     = "furiosa-device-plugin"
	cmdShort   = "Furiosa Device Plugin for Kubernetes"
	cmdExample = "furiosa-device-plugin"
)

func NewDevicePluginCommand() *cobra.Command {
	devicePluginCmd := &cobra.Command{
		Use:     cmdUse,
		Short:   cmdShort,
		Example: cmdExample,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return start(cmd.Context())
		},
	}
	return devicePluginCmd
}

func start(ctx context.Context) error {
	// create core loop logger
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("subject", "core_loop").Logger()
	_ = logger.WithContext(ctx)

	//filesystem event listener for kubelet socket change by kubelet restart and configuration update
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// watch device-plugin path for kubelet restart
	fsErr := fsWatcher.Add(devicePluginAPIv1Beta1.DevicePluginPath)
	if fsErr != nil {
		logger.Err(fsErr).Msg(fmt.Sprintf("couldn't watch the path %s", devicePluginAPIv1Beta1.DevicePluginPath))
		return nil
	}

	//os signal listener
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	//grpc server panic listener
	grpcErrChan := make(chan error, 1)

	confUpdateChan := make(chan *config.ConfigChangeEvent, 1)
	conf, err := config.GetConfigWithWatcher(config.GlobalConfigMountPath, confUpdateChan)
	if err != nil {
		logger.Err(err).Msg("couldn't parse configuration")
		return err
	}

	defer func() {
		logger.Info().Msg("closing channels")
		_ = fsWatcher.Close()
		signal.Stop(sigChan)
		close(sigChan)
		close(grpcErrChan)
		close(confUpdateChan)
	}()

	deviceMap, err := server.BuildDeviceMap()
	if err != nil {
		logger.Err(err).Msg("couldn't build device-map with device-api")
		return err
	}

	if len(deviceMap) == 0 {
		noDeviceError := fmt.Errorf("couldn't recognize any furiosa devices")
		logger.Err(noDeviceError).Msg("no device detected")
		return noDeviceError
	}

	var pluginServers []server.PluginServer
	for arch, devices := range deviceMap {
		//FIXME(@bg): handle unknown arch case

		//get disabled Device for the current node
		nodeName := config.NewNodeNameGetter().GetNodename()
		disabledDeviceUUIDList := conf.DisabledDeviceUUIDListMap[nodeName]

		deviceManager, err := device_manager.NewDeviceManager(arch, devices, conf.ResourceStrategy, disabledDeviceUUIDList, conf.DebugMode)
		if err != nil {
			logger.Err(err).Msg(fmt.Sprintf("couldn't initialize device manager for %s arch", arch.ToString()))
			return err
		}
		logger.Info().Msg(fmt.Sprintf("starting new plugin server for %s", deviceManager.ResourceName()))

		newPluginServerCtx, newPluginServerCancelFunc := context.WithCancel(context.Background())
		newPluginServerLogger := zerolog.New(os.Stdout).With().Timestamp().Str("subject", "plugin_server_"+deviceManager.ResourceName()).Logger()
		newPluginServerCtx = newPluginServerLogger.WithContext(newPluginServerCtx)

		pluginServer := server.NewPluginServerWithContext(newPluginServerCtx, newPluginServerCancelFunc, deviceManager, conf)
		if err = startServerWithContext(newPluginServerCtx, pluginServer, grpcErrChan); err != nil {
			logger.Err(err).Msg(fmt.Sprintf("couldn't start plugin server for %s", deviceManager.ResourceName()))
			return err
		}

		pluginServers = append(pluginServers, pluginServer)
	}

	logger.Info().Msg("start event loop")

Loop:
	for {
		select {
		case fsEvent := <-fsWatcher.Events:
			// Note(@bg): the device-plugin should be re-registered to kubelet if the kubelet is restarted.
			// https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/#handling-kubelet-restarts
			if fsEvent.Name == devicePluginAPIv1Beta1.KubeletSocket && fsEvent.Has(fsnotify.Create) {
				logger.Err(err).Msg("kubelet socket is newly created, the device plugin should be restarted.")
				break Loop
			}
		case sig := <-sigChan:
			logger.Err(err).Msg(fmt.Sprintf("signal %d recevied.", sig))
			break Loop
		case grpcErr := <-grpcErrChan:
			logger.Err(grpcErr).Msg("error received from grpc server error channel")
			break Loop
		case confChangedEvent := <-confUpdateChan:
			if confChangedEvent.IsError {
				logger.Err(err).Msg(fmt.Sprintf("configuration file %s has been changed: %s, restarting device-plugin", confChangedEvent.Filename, confChangedEvent.Detail))
			} else {
				logger.Info().Msg(fmt.Sprintf("failed to watch file %s: %s, restarting device-plugin", confChangedEvent.Filename, confChangedEvent.Detail))
			}
			break Loop
		}
	}

	logger.Info().Msg("stopping pluginServers")
	for _, pluginServer := range pluginServers {
		if err := stopServer(pluginServer); err != nil {
			return err
		}
	}

	return nil
}

func startServerWithContext(ctx context.Context, server server.PluginServer, grpcErrChan chan error) error {
	return server.StartWithContext(ctx, grpcErrChan)
}
func stopServer(server server.PluginServer) error {
	return server.Stop()
}
