/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Kubernetes Service, 5737-D43
 * (C) Copyright IBM Corp. 2022 All Rights Reserved.
 * The source code for this program is not published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

// Package watcher ...
package watcher

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/utils/config"
	"github.com/coreos/go-systemd/v22/dbus"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	//VOLID ...
	VOLID = "ibm.io/volID"
	//ATTACHSTATUS ..
	ATTACHSTATUS = "ibm.io/attachstatus"
	// IQN ..
	IQN = "ibm.io/iqn"
	//USERNAME ...
	USERNAME = "ibm.io/username"
	//PASSWORD ...
	PASSWORD = "ibm.io/password"
	//TARGET ...
	TARGET = "ibm.io/targetip"
	//LUNID ...
	LUNID = "ibm.io/lunid"
	//NODEIP ...
	NODEIP = "ibm.io/nodeip"
	//DMPATH ...
	DMPATH = "ibm.io/dm"
	//MULTIPATH ...
	MULTIPATH = "ibm.io/mpath"
	//ATTACH ...
	ATTACH = "attach"
	//DETACH ...
	DETACH = "detach"
	//STORAGECLASS ...
	STORAGECLASS = "ibmc-block-attacher"
	//STATUS_ATTACHING ...
	STATUS_ATTACHING = "attaching" //nolint readability
	//STATUS_ATTACHED ...
	STATUS_ATTACHED = "attached" //nolint readability
	//STATUS_FAILED ...
	STATUS_FAILED = "failed" //nolint readability
	//INVALID_PARAMS ...
	INVALID_PARAMS = "invalid_params" //nolint readability
	//BLOCK_CONF ...
	BLOCK_CONF = "/host/etc/iscsi-block-volume.conf" //nolint readability
	//ATTACHER_SERVICE ...
	ATTACHER_SERVICE = "ibmc-block-attacher.service" //nolint readability
)

var clientset kubernetes.Interface
var lgr zap.Logger
var mutex = &sync.Mutex{}
var volumeQueue workqueue.RateLimitingInterface

// WatchPersistentVolumes ...
func WatchPersistentVolumes(client kubernetes.Interface, log zap.Logger) {
	clientset = client
	lgr = log
	ratelimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(15*time.Second, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
	)
	volumeQueue = workqueue.NewNamedRateLimitingQueue(ratelimiter, "volumes")

	volumeSource := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return clientset.CoreV1().PersistentVolumes().List(context.TODO(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			return clientset.CoreV1().PersistentVolumes().Watch(context.TODO(), options)
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
	go runVolumeWorker(stopch)
	lgr.Info("Running volume worker")
	<-stopch
}

func runVolumeWorker(_ <-chan struct{}) {
	for processNextVolume() {
	}
}

// processNextVolume processes items from volumeQueue
func processNextVolume() bool {
	obj, shutdown := volumeQueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer volumeQueue.Done(obj)
		var key *v1.PersistentVolume
		var ok bool
		if key, ok = obj.(*v1.PersistentVolume); !ok {
			volumeQueue.Forget(obj)
			return fmt.Errorf("expected string in workqueue but got %#v", obj)
		}

		if isRetryRequired, err := ModifyAttachConfig(key); isRetryRequired {
			volumeQueue.AddRateLimited(obj)
			lgr.Info("Retrying to attach storage", zap.String("Name", key.Name))
			return fmt.Errorf("retrying to attach storage %q: %s", key, err.Error())
		}

		volumeQueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		lgr.Error("Attach Error", zap.Error(err))
	}
	return true
}

// AttachVolume ...
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
	if volumeQueue.NumRequeues(pv) == 0 {
		volumeQueue.Add(pv)
		lgr.Info("Added storage to queue", zap.String("PV Name", pv.Name))
	}
}

