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

	"github.com/vmware/photon-controller-cli/photon/cli/client"
	"github.com/vmware/photon-controller-cli/photon/cli/mocks"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
)

type MockTasksPage struct {
	Items            []photon.Task `json:"items"`
	NextPageLink     string        `json:"nextPageLink"`
	PreviousPageLink string        `json:"previousPageLink"`
}

func TestListTasks(t *testing.T) {
	taskList := MockTasksPage{
		Items: []photon.Task{
			photon.Task{
				Operation: "CREATE_FLAVOR",
				State:     "COMPLETED",
				ID:        "fake-flavor-task-id",
				Entity:    photon.Entity{ID: "fake-flavor-id", Kind: "vm"},
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}
	response, err := json.Marshal(taskList)
	if err != nil {
		t.Error("Not expecting error serializaing expected taskLists")
	}

	server := mocks.NewTestServer()

	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks",
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
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	cxt := cli.NewContext(nil, set, nil)

	err = listTasks(cxt)
	if err != nil {
		t.Error("Not expecting error listing tasks: " + err.Error())
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks?entityId=fake-flavor-id&entityKind=vm",
		mocks.CreateResponder(200, string(response[:])))

	set = flag.NewFlagSet("test", 0)
	set.String("entityId", "fake-flavor-id", "entity ID")
	set.String("entityKind", "vm", "entity kind")
	cxt = cli.NewContext(nil, set, nil)

	err = listTasks(cxt)
	if err != nil {
		t.Error("Not expecting error listing tasks with filter options: " + err.Error())
	}
}

func TestShowMonitorTask(t *testing.T) {
	task := photon.Task{
		Operation: "CREATE_FLAVOR",
		State:     "COMPLETED",
		ID:        "fake-flavor-task-id",
		Entity:    photon.Entity{ID: "fake-flavor-id", Kind: "vm"},
	}
	response, err := json.Marshal(task)
	if err != nil {
		t.Error("Not expecting error serializaing expected taskLists")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/fake-flavor-id",
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake-flavor-id"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = showTask(cxt)
	if err != nil {
		t.Error("Not expecting error showing task: " + err.Error())
	}
	err = monitorTask(cxt)
	if err != nil {
		t.Error("Not expecting error monitoring task: " + err.Error())
	}
}
