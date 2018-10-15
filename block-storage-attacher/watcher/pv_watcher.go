/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Container Service, 5737-D43
 * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
 * The source code for this program is not  published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

package watcher

import (
	"fmt"
	//"encoding/json"
	"github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/utils/config"
	"github.com/coreos/go-systemd/dbus"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	//types "k8s.io/apimachinery/pkg/types"
	//"k8s.io/client-go/pkg/api/v1"
	"io/ioutil"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	ATTACHSTATUS = "ibm.io/attachstatus"
	IQN          = "ibm.io/iqn"
	USERNAME     = "ibm.io/username"
	PASSWORD     = "ibm.io/password"
	TARGET       = "ibm.io/targetip"
	LUNID        = "ibm.io/lunid"
	NODEIP       = "ibm.io/nodeip"
	DMPATH       = "ibm.io/dm"
	MULTIPATH    = "ibm.io/mpath"
	ATTACH       = "attach"
	DETACH       = "detach"

	STORAGECLASS     = "ibmc-block-attacher"
	STATUS_ATTACHING = "attaching"
	STATUS_ATTACHED  = "attached"
	STATUS_FAILED    = "failed"
	PX_CONF          = "/host/etc/iscsi-portworx-volume.conf"
	PX_SERVICE       = "ibmc-portworx.service"
)

var clientset kubernetes.Interface
var lgr zap.Logger

func WatchPersistentVolumes(client kubernetes.Interface, log zap.Logger) {
	clientset = client
	lgr = log
	volumeSource := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return clientset.CoreV1().PersistentVolumes().List(options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return clientset.CoreV1().PersistentVolumes().Watch(options)
		},
	}
	_, controller := cache.NewInformer(volumeSource, &v1.PersistentVolume{}, time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    AttachVolume,
			DeleteFunc: DetachVolume,
			UpdateFunc: nil,
		},
	)
	stopch := wait.NeverStop
	go controller.Run(stopch)
	lgr.Info("Watching persistent volumes for volume attach")
	<-stopch
}

func AttachVolume(obj interface{}) {
	pv, ok := obj.(*v1.PersistentVolume)
	if !ok {
		lgr.Error("Error in reading watcher event data of persistent volume")
		return
	}
	if pv.Spec.StorageClassName != STORAGECLASS {
		lgr.Info("Persistent volume does not belong to storage class: ", zap.String("Name", pv.Name), zap.String("Storage_Class", pv.Spec.StorageClassName))
		return
	}
	ModifyAttachConfig(pv)
}

