// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

type MockSubnetsPage struct {
	Items            []photon.Subnet `json:"items"`
	NextPageLink     string          `json:"nextPageLink"`
	PreviousPageLink string          `json:"previousPageLink"`
}

func TestCreateDeleteVirtualSubnet(t *testing.T) {
	routerStruct := photon.Router{
		Name: "fake_router_name",
		ID:   "fake_router_ID",
	}
	routerResponse, err := json.Marshal(routerStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected tenantStruct")
	}

	queuedTask := &photon.Task{
		Operation: "CREATE_SUBNET",
		State:     "QUEUED",
		ID:        "fake-subnet-task-id",
		Entity:    photon.Entity{ID: "fake-subnet-id"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "CREATE_SUBNET",
		State:     "COMPLETED",
		ID:        "fake-subnet-task-id",
		Entity:    photon.Entity{ID: "fake-subnet-id"},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	server = mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/routers/"+"fake_router_ID",
		mocks.CreateResponder(200, string(routerResponse[:])))
	mocks.RegisterResponder(
		"POST",
		server.URL+rootUrl+"/routers/"+"fake_router_ID"+"/subnets",
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
	set.String("name", "fake_subnet_name", "Subnet name")
	set.String("description", "test subnet", "Subnet description")
	set.String("privateIpCidr", "fake_subnet_privateIpCidr", "Subnet privateIpCidr")
	set.String("router", "fake_router_ID", "Router id")
	set.String("dns-server-addresses", "dnsServerAddress1,dnsServerAddress2", "Comma separated DNS server "+
		"addresses")

	cxt := cli.NewContext(nil, set, globalCtx)

	err = createSubnet(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error creating subnet: " + err.Error())
	}

	queuedTask = &photon.Task{
		Operation: "DELETE_SUBNET",
		State:     "QUEUED",
		ID:        "fake-subnet-task-id",
		Entity:    photon.Entity{ID: "fake-subnet-id"},
	}
	taskResponse, err = json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}

	completedTask = &photon.Task{
		Operation: "DELETE_SUBNET",
		State:     "COMPLETED",
		ID:        "fake-subnet-task-id",
		Entity:    photon.Entity{ID: "fake-subnet-id"},
	}
	response, err = json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completedTask")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+rootUrl+"/subnets/"+"fake-subnet-id",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set = flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake-subnet-id"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	cxt = cli.NewContext(nil, set, nil)
	err = deleteSubnet(cxt)
	if err != nil {
		t.Error("Not expecting error deleting subnet: " + err.Error())
	}
}

func TestCreateDeletePhysicalSubnet(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "CREATE_SUBNET",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "subnet-ID"},
	}
	completedTask := &photon.Task{
		Operation: "CREATE_SUBNET",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "subnet-ID"},
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
	set.String("name", "subnet_name", "subnet name")
	set.String("portgroups", "portgroup, portgroup2", "portgroups")

	cxt := cli.NewContext(nil, set, globalCtx)

	err = createSubnet(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting create subnet to fail", err)
	}

	queuedTask = &photon.Task{
		Operation: "DELETE_SUBNET",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "subnet-ID"},
	}
	completedTask = &photon.Task{
		Operation: "DELETE_SUBNET",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "subnet-ID"},
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
		t.Error("Not expecting delete subnet to fail")
	}
}

func TestUpdateSubnet(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "UPDATE_SUBNET",
		State:     "QUEUED",
		ID:        "fake-task-id",
		Entity:    photon.Entity{ID: "fake-subnet-id"},
	}
	completedTask := &photon.Task{
		Operation: "UPDATE_SUBNET",
		State:     "COMPLETED",
		ID:        "fake-task-id",
		Entity:    photon.Entity{ID: "fake-subnet-id"},
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
		server.URL+rootUrl+"/subnets"+"/fake-subnet-id",
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
	err = set.Parse([]string{"fake-subnet-id"})
	set.String("name", "subnet-1", "subnet name")
	cxt := cli.NewContext(nil, set, nil)

	err = updateSubnet(cxt, os.Stdout)
	if err != nil {
		t.Error("Error updating subnet: " + err.Error())
	}
}

