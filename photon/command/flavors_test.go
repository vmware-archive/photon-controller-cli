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

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
)

type MockFlavorsPage struct {
	Items            []photon.Flavor `json:"items"`
	NextPageLink     string          `json:"nextPageLink"`
	PreviousPageLink string          `json:"previousPageLink"`
}

func TestCreateDeleteFlavor(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "CREATE_FLAVOR",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "fake-id"},
	}
	completedTask := &photon.Task{
		Operation: "CREATE_FLAVOR",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "fake-id"},
	}

	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}
	taskresponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+"/flavors",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", true, "doc")
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set := flag.NewFlagSet("test", 0)
	set.String("name", "name", "flavor name")
	set.String("kind", "vm", "flavor kind")
	set.String("cost", "vm.test1 1 B, vm.test2 1 GB", "flavor cost")
	cxt := cli.NewContext(nil, set, globalCtx)

	err = createFlavor(cxt)
	if err != nil {
		t.Error("Not expecting error creating host: " + err.Error())
	}

	expectedStruct := photon.FlavorList{
		Items: []photon.Flavor{
			{
				Name: "testname",
				Kind: "vm",
				Cost: []photon.QuotaLineItem{{Key: "k", Value: 1, Unit: "B"}},
				ID:   "1",
			},
		},
	}

	response, err = json.Marshal(expectedStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected status")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/flavors",
		mocks.CreateResponder(200, string(response[:])))

	set = flag.NewFlagSet("test", 0)
	cxt = cli.NewContext(nil, set, nil)
	err = listFlavors(cxt)
	if err != nil {
		t.Error("Not expecting list deployment to fail")
	}

	queuedTask = &photon.Task{
		Operation: "DELETE_FLAVOR",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "fake-id"},
	}
	completedTask = &photon.Task{
		Operation: "DELETE_FLAVOR",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "fake-id"},
	}

	response, err = json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected deletedTask")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/flavors/"+queuedTask.Entity.ID,
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))

	set = flag.NewFlagSet("test", 0)
	err = set.Parse([]string{queuedTask.Entity.ID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	cxt = cli.NewContext(nil, set, nil)
	err = deleteFlavor(cxt)
	if err != nil {
		t.Error("Not expecting error deleting host: " + err.Error())
	}
}

func TestShowFlavor(t *testing.T) {
	getStruct := photon.Flavor{
		Name: "testname",
		ID:   "1",
		Kind: "persistent-disk",
		Cost: []photon.QuotaLineItem{{Key: "k", Value: 1, Unit: "B"}},
	}

	response, err := json.Marshal(getStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/flavors/"+getStruct.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{getStruct.ID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = showFlavor(cxt)
	if err != nil {
		t.Error("Not expecting get deployment to fail")
	}
}

func TestFlavorTasks(t *testing.T) {
	taskList := MockTasksPage{
		Items: []photon.Task{
			{
				Operation: "CREATE_FLAVOR",
				State:     "COMPLETED",
				ID:        "1",
				Entity:    photon.Entity{ID: "fake-id", Kind: "flavor"},
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}

	response, err := json.Marshal(taskList)
	if err != nil {
		t.Error("Not expecting error serializaing expected status")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/flavors/fake-id/tasks",
		mocks.CreateResponder(200, string(response[:])))

	taskList = MockTasksPage{
		Items:            []photon.Task{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}
	response, err = json.Marshal(taskList)
	if err != nil {
		t.Error("Not expecting error serializing expected taskLists")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/fake-next-page-link",
		mocks.CreateResponder(200, string(response[:])))

	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake-id"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = getFlavorTasks(cxt)
	if err != nil {
		t.Error("Not expecting error retrieving tenant tasks")
	}
}

func TestListFlavors(t *testing.T) {
	flavorLists := MockFlavorsPage{
		Items: []photon.Flavor{
			{
				Name: "f1",
				Kind: "vm",
			},
			{
				Name: "f2",
				Kind: "disk",
			},
		},
		NextPageLink:     "fake-next-page-link",
		PreviousPageLink: "",
	}

	response, err := json.Marshal(flavorLists)
	if err != nil {
		t.Error("Not expecting error serializing flavors list")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/flavors",
		mocks.CreateResponder(200, string(response[:])))

	flavorLists = MockFlavorsPage{
		Items:            []photon.Flavor{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}

	response, err = json.Marshal(flavorLists)
	if err != nil {
		t.Error("Not expecting error serializing flavors list")
	}
	mocks.RegisterResponder(
		"GET",
		server.URL+"fake-next-page-link",
		mocks.CreateResponder(200, string(response[:])))

	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	cxt := cli.NewContext(nil, set, nil)

	err = listFlavors(cxt)
	if err != nil {
		t.Error("Not expecting error listing flavors: " + err.Error())
	}
}
