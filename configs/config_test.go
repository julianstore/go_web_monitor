package configs

import (
	"fmt"
	"testing"
)

// Unit test function for CheckFileExist
func TestCheckFileExist(t *testing.T) {
	testCases := []struct {
		filePath      string
		expectedValue bool
	}{
		{
			filePath:      "/config.conf",
			expectedValue: true,
		},
		{
			filePath:      "/config1.conf",
			expectedValue: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		if CheckFileExist(testCase.filePath) == testCase.expectedValue {
			t.Logf("CheckFileExist Function OK")
		} else {
			t.Fatalf(fmt.Sprintf("CheckFileExist Function Failed: filePath=%s, expectedValue = %t , result = %t", testCase.filePath, testCase.expectedValue, CheckFileExist(testCase.filePath)))
		}
	}

}

// Unit test function for GetConfig
func TestGetConfig(t *testing.T) {
	config := GetConfig()
	if (config == nil) {
		t.Logf("GetConfig Failed")
	} else {
		t.Logf("GetConfig Function OK")
	}
	
}

