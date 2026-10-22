package device_manager

import (
	"testing"

	"github.com/furiosa-ai/furiosa-smi-go/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestBuildAndValidateFullResourceEndpointName(t *testing.T) {
	name, err := buildAndValidateFullResourceEndpointName(api.ArchRNGD)

	assert.NoError(t, err)
	assert.Equal(t, "furiosa.ai/rngd", name)
}
