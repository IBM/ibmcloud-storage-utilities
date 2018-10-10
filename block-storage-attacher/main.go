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
	"flag"
	cfg "github.ibm.com/alchemy-containers/block-storage-attacher/utils/config"
	"github.ibm.com/alchemy-containers/block-storage-attacher/utils/logger"
	"github.ibm.com/alchemy-containers/block-storage-attacher/watcher"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	// Logger
	logger, _ := logger.GetZapLogger()
	loggerLevel := zap.NewAtomicLevel()

	//Enable debug trace
	debug_trace := cfg.GetConfigBool("DEBUG_TRACE", false, *logger)
	if debug_trace {
		loggerLevel.SetLevel(zap.DebugLevel)
	}

	var config *rest.Config
	var err error
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
