/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Kubernetes Service, 5737-D43
 * (C) Copyright IBM Corp. 2022, 2024 All Rights Reserved.
 * The source code for this program is not published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package crn

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testPathDirectory = "/tmp"
)

func TestSanitize(t *testing.T) {
	toConvert := "armada-service-whichever&*()"
	expected := "armada_service_whichever____"
	res := Sanitize(toConvert)
	assert.Equal(t, expected, res)

	toConvert = ""
	expected = ""
	res = Sanitize(toConvert)
	assert.Equal(t, expected, res)

	toConvert = "this one has spaces"
	expected = "this_one_has_spaces"
	res = Sanitize(toConvert)
	assert.Equal(t, expected, res)

	toConvert = "thisIsJustFine"
	expected = "thisIsJustFine"
	res = Sanitize(toConvert)
	assert.Equal(t, expected, res)
}

func TestGetServiceCRNStructNothingSet(t *testing.T) {
	expectedCRNString := fmt.Sprintf("crn:CRN_CNAME:CRN_VERSION:CRN_CTYPE:CRN_SERVICENAME:CRN_REGION:CLUSTER_ID:storage_file:HOSTNAME")
	crn, err := GetServiceCRNStruct(crnPath)
	assert.Nil(t, err)
	assert.Equal(t, expectedCRNString, crn.String())
}

func TestGetServiceCRN(t *testing.T) {
	expectedCRNString := fmt.Sprintf("crn:CRN_CNAME:CRN_VERSION:CRN_CTYPE:CRN_SERVICENAME:CRN_REGION:CLUSTER_ID:storage_file:HOSTNAME")
	var testCM string

	testCM, err := GetServiceCRN()
	assert.Nil(t, err)
	assert.Equal(t, testCM, expectedCRNString)
}

func TestGetServiceCRNError(t *testing.T) {
	expectedCRNString := fmt.Sprintf("crn:CRN_CNAME:CRN_VERSION:CRN_CTYPE:CRN_SERVICENAME:CRN_REGION:CLUSTER_ID:storage_file:HOSTNAME")
	var testCM string
	fullPath := testPathDirectory + "/" + CRNCnameProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray := []byte(CRNCnameProp)
	os.WriteFile(fullPath, valueAsByteArray, 0000)
	testCM, err := GetServiceCRN()
	assert.Nil(t, err)
	assert.Equal(t, testCM, expectedCRNString)
	cleanupConfigMapMount(fullPath)
}

func TestGetCRNValueFromConfigMapMountNoFile(t *testing.T) {
	testFile := "testFile"
	defaultValue := "testDefault"
	res, err := getCRNValueFromConfigMapMount(testFile, defaultValue)
	assert.Nil(t, err, "Got error back from getCRNValueFromConfigMap")
	assert.Equal(t, defaultValue, res)

	testFile = "testFile"
	defaultValue = ""
	res, err = getCRNValueFromConfigMapMount(testFile, defaultValue)
	assert.Nil(t, err, "Got error back from getCRNValueFromConfigMap")
	assert.Equal(t, defaultValue, res)
}

func TestGetCRNValueFromConfigMapMountNoCName(t *testing.T) {
	testFile := ""
	defaultValue := "testDefault"
	_, err := getCRNValueFromConfigMapMount(testFile, defaultValue)
	assert.NotNil(t, err)
	assert.Equal(t, "filePath must have a value", err.Error())
}

func TestGetCRNValueFromConfigMapMount(t *testing.T) {
	testFile := "cname"
	testValue := "testCname"
	// make sure this file isn't already hanging out
	os.Remove(testFile)
	cnameValue := []byte(testValue)
	err := os.WriteFile(testFile, cnameValue, 0644)
	assert.Nil(t, err, "Error writing config map volume file.")
	expected := testValue
	defaultValue := "testDefault"
	res, err := getCRNValueFromConfigMapMount(testFile, defaultValue)
	assert.Nil(t, err)
	assert.Equal(t, expected, res)
	os.Remove(testFile)
}

