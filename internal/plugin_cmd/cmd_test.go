package plugin_cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

const (
	expectedHelpOutput = `Furiosa Device Plugin for Kubernetes

Usage:
  furiosa-device-plugin [flags]

Examples:
furiosa-device-plugin -localConfigPath = {path}

Flags:
  -h, --help                     help for furiosa-device-plugin
  -l, --localConfigPath string   [optional] local configuration file path
`
	expectedLocalConfigPathOutput = `Error: flag needs an argument: 'l' in -l
Usage:
  furiosa-device-plugin [flags]

Examples:
furiosa-device-plugin -localConfigPath = {path}

Flags:
  -h, --help                     help for furiosa-device-plugin
  -l, --localConfigPath string   [optional] local configuration file path
`
	expectedLocalConfigPathErrorOutput = "flag needs an argument: 'l' in -l"
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
		{
			description:    "test cmd -l",
			args:           []string{"-l"},
			expectedResult: expectedLocalConfigPathOutput,
			expectedError:  fmt.Errorf(expectedLocalConfigPathErrorOutput),
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
			if strings.TrimSpace(safeError(err)) != strings.TrimSpace(safeError(tc.expectedError)) {
				t.Errorf("expected %t but got actual %t", err, tc.expectedError)
			}
		}

		output := buf.String()

		if strings.TrimSpace(output) != strings.TrimSpace(tc.expectedResult) {
			t.Errorf("actual result does not match to expected result")
		}
	}
}
