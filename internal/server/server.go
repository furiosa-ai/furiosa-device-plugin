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
	DevicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const socketPathExp = DevicePluginAPIv1Beta1.DevicePluginPath + "%s" + ".sock"

var _ DevicePluginAPIv1Beta1.DevicePluginServer = (*PluginServer)(nil)

type PluginServer struct {
	config                *config.Config
	cancelCtxFunc         context.CancelFunc
	deviceManager         device_manager.DeviceManager
	socket                string
	server                *grpc.Server
	deviceHealthCheckChan chan error
	//stop
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
	DevicePluginAPIv1Beta1.RegisterDevicePluginServer(p.server, p)

	err := os.Remove(p.socket)
	if err != nil && !os.IsNotExist(err) {
		logger.Err(err).Msg(fmt.Sprintf("couldn't remove existing socket %s", p.socket))
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

	// check server liveliness
	conn, err := dialWithTimeout(p.socket, 5*time.Second)
	if err != nil {
		logger.Err(err).Msg(fmt.Sprintf("error received from %s dialer", p.socket))
		return err
	}
	_ = conn.Close()

	// register to kubelet
	conn, err = dialWithTimeout(DevicePluginAPIv1Beta1.KubeletSocket, 5*time.Second)
	if err != nil {
		logger.Err(err).Msg(fmt.Sprintf("error received from %s dialer", DevicePluginAPIv1Beta1.KubeletSocket))
		return err
	}

	if _, err = DevicePluginAPIv1Beta1.NewRegistrationClient(conn).Register(ctx, &DevicePluginAPIv1Beta1.RegisterRequest{
		Version:      DevicePluginAPIv1Beta1.Version,
		Endpoint:     path.Base(p.socket),
		ResourceName: p.deviceManager.ResourceName(),
		Options: &DevicePluginAPIv1Beta1.DevicePluginOptions{
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

func (p *PluginServer) GetDevicePluginOptions(_ context.Context, empty *DevicePluginAPIv1Beta1.Empty) (*DevicePluginAPIv1Beta1.DevicePluginOptions, error) {
	return &DevicePluginAPIv1Beta1.DevicePluginOptions{
		PreStartRequired:                false,
		GetPreferredAllocationAvailable: true,
	}, nil
}

func (p *PluginServer) ListAndWatch(empty *DevicePluginAPIv1Beta1.Empty, server DevicePluginAPIv1Beta1.DevicePlugin_ListAndWatchServer) error {
	//logger := zerolog.Ctx(ctx)
	//TODO(@bg): handle health check error here
	//TODO implement me
	panic("implement me")
}

func (p *PluginServer) GetPreferredAllocation(ctx context.Context, request *DevicePluginAPIv1Beta1.PreferredAllocationRequest) (*DevicePluginAPIv1Beta1.PreferredAllocationResponse, error) {
	//logger := zerolog.Ctx(ctx)
	//TODO implement me
	panic("implement me")
}

func (p *PluginServer) Allocate(ctx context.Context, request *DevicePluginAPIv1Beta1.AllocateRequest) (*DevicePluginAPIv1Beta1.AllocateResponse, error) {
	//logger := zerolog.Ctx(ctx)
	//TODO implement me
	panic("implement me")
}

func (p *PluginServer) PreStartContainer(ctx context.Context, request *DevicePluginAPIv1Beta1.PreStartContainerRequest) (*DevicePluginAPIv1Beta1.PreStartContainerResponse, error) {
	//logger := zerolog.Ctx(ctx)
	//TODO implement me
	panic("implement me")
}

func NewPluginServerWithContext(ctx context.Context, cancelFunc context.CancelFunc, deviceManager device_manager.DeviceManager, config *config.Config) PluginServer {
	// comment(@bg): full resource name is already validated
	split := strings.SplitN(deviceManager.ResourceName(), "/", 2)
	resNameWithoutPrefix := split[1]

	return PluginServer{
		socket:        fmt.Sprintf(socketPathExp, resNameWithoutPrefix),
		config:        config,
		cancelCtxFunc: cancelFunc,
		deviceManager: deviceManager,
		server:        grpc.NewServer(grpc.UnaryInterceptor(NewGrpcMiddleWareLogger(ctx))),
	}
}
