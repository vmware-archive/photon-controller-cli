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
	"bytes"
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

type MockVirtualNetworksPage struct {
	Items            []photon.VirtualSubnet `json:"items"`
	NextPageLink     string                 `json:"nextPageLink"`
	PreviousPageLink string                 `json:"previousPageLink"`
}

func TestCreateVirtualSubnet(t *testing.T) {
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
		server.URL+"/projects/project_id/subnets",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
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
	set.String("routingType", "ROUTED", "routing type")
	set.Int("size", 256, "network size")
	set.Int("staticIpSize", 0, "reserved static ip size")
	set.String("projectId", "project_id", "project ID")

	cxt := cli.NewContext(nil, set, globalCtx)

	err = createVirtualNetwork(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting create network to fail", err)
	}
}

func TestDeleteVirtualSubnet(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "DELETE_NETWORK",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "network-ID"},
	}
	completedTask := &photon.Task{
		Operation: "DELETE_NETWORK",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "network-ID"},
	}
	info := &photon.Info{
		NetworkType: SOFTWARE_DEFINED,
	}

	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}
	taskresponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completedTask")
	}
	infoString, err := json.Marshal(info)
	if err != nil {
		t.Error("Not expecting error when serializing expected info")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/subnets/"+queuedTask.Entity.ID,
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/info",
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

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{queuedTask.Entity.ID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, globalCtx)
	err = deleteNetwork(cxt)
	if err != nil {
		t.Error("Not expecting delete network to fail")
	}
}

func TestListVirtualNetworks(t *testing.T) {
	server := mocks.NewTestServer()
	defer server.Close()

	expectedList := MockVirtualNetworksPage{
		Items: []photon.VirtualSubnet{
			{
				ID:            "network_id",
				Name:          "network_name",
				IsDefault:     false,
				State:         "READY",
				RoutingType:   "ROUTED",
				Cidr:          "192.168.0.0/24",
				LowIpDynamic:  "192.168.0.1",
				HighIpDynamic: "192.168.0.200",
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}

	response, err := json.Marshal(expectedList)
	if err != nil {
		t.Error("Not expecting error serializing expected response")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/projects/project_id/subnets",
		mocks.CreateResponder(200, string(response[:])))

	expectedList = MockVirtualNetworksPage{
		Items:            []photon.VirtualSubnet{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}

	response, err = json.Marshal(expectedList)
	if err != nil {
		t.Error("Not expecting error serializing expected response")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/fake-next-page-link",
		mocks.CreateResponder(200, string(response[:])))

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalFlags := flag.NewFlagSet("global-flags", flag.ContinueOnError)
	globalFlags.String("output", "json", "output")
	err = globalFlags.Parse([]string{"--output=json"})
	if err != nil {
		t.Error(err)
	}
	globalCxt := cli.NewContext(nil, globalFlags, nil)
	commandFlags := flag.NewFlagSet("command-flags", flag.ContinueOnError)
	commandFlags.String("projectId", "project_id", "project ID")
	err = commandFlags.Parse([]string{})
	if err != nil {
		t.Error(err)
	}
	cxt := cli.NewContext(nil, commandFlags, globalCxt)
	var output bytes.Buffer

	err = listVirtualNetworks(cxt, &output)
	if err != nil {
		t.Error("Error listing networks: " + err.Error())
	}

	// Verify we printed a list of networks starting with a bracket
	err = checkRegExp(`^\s*\[`, output)
	if err != nil {
		t.Errorf("List networks didn't produce a JSON list that starts with a bracket (list): %s", err)
	}
	// and end with a bracket (two regular expressions because it's multiline, it's easier)
	err = checkRegExp(`\]\s*$`, output)
	if err != nil {
		t.Errorf("List networks didn't produce JSON that ended in a bracket (list): %s", err)
	}
	// And spot check that we have the "id" field
	err = checkRegExp(`\"id\":\s*\".*\"`, output)
	if err != nil {
		t.Errorf("List networks didn't produce a JSON field named 'id': %s", err)
	}
}

func TestShowVirtuallNetwork(t *testing.T) {
	expectedVirtualSubnet := photon.VirtualSubnet{
		ID:            "network_id",
		Name:          "network_name",
		IsDefault:     false,
		State:         "READY",
		RoutingType:   "ROUTED",
		Cidr:          "192.168.0.0/24",
		LowIpDynamic:  "192.168.0.1",
		HighIpDynamic: "192.168.0.200",
	}

	response, err := json.Marshal(expectedVirtualSubnet)
	if err != nil {
		t.Error("Not expecting error serializing expected response")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/subnets/"+expectedVirtualSubnet.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{expectedVirtualSubnet.ID})
	cxt := cli.NewContext(nil, set, nil)
	err = showVirtualNetwork(cxt, os.Stdout)
	if err != nil {
		t.Error("Error showing networks: " + err.Error())
	}
}

func TestSetDefaultVirtualNetwork(t *testing.T) {
	completedTask := &photon.Task{
		Operation: "SET_DEFAULT_NETWORK",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "id"},
	}

	taskresponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completedTask")
	}

	info := &photon.Info{
		NetworkType: SOFTWARE_DEFINED,
	}

	infoString, err := json.Marshal(info)
	if err != nil {
		t.Error("Not expecting error when serializing expected info")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+"/subnets/"+completedTask.Entity.ID+"/set_default",
		mocks.CreateResponder(200, string(taskresponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+completedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/info",
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

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{completedTask.Entity.ID})
	cxt := cli.NewContext(nil, set, globalCtx)

	err = setDefaultNetwork(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting set default network to fail", err)
	}
}
