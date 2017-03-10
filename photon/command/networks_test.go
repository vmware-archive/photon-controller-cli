// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package command

import (
	"encoding/json"
	"flag"
	"net/http"
	"testing"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

func TestCheckSoftwareDefinedNetworkEnabled(t *testing.T) {
	info := &photon.Info{
		NetworkType: SOFTWARE_DEFINED,
	}
	infoString, err := json.Marshal(info)
	if err != nil {
		t.Error("Not expecting error when serializing info")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/info",
		mocks.CreateResponder(200, string(infoString[:])))

	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", true, "doc")
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	cxt := cli.NewContext(nil, nil, globalCtx)

	sdnEnabled, err := isSoftwareDefinedNetwork(cxt)
	if err != nil {
		t.Error("Not expecting checking if a network is software-defined to fail", err)
	}
	if !sdnEnabled {
		t.Error("This network should be software-defined")
	}
}

func TestCheckNetworkTypeNotDefined(t *testing.T) {
	info := &photon.Info{
		NetworkType: NOT_AVAILABLE,
	}
	infoString, err := json.Marshal(info)
	if err != nil {
		t.Error("Not expecting error when serializing info")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/info",
		mocks.CreateResponder(200, string(infoString[:])))

	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", true, "doc")
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	cxt := cli.NewContext(nil, nil, globalCtx)

	expectedErrMsg := "Network type is missing"
	_, err = isSoftwareDefinedNetwork(cxt)
	if err == nil || err.Error() != expectedErrMsg {
		t.Error("Error should have happened due to missing network type")
	}
}
