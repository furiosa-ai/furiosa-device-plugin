package device_manager

import (
	"fmt"
	"maps"
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
	Contains(deviceIDs []string) (bool, []string)
	GetListAndWatchResponse() *devicePluginAPIv1Beta1.ListAndWatchResponse
	GetContainerPreferredAllocationResponse(available []string, required []string, request int) (*devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse, error)
	GetContainerAllocateResponse(deviceIDs []string) (*devicePluginAPIv1Beta1.ContainerAllocateResponse, error)
}

type newDeviceFunc func(originDevice device.Device, isDisabled bool) ([]FuriosaDevice, error)

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
			return fmt.Errorf("device %s is not healthy", dev.DeviceID())
		}
	}

	return nil
}

func (d *deviceManager) Contains(deviceIDs []string) (bool, []string) {
	var missing []string

	if len(deviceIDs) == 0 {
		return false, nil
	}

	for _, id := range deviceIDs {
		if _, ok := d.furiosaDevices[id]; !ok {
			missing = append(missing, id)
		}
	}

	if len(missing) > 0 {
		return false, missing
	}

	return true, nil
}

func fetchByID(furiosaDevices map[string]FuriosaDevice, IDs []string) ([]FuriosaDevice, error) {
	var found []FuriosaDevice
	var missing []string
	for _, id := range IDs {
		if furiosaDevice, exist := furiosaDevices[id]; exist {
			found = append(found, furiosaDevice)
		} else {
			missing = append(missing, id)
		}
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("couldn't found device(s) for device id(s) %s", strings.Join(missing, ", "))
	}

	return found, nil
}

func fetchDevicesByID(furiosaDevices map[string]FuriosaDevice, IDs []string) ([]npu_allocator.Device, error) {
	found, err := fetchByID(furiosaDevices, IDs)
	if err != nil {
		return nil, err
	}

	var devices []npu_allocator.Device
	for _, furiosaDevice := range found {
		devices = append(devices, furiosaDevice)
	}

	return devices, nil
}

func (d *deviceManager) GetContainerPreferredAllocationResponse(available []string, required []string, request int) (*devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse, error) {
	availableDevices, err := fetchDevicesByID(d.furiosaDevices, available)
	if err != nil {
		return nil, err
	}

	requiredDevices, err := fetchDevicesByID(d.furiosaDevices, required)
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
	deviceRequests, err := fetchByID(d.furiosaDevices, deviceIDs)
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

func (d *deviceManager) GetListAndWatchResponse() *devicePluginAPIv1Beta1.ListAndWatchResponse {
	var resp []*devicePluginAPIv1Beta1.Device

	for _, dev := range d.furiosaDevices {
		var health = devicePluginAPIv1Beta1.Healthy
		isHealthy, err := dev.IsHealthy()
		if err != nil || !isHealthy {
			health = devicePluginAPIv1Beta1.Unhealthy
		}

		resp = append(resp, &devicePluginAPIv1Beta1.Device{
			ID:     dev.DeviceID(),
			Health: health,
			Topology: &devicePluginAPIv1Beta1.TopologyInfo{
				Nodes: []*devicePluginAPIv1Beta1.NUMANode{
					{
						ID: int64(dev.NUMANode()),
					},
				},
			},
		})
	}

	return &devicePluginAPIv1Beta1.ListAndWatchResponse{
		Devices: resp,
	}
}

func (d *deviceManager) ResourceName() string {
	return d.resourceName
}

func newDeviceFuncResolver(strategy config.ResourceUnitStrategy) (ret newDeviceFunc) {
	// Note: config validation ensure that there is no exception other than listed strategies.
	switch strategy {
	case config.LegacyStrategy, config.GenericStrategy:
		ret = func(originDevice device.Device, isDisabled bool) ([]FuriosaDevice, error) {
			newExclusiveDevice, err := NewExclusiveDevice(originDevice, isDisabled)
			if err != nil {
				return nil, err
			}

			return []FuriosaDevice{newExclusiveDevice}, nil
		}
	case config.SingleCoreStrategy, config.DualCoreStrategy, config.QuadCoreStrategy:
		ret = NewPartitionedDevices
	}

	return ret
}

func buildFuriosaDevices(devices []device.Device, blockedList []string, newDevFunc newDeviceFunc) (map[string]FuriosaDevice, error) {
	furiosaDevices := map[string]FuriosaDevice{}
	for _, origin := range devices {
		devUUID, err := origin.DeviceUUID()
		if err != nil {
			return nil, err
		}
		isDisabled := contains(blockedList, devUUID)
		devices, err := newDevFunc(origin, isDisabled)
		if err != nil {
			return nil, err
		}

		for _, d := range devices {
			furiosaDevices[d.DeviceID()] = d
		}
	}
	return furiosaDevices, nil
}

func NewDeviceManager(devices []device.Device, strategy config.ResourceUnitStrategy, blockedList []string, debugMode bool) (DeviceManager, error) {
	resName, err := buildAndValidateFullResourceEndpointName(devices[0].Arch(), strategy)
	if err != nil {
		return nil, err
	}

	furiosaDevices, err := buildFuriosaDevices(devices, blockedList, newDeviceFuncResolver(strategy))
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
