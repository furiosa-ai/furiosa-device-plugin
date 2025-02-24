package main

import (
	"os"

	"github.com/furiosa-ai/furiosa-device-plugin/internal/plugin_cmd"
)

func main() {
	cli := plugin_cmd.NewDevicePluginCommand()
	err := cli.Execute()
	if err != nil {
		os.Exit(1)
	}
}
