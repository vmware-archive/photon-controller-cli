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
		server.URL+rootUrl+"/availabilityzones",
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
	set.String("name", "fake_availabilityZone", "availability zone name")
	cxt := cli.NewContext(nil, set, nil)

	err = createAvailabilityZone(cxt, os.Stdout)
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
		server.URL+rootUrl+"/availabilityzones/"+queuedTask.Entity.ID,
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
		server.URL+rootUrl+"/availabilityzones/"+expectedStruct.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{expectedStruct.ID})
	cxt := cli.NewContext(nil, set, nil)
	err = showAvailabilityZone(cxt, os.Stdout)
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
		server.URL+rootUrl+"/availabilityzones",
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
	err = listAvailabilityZones(cxt, &output)
	if err != nil {
		t.Error("Not expecting list availabilityzone to fail: " + err.Error())
	}

	// Verify we printed a list of availability zones starting with a bracket
	err = checkRegExp(`^\s*\[`, output)
	if err != nil {
		t.Errorf("List availability zones didn't produce a JSON list that starts with a bracket (list): %s", err)
	}
	// and end with a bracket (two regular expressions because it's multiline, it's easier)
	err = checkRegExp(`\]\s*$`, output)
	if err != nil {
		t.Errorf("List availability zones didn't produce JSON that ended in a bracket (list): %s", err)
	}
	// And spot check that we have the "id" field
	err = checkRegExp(`\"id\":\s*\".*\"`, output)
	if err != nil {
		t.Errorf("List availability zones didn't produce a JSON field named 'id': %s", err)
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
		server.URL+rootUrl+"/availabilityzones/1/tasks",
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
	err = getAvailabilityZoneTasks(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting retrieving availabilityzone tasks to fail: " + err.Error())
	}
}
