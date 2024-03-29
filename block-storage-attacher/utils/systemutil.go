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
	"fmt"

	"github.com/coreos/go-systemd/v22/dbus"
)

const SYSTEMD_UNIT_FILE_PATH = "/host/systemd/system/" //nolint readability

func main() {
	var target = flag.String("target", "kubelet.service", "The name of systemctl unit/ service")
	var action = flag.String("action", "restart", "The action ( start/stop/restart/enable) for the service")
	flag.Parse()
	fmt.Println("parameters -target = ", *target, " -action = ", *action)

	dbConn, connErr := dbus.New()
	if connErr != nil {
		fmt.Println("Error: Unable to connect!", connErr)
		return
	}

	reschan := make(chan string)

	if *action == "reload" {
		reloadErr := dbConn.Reload()
		if reloadErr != nil {
			fmt.Println("Error: Unable to reload daemon!", reloadErr)
		} else {
			fmt.Println("Info: Daemon reload done!")
		}
	}
	if *action == "start" {
		_, startErr := dbConn.StartUnit(*target, "fail", reschan)
		if startErr != nil {
			fmt.Println("Error: Unable to start target", startErr)
			return
		}
		fmt.Println("Unit started !!")
		job := <-reschan
		if job != "done" {
			fmt.Print("Error: Start of service is not done:", job)
		}
	}
	if *action == "restart" {
		_, restartErr := dbConn.RestartUnit(*target, "fail", reschan)
		if restartErr != nil {
			fmt.Println("Error: Unable to restart target", restartErr)
			return
		}
		fmt.Println("Unit Restarted !!")
		job := <-reschan
		if job != "done" {
			fmt.Print("Error: Restart of service is not done:", job)
		}
	}
	if *action == "enable" {
		// This does NOT work from container
		unitFile := SYSTEMD_UNIT_FILE_PATH + *target
		fmt.Println("Trying to enable unit file ", unitFile)
		_, change, err := dbConn.EnableUnitFiles([]string{unitFile}, false, true)
		if err != nil {
			fmt.Println("Error: Unable to enable target ", unitFile, err)
			return
		}
		fmt.Println("Unit  enabled  !!", *target, "Change:", change)
	}
	dbConn.Close()
}
