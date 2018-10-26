package watcher

import (

	//"encoding/json"
	//"github.com/coreos/go-systemd/dbus"
	"github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/utils/config"
	"github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/utils/logger"
	"go.uber.org/zap"
	//"k8s.io/apimachinery/pkg/util/wait"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apimachinery/pkg/api/resource"
	//"k8s.io/client-go/kubernetes"
	//types "k8s.io/apimachinery/pkg/types"
	//"k8s.io/client-go/pkg/api/v1"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"
	//"k8s.io/client-go/tools/cache"
	"io/ioutil"
	"os"
	"testing"
)

var SUBNET_CONFIG = "subnetconfig"
var SUBNET_NS = "ibm_namespace"

const annDynamicallyProvisioned = "pv.kubernetes.io/provisioned-by"

func TestAttachVolume(t *testing.T) {

	nlgr, _ := logger.GetZapLogger()

	lgr = *nlgr

	lgr.Info("Testing attachVolume")

	pv := newVolumeType("volume-1")
	clientset = fake.NewSimpleClientset()

	objs := []runtime.Object{pv}
	AttachVolume(objs)

	createOutPathsFile()
	createMultiPathsFile()
	pv = newVolumeType("volume-1")
	objs = []runtime.Object{pv}

	clientset = fake.NewSimpleClientset(objs...)
	AttachVolume(objs)

}

func TestUpdatePersistentVolume(t *testing.T) {

	var volume config.Volume

	pv := newVolumeType("volume-1")
	objs := []runtime.Object{pv}
	lgr.Info("Testing UpdatePersistentVolume")
	volume = config.Volume{

		Iqn:      "testiq",
		Username: "testusername",
		Password: "testpassword",
		Target:   "testtarget",
		Lunid:    2,
		Nodeip:   "9.2.3.4",
	}
	createOutPathsFile()
	createMultiPathsFile()

	clientset = fake.NewSimpleClientset(objs...)
	UpdatePersistentVolume(volume, pv)

}
func TestModifyAttachConfig(t *testing.T) {
	lgr.Info("Testing TestModifyAttachConfig")

	pv := newVolumeType("volume-1")
	objs := []runtime.Object{pv}
	clientset = fake.NewSimpleClientset(objs...)

	/* Node IP is not matching */
	ModifyAttachConfig(pv)
	os.Setenv("NODE_IP", "9.2.3.4")

	/* Config file is not present */
	ModifyAttachConfig(pv)

	/* Config file with write error */

	createConfigFileWithError()
	ModifyAttachConfig(pv)

	createConfigFile()
	createOutPathsFile()
	createMultiPathsFile()
	clientset = fake.NewSimpleClientset(objs...)
	ModifyAttachConfig(pv)

}
func TestModifyDetachConfig(t *testing.T) {
	lgr.Info("Testing TestModifyDttachConfig")

	pv := newVolumeType("volume-1")
	os.Setenv("NODE_IP", "9.2.3.4")

	createOutPathsFile()
	createMultiPathsFile()
	objs := []runtime.Object{pv}
	clientset = fake.NewSimpleClientset(objs...)
	ModifyDetachConfig(pv)

}
func TestDetachVolume(t *testing.T) {
	lgr.Info("Testing DetachVolume")

	pv := newVolumeType("volume-1")

	os.Setenv("NODE_IP", "9.2.3.4")

	createOutPathsFile()
	createMultiPathsFile()
	objs := []runtime.Object{pv}
	clientset = fake.NewSimpleClientset(objs...)
	DetachVolume(objs)

}

func TestValidate(t *testing.T) {

	lgr.Info("Testing Validate Method")

	pv := NoAnnotationVolumeType("emptyvolume")

	objs := []runtime.Object{pv}
	clientset = fake.NewSimpleClientset(objs...)

	err := Validate(pv)
	assert.NotNil(t, err)
	os.RemoveAll("/host/etc/iscsi-portworx-volume.conf")

}

