package device_manager

import (
	"fmt"
	"maps"
	"strings"

	devicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/furiosa_device"
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

var _ DeviceManager = (*deviceManager)(nil)

type deviceManager struct {
	origin         []smi.Device
	furiosaDevices map[string]furiosa_device.FuriosaDevice
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

func fetchByID(furiosaDevices map[string]furiosa_device.FuriosaDevice, IDs []string) ([]furiosa_device.FuriosaDevice, error) {
	var found []furiosa_device.FuriosaDevice
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

func fetchDevicesByID(furiosaDevices map[string]furiosa_device.FuriosaDevice, IDs []string) ([]npu_allocator.Device, error) {
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
	allocatedDeviceSet := d.allocator.Allocate(npu_allocator.NewDeviceSet(availableDevices...), npu_allocator.NewDeviceSet(requiredDevices...), request)
	for _, allocatedDevice := range allocatedDeviceSet.Devices() {
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

func NewDeviceManager(arch smi.Arch, devices []smi.Device, strategy config.ResourceUnitStrategy, blockedList []string, debugMode bool) (DeviceManager, error) {
	resName, err := buildAndValidateFullResourceEndpointName(arch, strategy)
	if err != nil {
		return nil, err
	}

	furiosaDevices, err := furiosa_device.NewFuriosaDevices(devices, blockedList, strategy.Policy())
	if err != nil {
		return nil, err
	}

	allocator, err := getNpuAllocatorByStrategy(devices, strategy)
	if err != nil {
		return nil, err
	}

	furiosaDevicesMap := map[string]furiosa_device.FuriosaDevice{}
	for _, d := range furiosaDevices {
		furiosaDevicesMap[d.DeviceID()] = d
	}

	return &deviceManager{
		origin:         devices,
		furiosaDevices: furiosaDevicesMap,
		resourceName:   resName,
		debugMode:      debugMode,
		allocator:      allocator,
	}, nil
}

func getNpuAllocatorByStrategy(devices []smi.Device, strategy config.ResourceUnitStrategy) (npu_allocator.NpuAllocator, error) {
	switch strategy {
	case config.GenericStrategy:
		return npu_allocator.NewScoreBasedOptimalNpuAllocator(devices)

	case config.SingleCoreStrategy, config.DualCoreStrategy, config.QuadCoreStrategy:
		return npu_allocator.NewBinPackingNpuAllocator(devices)

	default:
		// should not reach here!
		return nil, fmt.Errorf("unknown resource unit strategy %v", strategy)
	}
}
