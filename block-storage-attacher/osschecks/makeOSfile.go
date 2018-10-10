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

	"github.ibm.com/alchemy-containers/armada-opensource-lib/utils"
)

const (
	opensourceFile = "OPENSOURCE"
)

func main() {
	opensourceContent, _ := utils.GenerateNewOPENSOURCE("glide.lock", "Dockerfile", "Dockerfile.build")
	err := ioutil.WriteFile(opensourceFile, []byte(opensourceContent), 0644)
	if err != nil {
		log.Println("Unable to write the OPENSOURCE file")
	}
}
