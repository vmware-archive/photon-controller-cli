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
	"net/http/httptest"
	"os"
	"testing"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

var tenantName string
var tenantID string
var rtName string
var rtID string
var server *httptest.Server

type MockResourceTicketsPage struct {
	Items            []photon.ResourceTicket `json:"items"`
	NextPageLink     string                  `json:"nextPageLink"`
	PreviousPageLink string                  `json:"previousPageLink"`
}

func TestCreateResourceTicket(t *testing.T) {
	tenantName = "fake_tenant_Name"
	tenantID = "fake_tenant_ID"
	rtName = "fake_rt_Name"
	rtID = "fake_rt_ID"

	tenantStruct := photon.Tenants{
		Items: []photon.Tenant{
			{
				Name: tenantName,
				ID:   tenantID,
			},
		},
	}
	tenantResponse, err := json.Marshal(tenantStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected tenantStruct")
	}

	queuedTask := &photon.Task{
		Operation: "CREATE_RESOURCE_TICKET",
		State:     "QUEUED",
		ID:        "fake-rt-task-id",
		Entity:    photon.Entity{ID: rtID},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "CREATE_RESOURCE_TICKET",
		State:     "COMPLETED",
		ID:        "fake-rt-task-id",
		Entity:    photon.Entity{ID: rtID},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	server = mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants",
		mocks.CreateResponder(200, string(tenantResponse[:])))
	mocks.RegisterResponder(
		"POST",
		server.URL+"/tenants/"+tenantID+"/resource-tickets",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
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
	set.String("name", "name", "rt name")
	set.String("tenant", tenantName, "tenant name")
	set.String("limits", "vm.test1 1 B, vm.test2 1 GB", "rt limits")
	cxt := cli.NewContext(nil, set, globalCtx)

	err = createResourceTicket(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error creating resource ticket: " + err.Error())
	}
}

func TestShowResourceTicket(t *testing.T) {
	rtListStruct := photon.ResourceList{
		Items: []photon.ResourceTicket{
			{
				Name:   rtName,
				ID:     rtID,
				Limits: []photon.QuotaLineItem{{Key: "k", Value: 1, Unit: "B"}},
				Usage:  []photon.QuotaLineItem{{Key: "k", Value: 0, Unit: "B"}},
			},
		},
	}
	listResponse, err := json.Marshal(rtListStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected rtLists")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants/"+tenantID+"/resource-tickets",
		mocks.CreateResponder(200, string(listResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants/"+tenantID+"/resource-tickets?name="+rtName,
		mocks.CreateResponder(200, string(listResponse[:])))

	set := flag.NewFlagSet("test", 0)
	set.String("tenant", tenantName, "tenant name")
	err = set.Parse([]string{rtName})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = showResourceTicket(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error showing resource ticket: " + err.Error())
	}
}

func TestListResourceTickets(t *testing.T) {
	rtList := MockResourceTicketsPage{
		Items: []photon.ResourceTicket{
			{
				Name:   rtName,
				ID:     rtID,
				Limits: []photon.QuotaLineItem{{Key: "k", Value: 1, Unit: "B"}},
				Usage:  []photon.QuotaLineItem{{Key: "k", Value: 0, Unit: "B"}},
			},
		},
		NextPageLink:     "fake-next-page-link",
		PreviousPageLink: "",
	}
	listResponse, err := json.Marshal(rtList)
	if err != nil {
		t.Error("Not expecting error serializaing expected rtLists")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants/"+tenantID+"/resource-tickets",
		mocks.CreateResponder(200, string(listResponse[:])))

	rtList = MockResourceTicketsPage{
		Items:            []photon.ResourceTicket{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}
	listResponse, err = json.Marshal(rtList)
	if err != nil {
		t.Error("Not expecting error serializaing expected rtLists")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"fake-next-page-link",
		mocks.CreateResponder(200, string(listResponse[:])))

	globalFlags := flag.NewFlagSet("global-flags", flag.ContinueOnError)
	globalFlags.String("output", "json", "output")
	err = globalFlags.Parse([]string{"--output=json"})
	if err != nil {
		t.Error(err)
	}
	globalCxt := cli.NewContext(nil, globalFlags, nil)
	commandFlags := flag.NewFlagSet("command-flags", flag.ContinueOnError)
	commandFlags.String("tenant", tenantName, "tenant name")
	err = commandFlags.Parse([]string{})
	if err != nil {
		t.Error(err)
	}
	cxt := cli.NewContext(nil, commandFlags, globalCxt)
	var output bytes.Buffer

	err = listResourceTickets(cxt, &output)
	if err != nil {
		t.Error("Not expecting error listing resource ticket: " + err.Error())
	}

	// Verify we printed a list of resource ticket starting with a bracket
	err = checkRegExp(`^\s*\[`, output)
	if err != nil {
		t.Errorf("List resource ticket didn't produce a JSON list that starts with a bracket (list): %s", err)
	}
	// and end with a bracket (two regular expressions because it's multiline, it's easier)
	err = checkRegExp(`\]\s*$`, output)
	if err != nil {
		t.Errorf("List resource ticket didn't produce JSON that ended in a bracket (list): %s", err)
	}
	// And spot check that we have the "id" field
	err = checkRegExp(`\"id\":\s*\".*\"`, output)
	if err != nil {
		t.Errorf("List resource ticket didn't produce a JSON field named 'id': %s", err)
	}
}

func TestListResourceTicketTasks(t *testing.T) {
	taskList := MockTasksPage{
		Items: []photon.Task{
			{
				Operation: "CREATE_RESOURCE_TICKET",
				State:     "COMPLETED",
				ID:        "fake-rt-task-id",
				Entity:    photon.Entity{ID: rtID},
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}
	response, err := json.Marshal(taskList)
	if err != nil {
		t.Error("Not expecting error serializaing expected taskLists")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/resource-tickets/"+rtID+"/tasks",
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

	set := flag.NewFlagSet("test", 0)
	set.String("tenant", tenantName, "tenant name")
	err = set.Parse([]string{rtName})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = getResourceTicketTasks(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error showing resource ticket tasks: " + err.Error())
	}
}
