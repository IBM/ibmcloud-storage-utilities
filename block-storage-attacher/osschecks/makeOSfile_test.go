/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package main

import (
	"io/ioutil"
	"log"
	"testing"

	"github.ibm.com/alchemy-containers/armada-opensource-lib/utils"
)

func Test_opensourceUpToDate(t *testing.T) {

	// Generate a new OPENSOURCE
	newOpensource, err := buildNewOPENSOURCE()
	if err != nil {
		t.Log("Error generating new OPENSOURCE. err", err)
		t.FailNow()
	}

	// Read in the existing OPENSOUCE
	existingOpensource, err := readExistingOPENSOURCE("../OPENSOURCE")
	if err != nil {
		t.Log("Error reading existing OPENSOURCE. err", err)
		t.FailNow()
	}

	// Compare the two. If there is a difference. BANG!
	if existingOpensource != newOpensource {
		t.Log("The OPENSOURCE has changed. Regenerate the OPENSOURCE and add the changes to your commit before delivering.")
		t.Log("Run the following from the armada-storage-file-plugin directory: go run osschecks/makeOSfile.go")
		t.FailNow()
	}
}

func buildNewOPENSOURCE() (string, error) {
	var OPENSOURCE string

	if utils.CheckFileExists("../glide.lock") {
		glideEntries, err := utils.GenerateNewOPENSOURCEfromGlideLock("../glide.lock")
		if err != nil {
			return "", err
		}
		OPENSOURCE += glideEntries
	}

	if utils.CheckFileExists("../Dockerfile") {
		dockerfileEntry, err := utils.GenerateDockerEntries("../Dockerfile")
		if err != nil {
			return "", err
		}
		OPENSOURCE += dockerfileEntry
	}

	if utils.CheckFileExists("../Dockerfile.build") {
		dockerfileBuildEntry, err := utils.GenerateDockerEntries("../Dockerfile.build")
		if err != nil {
			return "", err
		}
		OPENSOURCE += dockerfileBuildEntry
	}

	return OPENSOURCE, nil
}

func readExistingOPENSOURCE(filename string) (string, error) {
	var opensource string

	// Read in the existing OPENSOURCE file
	opensourceFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("Error detected reading existing OPENSOURCE: err   #%v ", err)
		return opensource, err
	}
	// convert it to string
	opensource = string(opensourceFile)

	return opensource, nil
}
