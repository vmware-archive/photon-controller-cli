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
	"os"
	"testing"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

type MockNetworksPage struct {
	Items            []photon.Subnet `json:"items"`
	NextPageLink     string          `json:"nextPageLink"`
	PreviousPageLink string          `json:"previousPageLink"`
}

func TestCreateDeletePhysicalNetwork(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "CREATE_NETWORK",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "network-ID"},
	}
	completedTask := &photon.Task{
		Operation: "CREATE_NETWORK",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "network-ID"},
	}
	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}
	taskresponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completedTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+rootUrl+"/subnets",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
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
	set := flag.NewFlagSet("test", 0)
	set.String("name", "network_name", "network name")
	set.String("portgroups", "portgroup, portgroup2", "portgroups")

	cxt := cli.NewContext(nil, set, globalCtx)

	err = createPhysicalNetwork(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting create network to fail", err)
	}

	queuedTask = &photon.Task{
		Operation: "DELETE_NETWORK",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "network-ID"},
	}
	completedTask = &photon.Task{
		Operation: "DELETE_NETWORK",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "network-ID"},
	}
	info := &photon.Info{
		NetworkType: PHYSICAL,
	}

	response, err = json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}
	taskresponse, err = json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completedTask")
	}
	infoString, err := json.Marshal(info)
	if err != nil {
		t.Error("Not expecting error when serializing expected info")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+rootUrl+"/subnets/"+queuedTask.Entity.ID,
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/info",
		mocks.CreateResponder(200, string(infoString[:])))

	set = flag.NewFlagSet("test", 0)
	err = set.Parse([]string{queuedTask.Entity.ID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt = cli.NewContext(nil, set, globalCtx)
	err = deleteSubnet(cxt)
	if err != nil {
		t.Error("Not expecting delete network to fail")
	}
}