func ModifyAttachConfig(pv *v1.PersistentVolume) {
	lgr.Info("Volume to be attached: ", zap.String("Name", pv.Name))

	validateErr := Validate(pv)
	if validateErr != nil {
		lgr.Error("", zap.Error(validateErr))
		return
	}
	volume := config.Volume{}
	volume.Iqn = pv.Annotations[IQN]
	volume.Username = pv.Annotations[USERNAME]
	volume.Password = pv.Annotations[PASSWORD]
	volume.Target = pv.Annotations[TARGET]
	volume.Lunid, _ = strconv.Atoi(pv.Annotations[LUNID])
	volume.Nodeip = pv.Annotations[NODEIP]

	worker_node := os.Getenv("NODE_IP")
	if worker_node != volume.Nodeip {
		lgr.Info("The volume attach is not requested for this worker node")
		return
	}
	var input []byte
	var err error
	if input, err = ioutil.ReadFile(PX_CONF); err != nil {
		lgr.Error("Could not read iscsi-portworx-volume.conf file")
		return
	} else {
		lines := strings.Split(string(input), "\n")
		for i, line := range lines {
			if strings.Contains(line, "iqn=") {
				lines[i] = "iqn=" + strings.TrimSpace(volume.Iqn)
			} else if strings.Contains(line, "username=") {
				lines[i] = "username=" + strings.TrimSpace(volume.Username)
			} else if strings.Contains(line, "password=") {
				lines[i] = "password=" + strings.TrimSpace(volume.Password)
			} else if strings.Contains(line, "target_ip=") {
				lines[i] = "target_ip=" + strings.TrimSpace(volume.Target)
			} else if strings.Contains(line, "lunid=") {
				lines[i] = "lunid=" + strconv.Itoa(volume.Lunid)
			} else if strings.Contains(line, "node_ip=") {
				lines[i] = "node_ip=" + strings.TrimSpace(volume.Nodeip)
			} else if strings.Contains(line, "op=") {
				lines[i] = "op=" + strings.TrimSpace(ATTACH)
			}
		}

		modifiedlines := []string{}
		modifiedlines = append(modifiedlines, lines...)
		output := strings.Join(modifiedlines, "\n")
		if err = ioutil.WriteFile(PX_CONF, []byte(output), 0644); err != nil {
			lgr.Error("Could not write to iscsi-portworx-volume.conf file")
			return
		}
	}

	pvUpdated := false
	for x := 0; x < 5; x++ {
		//Adding sleep since kubernetes will be still modifying the PV object
		time.Sleep(5 * time.Second)

		//Fetch the latest version of the PV from Kubernetes apiserver
		latestPV, pvErr := clientset.CoreV1().PersistentVolumes().Get(pv.Name, metav1.GetOptions{})
		if pvErr != nil {
			lgr.Warn("Failed to fetch PV from apiserver:", zap.String("pvname", pv.Name), zap.Error(pvErr))
			continue
		}
		//annotations := pv.ObjectMeta.Annotations
		//annotations[ATTACHSTATUS] = STATUS_ATTACHING
		//jsonAnnotations, _ := json.Marshal(annotations)
		//patchData := "{\"metadata\": {\"annotations\":" + string(jsonAnnotations) + "}}"
		//pv, err = clientset.CoreV1().PersistentVolumes().Patch(pv.ObjectMeta.Name, types.MergePatchType, []byte(patchData))
		latestPV.Annotations[ATTACHSTATUS] = STATUS_ATTACHING
		_, pvErr = clientset.CoreV1().PersistentVolumes().Update(latestPV)
		if pvErr == nil {
			pvUpdated = true
			break
		}
		lgr.Warn("Failed to update PV from apiserver:", zap.String("pvname", pv.Name), zap.Error(pvErr))
	}
	if !pvUpdated {
		return
	}

	// Restart ibmc-portworx service so volume can be attached
	dbConn, connErr := dbus.New()
	if connErr != nil {
		lgr.Error("Error: Unable to connect!", zap.Error(connErr))
		return
	}
	reschan := make(chan string)
	_, restartErr := dbConn.RestartUnit(PX_SERVICE, "fail", reschan)
	if restartErr != nil {
		lgr.Error("Error: Unable to restart target", zap.Error(restartErr))
		return
	} else {
		lgr.Info("Unit Restarted !!")
	}
	job := <-reschan
	if job != "done" {
		lgr.Error("Error: Restart of service is not done: " + job)
		return
	}
	UpdatePersistentVolume(volume, pv)
}

