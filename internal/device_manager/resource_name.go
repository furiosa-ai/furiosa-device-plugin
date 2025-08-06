package device_manager

import (
	"fmt"
	"strings"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/smi"
	"k8s.io/apimachinery/pkg/api/validation"
)

const (
	defaultDomain   = "furiosa.ai"
	fullResourceExp = "%s/%s"
)

func buildResourceEndpointName(arch smi.Arch) string {
	return strings.ToLower(arch.ToString())
}

func buildAndValidateFullResourceEndpointName(arch smi.Arch) (string, error) {
	endpointName := buildResourceEndpointName(arch)
	errs := validation.NameIsDNSSubdomain(endpointName, false)
	if len(errs) != 0 {
		return "", fmt.Errorf("resource name %s is not valid %v", endpointName, errs)
	}

	return fmt.Sprintf(fullResourceExp, defaultDomain, endpointName), nil
}
