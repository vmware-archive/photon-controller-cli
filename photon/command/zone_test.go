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

type MockZonePage struct {
	Items            []photon.Zone `json:"items"`
	NextPageLink     string        `json:"nextPageLink"`
	PreviousPageLink string        `json:"previousPageLink"`
}

func TestCreateDeleteZone(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "CREATE_ZONE",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask := &photon.Task{
		Operation: "CREATE_ZONE",
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
		server.URL+rootUrl+"/zones",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	set.String("name", "fake_zone", "zone name")
	cxt := cli.NewContext(nil, set, nil)

	err = createZone(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting create zone to fail: " + err.Error())
	}

	queuedTask = &photon.Task{
		Operation: "DELETE_ZONE",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask = &photon.Task{
		Operation: "DELETE_ZONE",
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
		server.URL+rootUrl+"/zones/"+queuedTask.Entity.ID,
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))

	set = flag.NewFlagSet("test", 0)
	err = set.Parse([]string{queuedTask.Entity.ID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt = cli.NewContext(nil, set, nil)
	err = deleteZone(cxt)
	if err != nil {
		t.Error("Not expecting delete zone to fail: " + err.Error())
	}
}

func TestShowZone(t *testing.T) {
	expectedStruct := photon.Zone{
		ID:   "zone_id",
		Name: "zone_name",
	}

	response, err := json.Marshal(expectedStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected response")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/zones/"+expectedStruct.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{expectedStruct.ID})
	cxt := cli.NewContext(nil, set, nil)
	err = showZone(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting show zone to fail: " + err.Error())
	}
}

func TestListZones(t *testing.T) {
	server := mocks.NewTestServer()
	defer server.Close()

	expectedList := MockZonePage{
		Items: []photon.Zone{
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
		t.Error("Not expecting error serializing expected zones")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/zones",
		mocks.CreateResponder(200, string(response[:])))

	expectedList = MockZonePage{
		Items:            []photon.Zone{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}

	response, err = json.Marshal(expectedList)
	if err != nil {
		t.Error("Not expecting error serializing expected zones")
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
	err = listZones(cxt, &output)
	if err != nil {
		t.Error("Not expecting list zone to fail: " + err.Error())
	}

	// Verify we printed a list of zones starting with a bracket
	err = checkRegExp(`^\s*\[`, output)
	if err != nil {
		t.Errorf("List zones didn't produce a JSON list that starts with a bracket (list): %s", err)
	}
	// and end with a bracket (two regular expressions because it's multiline, it's easier)
	err = checkRegExp(`\]\s*$`, output)
	if err != nil {
		t.Errorf("List zones didn't produce JSON that ended in a bracket (list): %s", err)
	}
	// And spot check that we have the "id" field
	err = checkRegExp(`\"id\":\s*\".*\"`, output)
	if err != nil {
		t.Errorf("List zones didn't produce a JSON field named 'id': %s", err)
	}
}

func TestZoneTasks(t *testing.T) {
	taskList := MockTasksPage{
		Items: []photon.Task{
			{
				Operation: "CREATE_ZONE",
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
		server.URL+rootUrl+"/zones/1/tasks",
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
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"1"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = getZoneTasks(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting retrieving zone tasks to fail: " + err.Error())
	}
}