func UpdatePersistentVolume(volume config.Volume, pv *v1.PersistentVolume) {
	pathsFile := "/host/lib/ibmc-portworx/out_paths"
	mpathsFile := "/host/lib/ibmc-portworx/out_multipaths"
	var fileExists bool
	var mpath string
	var devicepath string
	var lunid int

	//Waiting for 625 secs here as the iscsi-attach script has a wait time of 600secs in total for the volume attach to finish
	for x := 0; x < 125; x++ {
		_, err1 := os.Stat(pathsFile)
		_, err2 := os.Stat(mpathsFile)
		if (!os.IsNotExist(err1)) && (!os.IsNotExist(err2)) {
			fileExists = true
			break
		}
		time.Sleep(5 * time.Second)
	}

	if fileExists {
		var input []byte
		var err error
		//Parse paths to fetch lun id as per below command and output
		/* multipathd show paths format "%w %i"
		uuid                              hcil
		3600a09803830445455244c4a38752d30 10:0:0:15  --> The last part of hcil is lun id
		3600a09803830445455244c4a38752d30 11:0:0:15
		*/
		if input, err = ioutil.ReadFile(pathsFile); err != nil {
			lgr.Error("Could not read " + pathsFile + " file")
		} else {
			lines := strings.Split(string(input), "\n")
			for _, line := range lines {
				space := regexp.MustCompile(`\s+`)
				line = space.ReplaceAllString(line, " ")
				line_parts := strings.Split(string(line), " ")
				lgr.Info("Line: ", zap.Strings("LINE", line_parts))
				if len(line_parts) >= 2 {
					// Parse the LUN ID from output
					lun := strings.Split(string(line_parts[1]), ":")
					if len(lun) == 4 {
						if lunid, _ = strconv.Atoi(lun[3]); lunid == volume.Lunid {
							mpath = line_parts[0]
							break
						}
					}
				}
			}
		}

		// Parse multipaths to fetch device path
		/* multipathd show multipaths
		name                              sysfs uuid
		3600a09803830445455244c4a38752d30 dm-0  3600a09803830445455244c4a38752d30
		*/
		if input, err = ioutil.ReadFile(mpathsFile); err != nil {
			lgr.Error("Could not read " + mpathsFile + " file")
		} else {
			lines := strings.Split(string(input), "\n")
			for _, line := range lines {
				space := regexp.MustCompile(`\s+`)
				line = space.ReplaceAllString(line, " ")
				line_parts := strings.Split(string(line), " ")
				lgr.Info("Mpath Line: ", zap.Strings("LINE", line_parts))
				lgr.Info("MPath: ", zap.String("MPATH", mpath))
				if line_parts[0] == mpath {
					// Device path sample is /dev/dm-0
					devicepath = "/dev/" + line_parts[1]
					break
				}
			}
		}
		lgr.Info("Device path and volume lun ID: ", zap.String("LUN_Id", strconv.Itoa(volume.Lunid)), zap.String("Device_Path", devicepath))

		// Delete the output files
		del_err := os.Remove(pathsFile)
		if del_err != nil {
			lgr.Error("Delete of "+pathsFile+" file failed ", zap.Error(del_err))
		}

		del_err = os.Remove(mpathsFile)
		if del_err != nil {
			lgr.Error("Delete of "+mpathsFile+" file failed ", zap.Error(del_err))
		}

		for x := 0; x < 5; x++ {
			//Adding sleep since kubernetes will be still modifying the PV object
			time.Sleep(5 * time.Second)

			//Fetch the latest version of the PV from Kubernetes apiserver
			latestPV, pvErr := clientset.CoreV1().PersistentVolumes().Get(pv.Name, metav1.GetOptions{})
			// Update PV with devicepath and multipath
			latestPV.Annotations[DMPATH] = devicepath
			latestPV.Annotations[MULTIPATH] = mpath
			latestPV.Annotations[ATTACHSTATUS] = STATUS_ATTACHED
			_, pvErr = clientset.CoreV1().PersistentVolumes().Update(latestPV)
			if pvErr == nil {
				break
			}
			lgr.Warn("Failed to update PV from apiserver:", zap.String("pvname", pv.Name), zap.Error(pvErr))
		}
		return
	}
	for x := 0; x < 5; x++ {
		//Adding sleep since kubernetes will be still modifying the PV object
		time.Sleep(5 * time.Second)

		//Fetch the latest version of the PV from Kubernetes apiserver
		latestPV, pvErr := clientset.CoreV1().PersistentVolumes().Get(pv.Name, metav1.GetOptions{})
		latestPV.Annotations[ATTACHSTATUS] = STATUS_FAILED
		_, pvErr = clientset.CoreV1().PersistentVolumes().Update(latestPV)
		if pvErr == nil {
			break
		}
		lgr.Warn("Failed to update PV from apiserver:", zap.String("pvname", pv.Name), zap.Error(pvErr))
	}
}

