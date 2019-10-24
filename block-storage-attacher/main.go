package main

import (
	"flag"
	cfg "github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/utils/config"
	"github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/utils/logger"
	"github.com/IBM/ibmcloud-storage-utilities/block-storage-attacher/watcher"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {

	var err error

	// Logger
	logger, _ := logger.GetZapLogger()
	loggerLevel := zap.NewAtomicLevel()

	//Enable debug trace
	debug_trace := cfg.GetConfigBool("DEBUG_TRACE", false, *logger)
	if debug_trace {
		loggerLevel.SetLevel(zap.DebugLevel)
	}

	_ = flag.Set("logtostderr", "true")
	flag.Parse()

	var config *rest.Config
	config, err = clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		logger.Fatal("Failed to create config:", zap.Error(err))
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Fatal("Failed to create client:", zap.Error(err))
	}

	// Start watcher for volumes on config map
	watcher.WatchPersistentVolumes(clientset, *logger)
}
