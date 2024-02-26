package device_manager

import (
	"fmt"
	"maps"
	"reflect"
	"strings"

	devicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/npu_allocator"
)

type DeviceManager interface {
	ResourceName() string
	Devices() []string
	HealthCheck() error
	Contains(deviceIDs []string) bool
	GetContainerPreferredAllocationResponse(available []string, required []string, request int) (*devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse, error)
	GetContainerAllocateResponse(deviceIDs []string) (*devicePluginAPIv1Beta1.ContainerAllocateResponse, error)
}

type newDeviceFunc func(originDevice device.Device) (FuriosaDevice, error)

var _ DeviceManager = (*deviceManager)(nil)

type deviceManager struct {
	origin         []device.Device
	furiosaDevices map[string]FuriosaDevice
	resourceName   string
	debugMode      bool
	allocator      npu_allocator.NpuAllocator
}

func (d *deviceManager) Devices() (ret []string) {
	for id := range d.furiosaDevices {
		ret = append(ret, id)
	}

	return ret
}

func (d *deviceManager) HealthCheck() error {
	for _, dev := range d.furiosaDevices {
		healthy, err := dev.IsHealthy()
		if err != nil {
			return err
		}

		if !healthy {
			return fmt.Errorf("device is not healthy")
		}
	}

	return nil
}

func (d *deviceManager) Contains(deviceIDs []string) bool {
	if len(deviceIDs) == 0 {
		return false
	}

	for _, id := range deviceIDs {
		if _, ok := d.furiosaDevices[id]; !ok {
			return false
		}
	}

	return true
}

// isFuriosaDevice checked whether given type T implement FuriosaDevice interface or sub interfaces
func isFuriosaDevice[T any]() bool {
	typeT := reflect.TypeOf((*T)(nil)).Elem()

	allowed := []reflect.Type{
		reflect.TypeOf((*FuriosaDevice)(nil)).Elem(),
		reflect.TypeOf((*DeviceInfo)(nil)).Elem(),
		reflect.TypeOf((*Manifest)(nil)).Elem(),
		reflect.TypeOf((*npu_allocator.Device)(nil)).Elem(),
	}

	for _, target := range allowed {
		if typeT.Implements(target) {
			return true
		}
	}

	return false
}

// NOTE: type T should be FuriosaDevice itself or an interface that FuriosaDevice implements.
// For example, DeviceInfo, Manifest, npu_allocator.Device
func fetchByID[T any](furiosaDevices map[string]FuriosaDevice, IDs []string) ([]T, error) {
	// if type T is an empty interface, we don't need to go further.
	typeName := reflect.TypeOf((*T)(nil)).Elem().Name()
	if !isFuriosaDevice[T]() {
		return nil, fmt.Errorf("the given type %s does not implement FuriosaDevice interface", typeName)
	}

	var found []T
	var missing []string
	for _, id := range IDs {
		if furiosaDevice, exist := furiosaDevices[id]; exist {
			t, ok := furiosaDevice.(T)
			if !ok {
				return nil, fmt.Errorf("couldn't convert furiosaDevice to %s", typeName)
			}
			found = append(found, t)
		} else {
			missing = append(missing, id)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("couldn't found device(s) for device id(s) %s", strings.Join(missing, ", "))
	}

	return found, nil
}

func (d *deviceManager) GetContainerPreferredAllocationResponse(available []string, required []string, request int) (*devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse, error) {
	availableDevices, err := fetchByID[npu_allocator.Device](d.furiosaDevices, available)
	if err != nil {
		return nil, err
	}

	requiredDevices, err := fetchByID[npu_allocator.Device](d.furiosaDevices, required)
	if err != nil {
		return nil, err
	}

	var allocated []string
	allocatedDeviceSet := d.allocator.Allocate(availableDevices, requiredDevices, request)
	for _, allocatedDevice := range allocatedDeviceSet {
		allocated = append(allocated, allocatedDevice.ID())
	}

	return &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
		DeviceIDs: allocated,
	}, nil
}

func (d *deviceManager) GetContainerAllocateResponse(deviceIDs []string) (*devicePluginAPIv1Beta1.ContainerAllocateResponse, error) {
	deviceRequests, err := fetchByID[FuriosaDevice](d.furiosaDevices, deviceIDs)
	if err != nil {
		return nil, err
	}

	// TODO(@bg): filter devices marked disabled in configuration and return error if request contains one of them

	resp := &devicePluginAPIv1Beta1.ContainerAllocateResponse{}
	for _, deviceRequest := range deviceRequests {
		maps.Copy(resp.Envs, deviceRequest.EnvVars())
		resp.Mounts = append(resp.Mounts, deviceRequest.Mounts()...)
		resp.Devices = append(resp.Devices, deviceRequest.DeviceSpecs()...)
		maps.Copy(resp.Annotations, deviceRequest.Annotations())
		//TODO(@bg): support CDI
	}

	return resp, nil
}

func (d *deviceManager) ResourceName() string {
	return d.resourceName
}

func newDeviceFuncResolver(strategy config.ResourceUnitStrategy) (ret newDeviceFunc) {
	// Note: config validation ensure that there is no exception other than listed strategies.
	switch strategy {
	case config.LegacyStrategy, config.GenericStrategy:
		ret = NewFullDevice
	case config.SingleCoreStrategy, config.DualCoreStrategy, config.QuadCoreStrategy:
		ret = NewPartialDevice
	}

	return ret
}

func buildFuriosaDevices(devices []device.Device, newDevFunc newDeviceFunc) (map[string]FuriosaDevice, error) {
	furiosaDevices := map[string]FuriosaDevice{}
	for _, origin := range devices {
		furiosaDevice, err := newDevFunc(origin)
		if err != nil {
			return nil, err
		}
		furiosaDevices[furiosaDevice.DeviceID()] = furiosaDevice

	}
	return furiosaDevices, nil
}

func NewDeviceManager(devices []device.Device, strategy config.ResourceUnitStrategy, debugMode bool) (DeviceManager, error) {
	resName, err := buildAndValidateFullResourceEndpointName(devices[0].Arch(), strategy)
	if err != nil {
		return nil, err
	}

	furiosaDevices, err := buildFuriosaDevices(devices, newDeviceFuncResolver(strategy))
	if err != nil {
		return nil, err
	}

	// NOTE(@bg): we may need to support configuration option for various allocators
	allocator, err := npu_allocator.NewScoreBasedOptimalNpuAllocator(devices)
	if err != nil {
		return nil, err
	}

	return &deviceManager{
		origin:         devices,
		furiosaDevices: furiosaDevices,
		resourceName:   resName,
		debugMode:      debugMode,
		allocator:      allocator,
	}, nil
}
