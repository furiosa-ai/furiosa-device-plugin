package device_manager

import (
	"fmt"
	"strings"
	"sync"

	devicePluginAPIv1Beta1 "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"github.com/furiosa-ai/furiosa-cdi/pkg/furiosa_device"
	"github.com/furiosa-ai/furiosa-device-plugin/internal/npu_allocator"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

type DeviceManager interface {
	ResourceName() string
	Devices() []string
	HealthCheck() error
	Contains(deviceIDs []string) (bool, []string)
	GetListAndWatchResponse() *devicePluginAPIv1Beta1.ListAndWatchResponse
	GetContainerPreferredAllocationResponse(available []string, required []string, request int) (*devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse, error)
	GetContainerAllocateResponse(deviceIDs []string) (*devicePluginAPIv1Beta1.ContainerAllocateResponse, error)
	SyncDevices() (changed bool, err error)
}

var _ DeviceManager = (*deviceManager)(nil)

type deviceManager struct {
	sync.RWMutex

	arch           api.Arch
	origin         []api.Device
	furiosaDevices map[string]furiosa_device.FuriosaDevice
	resourceName   string
	debugMode      bool
	allocator      npu_allocator.NpuAllocator
}

// snapshot returns the current device set; SyncDevices replaces the map and
// the allocator wholesale, so callers can use them without holding the lock.
func (d *deviceManager) snapshot() (map[string]furiosa_device.FuriosaDevice, npu_allocator.NpuAllocator) {
	d.RLock()
	defer d.RUnlock()

	return d.furiosaDevices, d.allocator
}

func (d *deviceManager) Devices() (ret []string) {
	furiosaDevices, _ := d.snapshot()
	for id := range furiosaDevices {
		ret = append(ret, id)
	}

	return ret
}

func (d *deviceManager) HealthCheck() error {
	furiosaDevices, _ := d.snapshot()
	for _, dev := range furiosaDevices {
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

	furiosaDevices, _ := d.snapshot()
	for _, id := range deviceIDs {
		if _, ok := furiosaDevices[id]; !ok {
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
		devices = append(devices, npu_allocator.NewDevice(furiosaDevice))
	}

	return devices, nil
}

func (d *deviceManager) GetContainerPreferredAllocationResponse(available []string, required []string, request int) (*devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse, error) {
	furiosaDevices, allocator := d.snapshot()
	availableDevices, err := fetchDevicesByID(furiosaDevices, available)
	if err != nil {
		return nil, err
	}

	requiredDevices, err := fetchDevicesByID(furiosaDevices, required)
	if err != nil {
		return nil, err
	}

	var allocated []string
	allocatedDeviceSet := allocator.Allocate(npu_allocator.NewDeviceSet(availableDevices...), npu_allocator.NewDeviceSet(requiredDevices...), request)
	for _, allocatedDevice := range allocatedDeviceSet.Devices() {
		allocated = append(allocated, allocatedDevice.ID())
	}

	return &devicePluginAPIv1Beta1.ContainerPreferredAllocationResponse{
		DeviceIDs: allocated,
	}, nil
}

func (d *deviceManager) GetContainerAllocateResponse(deviceIDs []string) (*devicePluginAPIv1Beta1.ContainerAllocateResponse, error) {
	furiosaDevices, _ := d.snapshot()
	deviceRequests, err := fetchByID(furiosaDevices, deviceIDs)
	if err != nil {
		return nil, err
	}

	//TODO(@bg): support CDI
	resp, err := buildDeviceSpecToContainerAllocateResponse(deviceRequests...)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (d *deviceManager) GetListAndWatchResponse() *devicePluginAPIv1Beta1.ListAndWatchResponse {
	var resp []*devicePluginAPIv1Beta1.Device

	furiosaDevices, _ := d.snapshot()
	for _, dev := range furiosaDevices {
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

// SyncDevices re-enumerates the devices of the manager's arch and replaces the
// device set when it differs from the current one. An enumeration error keeps
// the current set; an empty result replaces it with an empty set.
func (d *deviceManager) SyncDevices() (changed bool, err error) {
	deviceMap, err := BuildDeviceMap()
	if err != nil {
		return false, err
	}

	d.RLock()
	origin := d.origin
	d.RUnlock()

	devices := deviceMap[d.arch]
	if sameDevices(origin, devices) {
		return false, nil
	}

	furiosaDevices, allocator, err := buildDeviceSet(devices)
	if err != nil {
		return false, err
	}

	d.Lock()
	defer d.Unlock()

	d.origin = devices
	d.furiosaDevices = furiosaDevices
	d.allocator = allocator

	return true, nil
}

// sameDevices reports whether a and b enumerate the same devices in the same
// order. Both must come from BuildDeviceMap, which sorts them. Device files are
// not compared: CDI specs are rendered on every call.
func sameDevices(a, b []api.Device) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].DeviceName() != b[i].DeviceName() ||
			a[i].UUID() != b[i].UUID() ||
			a[i].BDF().String() != b[i].BDF().String() ||
			a[i].NUMANode() != b[i].NUMANode() {
			return false
		}
	}

	return true
}

func buildDeviceSet(devices []api.Device) (map[string]furiosa_device.FuriosaDevice, npu_allocator.NpuAllocator, error) {
	furiosaDevices, err := furiosa_device.NewFuriosaDevices(devices, nil, furiosa_device.NonePolicy)
	if err != nil {
		return nil, nil, err
	}

	allocator, err := npu_allocator.NewScoreBasedOptimalNpuAllocator(devices)
	if err != nil {
		return nil, nil, err
	}

	furiosaDevicesMap := map[string]furiosa_device.FuriosaDevice{}
	for _, d := range furiosaDevices {
		furiosaDevicesMap[d.DeviceID()] = d
	}

	return furiosaDevicesMap, allocator, nil
}

func NewDeviceManager(arch api.Arch, devices []api.Device, debugMode bool) (DeviceManager, error) {
	resName, err := buildAndValidateFullResourceEndpointName(arch)
	if err != nil {
		return nil, err
	}

	furiosaDevices, allocator, err := buildDeviceSet(devices)
	if err != nil {
		return nil, err
	}

	return &deviceManager{
		arch:           arch,
		origin:         devices,
		furiosaDevices: furiosaDevices,
		resourceName:   resName,
		debugMode:      debugMode,
		allocator:      allocator,
	}, nil
}
