package device_manager

import (
	"fmt"
	"strings"

	"github.com/furiosa-ai/furiosa-device-plugin/internal/config"
	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/device"

	"k8s.io/apimachinery/pkg/api/validation"
)

func buildDomainName(strategy config.ResourceUnitStrategy) string {
	if strategy == config.LegacyStrategy {
		return "alpha.furiosa.ai"
	}

	return "furiosa.ai"
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

func validateCoreUnit(arch device.Arch, coreUnit int) error {
	switch arch {
	case device.ArchWarboy:
		if err := coreUnitValidator(1, 2, coreUnit); err != nil {
			return err
		}
	case device.ArchRenegade:
		if err := coreUnitValidator(1, 4, coreUnit); err != nil {
			return err
		}
	default:
		return fmt.Errorf("wrong architecture")
	}

	return nil
}

func buildResourceEndpointCoreUnitTag(coreUnit int) string {
	return fmt.Sprintf("%dcore", coreUnit)
}

func buildResourceEndpointMemoryTag(arch device.Arch, coreUnit int) string {
	if arch == device.ArchWarboy {
		return fmt.Sprintf("%d", 16/(2/coreUnit))
	}

	return fmt.Sprintf("%d", 48/(8/coreUnit))
}

func buildResourceEndpointName(arch device.Arch, strategy config.ResourceUnitStrategy) string {
	if strategy == config.LegacyStrategy {
		return "npu"
	}
	return strings.ToLower(string(arch))
}

func strategyToCoreUnit(strategy config.ResourceUnitStrategy) int {
	if strategy == config.SingleCoreStrategy {
		return 1
	} else if strategy == config.DualCoreStrategy {
		return 2
	}

	return 4
}

func buildFullEndpoint(arch device.Arch, strategy config.ResourceUnitStrategy) (string, error) {
	endpointName := buildResourceEndpointName(arch, strategy)

	if strategy == config.LegacyStrategy || strategy == config.GenericStrategy {
		return endpointName, nil
	}

	coreUnit := strategyToCoreUnit(strategy)
	if err := validateCoreUnit(arch, coreUnit); err != nil {
		return "", err
	}

	coreUnitTag := buildResourceEndpointCoreUnitTag(coreUnit)
	memoryTag := buildResourceEndpointMemoryTag(arch, coreUnit)

	return fmt.Sprintf("%s-%s.%sgb", endpointName, coreUnitTag, memoryTag), nil
}

func buildAndValidateFullResourceEndpointName(arch device.Arch, strategy config.ResourceUnitStrategy) (string, error) {
	domainName := buildDomainName(strategy)

	fullEndpoint, err := buildFullEndpoint(arch, strategy)
	if err != nil {
		return "", err
	}

	errs := validation.NameIsDNSSubdomain(fullEndpoint, false)
	if len(errs) != 0 {
		return "", fmt.Errorf("resource name %s is not valid %v", fullEndpoint, errs)
	}

	return fmt.Sprintf("%s/%s", domainName, fullEndpoint), nil
}
