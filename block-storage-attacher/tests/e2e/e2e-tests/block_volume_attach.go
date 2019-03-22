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
	commontest "github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/tests/e2e/common"
	"github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/tests/e2e/framework"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	volumeid          = ""
	pvname            = ""
	pvcname           = ""
	clusterName       = ""
	pvfilepath        = ""
	pv                *v1.PersistentVolume
	e2epath           = "src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/tests/e2e/e2e-tests/"
	scriptspath       = "src/github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/scripts/"
	pvscriptpath      = ""
	ymlscriptpath     = ""
	ymlgenpath        = ""
	expvname          = ""
	testfilepath      = ""
	portworxscpath    = ""
	portworxclassname = ""
	expv              *v1.PersistentVolume
	c                 clientset.Interface
	ns                string
	fpointer          *os.File
	perr              error
)
var _ = framework.KubeDescribe("[Feature:Block_Volume_Attach_E2E]", func() {
	f := framework.NewDefaultFramework("block-volume-attach")
	// filled in BeforeEach

	BeforeEach(func() {
		c = f.ClientSet
		ns = f.Namespace.Name
		pvscriptpath = e2epath + "utilscript.sh"
		ymlscriptpath = scriptspath + "mkpvyaml"
		ymlgenpath = scriptspath + "yamlgen.yaml"
		testfilepath = e2epath + "e2eTests.txt"
		portworxscpath = e2epath + "portworx_storageclass.yaml"
		portworxclassname = "portworx-sc"

	})

	framework.KubeDescribe("Block_Volume_Attach E2E ", func() {
		It("Block Volume attach E2e Testcases", func() {
			By("Volume Creation")
			gopath := os.Getenv("GOPATH")
			testfilepath = gopath + "/" + testfilepath
			fpointer, perr = os.OpenFile(testfilepath, os.O_APPEND|os.O_WRONLY, 0644)
			if perr != nil {
				panic(perr)
			}
			defer fpointer.Close()
			clusterName, err := getCluster(gopath + "/" + ymlgenpath)
			if err != nil {
				logResult("BlockVolumeAttacher-Volume-Test: Getting Cluster Details: FAIL\n")
			} else {
				logResult("BlockVolumeAttacher-Volume-Test: Getting Cluster Details: PASS\n")
			}
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
			if err != nil {
				logResult("BlockVolumeAttacher-Volume-Test: Volume Creation: FAIL\n")
			} else {
				logResult("BlockVolumeAttacher-Volume-Test: Volume Creaiton: PASS\n")
			}
			Expect(err).NotTo(HaveOccurred())

			/* Static PV Creation */

			By("Static PV  Creation")
			if filestatus == true {
				expvname, _ := getPVName(pvfilepath)
				fmt.Printf("expvname:\n%s\n", expvname)
				pvscriptpath = gopath + "/" + pvscriptpath
				filepatharg := fmt.Sprintf("%s", pvfilepath)
				expv, err = c.Core().PersistentVolumes().Get(expvname)
				if err == nil {
					cleanUP(expvname, expv)
				}
				cmd := exec.Command(pvscriptpath, filepatharg, "pvcreate")
				var stdout, stderr bytes.Buffer
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
					fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
					logResult("BlockVolumeAttacher-Volume-Test: PV Creation: FAIL\n")
				} else {
					logResult("BlockVolumeAttacher-Volume-Test: PV Creation: PASS\n")
				}
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
				fmt.Printf("Created Static PV \n%s\n", pvname)
				attachStatus, err := getAttchStatus()
				if err != nil {
					cleanUP(pvname, pv)
				}
				Expect(err).NotTo(HaveOccurred())
				devicePath := pv.ObjectMeta.Annotations["ibm.io/dm"]
				if !strings.Contains(devicePath, "/dev/dm-") {
					err := errors.New("Device path is not attached")
					logResult("BlockVolumeAttacher-Volume-Test: Device Attach: FAIL\n")
					Expect(err).NotTo(HaveOccurred())
				}
				Expect(attachStatus).To(Equal("attached"))
				logResult("BlockVolumeAttacher-Volume-Test: Attach: PASS\n")
			}

			/* Port Worx PVC creation */

			By("PVC  Creation")
			portworxscpath = gopath + "/" + portworxscpath
			filestatus, err = fileExists(portworxscpath)
			if filestatus == true {
				filepatharg2 := fmt.Sprintf("%s", portworxscpath)
				cmd := exec.Command(pvscriptpath, filepatharg2, "portworxpvcreate")
				var stdout, stderr bytes.Buffer
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr
				err = cmd.Run()
				if err != nil {
					outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
					fmt.Printf("out:\n%s\nerr:\n%s\n", outStr, errStr)
					cleanUP(pvname, pv)
					portworxcleanup(portworxclassname)
					logResult("BlockVolumeAttacher-Volume-Test: Portworx Installtion: FAIL\n")
				} else {
					logResult("BlockVolumeAttacher-Volume-Test: Portworx installtion: PASS\n")
				}
				Expect(err).NotTo(HaveOccurred())
			}
			framework.KubeDescribe("Test mount point: Create mount point, read, write...", func() {
				//It("PRESTAGE: Test mount point: Create mount point, read, write... [Serial]", func() {
				By("Creating a claim with a dynamic provisioning annotation")
				claim := commontest.NewClaim(ns, "portworx-sc", "2Gi")
				defer func() {
					c.Core().PersistentVolumeClaims(ns).Delete(claim.Name, nil)
				}()
				claim, err := c.Core().PersistentVolumeClaims(ns).Create(claim)
				if err != nil {
					cleanUP(pvname, pv)
					portworxcleanup(portworxclassname)
					logResult("BlockVolumeAttacher-Volume-Test: Portworx PVC creation: FAIL\n")
				} else {
					logResult("BlockVolumeAttacher-Volume-Test: Portworx PVC creation: PASS\n")
				}
				Expect(err).NotTo(HaveOccurred())

				//pv := commontest.TestCreate(c, claim)
				commontest.TestWrite(c, claim)
				commontest.TestRead(c, claim)
				//commontest.TestDelete(c, claim, pv)
				// })
			})

			/* Portworx deleteion */

			filepatharg := fmt.Sprintf("%s", portworxclassname)
			cmd = exec.Command(pvscriptpath, filepatharg, "portworxdelete")
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()

			/* Stativ PV  Deletion */

			By("Static PV Deletion ")
			err = c.Core().PersistentVolumes().Delete(pvname, nil)
			if err != nil {
				logResult("BlockVolumeAttacher-Volume-Test: PV Deletion: FAIL\n")
			} else {
				logResult("BlockVolumeAttacher-Volume-Test: PV Deletion: PASS\n")
			}
			Expect(err).NotTo(HaveOccurred())

			/* Volume deletion */

			By("Volume Deletion  ")
			volumeid = pv.ObjectMeta.Annotations["ibm.io/volID"]
			volidarg := fmt.Sprintf("%s", volumeid)
			nodeip := pv.ObjectMeta.Annotations["ibm.io/nodeip"]
			nodeiparg := fmt.Sprintf("%s", nodeip)
			cmd = exec.Command(pvscriptpath, volidarg, "voldelete", nodeiparg)
			//var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				logResult("BlockVolumeAttacher-Volume-Test: Volume Deletion: FAIL\n")
			} else {
				logResult("BlockVolumeAttacher-Volume-Test: VOlume Deletion: PASS\n")
			}
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
		fmt.Printf("\n%s\n", attachStatus)
		time.Sleep(1 * time.Minute)
		if attachStatus == "failed" {
			return attachStatus, err
		} else if attachStatus == "attached" {
			return attachStatus, nil
		}
	}
	return attachStatus, err
}

func logResult(logdata string) {

	if _, err := fpointer.WriteString(logdata); err != nil {
		panic(err)
	}

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

func getPVName(filename string) (string, error) {

	var line = ""
	var pvname = ""

	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line = scanner.Text()
		fmt.Printf("line:\n%s\n", line)
		if strings.Contains(line, ":") {
			value := strings.Split(line, ":")
			if strings.TrimSpace(value[0]) == "name" {
				pvname = strings.TrimSpace(value[1])
				break
			}

		}
	}
	return pvname, nil
}

func cleanUP(expvname string, pvobj *v1.PersistentVolume) {
	volumeid = pvobj.ObjectMeta.Annotations["ibm.io/volID"]
	volidarg := fmt.Sprintf("%s", volumeid)
	nodeip := pvobj.ObjectMeta.Annotations["ibm.io/nodeip"]
	nodeiparg := fmt.Sprintf("%s", nodeip)
	cmd := exec.Command(pvscriptpath, volidarg, "voldelete", nodeiparg)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())
	fmt.Printf("CleanUP: out:\n%s\n CleanUP err:\n%s\n", outStr, errStr)
	c.Core().PersistentVolumes().Delete(expvname, nil)

}

func portworxcleanup(classname string) {

	filepatharg := fmt.Sprintf("%s", classname)
	cmd := exec.Command(pvscriptpath, filepatharg, "portworxdelete")
	cmd.Run()
}