func Validate(pv *v1.PersistentVolume) error {
	volDetails := make([]string, 0)
	if pv.Annotations == nil {
		lgr.Error("The PV has no volume details given to perform attach.")
		pv.Annotations[ATTACHSTATUS] = STATUS_FAILED
	} else {
		if _, present := pv.Annotations[IQN]; !present {
			volDetails = append(volDetails, IQN)
			pv.Annotations[ATTACHSTATUS] = STATUS_FAILED
		}
		if _, present := pv.Annotations[USERNAME]; !present {
			volDetails = append(volDetails, USERNAME)
			pv.Annotations[ATTACHSTATUS] = STATUS_FAILED
		}
		if _, present := pv.Annotations[PASSWORD]; !present {
			volDetails = append(volDetails, PASSWORD)
			pv.Annotations[ATTACHSTATUS] = STATUS_FAILED
		}
		if _, present := pv.Annotations[TARGET]; !present {
			volDetails = append(volDetails, TARGET)
			pv.Annotations[ATTACHSTATUS] = STATUS_FAILED
		}
		if _, present := pv.Annotations[LUNID]; !present {
			volDetails = append(volDetails, LUNID)
			pv.Annotations[ATTACHSTATUS] = STATUS_FAILED
		} else {
			if _, err := strconv.Atoi(pv.Annotations[LUNID]); err != nil {
				volDetails = append(volDetails, LUNID)
				pv.Annotations[ATTACHSTATUS] = STATUS_FAILED
			}
		}
		if _, present := pv.Annotations[NODEIP]; !present {
			volDetails = append(volDetails, NODEIP)
			pv.Annotations[ATTACHSTATUS] = STATUS_FAILED
		}
	}
	if pv.Annotations[ATTACHSTATUS] == STATUS_FAILED {
		lgr.Warn("Either no annotations are given or the following volume attributes are not valid in the PV:", zap.Strings("vol_attach_attrs", volDetails))

		for x := 0; x < 5; x++ {
			time.Sleep(5 * time.Second)

			//Fetch the latest version of the PV from Kubernetes apiserver
			latestPV, err := clientset.CoreV1().PersistentVolumes().Get(pv.Name, metav1.GetOptions{})
			latestPV.Annotations[ATTACHSTATUS] = STATUS_FAILED
			_, err = clientset.CoreV1().PersistentVolumes().Update(latestPV)
			if err == nil {
				break
			}
			lgr.Error("Failed to update PV from apiserver:", zap.String("pvname", pv.Name), zap.Error(err))
		}
		return fmt.Errorf("Error while validating the PV annotations %s", pv.Name)
	}
	return nil
}

func DetachVolume(obj interface{}) {
	pv, ok := obj.(*v1.PersistentVolume)
	if !ok {
		lgr.Error("Error in reading watcher event data of persistent volume")
		return
	}
	if pv.Spec.StorageClassName != STORAGECLASS {
		lgr.Info("Persistent volume does not belong to storage class: ", zap.String("Name", pv.Name), zap.String("Storage_Class", pv.Spec.StorageClassName))
		return
	}
	ModifyDetachConfig(pv)
}

func ModifyDetachConfig(pv *v1.PersistentVolume) {
	lgr.Info("Volume to be detached: ", zap.String("Name", pv.Name))

	volDetails := make([]string, 0)
	if pv.Annotations == nil {
		lgr.Error("The PV has no volume details given to perform detach.")
		return
	} else {
		if _, present := pv.Annotations[DMPATH]; !present {
			volDetails = append(volDetails, DMPATH)
		}
		if _, present := pv.Annotations[MULTIPATH]; !present {
			volDetails = append(volDetails, MULTIPATH)
		}
		if len(volDetails) > 0 {
			lgr.Error("Either no annotations are given or the following volume attributes are not valid in the PV:", zap.Strings("vol_detach_attrs", volDetails))
			return
		}
	}

	dev_path := strings.Split(pv.Annotations[DMPATH], "/")
	var input []byte
	var err error
	if input, err = ioutil.ReadFile(PX_CONF); err != nil {
		lgr.Error("Could not read iscsi-portworx-volume.conf file")
		return
	} else {
		lines := strings.Split(string(input), "\n")
		for i, line := range lines {
			if strings.Contains(line, "dm=") {
				lines[i] = "dm=" + strings.TrimSpace(dev_path[2])
			} else if strings.Contains(line, "mpath=") {
				lines[i] = "mpath=" + strings.TrimSpace(pv.Annotations[MULTIPATH])
			} else if strings.Contains(line, "op=") {
				lines[i] = "op=" + strings.TrimSpace(DETACH)
			}
		}

		modifiedlines := []string{}
		modifiedlines = append(modifiedlines, lines...)
		output := strings.Join(modifiedlines, "\n")
		if err = ioutil.WriteFile(PX_CONF, []byte(output), 0644); err != nil {
			lgr.Error("Could not write to iscsi-portworx-volume.conf file")
			return
		}
	}

	// Restart ibmc-portworx service so volume can be attached
	dbConn, connErr := dbus.New()
	if connErr != nil {
		lgr.Error("Error: Unable to connect!", zap.Error(connErr))
		return
	}
	reschan := make(chan string)
	_, restartErr := dbConn.RestartUnit(PX_SERVICE, "fail", reschan)
	if restartErr != nil {
		lgr.Error("Error: Unable to restart target", zap.Error(restartErr))
		return
	} else {
		lgr.Info("Unit Restarted !!")
	}
	job := <-reschan
	if job != "done" {
		lgr.Error("Error: Restart of service is not done: " + job)
		return
	}
}