func TestShowSubnet(t *testing.T) {
	getStruct := photon.Subnet{
		Name:          "subnetname",
		ID:            "1",
		Description:   "test subnet",
		Kind:          "subnet",
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
		server.URL+rootUrl+"/subnets/"+getStruct.ID,
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

	err = showSubnet(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting get deployment to fail")
	}
}

func TestListSubnetsUnderRouter(t *testing.T) {
	subnetList := MockSubnetsPage{
		Items: []photon.Subnet{
			{
				Name:          "fake_subnet_name",
				ID:            "fake_subnet_ID",
				Kind:          "fake_subnet_kind",
				PrivateIpCidr: "fake_cidr",
				State:         "READY",
				IsDefault:     false,
			},
		},
		NextPageLink:     "fake-next-page-link",
		PreviousPageLink: "",
	}
	listResponse, err := json.Marshal(subnetList)
	if err != nil {
		t.Error("Not expecting error serializaing expected subnetList")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/routers/"+"fake-router-id"+"/subnets",
		mocks.CreateResponder(200, string(listResponse[:])))

	subnetList = MockSubnetsPage{
		Items:            []photon.Subnet{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}
	listResponse, err = json.Marshal(subnetList)
	if err != nil {
		t.Error("Not expecting error serializaing expected subnetList")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"fake-next-page-link",
		mocks.CreateResponder(200, string(listResponse[:])))

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	set.String("router", "fake-router-id", "Router id")
	cxt := cli.NewContext(nil, set, nil)

	err = listSubnets(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error listing subnets: " + err.Error())
	}
}

func TestListSubnetsUnderNetwork(t *testing.T) {
	subnetList := MockSubnetsPage{
		Items: []photon.Subnet{
			{
				Name:          "fake_subnet_name",
				ID:            "fake_subnet_ID",
				Kind:          "fake_subnet_kind",
				PrivateIpCidr: "fake_cidr",
				State:         "READY",
				IsDefault:     false,
			},
		},
		NextPageLink:     "fake-next-page-link",
		PreviousPageLink: "",
	}
	listResponse, err := json.Marshal(subnetList)
	if err != nil {
		t.Error("Not expecting error serializaing expected subnetList")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/networks/"+"fake-network-id"+"/subnets",
		mocks.CreateResponder(200, string(listResponse[:])))

	subnetList = MockSubnetsPage{
		Items:            []photon.Subnet{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}
	listResponse, err = json.Marshal(subnetList)
	if err != nil {
		t.Error("Not expecting error serializaing expected subnetList")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"fake-next-page-link",
		mocks.CreateResponder(200, string(listResponse[:])))

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	set.String("network", "fake-network-id", "Network id")
	cxt := cli.NewContext(nil, set, nil)

	err = listSubnets(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error listing subnets: " + err.Error())
	}
}

func TestSetDefaultSubnet(t *testing.T) {
	completedTask := &photon.Task{
		Operation: "SET_DEFAULT_SUBNET",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "id"},
	}

	taskresponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completedTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+rootUrl+"/subnets/"+completedTask.Entity.ID+"/set_default",
		mocks.CreateResponder(200, string(taskresponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+completedTask.ID,
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
	err = set.Parse([]string{completedTask.Entity.ID})
	cxt := cli.NewContext(nil, set, globalCtx)

	err = setDefaultSubnet(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting set default subnet to fail", err)
	}
}

func TestListPortGroups(t *testing.T) {
	server := mocks.NewTestServer()
	defer server.Close()

	expectedList := MockSubnetsPage{
		Items: []photon.Subnet{
			{
				ID:         "subnet_id",
				Name:       "subnet_name",
				PortGroups: photon.PortGroups{Names: []string{"port", "group"}},
				IsDefault:  false,
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
		server.URL+rootUrl+"/subnets",
		mocks.CreateResponder(200, string(response[:])))

	expectedList = MockSubnetsPage{
		Items:            []photon.Subnet{},
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
	err = commandFlags.Parse([]string{})
	if err != nil {
		t.Error(err)
	}
	cxt := cli.NewContext(nil, commandFlags, globalCxt)
	var output bytes.Buffer

	err = listSubnets(cxt, &output)
	if err != nil {
		t.Error("Error listing subnets: " + err.Error())
	}
}