// ModifyAttachConfig ...
func ModifyAttachConfig(pv *v1.PersistentVolume) (bool, error) {
	lgr.Info("Waiting for mutex lock in ATTACH", zap.String("Name", pv.Name))
	mutex.Lock()
	lgr.Info("Acquired mutex lock in ATTACH", zap.String("Name", pv.Name))
	defer mutex.Unlock()

	//Check if the PV exists using Kubernetes apiserver
	_, volErr := clientset.CoreV1().PersistentVolumes().Get(context.TODO(), pv.Name, metav1.GetOptions{})
	if volErr != nil {
		lgr.Warn("Failed to fetch PV from apiserver:", zap.String("pvname", pv.Name), zap.Error(volErr))
		return false, fmt.Errorf("error while fetching persistent volume %s. Error: %v", pv.Name, volErr)
	}
	lgr.Info("Volume to be attached: ", zap.String("Name", pv.Name))

	retry, validateErr := Validate(pv)
	if validateErr != nil {
		lgr.Error("Validation Error", zap.Error(validateErr))
		return retry, fmt.Errorf("error while validating PV attributes %s. Error: %v", pv.Name, validateErr)
	}
	volume := config.Volume{}
	volume.VolID = pv.Annotations[VOLID]
	volume.Iqn = pv.Annotations[IQN]
	volume.Username = pv.Annotations[USERNAME]
	volume.Password = pv.Annotations[PASSWORD]
	volume.Target = pv.Annotations[TARGET]
	lunid, errConv := strconv.Atoi(pv.Annotations[LUNID])
	if errConv != nil {
		return false, errConv
	}
	volume.Lunid = lunid
	volume.Nodeip = pv.Annotations[NODEIP]

	workerNode := os.Getenv("NODE_IP")
	if workerNode != volume.Nodeip {
		lgr.Info("The volume attach is not requested for this worker node")
		return false, fmt.Errorf("the volume attach is not requested for this worker node")
	}
	var input []byte
	var err error
	if input, err = os.ReadFile(BLOCK_CONF); err != nil {
		lgr.Error("Could not read iscsi-block-volume.conf file")
		return false, fmt.Errorf("could not read iscsi-block-volume.conf file. Error: %v", err)
	}
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
	if err = os.WriteFile(BLOCK_CONF, []byte(output), 0600); err != nil {
		lgr.Error("Could not write to iscsi-block-volume.conf file")
		return false, fmt.Errorf("could not write to iscsi-block-volume.conf file. Error: %v", err)
	}

	pvUpdated := false
	for x := 0; x < 5; x++ {
		//Adding sleep since kubernetes will be still modifying the PV object
		time.Sleep(5 * time.Second)

		//Fetch the latest version of the PV from Kubernetes apiserver
		latestPV, pvErr := clientset.CoreV1().PersistentVolumes().Get(context.TODO(), pv.Name, metav1.GetOptions{})
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
		_, pvErr = clientset.CoreV1().PersistentVolumes().Update(context.TODO(), latestPV, metav1.UpdateOptions{})
		if pvErr == nil {
			pvUpdated = true
			break
		}
		lgr.Warn("Failed to update PV from apiserver:", zap.String("pvname", pv.Name), zap.Error(pvErr))
	}
	if !pvUpdated {
		return true, fmt.Errorf("failed to update PV %s", pv.Name)
	}

	// Restart ibmc-block-attacher service so volume can be attached
	dbConn, connErr := dbus.New()
	if connErr != nil {
		lgr.Error("Error: Unable to connect!", zap.Error(connErr))
		return true, fmt.Errorf("error: Unable to connect. %v", connErr)
	}
	reschan := make(chan string)
	_, restartErr := dbConn.RestartUnit(ATTACHER_SERVICE, "fail", reschan)
	if restartErr != nil {
		lgr.Error("Error: Unable to restart target", zap.Error(restartErr))
		return true, fmt.Errorf("error: Unable to restart target. %v", restartErr)
	}
	lgr.Info("Unit Restarted !!")
	job := <-reschan
	if job != "done" {
		lgr.Error("Error: Restart of service is not done: " + job)
		return true, fmt.Errorf("error: Restart of service is not done")
	}
	retry, attErr := UpdatePersistentVolume(volume, pv)
	return retry, attErr
}

