package device_manager

import (
	"fmt"
	"regexp"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
)

const (
	bdfPattern   = `^(?P<domain>[0-9a-fA-F]{1,4}):(?P<bus>[0-9a-fA-F]+):(?P<function>[0-9a-fA-F]+\.[0-9])$`
	subExpKeyBus = "bus"
)

var (
	bdfRegExp = regexp.MustCompile(bdfPattern)
)

func parseBusIDfromBDF(bdf string) (string, error) {
	matches := bdfRegExp.FindStringSubmatch(bdf)
	if matches == nil {
		return "", fmt.Errorf("couldn't parse the given string %s with bdf regex pattern: %s", bdf, bdfPattern)
	}

	subExpIndex := bdfRegExp.SubexpIndex(subExpKeyBus)
	if subExpIndex == -1 {
		return "", fmt.Errorf("couldn't parse bus id from the given bdf expression %s", bdf)
	}

	return matches[subExpIndex], nil
}

// parseOriginDeviceInfo returns below information.
//   - arch: ArchWarboy, ArchRngd, ArchRngdMax, ArchRngdS
//   - uuid: UUID string of origin device (board)
//   - pciBusID
//   - numaNode
func parseOriginDeviceInfo(originDevice smi.Device) (arch smi.Arch, uuid, pciBusID string, numaNode uint, err error) {
	info, err := originDevice.DeviceInfo()
	if err != nil {
		return 0, "", "", 0, err
	}

	arch = info.Arch()
	uuid = info.UUID()
	pciBusID, err = parseBusIDfromBDF(info.BDF())
	numaNode = uint(info.NumaNode())

	return arch, uuid, pciBusID, numaNode, err
}

func contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
