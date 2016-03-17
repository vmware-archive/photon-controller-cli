// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package client

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"

	cf "github.com/vmware/photon-controller-cli/photon/cli/configuration"
)

// Global variable pointing to photon client, can be assigned to mock client in tests
var Esxclient *photon.Client

var logger *log.Logger = nil
var logFile *os.File = nil

// Read from local config file and create a new photon client using target
func get() (*photon.Client, error) {
	var options *photon.ClientOptions
	config, err := cf.LoadConfig()
	if err != nil {
		return nil, err
	}

	if len(config.CloudTarget) == 0 {
		return nil, errors.New("Specify a Photon Controller endpoint by running 'target set' command")
	}

	options = &photon.ClientOptions{
		TokenOptions: &photon.TokenOptions{
			AccessToken: config.Token,
		},
	}

	//
	// If target is https, check if we could ignore client side cert check
	// If we can't ignore client side cert check, try setting the root certs
	//
	u, err := url.Parse(config.CloudTarget)
	if err == nil && u.Scheme == "https" {
		if config.IgnoreCertificate == true {
			options.IgnoreCertificate = true
		} else {
			roots, err := cf.GetCertsFromLocalStore()
			if err == nil {
				options.RootCAs = roots
			} else {
				return nil, err
			}
		}
	}

	esxclient := photon.NewClient(config.CloudTarget, options, logger)
	return esxclient, nil
}

// Returns the photon client, if not set, it will read a config file.
func GetClient(isScripting bool) (*photon.Client, error) {
	if Esxclient == nil {
		var err error
		Esxclient, err = get()
		if err != nil {
			return nil, err
		}
	}

	if !isScripting {
		fmt.Printf("Using target '%s'\n", Esxclient.Endpoint)
	}

	return Esxclient, nil
}

func InitializeLogging(logFileName string) error {
	var output io.Writer
	var err error
	if logFileName != "" {
		logFile, err = os.OpenFile(
			logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		output = logFile
		logger = log.New(
			output,
			"[photon-cli] ", // prefix for each log statement
			log.LstdFlags)   // standard flags. prints date and time for each log statement
	}

	return nil
}

func CleanupLogging() error {
	// Close the logging file if it was created
	// for Verbose logging
	if logFile != nil {
		err := logFile.Close()
		if err != nil {
			fmt.Println(err)
		}
		logFile = nil
		logger = nil
	}
	return nil
}
