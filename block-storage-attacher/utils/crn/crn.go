/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Kubernetes Service, 5737-D43
 * (C) Copyright IBM Corp. 2022 All Rights Reserved.
 * The source code for this program is not published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

// Package crn ...
package crn

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	crnBase = "crn:%s:%s:%s:%s:%s:%s:storage_file:%s"

	// CRN MountPath
	crnPath = "/etc/crn_info_ibmc"

	// CRN MountPath ENV variable
	crnENVVariable = "CRN_CONFIGMAP_PATH"

	/**
	CRN constants
	*/

	// CRNCnameProp is the property name that contains the cname value
	CRNCnameProp = "CRN_CNAME"

	// CRNCtypeProp is the property name that contains the ctype value
	CRNCtypeProp = "CRN_CTYPE"

	// CRNVersionProp is the property name that contains the ctype value
	CRNVersionProp = "CRN_VERSION"

	// CRNRegionProp is the property name that contains the region value
	CRNRegionProp = "CRN_REGION"

	// CRNClusterIDProp is the property name that contains the infrastructure id value
	CRNClusterIDProp = "CLUSTER_ID"

	// CRNServiceNameProp is the property name that contains the service name value
	CRNServiceNameProp = "CRN_SERVICENAME"

	// CRNServiceIDProp is the property name that contains the service id value
	CRNServiceIDProp = "HOSTNAME"
)

// CRN struct is used to house all information that builds the CRN name
type CRN struct {
	Cname       string
	Ctype       string
	Cversion    string
	Region      string
	ClusterID   string
	ServiceName string
	ServiceID   string
}

func (c CRN) String() string {
	return fmt.Sprintf(crnBase, c.Cname, c.Cversion, c.Ctype, c.ServiceName, c.Region, c.ClusterID, c.ServiceID)
}

func getEnv(key string) string {
	return os.Getenv(strings.ToUpper(key))
}

// GetServiceCRN retrieves the crn for the given service and returns a string.
func GetServiceCRN() (string, error) {
	var crnPathDir string
	crnPathDir = crnPath
	if crnPath := getEnv(crnENVVariable); crnPath != "" {
		crnPathDir = crnPath
	}

	crnStruct, err := GetServiceCRNStruct(crnPathDir)
	if err != nil {
		return "", err
	}
	return crnStruct.String(), nil
}

// populateFromMountedConfigMap populates the existing CRN struct with values from
// a mounted config map. The existing CRN value is overwritten if a new value exists
// in the mounted config map
func populateFromMountedConfigMap(pathToDirectory string, c *CRN) error {
	var (
		cname       string
		cversion    string
		ctype       string
		region      string
		clusterID   string
		serviceName string
		serviceID   string
		err         error
	)

	var filePath string
	filePath = filepath.Join(pathToDirectory, CRNCnameProp)
	cname, err = getCRNValueFromConfigMapMount(filePath, c.Cname)
	if err != nil {
		return err
	}

	filePath = filepath.Join(pathToDirectory, CRNVersionProp)
	cversion, err = getCRNValueFromConfigMapMount(filePath, c.Cversion)
	if err != nil {
		return err
	}

	filePath = filepath.Join(pathToDirectory, CRNCtypeProp)
	ctype, err = getCRNValueFromConfigMapMount(filePath, c.Ctype)
	if err != nil {
		return err
	}

	filePath = filepath.Join(pathToDirectory, CRNRegionProp)
	region, err = getCRNValueFromConfigMapMount(filePath, c.Region)
	if err != nil {
		return err
	}

	clusterID = getCRNValueFromEnvVar(CRNClusterIDProp, c.ClusterID)

	filePath = filepath.Join(pathToDirectory, CRNServiceNameProp)
	serviceName, err = getCRNValueFromConfigMapMount(filePath, c.ServiceName)
	if err != nil {
		return err
	}

	serviceID = getCRNValueFromEnvVar(CRNServiceIDProp, c.ServiceID)

	c.Cname = cname
	c.Cversion = cversion
	c.Ctype = ctype
	c.Region = region
	c.ClusterID = clusterID
	c.ServiceName = serviceName
	c.ServiceID = serviceID

	return nil
}

// getCRNValueFromConfigMapMount gets the crn value from the config map mount. If the
// crn property doens't exist in the config map mount then the default value provided is returned
func getCRNValueFromConfigMapMount(filePath string, defaultValue string) (string, error) {
	if filePath == "" {
		return defaultValue, errors.New("filePath must have a value")
	}

	crnFile, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		if os.IsNotExist(err) {
			return defaultValue, nil
		}
		return defaultValue, err
	}
	data := make([]byte, 200)
	count, readErr := crnFile.Read(data)
	if readErr != nil {
		return defaultValue, readErr
	}
	noSpace := strings.TrimSpace(string(data[:count]))
	return Sanitize(noSpace), nil
}

// GetServiceCRNStruct retrieves the crn for the given service from a mounted config map.
// Mounting a config map populates a directory with files that correspond to each field
// in the config map. These files contain the data for the field.
func GetServiceCRNStruct(pathToDirectory string) (CRN, error) {
	var err error
	// default crn struct
	crnStruct := CRN{
		Cname:       CRNCnameProp,
		Ctype:       CRNCtypeProp,
		Cversion:    CRNVersionProp,
		Region:      CRNRegionProp,
		ClusterID:   CRNClusterIDProp,
		ServiceName: CRNServiceNameProp,
		ServiceID:   CRNServiceIDProp,
	}
	err = populateFromMountedConfigMap(pathToDirectory, &crnStruct)
	return crnStruct, err
}

// getCRNValueFromEnvVar gets the crn value from an environment variable, returns the
// default value supplied  if the environment variable does not exist
func getCRNValueFromEnvVar(name string, defaultValue string) string {
	if name == "" {
		return defaultValue
	}
	value, ok := os.LookupEnv(name)
	if ok {
		value = Sanitize(value)
	} else {
		// the environment variable doesn't exist so take the default value
		value = defaultValue
	}
	return value
}

// Sanitize removes characters from a string that aren't supported by metrics collection
func Sanitize(metric string) string {
	if metric == "" {
		return metric
	}
	newMetric := metric
	for i, b := range metric {
		if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_' || b == ':' || (b >= '0' && b <= '9' && i > 0)) {
			invalidChar := fmt.Sprintf("%c", b)
			newMetric = strings.Replace(newMetric, invalidChar, "_", 1)
		}
	}
	return newMetric
}