func newVolumeType(name string) *v1.PersistentVolume {
	//Create a volume
	labels := map[string]string{}
	labels["volumeId"] = "12345"
	labels["Username"] = "IBM_12345"
	labels["server"] = "testserver.nfs.com"
	labels["path"] = "/IBM_12345/data01"
	labels["CapacityGb"] = "20"
	labels["Iops"] = "4"
	annotations := map[string]string{"ibm.io/iqn": "iqn.2018-04.com.ibm:ibm02su1186049-i71967875", "ibm.io/username": "testusername", "ibm.io/password": "testpassword", "ibm.io/targetip": "1.1.1.2", "ibm.io/lunid": "2", "ibm.io/nodeip": "9.2.3.4", "ibm.io/dm": "/dev/dm-0", "ibm.io/mpath": "/host/lib/ibmc-portworx/out_multipaths"}

	return newVolume(name, v1.VolumeReleased, v1.PersistentVolumeReclaimDelete, labels, annotations)
}
func NoAnnotationVolumeType(name string) *v1.PersistentVolume {
	//Create a volume
	labels := map[string]string{}
	labels["volumeId"] = "12345"
	labels["Username"] = "IBM_12345"
	labels["server"] = "testserver.nfs.com"
	labels["path"] = "/IBM_12345/data01"
	labels["CapacityGb"] = "20"
	labels["Iops"] = "4"
	//annotations := map[string]string{"ibm.io/iqn": "", "ibm.io/username": "", "ibm.io/password": "", "ibm.io/targetip": "", "ibm.io/lunid": "", "ibm.io/nodeip": "", "ibm.io/attachstatus": ""}
	annotations := map[string]string{}

	return newVolume(name, v1.VolumeReleased, v1.PersistentVolumeReclaimDelete, labels, annotations)
}
func newVolume(name string, phase v1.PersistentVolumePhase, policy v1.PersistentVolumeReclaimPolicy, labels map[string]string, annotations map[string]string) *v1.PersistentVolume {
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: policy,
			AccessModes:                   []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce, v1.ReadOnlyMany},
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse("20Gi"),
			},
			PersistentVolumeSource: v1.PersistentVolumeSource{
				NFS: &v1.NFSVolumeSource{
					Server:   "nfsserver.softlayer.com",
					Path:     "/IBM_12345/data01",
					ReadOnly: false,
				},
			},
			ClaimRef: &v1.ObjectReference{
				Kind:      "PersistentVolumeClaim",
				Namespace: "default",
				Name:      "pvcNameXYZ",
			},
			//storageClassName: "pxclass",
			//hostpath: "/",

		},
		Status: v1.PersistentVolumeStatus{
			Phase: phase,
		},
	}

	return pv
}

func getNewLogger() zap.Logger {
	nlgr, _ := logger.GetZapLogger()
	return *nlgr
}

func createOutPathsFile() {
	serviceDir := "/tmp/host/lib/ibmc-portworx"
	os.Setenv("service_dir", serviceDir)
	os.MkdirAll(serviceDir, 0777)
	//	os.Mkdir("/host", 0777)
	//	os.Mkdir("/host/lib", 0777)
	//	os.Mkdir("/host/lib/ibmc-portworx", 0777)

	data := []byte("3600a09803830445455244c4a38752d30 10:0:0:2")
	out_paths := serviceDir + "/out_paths"
	os.Create(out_paths)
	ioutil.WriteFile(out_paths, data, 0666)
}
func createMultiPathsFile() {
	serviceDir := "/tmp/host/lib/ibmc-portworx"
	os.Setenv("service_dir", serviceDir)
	os.MkdirAll(serviceDir, 0777)
	//	os.Mkdir("/host", 0777)
	//	os.Mkdir("/host/lib", 0777)
	//	os.Mkdir("/host/lib/ibmc-portworx", 0777)

	data := []byte("3600a09803830445455244c4a38752d30 dm-0  3600a09803830445455244c4a38752d30")
	out_multipaths := serviceDir + "/out_multipaths"
	os.Create(out_multipaths)
	ioutil.WriteFile(out_multipaths, data, 0666)
}
func createConfigFile() {
	os.Mkdir("/host", 0777)
	os.Mkdir("/host/etc", 0777)

	data := []byte("iqn=iqn.2018-04.com.ibm:ibm02su1186049-i71967875\nusername=testusername\npassword=testpassword\ntarget_ip=1.1.1.2\nlunid=2\nnode_ip=9.2.3.4\nop=detach\ndm=dm-0\nmpath=/host/lib/ibmc-portworx/out_multipaths")
	os.Create("/host/etc/iscsi-portworx-volume.conf")
	ioutil.WriteFile("/host/etc/iscsi-portworx-volume.conf", data, 0666)
}
func createConfigFileWithError() {
	os.Mkdir("/host", 0644)
	os.Mkdir("/host/etc", 0644)

	//   data := []byte("iqn=iqn.2018-04.com.ibm:ibm02su1186049-i71967875\nusername=testusername\npassword=testpassword\ntarget_ip=1.1.1.2\nlunid=2\nnode_ip=9.2.3.4\nop=detach\ndm=dm-0\nmpath=/host/lib/ibmc-portworx/out_multipaths")
	os.Create("/host/etc/iscsi-portworx-volume.conf")
	os.Chmod("/host/etc/iscsi-portworx-volume.conf", 0400)
	// ioutil.WriteFile("/host/etc/iscsi-portworx-volume.conf", data, 0000)
}