// UpdatePersistentVolume ...
func UpdatePersistentVolume(volume config.Volume, pv *v1.PersistentVolume) (bool, error) {
	folder := "/host/lib/ibmc-block-attacher"
	if val := os.Getenv("service_dir"); val != "" {
		folder = os.Getenv("service_dir")
	}
	pathsFile := folder + "/out_paths"
	mpathsFile := folder + "/out_multipaths"
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
		/* multipathd show paths format "%w %i %C"
		uuid                              hcil      next_check
		3600a09803830476d6f3f4f435751684f 20:0:0:37 orphan
		3600a09803830445455244c4a38752d30 10:0:0:15 XXXXXXXXX. 18/20 --> The last part of hcil is lun id
		3600a09803830445455244c4a38752d30 11:0:0:15 XXXX...... 9/20
		*/
		if input, err = os.ReadFile(pathsFile); err != nil {
			lgr.Error("Could not read " + pathsFile + " file")
		} else {
			lines := strings.Split(string(input), "\n")
			for _, line := range lines {
				space := regexp.MustCompile(`\s+`)
				line = space.ReplaceAllString(line, " ")
				lineParts := strings.Split(line, " ")
				lgr.Info("Line: ", zap.Strings("LINE", lineParts))
				// We ignore the orphan multipaths
				if (len(lineParts) >= 3) && (strings.TrimSpace(lineParts[2]) != "orphan") {
					// Parse the LUN ID from output
					lun := strings.Split(lineParts[1], ":")
					if len(lun) == 4 {
						lunid, err = strconv.Atoi(lun[3])
						if err != nil {
							return true, err
						}
						if lunid == volume.Lunid {
							mpath = lineParts[0]
							break
						}
					}
				}
			}
		}
		if len(mpath) == 0 {
			lgr.Error("Multipaths are taking time to load")
			return true, fmt.Errorf("multipaths are taking time to load for storage %s", volume.VolID)
		}

		// Parse multipaths to fetch device path
		/* multipathd show multipaths
		name                              sysfs uuid
		3600a09803830445455244c4a38752d30 dm-0  3600a09803830445455244c4a38752d30
		*/
		if input, err = os.ReadFile(mpathsFile); err != nil {
			lgr.Error("Could not read " + mpathsFile + " file")
		} else {
			lines := strings.Split(string(input), "\n")
			for _, line := range lines {
				space := regexp.MustCompile(`\s+`)
				line = space.ReplaceAllString(line, " ")
				lineParts := strings.Split(line, " ")
				lgr.Info("Mpath Line: ", zap.Strings("LINE", lineParts))
				lgr.Info("MPath: ", zap.String("MPATH", mpath))
				if lineParts[0] == mpath {
					// Device path sample is /dev/dm-0
					devicepath = "/dev/" + lineParts[1]
					break
				}
			}
		}
		if len(devicepath) == 0 {
			lgr.Error("Device path is taking time to load")
			return true, fmt.Errorf("device path is taking time to load for storage %s", volume.VolID)
		}
		lgr.Info("Device path and volume lun ID: ", zap.String("LUN_Id", strconv.Itoa(volume.Lunid)), zap.String("Device_Path", devicepath))

		// Delete the output files
		delErr := os.Remove(pathsFile)
		if delErr != nil {
			lgr.Error("Delete of "+pathsFile+" file failed ", zap.Error(delErr))
		}

		delErr = os.Remove(mpathsFile)
		if delErr != nil {
			lgr.Error("Delete of "+mpathsFile+" file failed ", zap.Error(delErr))
		}

		pvUpdated := false
		for x := 0; x < 5; x++ {
			//Adding sleep since kubernetes will be still modifying the PV object
			time.Sleep(5 * time.Second)

			//Fetch the latest version of the PV from Kubernetes apiserver
			latestPV, pvErr := clientset.CoreV1().PersistentVolumes().Get(context.TODO(), pv.Name, metav1.GetOptions{})
			if pvErr != nil {
				lgr.Warn("Failed to fetch PV from apiserver:", zap.String("pvname", pv.Name), zap.Error(pvErr))
				continue
			}
			// Update PV with devicepath and multipath
			latestPV.Annotations[DMPATH] = devicepath
			latestPV.Annotations[MULTIPATH] = mpath
			latestPV.Annotations[ATTACHSTATUS] = STATUS_ATTACHED
			_, pvErr = clientset.CoreV1().PersistentVolumes().Update(context.TODO(), latestPV, metav1.UpdateOptions{})
			if pvErr == nil {
				pvUpdated = true
				break
			}
			lgr.Warn("Failed to update PV from apiserver:", zap.String("pvname", pv.Name), zap.Error(pvErr))
		}
		if !pvUpdated {
			return true, fmt.Errorf("failed to update PV %s", pv.Name)
		}
		return false, nil
	}
	for x := 0; x < 5; x++ {
		//Adding sleep since kubernetes will be still modifying the PV object
		time.Sleep(5 * time.Second)

		//Fetch the latest version of the PV from Kubernetes apiserver
		latestPV, pvErr := clientset.CoreV1().PersistentVolumes().Get(context.TODO(), pv.Name, metav1.GetOptions{})
		if pvErr != nil {
			lgr.Warn("Failed to fetch PV from apiserver:", zap.String("pvname", pv.Name), zap.Error(pvErr))
			continue
		}
		latestPV.Annotations[ATTACHSTATUS] = STATUS_FAILED + " --- Issue in iscsi attach. Retrying..."
		_, pvErr = clientset.CoreV1().PersistentVolumes().Update(context.TODO(), latestPV, metav1.UpdateOptions{})
		if pvErr == nil {
			break
		}
		lgr.Warn("Failed to update PV from apiserver:", zap.String("pvname", pv.Name), zap.Error(pvErr))
	}
	return true, fmt.Errorf("error while attaching storage %s", volume.VolID)
}

