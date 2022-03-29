/*******************************************************************************
 * IBM Confidential
 * OCO Source Materials
 * IBM Cloud Kubernetes Service, 5737-D43
 * (C) Copyright IBM Corp. 2022 All Rights Reserved.
 * The source code for this program is not published or otherwise divested of
 * its trade secrets, irrespective of what has been deposited with
 * the U.S. Copyright Office.
 ******************************************************************************/

// Package main ...
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
	logger, _ := logger.GetZapLogger() //#nosec G104 notCritical
	loggerLevel := zap.NewAtomicLevel()

	//Enable debug trace
	debugTrace := cfg.GetConfigBool("DEBUG_TRACE", false, *logger)
	if debugTrace {
		loggerLevel.SetLevel(zap.DebugLevel)
	}

	_ = flag.Set("logtostderr", "true") //#nosec G104 notCritical
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
