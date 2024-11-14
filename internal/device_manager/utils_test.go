package device_manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBusIDfromBDF(t *testing.T) {
	tests := []struct {
		description    string
		bdf            string
		expectedResult string
		expectedError  bool
	}{
		{
			description:    "test positive1",
			bdf:            "0000:51:00.0",
			expectedResult: "51",
			expectedError:  false,
		},
		{
			description:    "test positive2",
			bdf:            "0011:9e:00.0",
			expectedResult: "9e",
			expectedError:  false,
		},
		{
			description:    "test wrong format",
			bdf:            "0000.af94:0.1",
			expectedResult: "",
			expectedError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			actualResult, actualErr := parseBusIDfromBDF(tc.bdf)
			if tc.expectedError {
				assert.Error(t, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tc.expectedResult, actualResult)
		})
	}
}
