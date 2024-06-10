package device_manager

import (
	"fmt"
	"strings"

	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/smi"
	"k8s.io/apimachinery/pkg/api/validation"
)

const (
	defaultDomain = "furiosa.ai"
	legacyDomain  = "alpha.furiosa.ai"

	coreUnitTagExp     = "%dcore"
	memoryUnitTagExp   = "%dgb"
	taggedResourceExp  = "%s-%s.%s"
	fullResourceExp    = "%s/%s"
	legacyResourceName = "npu"

	warboyMaxMemory = 16
	warboyMaxCores  = 2
	rngdMaxMemory   = 48
	rngdMaxCores    = 8

	singleCore = 1
	dualCore   = 2
	quadCore   = 4
)

func buildDomainName(strategy config.ResourceUnitStrategy) string {
	if strategy == config.LegacyStrategy {
		return legacyDomain
	}

	return defaultDomain
}

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
	if arch == smi.ArchWarboy {
		return fmt.Sprintf(memoryUnitTagExp, coreUnit*(warboyMaxMemory/warboyMaxCores))
	}

	return fmt.Sprintf(memoryUnitTagExp, coreUnit*(rngdMaxMemory/rngdMaxCores))
}

func buildResourceEndpointName(arch smi.Arch, strategy config.ResourceUnitStrategy) string {
	if strategy == config.LegacyStrategy {
		return legacyResourceName
	}
	return strings.ToLower(arch.ToString())
}

func strategyToCoreUnit(strategy config.ResourceUnitStrategy) int {
	if strategy == config.SingleCoreStrategy {
		return singleCore
	} else if strategy == config.DualCoreStrategy {
		return dualCore
	}

	return quadCore
}

func buildFullEndpoint(arch smi.Arch, strategy config.ResourceUnitStrategy) (string, error) {
	endpointName := buildResourceEndpointName(arch, strategy)

	if strategy == config.LegacyStrategy || strategy == config.GenericStrategy {
		return endpointName, nil
	}

	coreUnit := strategyToCoreUnit(strategy)
	if err := validateCoreUnit(arch, coreUnit); err != nil {
		return "", err
	}

	coreUnitTag := buildResourceEndpointCoreUnitTag(coreUnit)
	memoryTag := buildResourceEndpointMemoryUnitTag(arch, coreUnit)

	return fmt.Sprintf(taggedResourceExp, endpointName, coreUnitTag, memoryTag), nil
}

func buildAndValidateFullResourceEndpointName(arch smi.Arch, strategy config.ResourceUnitStrategy) (string, error) {
	domainName := buildDomainName(strategy)

	fullEndpoint, err := buildFullEndpoint(arch, strategy)
	if err != nil {
		return "", err
	}

	errs := validation.NameIsDNSSubdomain(fullEndpoint, false)
	if len(errs) != 0 {
		return "", fmt.Errorf("resource name %s is not valid %v", fullEndpoint, errs)
	}

	return fmt.Sprintf(fullResourceExp, domainName, fullEndpoint), nil
}
