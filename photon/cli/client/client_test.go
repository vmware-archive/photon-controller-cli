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
	"testing"

	cf "github.com/vmware/photon-controller-cli/photon/cli/configuration"
)

func TestGet(t *testing.T) {
	configOri, err := cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	// test Get when config file doesn't exist
	err = cf.RemoveConfigFile()
	if err != nil {
		t.Error("Not expecting error removing config file")
	}
	_, err = get()
	if err == nil {
		t.Error("Expected to receive error trying to get client when config file does not exist")
	}

	// test GetClient with valid endpoint in config file
	endpoint := "http://localhost:9080"
	testGetEndpoint(t, endpoint, false, true)

	endpoint = "https://172.31.253.66:443"
	//test Get for a https endpoint with secure endpoint = true & skipping verify = true
	testGetEndpoint(t, endpoint, true, true)

	//test Get for a https endpoint with secure endpoint = true & skipping verify = false
	testGetEndpoint(t, endpoint, true, false)

	//Restore the original configuration
	err = cf.SaveConfig(configOri)
	if err != nil {
		t.Error("Not expecting error when saving config file")
	}
}

func testGetEndpoint(t *testing.T, endpoint string, ephttps bool, skipVerify bool) {
	token := "fake-token"
	var configExpected *cf.Configuration
	if ephttps == false {
		//this is http case
		configExpected = &cf.Configuration{
			CloudTarget: endpoint,
			Token:       token,
		}
	} else if skipVerify == true {
		configExpected = &cf.Configuration{
			CloudTarget:       endpoint,
			Token:             token,
			IgnoreCertificate: true,
		}
	} else {
		configExpected = &cf.Configuration{
			CloudTarget:       endpoint,
			Token:             token,
			IgnoreCertificate: false,
		}
	}
	err := cf.SaveConfig(configExpected)
	if err != nil {
		t.Error("Not expecting error saving config file")
	}

	Esxclient, err = get()
	if err != nil {
		t.Error("Not expecting error trying to get client when config file has valid endpoint")
	}

	if Esxclient.Endpoint != endpoint {
		t.Error("Endpoint of client not match endpoint in config file")
	}
}
