package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/furiosa-device-plugin/internal/device_manager"
	"github.com/rs/zerolog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"k8s.io/apimachinery/pkg/util/wait"
	devicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	pluginRegistrationV1 "k8s.io/kubelet/pkg/apis/pluginregistration/v1"
)

const (
	socketPathExp        = devicePluginAPIv1Beta1.DevicePluginPath + "%s" + ".sock"
	socketSymlinkPathExp = "/var/lib/kubelet/plugins_registry/%s.sock"
)

var _ devicePluginAPIv1Beta1.DevicePluginServer = (*PluginServer)(nil)
var _ pluginRegistrationV1.RegistrationServer = (*PluginServer)(nil)

type PluginServer struct {
	config                *config.Config
	cancelCtxFunc         context.CancelFunc
	deviceManager         device_manager.DeviceManager
	socket                string
	socketSymLink         string
	server                *grpc.Server
	deviceHealthCheckChan chan error
}

func dialWithTimeout(socket string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()

	conn, err := grpc.DialContext(ctx, socket,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		//TODO(@bg): pass dialFn for mocking if needed
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return net.Dial("unix", addr)
		}))

	ctxErr := ctx.Err()
	if errors.Is(ctxErr, context.DeadlineExceeded) {
		return nil, ctxErr
	}

	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (p *PluginServer) StartWithContext(ctx context.Context, grpcErrChan chan error) error {
	logger := zerolog.Ctx(ctx)
	devicePluginAPIv1Beta1.RegisterDevicePluginServer(p.server, p)
	pluginRegistrationV1.RegisterRegistrationServer(p.server, p)

	err := os.Remove(p.socket)
	if err != nil && !os.IsNotExist(err) {
		logger.Err(err).Msg(fmt.Sprintf("couldn't remove existing socket %s", p.socket))
		return err
	}

	err = os.Remove(p.socketSymLink)
	if err != nil && !os.IsNotExist(err) {
		logger.Err(err).Msg(fmt.Sprintf("couldn't remove existing socket symbolic link %s", p.socketSymLink))
		return err
	}

	// listen unix socket
	sock, err := net.Listen("unix", p.socket)
	if err != nil {
		logger.Err(err).Msg(fmt.Sprintf("couldn't listen socket %s", p.socket))
		return err
	}

	// run grpc server.serve in a new goroutine
	go func() {
		logger.Info().Msg(fmt.Sprintf("start listening %s", sock))
		if serveRrr := p.server.Serve(sock); serveRrr != nil {
			logger.Err(serveRrr).Msg("error received from grpc serving framework, the device-plugin will be restarted for recovery")
			grpcErrChan <- serveRrr
		}
	}()

	// create symbolic link from unix socket
	if err := os.Symlink(p.socket, p.socketSymLink); err != nil {
		logger.Err(err).Msg(fmt.Sprintf("couldn't create symbolic link %s", p.socketSymLink))
		return err
	}

	// check server liveliness
	conn, err := dialWithTimeout(p.socket, 5*time.Second)
	if err != nil {
		logger.Err(err).Msg(fmt.Sprintf("error received from %s dialer", p.socket))
		return err
	}
	_ = conn.Close()

	// register to kubelet
	conn, err = dialWithTimeout(devicePluginAPIv1Beta1.KubeletSocket, 5*time.Second)
	if err != nil {
		logger.Err(err).Msg(fmt.Sprintf("error received from %s dialer", devicePluginAPIv1Beta1.KubeletSocket))
		return err
	}

	if _, err = devicePluginAPIv1Beta1.NewRegistrationClient(conn).Register(ctx, &devicePluginAPIv1Beta1.RegisterRequest{
		Version:      devicePluginAPIv1Beta1.Version,
		Endpoint:     path.Base(p.socket),
		ResourceName: p.deviceManager.ResourceName(),
		Options: &devicePluginAPIv1Beta1.DevicePluginOptions{
			PreStartRequired:                false,
			GetPreferredAllocationAvailable: true,
		},
	}); err != nil {
		logger.Err(err).Msg(fmt.Sprintf("couldn't register resource %s", p.deviceManager.ResourceName()))
		return err
	}

	logger.Info().Msg(fmt.Sprintf("resource %s is registered to kubelet", p.deviceManager.ResourceName()))

	_ = conn.Close()

	// start health check loop
	logger.Info().Msg(fmt.Sprintf("start health check loop for the resource %s", p.deviceManager.ResourceName()))
	//TODO(@bg): parse duration from configuration
	go wait.UntilWithContext(ctx, func(ctx context.Context) {
		healthCheckLogger := zerolog.Ctx(ctx)
		if healthCheckErr := p.deviceManager.HealthCheck(); healthCheckErr != nil {
			healthCheckLogger.Err(healthCheckErr).Msg("device health check fail")
			// send error to health check channel, list watcher will process this event
			p.deviceHealthCheckChan <- err
		}
	}, 5*time.Second)

	return nil
}

func (p *PluginServer) Stop() error {
	// stop grpc server
	p.server.Stop()
	// remove socket
	_ = os.Remove(p.socket)
	// stop goroutines associated to plugin server
	p.cancelCtxFunc()
	return nil
}

func (p *PluginServer) GetDevicePluginOptions(_ context.Context, _ *devicePluginAPIv1Beta1.Empty) (*devicePluginAPIv1Beta1.DevicePluginOptions, error) {
	return &devicePluginAPIv1Beta1.DevicePluginOptions{
		PreStartRequired:                false,
		GetPreferredAllocationAvailable: true,
	}, nil
}