func TestGetServiceCRNStruct(t *testing.T) {
	testPathDirectory := "/tmp"
	testCases := []struct {
		crnMap      map[string]string
		expectedCrn CRN
		shouldError bool
	}{
		{
			crnMap: map[string]string{
				CRNCnameProp:       "mycname",
				CRNVersionProp:     "mycversion",
				CRNCtypeProp:       "myctype",
				CRNRegionProp:      "myregion",
				CRNClusterIDProp:   "CLUSTER_ID",
				CRNServiceNameProp: "myservicename",
				CRNServiceIDProp:   "HOSTNAME",
			},
			expectedCrn: CRN{
				Cname:       "mycname",
				Cversion:    "mycversion",
				Ctype:       "myctype",
				Region:      "myregion",
				ClusterID:   "CLUSTER_ID",
				ServiceName: "myservicename",
				ServiceID:   "HOSTNAME",
			},
			shouldError: false,
		},
	}
	for _, testCase := range testCases {
		for k, v := range testCase.crnMap {
			fullPath := filepath.Join(testPathDirectory, k)
			cleanupConfigMapMount(fullPath)
			err := prepConfigMapMount(fullPath, v)
			assert.NoError(t, err)
		}
		crnData, err := GetServiceCRNStruct(testPathDirectory)
		if testCase.shouldError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedCrn, crnData)
		}
	}
}

func TestGetServiceCRNStructCNameError(t *testing.T) {
	fullPath := testPathDirectory + "/" + CRNCnameProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray := []byte(CRNCnameProp)
	os.WriteFile(fullPath, valueAsByteArray, 0000)
	_, err := GetServiceCRNStruct(testPathDirectory)
	assert.Error(t, err)
	cleanupConfigMapMount(fullPath)
}

func TestGetServiceCRNStructCVersionError(t *testing.T) {
	fullPath := testPathDirectory + "/" + CRNCnameProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray := []byte(CRNCnameProp)
	os.WriteFile(fullPath, valueAsByteArray, 0777)
	_, err := GetServiceCRNStruct(testPathDirectory)
	assert.Nil(t, err)

	fullPath = testPathDirectory + "/" + CRNVersionProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray = []byte(CRNVersionProp)
	os.WriteFile(fullPath, valueAsByteArray, 0000)
	_, err = GetServiceCRNStruct(testPathDirectory)
	assert.Error(t, err)
	cleanupConfigMapMount(fullPath)
}

func TestGetServiceCRNStructCTypeError(t *testing.T) {
	fullPath := testPathDirectory + "/" + CRNCnameProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray := []byte(CRNCnameProp)
	os.WriteFile(fullPath, valueAsByteArray, 0777)
	_, err := GetServiceCRNStruct(testPathDirectory)
	assert.Nil(t, err)

	fullPath = testPathDirectory + "/" + CRNVersionProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray = []byte(CRNVersionProp)
	os.WriteFile(fullPath, valueAsByteArray, 0777)
	_, err = GetServiceCRNStruct(testPathDirectory)
	assert.Nil(t, err)

	fullPath = testPathDirectory + "/" + CRNCtypeProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray = []byte(CRNCtypeProp)
	os.WriteFile(fullPath, valueAsByteArray, 0000)
	_, err = GetServiceCRNStruct(testPathDirectory)
	assert.Error(t, err)
	cleanupConfigMapMount(fullPath)
}

func TestGetServiceCRNStructVersionError(t *testing.T) {
	testPathDirectory := "/tmp"
	fullPath := testPathDirectory + "/" + CRNCnameProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray := []byte(CRNCnameProp)
	os.WriteFile(fullPath, valueAsByteArray, 0000)
	_, err := GetServiceCRNStruct(testPathDirectory)
	assert.Error(t, err)
}

func TestGetServiceCRNStructCRegionError(t *testing.T) {
	fullPath := testPathDirectory + "/" + CRNCnameProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray := []byte(CRNCnameProp)
	os.WriteFile(fullPath, valueAsByteArray, 0777)
	_, err := GetServiceCRNStruct(testPathDirectory)
	assert.Nil(t, err)

	fullPath = testPathDirectory + "/" + CRNVersionProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray = []byte(CRNVersionProp)
	os.WriteFile(fullPath, valueAsByteArray, 0777)
	_, err = GetServiceCRNStruct(testPathDirectory)
	assert.Nil(t, err)

	fullPath = testPathDirectory + "/" + CRNCtypeProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray = []byte(CRNCtypeProp)
	os.WriteFile(fullPath, valueAsByteArray, 0777)
	_, err = GetServiceCRNStruct(testPathDirectory)
	assert.Nil(t, err)

	fullPath = testPathDirectory + "/" + CRNRegionProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray = []byte(CRNRegionProp)
	os.WriteFile(fullPath, valueAsByteArray, 0000)
	_, err = GetServiceCRNStruct(testPathDirectory)
	assert.Error(t, err)
	cleanupConfigMapMount(fullPath)
}

