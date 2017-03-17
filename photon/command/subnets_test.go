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

func TestCreateDeleteSubnet(t *testing.T) {
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
		server.URL+"/subnets/"+"fake-subnet-id",
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
		"PUT",
		server.URL+"/subnets"+"/fake-subnet-id",
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
	}

	response, err := json.Marshal(getStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/subnets/"+getStruct.ID,
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

func TestListSubnets(t *testing.T) {
	subnetList := MockSubnetsPage{
		Items: []photon.Subnet{
			{
				Name:          "fake_subnet_name",
				ID:            "fake_subnet_ID",
				Kind:          "fake_subnet_kind",
				PrivateIpCidr: "fake_cidr",
				State:         "READY",
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
	err = set.Parse([]string{"fake-router-id"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = listSubnets(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error listing subnets: " + err.Error())
	}
}