// Validate ...
func Validate(pv *v1.PersistentVolume) (bool, error) {
	volDetails := make([]string, 0)
	if pv.Annotations == nil {
		lgr.Error("The PV has no volume details given to perform attach.")
		pv.Annotations[ATTACHSTATUS] = INVALID_PARAMS
	} else {
		if _, present := pv.Annotations[IQN]; !present {
			volDetails = append(volDetails, IQN)
			pv.Annotations[ATTACHSTATUS] = INVALID_PARAMS
		}
		if _, present := pv.Annotations[USERNAME]; !present {
			volDetails = append(volDetails, USERNAME)
			pv.Annotations[ATTACHSTATUS] = INVALID_PARAMS
		}
		if _, present := pv.Annotations[PASSWORD]; !present {
			volDetails = append(volDetails, PASSWORD)
			pv.Annotations[ATTACHSTATUS] = INVALID_PARAMS
		}
		if _, present := pv.Annotations[TARGET]; !present {
			volDetails = append(volDetails, TARGET)
			pv.Annotations[ATTACHSTATUS] = INVALID_PARAMS
		}
		if _, present := pv.Annotations[LUNID]; !present {
			volDetails = append(volDetails, LUNID)
			pv.Annotations[ATTACHSTATUS] = INVALID_PARAMS
		} else {
			if _, err := strconv.Atoi(pv.Annotations[LUNID]); err != nil {
				volDetails = append(volDetails, LUNID)
				pv.Annotations[ATTACHSTATUS] = INVALID_PARAMS
			}
		}
		if _, present := pv.Annotations[NODEIP]; !present {
			volDetails = append(volDetails, NODEIP)
			pv.Annotations[ATTACHSTATUS] = INVALID_PARAMS
		}
	}
	if pv.Annotations[ATTACHSTATUS] == INVALID_PARAMS {
		lgr.Warn("Either no annotations are given or the following volume attributes are not valid in the PV:", zap.Strings("vol_attach_attrs", volDetails))

		pvUpdated := false
		for x := 0; x < 5; x++ {
			time.Sleep(5 * time.Second)

			//Fetch the latest version of the PV from Kubernetes apiserver
			latestPV, err := clientset.CoreV1().PersistentVolumes().Get(context.TODO(), pv.Name, metav1.GetOptions{})
			if err != nil {
				lgr.Warn("Failed to fetch PV from apiserver:", zap.String("pvname", pv.Name), zap.Error(err))
				continue
			}
			latestPV.Annotations[ATTACHSTATUS] = INVALID_PARAMS
			_, err = clientset.CoreV1().PersistentVolumes().Update(context.TODO(), latestPV, metav1.UpdateOptions{})
			if err == nil {
				pvUpdated = true
				break
			}
			lgr.Error("Failed to update PV from apiserver:", zap.String("pvname", pv.Name), zap.Error(err))
		}
		if !pvUpdated {
			return true, fmt.Errorf("failed to update PV %s", pv.Name)
		}
		return false, fmt.Errorf("error while validating the PV annotations %s", pv.Name)
	}
	return false, nil
}

