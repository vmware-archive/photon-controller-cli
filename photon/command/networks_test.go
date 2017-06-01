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
	"os"
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
		server.URL+rootUrl+"/system/info",
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
		server.URL+rootUrl+"/system/info",
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

type MockNetworksPage struct {
	Items            []photon.Network `json:"items"`
	NextPageLink     string           `json:"nextPageLink"`
	PreviousPageLink string           `json:"previousPageLink"`
}

func TestCreateDeleteNetwork(t *testing.T) {
	tenantStruct := photon.Tenants{
		Items: []photon.Tenant{
			{
				Name: "fake_tenant_name",
				ID:   "fake_tenant_ID",
			},
		},
	}
	tenantResponse, err := json.Marshal(tenantStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected tenantStruct")
	}

	projectStruct := photon.ProjectList{
		Items: []photon.ProjectCompact{
			{
				Name: "fake_project_name",
				ID:   "fake_project_ID",
			},
		},
	}
	projectResponse, err := json.Marshal(projectStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected projectStruct")
	}

	queuedTask := &photon.Task{
		Operation: "CREATE_ROUTER",
		State:     "QUEUED",
		ID:        "fake-network-task-id",
		Entity:    photon.Entity{ID: "fake-network-id"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "CREATE_ROUTER",
		State:     "COMPLETED",
		ID:        "fake-network-task-id",
		Entity:    photon.Entity{ID: "fake-network-id"},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	server = mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tenants",
		mocks.CreateResponder(200, string(tenantResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tenants/"+"fake_tenant_ID"+"/projects?name="+"fake_project_name",
		mocks.CreateResponder(200, string(projectResponse[:])))
	mocks.RegisterResponder(
		"POST",
		server.URL+rootUrl+"/projects/"+"fake_project_ID"+"/networks",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
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
	set.String("name", "fake_network_name", "Network name")
	set.String("privateIpCidr", "fake_network_privateIpCidr", "Network privateIpCidr")
	set.String("tenant", "fake_tenant_name", "tenant name")
	set.String("project", "fake_project_name", "project name")
	cxt := cli.NewContext(nil, set, globalCtx)

	err = createNetwork(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error creating Network: " + err.Error())
	}

	queuedTask = &photon.Task{
		Operation: "DELETE_ROUTER",
		State:     "QUEUED",
		ID:        "fake-network-task-id",
		Entity:    photon.Entity{ID: "fake-network-id"},
	}
	taskResponse, err = json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}

	completedTask = &photon.Task{
		Operation: "DELETE_ROUTER",
		State:     "COMPLETED",
		ID:        "fake-network-task-id",
		Entity:    photon.Entity{ID: "fake-network-id"},
	}
	response, err = json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completedTask")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+rootUrl+"/networks/"+"fake-network-id",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set = flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake-network-id"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	cxt = cli.NewContext(nil, set, nil)
	err = deleteNetwork(cxt)
	if err != nil {
		t.Error("Not expecting error deleting network: " + err.Error())
	}
}

func TestUpdateNetwork(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "UPDATE_ROUTER",
		State:     "QUEUED",
		ID:        "fake-task-id",
		Entity:    photon.Entity{ID: "fake-network-id"},
	}
	completedTask := &photon.Task{
		Operation: "UPDATE_ROUTER",
		State:     "COMPLETED",
		ID:        "fake-task-id",
		Entity:    photon.Entity{ID: "fake-network-id"},
	}

	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected createTask")
	}
	taskResponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"PATCH",
		server.URL+rootUrl+"/networks"+"/fake-network-id",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskResponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake-network-id"})
	set.String("name", "network-1", "network name")
	cxt := cli.NewContext(nil, set, nil)

	err = updateNetwork(cxt, os.Stdout)
	if err != nil {
		t.Error("Error listing networks: " + err.Error())
	}
}

func TestShowNetwork(t *testing.T) {
	getStruct := photon.Network{
		Name:          "networkname",
		ID:            "1",
		Kind:          "network",
		PrivateIpCidr: "cidr1",
		IsDefault:     false,
	}

	response, err := json.Marshal(getStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/networks/"+getStruct.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{getStruct.ID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = showNetwork(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting get deployment to fail")
	}
}

func TestListNetworks(t *testing.T) {
	tenantStruct := photon.Tenants{
		Items: []photon.Tenant{
			{
				Name: "fake_tenant_name",
				ID:   "fake_tenant_ID",
			},
		},
	}
	tenantResponse, err := json.Marshal(tenantStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected tenantStruct")
	}

	projectStruct := photon.ProjectList{
		Items: []photon.ProjectCompact{
			{
				Name: "fake_project_name",
				ID:   "fake_project_ID",
			},
		},
	}
	projectResponse, err := json.Marshal(projectStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected projectStruct")
	}

	networkList := MockNetworksPage{
		Items: []photon.Network{
			{
				Name:          "fake_network_name",
				ID:            "fake_network_ID",
				Kind:          "fake_network_kind",
				PrivateIpCidr: "fake_cidr",
				IsDefault:     false,
			},
		},
		NextPageLink:     "fake-next-page-link",
		PreviousPageLink: "",
	}
	listResponse, err := json.Marshal(networkList)
	if err != nil {
		t.Error("Not expecting error serializaing expected networkList")
	}

	server := mocks.NewTestServer()
	defer server.Close()

	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tenants",
		mocks.CreateResponder(200, string(tenantResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tenants/"+"fake_tenant_ID"+"/projects?name="+"fake_project_name",
		mocks.CreateResponder(200, string(projectResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/projects/"+"fake_project_ID"+"/networks",
		mocks.CreateResponder(200, string(listResponse[:])))

	networkList = MockNetworksPage{
		Items:            []photon.Network{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}
	listResponse, err = json.Marshal(networkList)
	if err != nil {
		t.Error("Not expecting error serializaing expected networkList")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"fake-next-page-link",
		mocks.CreateResponder(200, string(listResponse[:])))

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	set.String("tenant", "fake_tenant_name", "tenant name")
	set.String("project", "fake_project_name", "project name")
	cxt := cli.NewContext(nil, set, nil)

	err = listNetworks(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error listing networks: " + err.Error())
	}
}
