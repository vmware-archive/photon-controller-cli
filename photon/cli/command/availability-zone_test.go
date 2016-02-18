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

type MockAvailZonePage struct {
	Items            []photon.AvailabilityZone `json:"items"`
	NextPageLink     string                    `json:"nextPageLink"`
	PreviousPageLink string                    `json:"previousPageLink"`
}

func TestCreateDeleteAvailabilityZone(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "CREATE_AVAILABILITYZONE",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask := &photon.Task{
		Operation: "CREATE_AVAILABILITYZONE",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "1"},
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
		server.URL+"/availabilityzones",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	set.String("name", "n", "testname")
	err = set.Parse([]string{"testname"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = createAvailabilityZone(cxt)
	if err != nil {
		t.Error("Not expecting create availability zone to fail: " + err.Error())
	}

	queuedTask = &photon.Task{
		Operation: "DELETE_AVAILABILITYZONE",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask = &photon.Task{
		Operation: "DELETE_AVAILABILITYZONE",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "1"},
	}

	response, err = json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}
	taskresponse, err = json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completedTask")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/availabilityzones/"+queuedTask.Entity.ID,
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
	err = deleteAvailabilityZone(cxt)
	if err != nil {
		t.Error("Not expecting delete availabilityzone to fail: " + err.Error())
	}
}

func TestShowAvailabilityZone(t *testing.T) {
	expectedStruct := photon.AvailabilityZone{
		ID:   "availabilityzone_id",
		Name: "availabilityzone_name",
	}

	response, err := json.Marshal(expectedStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected response")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/availabilityzones/"+expectedStruct.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{expectedStruct.ID})
	cxt := cli.NewContext(nil, set, nil)
	err = showAvailabilityZone(cxt)
	if err != nil {
		t.Error("Not expecting show availabilityzone to fail: " + err.Error())
	}
}

func TestListAvailabilityZones(t *testing.T) {
	server := mocks.NewTestServer()
	defer server.Close()

	expectedList := MockAvailZonePage{
		Items: []photon.AvailabilityZone{
			{
				Name: "testname",
				ID:   "1",
			},
			{
				Name: "secondname",
				ID:   "2",
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}

	response, err := json.Marshal(expectedList)
	if err != nil {
		t.Error("Not expecting error serializing expected availabilityzones")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/availabilityzones",
		mocks.CreateResponder(200, string(response[:])))

	expectedList = MockAvailZonePage{
		Items:            []photon.AvailabilityZone{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}

	response, err = json.Marshal(expectedList)
	if err != nil {
		t.Error("Not expecting error serializing expected availabilityzones")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/fake-next-page-link",
		mocks.CreateResponder(200, string(response[:])))

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	cxt := cli.NewContext(nil, set, nil)
	err = listAvailabilityZones(cxt)
	if err != nil {
		t.Error("Not expecting list availabilityzone to fail: " + err.Error())
	}
}

func TestAvailabilityZoneTasks(t *testing.T) {
	taskList := MockTasksPage{
		Items: []photon.Task{
			{
				Operation: "CREATE_AVAILABILITYZONE",
				State:     "COMPLETED",
				ID:        "1",
				Entity:    photon.Entity{ID: "1"},
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}

	response, err := json.Marshal(taskList)
	if err != nil {
		t.Error("Not expecting error serializing expected status")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/availabilityzones/1/tasks",
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
	err = set.Parse([]string{"1"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = getAvailabilityZoneTasks(cxt)
	if err != nil {
		t.Error("Not expecting retrieving availabilityzone tasks to fail: " + err.Error())
	}
}
