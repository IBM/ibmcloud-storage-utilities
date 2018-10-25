/*
# Licensed Materials - Property of IBM
#
# (C) Copyright IBM Corp. 2017 All Rights Reserved
#
# US Government Users Restricted Rights - Use, duplicate or
# disclosure restricted by GSA ADP Schedule Contract with
# IBM Corp.
# encoding: utf-8
*/

package e2e

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.ibm.com/alchemy-containers/ibmcloud-storage-utilities/block-storage-attacher/tests/e2e/framework"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	volumeid      = ""
	pvname        = ""
	clusterName   = ""
	pvfilepath    = ""
	pv            *v1.PersistentVolume
	e2epath       = "src/github.ibm.com/alchemy-containers/ibmcloud-storage-utilities/block-storage-attacher/tests/e2e/e2e-tests/"
	pvscriptpath  = ""
	ymlscriptpath = ""
	ymlgenpath    = ""
	c             clientset.Interface
)
var _ = framework.KubeDescribe("[Feature:Block_Volume_Attach_E2E]", func() {
	f := framework.NewDefaultFramework("block-volume-attach")
	// filled in BeforeEach
	var ns string

	BeforeEach(func() {
		c = f.ClientSet
		ns = f.Namespace.Name
		pvscriptpath = e2epath + "utilscript.sh"
		ymlscriptpath = e2epath + "mkpvyaml"
		ymlgenpath = e2epath + "yamlgen.yaml"
	})

	framework.KubeDescribe("Block_Volume_Attach E2E ", func() {
		It("Block Volume attach E2e Testcases", func() {
			By("Volume Creation")
			gopath := os.Getenv("GOPATH")
			clusterName, err := getCluster(gopath + "/" + ymlgenpath)
			Expect(err).NotTo(HaveOccurred())
			pvfilepath = gopath + "/" + e2epath + "pv-" + clusterName + ".yaml"
			filestatus, err := fileExists(pvfilepath)
			if filestatus == true {
				os.Remove(pvfilepath)
			}
			ymlscriptpath = gopath + "/" + ymlscriptpath
			cmd := exec.Command(ymlscriptpath)
			cmd.Stdout = os.Stdout
			cmd.Env = os.Environ()
			cmd.Stderr = os.Stderr
			By("Volume Creation1")
			cmd.Run()

			filestatus, err = fileExists(pvfilepath)
			Expect(err).NotTo(HaveOccurred())

			/* Static PV Creation */

			By("Static PV  Creation")
			if filestatus == true {
				pvscriptpath = gopath + "/" + pvscriptpath
				filepatharg := fmt.Sprintf("%s", pvfilepath)
				cmd := exec.Command(pvscriptpath, filepatharg, "pvcreate")
				var stdout, stderr bytes.Buffer
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr
				err := cmd.Run()
				Expect(err).NotTo(HaveOccurred())
				outStr, _ := string(stdout.Bytes()), string(stderr.Bytes())
				if strings.Contains(outStr, "/") {
					pvstring := strings.Split(outStr, "/")
					pvnamestring := strings.Split(pvstring[1], " ")
					pvname = pvnamestring[0]
				} else {
					pvstring := strings.Split(outStr, " ")
					pvname = strings.Trim(pvstring[1], "\"")
				}

				pv, err = c.Core().PersistentVolumes().Get(pvname)
				Expect(err).NotTo(HaveOccurred())
				attachStatus, err := getAttchStatus()
				Expect(err).NotTo(HaveOccurred())
				devicePath := pv.ObjectMeta.Annotations["ibm.io/dm"]
				if !strings.Contains(devicePath, "/dev/dm-") {
					err := errors.New("Device path is not attached")
					Expect(err).NotTo(HaveOccurred())
				}
				Expect(attachStatus).To(Equal("attached"))
			}

			/* Stativ PV  Deletion */

			By("Static PV Deletion ")
			err = c.Core().PersistentVolumes().Delete(pvname, nil)
			Expect(err).NotTo(HaveOccurred())

			/* Volume deletion */

			By("Volume Deletion  ")
			volumeid = pv.ObjectMeta.Annotations["ibm.io/volID"]
			volidarg := fmt.Sprintf("%s", volumeid)
			nodeip := pv.ObjectMeta.Annotations["ibm.io/nodeip"]
			nodeiparg := fmt.Sprintf("%s", nodeip)
			cmd = exec.Command(pvscriptpath, volidarg, "voldelete", nodeiparg)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			Expect(err).NotTo(HaveOccurred())
			outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
			fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)

			filestatus, err = fileExists(pvfilepath)
			if filestatus == true {
				os.Remove(pvfilepath)
			}

		})
	})
})

func fileExists(filename string) (bool, error) {
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			return false, err
		}
	}
	return true, nil
}

func getAttchStatus() (string, error) {
	attachStatus := "attaching"
	err := errors.New("Timed out in PV creation")
	for start := time.Now(); time.Since(start) < (15 * time.Minute); {
		pv, _ = c.Core().PersistentVolumes().Get(pvname)
		attachStatus = pv.ObjectMeta.Annotations["ibm.io/attachstatus"]
		time.Sleep(1 * time.Minute)
		if attachStatus == "attached" || attachStatus == "failed" {
			return attachStatus, nil
		}
	}
	return attachStatus, err
}

func getCluster(filename string) (string, error) {

	var line = ""
	var clustername = ""

	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line = scanner.Text()
		value := strings.Split(line, ":")
		fmt.Printf("Value[0], Valu[1]:\n%s\n%s\n", value[0], value[1])
		if value[0] == "cluster" {
			if strings.Contains(value[1], "#") {
				value = strings.Split(value[1], "#")
				fmt.Printf("Value[0], Valu[1]:\n%s\n%s\n", value[0], value[1])
				clustername = strings.TrimSpace(value[0])
			} else {
				clustername = strings.TrimSpace(value[1])
			}
			fmt.Printf("cluster:\n%s\n", clustername)
			break
		}

	}
	return clustername, nil
}
