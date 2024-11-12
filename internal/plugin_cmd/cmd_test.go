package plugin_cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	expectedHelpOutput = `Furiosa Device Plugin for Kubernetes

Usage:
  furiosa-device-plugin [flags]

Examples:
furiosa-device-plugin

Flags:
  -h, --help   help for furiosa-device-plugin
`
)

func safeError(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

func TestDevicePluginCommand(t *testing.T) {
	tests := []struct {
		description    string
		args           []string
		expectedResult string
		expectedError  error
	}{
		{
			description:    "test cmd -h",
			args:           []string{"-h"},
			expectedResult: expectedHelpOutput,
			expectedError:  nil,
		},
	}

	for _, tc := range tests {
		cmd := NewDevicePluginCommand()

		buf := new(bytes.Buffer)
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs(tc.args)

		err := cmd.Execute()
		if err != nil || tc.expectedError != nil {
			assert.Equal(t, strings.TrimSpace(safeError(err)), strings.TrimSpace(safeError(tc.expectedError)))
		}

		output := buf.String()

		assert.Equal(t, strings.TrimSpace(tc.expectedResult), strings.TrimSpace(output))
	}
}
