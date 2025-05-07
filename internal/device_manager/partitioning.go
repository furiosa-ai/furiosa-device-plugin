package device_manager

import (
	"fmt"
	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/furiosa_device"
)

func TransformPartitioningConfig(policy string) (furiosa_device.PartitioningPolicy, error) {
	switch policy {
	case config.NonePolicyStr:
		return furiosa_device.NonePolicy, nil
	case config.Rngd2Core12GbStr:
		return furiosa_device.DualCorePolicy, nil
	case config.Rngd4Core24GbStr:
		return furiosa_device.QuadCorePolicy, nil
	default:
		return furiosa_device.NonePolicy, fmt.Errorf("invalid partitioning policy: %s", policy)
	}
}