func (p *PluginServer) ListAndWatch(_ *devicePluginAPIv1Beta1.Empty, deviceMgrSrv devicePluginAPIv1Beta1.DevicePlugin_ListAndWatchServer) error {
	logger := zerolog.Ctx(deviceMgrSrv.Context())
	logger.Info().Msg(fmt.Sprintf("register devices and report initial states for devices %s", strings.Join(p.deviceManager.Devices(), ", ")))
	if err := deviceMgrSrv.Send(p.deviceManager.GetListAndWatchResponse()); err != nil {
		return err
	}

	for healthCheckErr := range p.deviceHealthCheckChan {
		logger.Info().Msg(fmt.Sprintf("device state updated %s", healthCheckErr))
		if err := deviceMgrSrv.Send(p.deviceManager.GetListAndWatchResponse()); err != nil {
			return err
		}
	}

	return nil
}

func (p *PluginServer) GetPreferredAllocation(ctx context.Context, request *devicePluginAPIv1Beta1.PreferredAllocationRequest) (*devicePluginAPIv1Beta1.PreferredAllocationResponse, error) {
	logger := zerolog.Ctx(ctx)
	var resp []*devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse

	for _, req := range request.ContainerRequests {
		logger.Info().Msg(fmt.Sprintf("received preferred allocation request, available: %s, include: %s, size: %d",
			strings.Join(req.AvailableDeviceIDs, ", "),
			strings.Join(req.MustIncludeDeviceIDs, ", "),
			req.AllocationSize))

		//FIXME(@bg): fix interfaces(Manager, Allocator) to use int32
		allocResp, err := p.deviceManager.GetContainerPreferredAllocationResponse(req.AvailableDeviceIDs, req.MustIncludeDeviceIDs, int(req.AllocationSize))
		if err != nil {
			return nil, err
		}
		resp = append(resp, allocResp)
	}

	return &devicePluginAPIv1Beta1.PreferredAllocationResponse{
		ContainerResponses: resp,
	}, nil
}

func (p *PluginServer) Allocate(ctx context.Context, request *devicePluginAPIv1Beta1.AllocateRequest) (*devicePluginAPIv1Beta1.AllocateResponse, error) {
	logger := zerolog.Ctx(ctx)
	var resp []*devicePluginAPIv1Beta1.ContainerAllocateResponse

	for _, req := range request.ContainerRequests {
		logger.Info().Msg(fmt.Sprintf("received device allocation request for device id(s) %s", strings.Join(req.DevicesIDs, ", ")))
		exist, missing := p.deviceManager.Contains(req.DevicesIDs)
		if !exist {
			return nil, fmt.Errorf("couldn't find device(s) for device id(s) %s", strings.Join(missing, ", "))
		}

		allocResp, err := p.deviceManager.GetContainerAllocateResponse(req.DevicesIDs)
		if err != nil {
			return nil, err
		}

		resp = append(resp, allocResp)
	}

	return &devicePluginAPIv1Beta1.AllocateResponse{
		ContainerResponses: resp,
	}, nil
}

func (p *PluginServer) PreStartContainer(_ context.Context, _ *devicePluginAPIv1Beta1.PreStartContainerRequest) (*devicePluginAPIv1Beta1.PreStartContainerResponse, error) {
	// NOTE(@bg): we don't need reinitialization of device.
	return &devicePluginAPIv1Beta1.PreStartContainerResponse{}, nil
}

func (p *PluginServer) GetInfo(ctx context.Context, _ *pluginRegistrationV1.InfoRequest) (*pluginRegistrationV1.PluginInfo, error) {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("plugin registration request using `plugins_registry` received")

	return &pluginRegistrationV1.PluginInfo{
		Type:              pluginRegistrationV1.DevicePlugin,
		Name:              p.deviceManager.ResourceName(),
		SupportedVersions: []string{devicePluginAPIv1Beta1.Version},
	}, nil
}

func (p *PluginServer) NotifyRegistrationStatus(ctx context.Context, registrationStatus *pluginRegistrationV1.RegistrationStatus) (*pluginRegistrationV1.RegistrationStatusResponse, error) {
	logger := zerolog.Ctx(ctx)
	if registrationStatus.PluginRegistered {
		logger.Info().Msg("successfully registered plugin using `plugins_registry`")
	} else {
		logger.Warn().Msg("failed to register plugin using `plugins_registry`")
	}

	return &pluginRegistrationV1.RegistrationStatusResponse{}, nil
}

func NewPluginServerWithContext(ctx context.Context, cancelFunc context.CancelFunc, deviceManager device_manager.DeviceManager, config *config.Config) PluginServer {
	// comment(@bg): full resource name is already validated
	split := strings.SplitN(deviceManager.ResourceName(), "/", 2)
	resNameWithoutPrefix := split[1]

	return PluginServer{
		socket:                fmt.Sprintf(socketPathExp, resNameWithoutPrefix),
		socketSymLink:         fmt.Sprintf(socketSymlinkPathExp, resNameWithoutPrefix),
		config:                config,
		cancelCtxFunc:         cancelFunc,
		deviceManager:         deviceManager,
		server:                grpc.NewServer(grpc.UnaryInterceptor(NewGrpcMiddleWareLogger(ctx))),
		deviceHealthCheckChan: nil,
	}
}
