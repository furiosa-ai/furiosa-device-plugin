package main

import (
	"fmt"
	"os"

	"github.com/furiosa-ai/furiosa-device-plugin/e2e"
	furiosaDevice "github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
)

func main() {
	test := "{\"devices\":[{\"arch\":\"Warboy\",\"dev\":\"npu0\",\"uuid\":\"A73A4486-3C40-445F-BDF7-3722462C87C8\",\"cores\":[0,1],\"device_files\":[{\"path\":\"/dev/npu0\",\"filename\":\"npu0\",\"device_index\":0,\"core_range\":{\"type\":\"CoreRangeTypeAll\",\"start\":0,\"end\":0},\"mode\":\"DeviceModeMultiCore\"},{\"path\":\"/dev/npu0pe0\",\"filename\":\"npu0pe0\",\"device_index\":0,\"core_range\":{\"type\":\"CoreRangeTypeRange\",\"start\":0,\"end\":0},\"mode\":\"DeviceModeSingle\"},{\"path\":\"/dev/npu0pe1\",\"filename\":\"npu0pe1\",\"device_index\":0,\"core_range\":{\"type\":\"CoreRangeTypeRange\",\"start\":1,\"end\":1},\"mode\":\"DeviceModeSingle\"},{\"path\":\"/dev/npu0pe0-1\",\"filename\":\"npu0pe0-1\",\"device_index\":0,\"core_range\":{\"type\":\"CoreRangeTypeRange\",\"start\":0,\"end\":1},\"mode\":\"DeviceModeFusion\"}]},{\"arch\":\"Warboy\",\"dev\":\"npu1\",\"uuid\":\"A73A5D5B-019E-4C1E-98FE-D4D1223117C8\",\"cores\":[0,1],\"device_files\":[{\"path\":\"/dev/npu1\",\"filename\":\"npu1\",\"device_index\":1,\"core_range\":{\"type\":\"CoreRangeTypeAll\",\"start\":0,\"end\":0},\"mode\":\"DeviceModeMultiCore\"},{\"path\":\"/dev/npu1pe0\",\"filename\":\"npu1pe0\",\"device_index\":1,\"core_range\":{\"type\":\"CoreRangeTypeRange\",\"start\":0,\"end\":0},\"mode\":\"DeviceModeSingle\"},{\"path\":\"/dev/npu1pe1\",\"filename\":\"npu1pe1\",\"device_index\":1,\"core_range\":{\"type\":\"CoreRangeTypeRange\",\"start\":1,\"end\":1},\"mode\":\"DeviceModeSingle\"},{\"path\":\"/dev/npu1pe0-1\",\"filename\":\"npu1pe0-1\",\"device_index\":1,\"core_range\":{\"type\":\"CoreRangeTypeRange\",\"start\":0,\"end\":1},\"mode\":\"DeviceModeFusion\"}]},{\"arch\":\"Warboy\",\"dev\":\"npu2\",\"uuid\":\"A73A4486-E968-41F2-8153-FD3A434087C8\",\"cores\":[0,1],\"device_files\":[{\"path\":\"/dev/npu2\",\"filename\":\"npu2\",\"device_index\":2,\"core_range\":{\"type\":\"CoreRangeTypeAll\",\"start\":0,\"end\":0},\"mode\":\"DeviceModeMultiCore\"},{\"path\":\"/dev/npu2pe0\",\"filename\":\"npu2pe0\",\"device_index\":2,\"core_range\":{\"type\":\"CoreRangeTypeRange\",\"start\":0,\"end\":0},\"mode\":\"DeviceModeSingle\"},{\"path\":\"/dev/npu2pe1\",\"filename\":\"npu2pe1\",\"device_index\":2,\"core_range\":{\"type\":\"CoreRangeTypeRange\",\"start\":1,\"end\":1},\"mode\":\"DeviceModeSingle\"},{\"path\":\"/dev/npu2pe0-1\",\"filename\":\"npu2pe0-1\",\"device_index\":2,\"core_range\":{\"type\":\"CoreRangeTypeRange\",\"start\":0,\"end\":1},\"mode\":\"DeviceModeFusion\"}]},{\"arch\":\"Warboy\",\"dev\":\"npu3\",\"uuid\":\"A73A5D5B-50A4-4CAD-B642-D853221C57C8\",\"cores\":[0,1],\"device_files\":[{\"path\":\"/dev/npu3\",\"filename\":\"npu3\",\"device_index\":3,\"core_range\":{\"type\":\"CoreRangeTypeAll\",\"start\":0,\"end\":0},\"mode\":\"DeviceModeMultiCore\"},{\"path\":\"/dev/npu3pe0\",\"filename\":\"npu3pe0\",\"device_index\":3,\"core_range\":{\"type\":\"CoreRangeTypeRange\",\"start\":0,\"end\":0},\"mode\":\"DeviceModeSingle\"},{\"path\":\"/dev/npu3pe1\",\"filename\":\"npu3pe1\",\"device_index\":3,\"core_range\":{\"type\":\"CoreRangeTypeRange\",\"start\":1,\"end\":1},\"mode\":\"DeviceModeSingle\"},{\"path\":\"/dev/npu3pe0-1\",\"filename\":\"npu3pe0-1\",\"device_index\":3,\"core_range\":{\"type\":\"CoreRangeTypeRange\",\"start\":0,\"end\":1},\"mode\":\"DeviceModeFusion\"}]}]}"
	unmarshaled, err := e2e.UnmarshalDevice([]byte(test))
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("%v\n\n\n", unmarshaled)

	devices, err := furiosaDevice.NewDeviceLister().ListDevices()
	if err != nil {
		os.Exit(1)
	}

	marshalDevices, err := e2e.MarshalDevices(devices)
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("%s", marshalDevices)

	// infinity loop
	select {}
}
