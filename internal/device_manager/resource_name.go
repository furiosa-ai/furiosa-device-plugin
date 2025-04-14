package device_manager

import (
	"fmt"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/furiosa_device"
	"strings"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"k8s.io/apimachinery/pkg/api/validation"
)

const (
	defaultDomain = "furiosa.ai"

	coreUnitTagExp    = "%dcore"
	memoryUnitTagExp  = "%dgb"
	taggedResourceExp = "%s-%s.%s"
	fullResourceExp   = "%s/%s"

	rngdMaxMemory = 48
	rngdMaxCores  = 8

	singleCore = 1
	dualCore   = 2
	quadCore   = 4
)

func coreUnitValidator(min, max, core int) error {
	if core < min || core > max {
		return fmt.Errorf("wrong core unit %d", core)
	}

	if core != 1 && core%2 != 0 {
		return fmt.Errorf("wrong core unit %d", core)
	}

	return nil
}

func validateCoreUnit(arch smi.Arch, coreUnit int) error {
	switch arch {
	case smi.ArchWarboy:
		if err := coreUnitValidator(1, 2, coreUnit); err != nil {
			return err
		}
	case smi.ArchRngd:
		if err := coreUnitValidator(1, 4, coreUnit); err != nil {
			return err
		}
	default:
		return fmt.Errorf("wrong architecture")
	}

	return nil
}

func buildResourceEndpointCoreUnitTag(coreUnit int) string {
	return fmt.Sprintf(coreUnitTagExp, coreUnit)
}

func buildResourceEndpointMemoryUnitTag(arch smi.Arch, coreUnit int) string {
	return fmt.Sprintf(memoryUnitTagExp, coreUnit*(rngdMaxMemory/rngdMaxCores))
}

func buildResourceEndpointName(arch smi.Arch) string {
	return strings.ToLower(arch.ToString())
}

func policyToCoreUnit(policy furiosa_device.PartitioningPolicy) int {
	if policy == furiosa_device.SingleCorePolicy {
		return singleCore
	} else if policy == furiosa_device.DualCorePolicy {
		return dualCore
	}

	return quadCore
}

func buildFullEndpoint(arch smi.Arch, policy furiosa_device.PartitioningPolicy) (string, error) {
	endpointName := buildResourceEndpointName(arch)

	if policy == furiosa_device.NonePolicy {
		return endpointName, nil
	}

	coreUnit := policyToCoreUnit(policy)
	if err := validateCoreUnit(arch, coreUnit); err != nil {
		return "", err
	}

	coreUnitTag := buildResourceEndpointCoreUnitTag(coreUnit)
	memoryTag := buildResourceEndpointMemoryUnitTag(arch, coreUnit)

	return fmt.Sprintf(taggedResourceExp, endpointName, coreUnitTag, memoryTag), nil
}

func buildAndValidateFullResourceEndpointName(arch smi.Arch, policy furiosa_device.PartitioningPolicy) (string, error) {
	fullEndpoint, err := buildFullEndpoint(arch, policy)
	if err != nil {
		return "", err
	}

	errs := validation.NameIsDNSSubdomain(fullEndpoint, false)
	if len(errs) != 0 {
		return "", fmt.Errorf("resource name %s is not valid %v", fullEndpoint, errs)
	}

	return fmt.Sprintf(fullResourceExp, defaultDomain, fullEndpoint), nil
}
