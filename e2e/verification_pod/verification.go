package main

import (
	"fmt"
	"os"

	"github.com/furiosa-ai/furiosa-device-plugin/e2e"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/smi"
)

func main() {
	devices, err := smi.GetDevices()
	if err != nil {
		os.Exit(1)
	}

	marshalDevices, err := e2e.MarshalDevices(devices)
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("%s", marshalDevices)
}