func TestGetServiceCRNStructCServiceNameError(t *testing.T) {
	fullPath := testPathDirectory + "/" + CRNCnameProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray := []byte(CRNCnameProp)
	os.WriteFile(fullPath, valueAsByteArray, 0777)
	_, err := GetServiceCRNStruct(testPathDirectory)
	assert.Nil(t, err)

	fullPath = testPathDirectory + "/" + CRNVersionProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray = []byte(CRNVersionProp)
	os.WriteFile(fullPath, valueAsByteArray, 0777)
	_, err = GetServiceCRNStruct(testPathDirectory)
	assert.Nil(t, err)

	fullPath = testPathDirectory + "/" + CRNCtypeProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray = []byte(CRNCtypeProp)
	os.WriteFile(fullPath, valueAsByteArray, 0777)
	_, err = GetServiceCRNStruct(testPathDirectory)
	assert.Nil(t, err)

	fullPath = testPathDirectory + "/" + CRNRegionProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray = []byte(CRNRegionProp)
	os.WriteFile(fullPath, valueAsByteArray, 0777)
	_, err = GetServiceCRNStruct(testPathDirectory)
	assert.Nil(t, err)

	fullPath = testPathDirectory + "/" + CRNClusterIDProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray = []byte(CRNClusterIDProp)
	os.WriteFile(fullPath, valueAsByteArray, 0777)
	_, err = GetServiceCRNStruct(testPathDirectory)
	assert.Nil(t, err)

	fullPath = testPathDirectory + "/" + CRNServiceNameProp
	cleanupConfigMapMount(fullPath)
	valueAsByteArray = []byte(CRNServiceNameProp)
	os.WriteFile(fullPath, valueAsByteArray, 0000)
	_, err = GetServiceCRNStruct(testPathDirectory)
	assert.Error(t, err)
	cleanupConfigMapMount(fullPath)
}

func prepConfigMapMount(fileName string, value string) error {
	cleanupConfigMapMount(fileName)
	valueAsByteArray := []byte(value)
	return os.WriteFile(fileName, valueAsByteArray, 0644)
}

func cleanupConfigMapMount(filename string) {
	os.Remove(filename)
}

func TestGetCRNValueFromEnvVarEmptyValue(t *testing.T) {
	t.Log("Testing getting ENV with empty key")
	value := getCRNValueFromEnvVar("", "ENVTEST")

	assert.Equal(t, value, "ENVTEST")
}

func TestGetCRNValueFromEnvVarWithValue(t *testing.T) {
	t.Log("Testing getting ENV with value")
	os.Setenv("ENVTEST", "SOMETHING")
	value := getCRNValueFromEnvVar("ENVTEST", "ENVTESTISEMPTY")

	assert.Equal(t, value, "SOMETHING")
}

func TestGetEnvUnset(t *testing.T) {
	t.Log("Testing getting ENV with unset")
	os.Unsetenv("ENVTEST")
	path := getEnv("ENVTEST")

	assert.Equal(t, path, "")
}

func TestGetEnv(t *testing.T) {
	t.Log("Testing getting ENV")
	crnPath := "/tmp"
	os.Setenv("ENVTEST", crnPath)
	path := getEnv("ENVTEST")

	assert.Equal(t, crnPath, path)
}

func TestGetGoPathNullPath(t *testing.T) {
	t.Log("Testing getting CRN_CONFIGMAP_PATH NULL Path")
	crnPath := ""
	os.Setenv(crnENVVariable, crnPath)
	path := getEnv(crnENVVariable)

	assert.Equal(t, crnPath, path)
}

func TestGetServiceCRNENV(t *testing.T) {
	expectedCRNString := fmt.Sprintf("crn:CRN_CNAME:CRN_VERSION:CRN_CTYPE:CRN_SERVICENAME:CRN_REGION:CLUSTER_ID:storage_file:HOSTNAME")
	var testCM string
	os.Setenv(crnENVVariable, "/tmp")
	testCM, err := GetServiceCRN()
	assert.Nil(t, err)
	assert.Equal(t, testCM, expectedCRNString)
}
