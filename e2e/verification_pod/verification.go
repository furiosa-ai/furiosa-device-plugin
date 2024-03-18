package main

import (
	"fmt"
	"os"

	"github.com/furiosa-ai/furiosa-device-plugin/e2e"
	furiosaDevice "github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"
)

func main() {
	devices, err := furiosaDevice.NewDeviceLister().ListDevices()
	if err != nil {
		os.Exit(1)
	}

	marshalDevices, err := e2e.MarshalDevices(devices)
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("%s", marshalDevices)
}