// DetachVolume ...
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

// ModifyDetachConfig ...
func ModifyDetachConfig(pv *v1.PersistentVolume) {
	lgr.Info("Waiting for mutex lock in DETACH", zap.String("Name", pv.Name))
	mutex.Lock()
	lgr.Info("Acquired mutex lock in DETACH", zap.String("Name", pv.Name))
	defer mutex.Unlock()

	lgr.Info("Volume to be detached: ", zap.String("Name", pv.Name))

	volDetails := make([]string, 0)
	if pv.Annotations == nil {
		lgr.Error("The PV has no volume details given to perform detach.")
		return
	}
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

	volume := config.Volume{}
	volume.Nodeip = pv.Annotations[NODEIP]
	workerNode := os.Getenv("NODE_IP")
	if workerNode != volume.Nodeip {
		lgr.Info("The volume detach is not requested for this worker node")
		return
	}

	devPath := strings.Split(pv.Annotations[DMPATH], "/")
	var input []byte
	var err error
	if input, err = os.ReadFile(BLOCK_CONF); err != nil {
		lgr.Error("Could not read iscsi-block-volume.conf file")
		return
	}
	lines := strings.Split(string(input), "\n")
	for i, line := range lines {
		if strings.Contains(line, "dm=") {
			lines[i] = "dm=" + strings.TrimSpace(devPath[2])
		} else if strings.Contains(line, "mpath=") {
			lines[i] = "mpath=" + strings.TrimSpace(pv.Annotations[MULTIPATH])
		} else if strings.Contains(line, "op=") {
			lines[i] = "op=" + strings.TrimSpace(DETACH)
		}
	}

	modifiedlines := []string{}
	modifiedlines = append(modifiedlines, lines...)
	output := strings.Join(modifiedlines, "\n")
	if err = os.WriteFile(BLOCK_CONF, []byte(output), 0600); err != nil {
		lgr.Error("Could not write to iscsi-block-volume.conf file")
		return
	}

	// Restart ibmc-block-attacher service so volume can be attached
	dbConn, connErr := dbus.New()
	if connErr != nil {
		lgr.Error("Error: Unable to connect!", zap.Error(connErr))
		return
	}
	reschan := make(chan string)
	_, restartErr := dbConn.RestartUnit(ATTACHER_SERVICE, "fail", reschan)
	if restartErr != nil {
		lgr.Error("Error: Unable to restart target", zap.Error(restartErr))
		return
	}
	lgr.Info("Unit Restarted !!")
	job := <-reschan
	if job != "done" {
		lgr.Error("Error: Restart of service is not done: " + job)
		return
	}
}
