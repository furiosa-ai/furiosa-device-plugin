package device_manager

import (
	"fmt"
	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/furiosa_device"
)

func transformPartitioningConfig(policy string) (furiosa_device.PartitioningPolicy, error) {
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

func PreparePartitioningPolicy(deviceMap DeviceMap, policy string) (furiosa_device.PartitioningPolicy, error) {
	transformed, err := transformPartitioningConfig(policy)
	if err != nil {
		return furiosa_device.NonePolicy, err
	}

	if err := validatePartitioningConfig(deviceMap, transformed); err != nil {
		return furiosa_device.NonePolicy, err
	}

	return transformed, nil
}

func validatePartitioningConfig(deviceMap DeviceMap, policy furiosa_device.PartitioningPolicy) error {
	if policy == furiosa_device.NonePolicy {
		return nil
	}

	for arch := range deviceMap {
		if arch == smi.ArchWarboy {
			return fmt.Errorf("partitioning is not supported for warboy")
		}
	}

	return nil
}
