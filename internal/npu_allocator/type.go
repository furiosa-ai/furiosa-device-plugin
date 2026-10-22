package npu_allocator

import (
	"github.com/furiosa-ai/furiosa-cdi/pkg/furiosa_device"
	cdiutil "github.com/furiosa-ai/furiosa-cdi/pkg/util"
	"github.com/furiosa-ai/furiosa-device-plugin/internal/util"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
)

type NpuAllocator interface {
	Allocate(available DeviceSet, required DeviceSet, size int) DeviceSet
}

type Device interface {
	// Index returns an index number of Device for sorting purpose.
	// Index must be injected from `furiosa-device-plugin`, and should not be modified by `npu_allocator`.
	Index() int

	// ID
	// ID returns a unique ID of Device to identify the device.
	ID() string

	// TopologyHintKey returns unique key to retrieve TopologyHint using TopologyHintProvider.
	TopologyHintKey() TopologyHintKey

	// Equal check whether source Device is identical to the target Device.
	Equal(target Device) bool
}

func NewDevice(origin furiosa_device.FuriosaDevice) Device {
	return &device{origin: origin}
}

type device struct {
	origin furiosa_device.FuriosaDevice
}

func (d *device) Index() int {
	return d.origin.Index()
}

func (d *device) ID() string {
	return d.origin.DeviceID()
}

func (d *device) TopologyHintKey() TopologyHintKey {
	return TopologyHintKey(d.origin.PCIBusID())
}

func (d *device) Equal(target Device) bool {
	return d.ID() == target.ID()
}

type DeviceSet interface {
	Contains(target ...Device) bool
	Equal(target ...Device) bool
	Difference(target ...Device) DeviceSet
	Union(target ...Device) DeviceSet
	Insert(target ...Device)
	Devices() []Device
	Len() int
}

type deviceSet struct {
	btreeSet *util.BtreeSet[Device]
}

func NewDeviceSet(devices ...Device) DeviceSet {
	btreeSet := util.NewBtreeSetWithLessFunc(len(devices), func(a, b Device) bool {
		idx1, idx2 := a.Index(), b.Index()
		if idx1 == idx2 {
			id1, id2 := a.ID(), b.ID()
			return id1 < id2
		}

		return idx1 < idx2
	})

	for _, device := range devices {
		btreeSet.Insert(device)
	}

	return &deviceSet{btreeSet: btreeSet}
}

// Contains checks whether source DeviceSet contains target DeviceSet.
func (source *deviceSet) Contains(target ...Device) bool {
	if source == nil || source.btreeSet == nil {
		return false
	}

	if source.Len() == 0 || len(target) == 0 {
		return false
	}

	for _, targetDevice := range target {
		if !source.btreeSet.Has(targetDevice) {
			return false
		}
	}

	return true
}

// Equal check whether source DeviceSet and target DeviceSet is identical regardless of element order.
func (source *deviceSet) Equal(target ...Device) bool {
	if source == nil || source.btreeSet == nil {
		return false
	}

	if source.Len() == 0 && len(target) == 0 {
		return true
	}

	if source.Len() != len(target) {
		return false
	}

	for _, targetDevice := range target {
		if !source.btreeSet.Has(targetDevice) {
			return false
		}
	}

	return true
}

// Difference returns a subset of the source DeviceSet that has no intersection with the target DeviceSet.
func (source *deviceSet) Difference(target ...Device) DeviceSet {
	if source == nil || source.btreeSet == nil {
		return nil
	}

	targetDeviceSet := NewDeviceSet(target...)

	difference := NewDeviceSet()
	for _, sourceDevice := range source.Devices() {
		if !targetDeviceSet.Contains(sourceDevice) {
			difference.Insert(sourceDevice)
		}
	}

	return difference
}

// Union returns new DeviceSet containing elements of source and target DeviceSets
func (source *deviceSet) Union(target ...Device) DeviceSet {
	if source == nil || source.btreeSet == nil {
		return nil
	}

	ds := NewDeviceSet(source.Devices()...)
	for _, targetDevice := range target {
		ds.Insert(targetDevice)
	}

	return ds
}

func (source *deviceSet) Insert(target ...Device) {
	if source == nil || source.btreeSet == nil {
		return
	}

	for _, device := range target {
		source.btreeSet.Insert(device)
	}
}

func (source *deviceSet) Devices() []Device {
	if source == nil || source.btreeSet == nil {
		return nil
	}

	return source.btreeSet.Items()
}

func (source *deviceSet) Len() int {
	if source == nil || source.btreeSet == nil {
		return 0
	}

	return source.btreeSet.Len()
}

// TODO(@bg): impl Intersection()

// TopologyHintProvider takes two devices as argument return topology hint.
// The hint would be score, distance, preference of two devices.
type TopologyHintProvider func(device1, device2 Device) uint

// TopologyHintKey is named type of string, used for TopologyHintMatrix
type TopologyHintKey string

// TopologyHintMatrix provides score of device to device based on api.Device api.LinkType.
type TopologyHintMatrix map[TopologyHintKey]map[TopologyHintKey]uint

// NewTopologyHintMatrix generates TopologyHintMatrix using list of api.Device.
func NewTopologyHintMatrix(devices []api.Device) (TopologyHintMatrix, error) {
	topologyHintMatrix := make(TopologyHintMatrix)

	for i := range devices {
		for j := range devices {
			linkType, err := devices[i].DeviceToDeviceLinkType(&devices[j])
			if err != nil {
				return nil, err
			}

			score := linkTypeScore(linkType)

			pciBusID1, err := cdiutil.ParseBusIDFromBDF(devices[i].BDF().String())
			if err != nil {
				return nil, err
			}

			pciBusID2, err := cdiutil.ParseBusIDFromBDF(devices[j].BDF().String())
			if err != nil {
				return nil, err
			}

			key1, key2 := TopologyHintKey(pciBusID1), TopologyHintKey(pciBusID2)
			if key1 > key2 {
				key1, key2 = key2, key1
			}

			if _, ok := topologyHintMatrix[key1]; !ok {
				topologyHintMatrix[key1] = make(map[TopologyHintKey]uint)
			}

			topologyHintMatrix[key1][key2] = score
		}
	}

	return topologyHintMatrix, nil
}
