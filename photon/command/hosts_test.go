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
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
)

func TestCreateDeleteHost(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "CREATE_HOST",
		State:     "QUEUED",
		ID:        "fake-task-id",
		Entity:    photon.Entity{ID: "fake-host-id"},
	}
	completedTask := &photon.Task{
		Operation: "CREATE_HOST",
		State:     "COMPLETED",
		ID:        "fake-task-id",
		Entity:    photon.Entity{ID: "fake-host-id"},
	}

	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected createTask")
	}
	taskresponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+"/deployments"+"/fake-deployment-id"+"/hosts",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	set.String("username", "u", "username")
	set.String("password", "p", "password")
	set.String("address", "192.168.1.1", "host ip")
	set.String("tag", "CLOUD, MGMT", "host tag")
	set.String("metadata", "{\"a\":\"b\", \"c\":\"d\"}", "MGMT host metadata")
	set.String("deployment_id", "fake-deployment-id", "deployment_id")
	cxt := cli.NewContext(nil, set, nil)

	expectedStruct := photon.Deployments{
		Items: []photon.Deployment{
			{
				ImageDatastores: []string{"testname"},
				ID:              "fake-deployment-id",
			},
		},
	}

	response, err = json.Marshal(expectedStruct)
	if err != nil {
		t.Error("Not expecting error serializing expected status")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments",
		mocks.CreateResponder(200, string(response[:])))

	err = createHost(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error creating host: " + err.Error())
	}

	queuedTask = &photon.Task{
		Operation: "DELETE_HOST",
		State:     "QUEUED",
		ID:        "fake-task-id",
		Entity:    photon.Entity{ID: "fake-host-id"},
	}
	completedTask = &photon.Task{
		Operation: "DELETE_HOST",
		State:     "COMPLETED",
		ID:        "fake-task-id",
		Entity:    photon.Entity{ID: "fake-host-id"},
	}

	response, err = json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected deletedTask")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/hosts/"+queuedTask.Entity.ID,
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
	err = deleteHost(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error deleting host: " + err.Error())
	}
}

func TestListHosts(t *testing.T) {
	server := mocks.NewTestServer()
	defer server.Close()

	// We first test that with exactly one deployment, we work as expected.
	// This is the expected case in a real installation
	err := mockHostsForList(t, server)
	if err != nil {
		t.Error("Failed to mock hosts: " + err.Error())
	}
	err = mockDeploymentsForList(t, server, 1)
	if err != nil {
		t.Error("Failed to mock one deployment: " + err.Error())
	}
	mocks.Activate(true)

	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalFlags := flag.NewFlagSet("global-flags", flag.ContinueOnError)
	globalFlags.String("output", "json", "output")
	err = globalFlags.Parse([]string{"--output=json"})
	if err != nil {
		t.Error(err)
	}
	globalCtx := cli.NewContext(nil, globalFlags, nil)
	commandFlags := flag.NewFlagSet("command-flags", flag.ContinueOnError)
	err = commandFlags.Parse([]string{})
	if err != nil {
		t.Error(err)
	}
	cxt := cli.NewContext(nil, commandFlags, globalCtx)
	var output bytes.Buffer
	err = listHosts(cxt, &output)
	if err != nil {
		t.Error("listHosts with one deployment failed unexpectedly: " + err.Error())
	}

	// Verify we printed a list of hosts: it should start with a bracket
	err = checkRegExp(`^\s*\[`, output)
	if err != nil {
		t.Errorf("List hosts didn't produce a JSON list that starts with a bracket (list): %s", err)
	}
	// and end with a bracket (two regular expressions because it's multiline, it's easier)
	err = checkRegExp(`\]\s*$`, output)
	if err != nil {
		t.Errorf("List hosts didn't produce JSON that ended in a bracket (list): %s", err)
	}
	// And spot check that we have the "id" field
	err = checkRegExp(`\"id\":\s*\".*\"`, output)
	if err != nil {
		t.Errorf("List hosts didn't produce a JSON field named 'id': %s", err)
	}

	// Now we verify that with zero deployments, we fail as expected
	err = mockDeploymentsForList(t, server, 0)
	if err != nil {
		t.Error("Failed to mock zero deployments: " + err.Error())
	}
	err = listHosts(cxt, os.Stdout)
	if err == nil {
		t.Error("listHosts with zero deployments succeeded unexpectedly: " + err.Error())
	} else if !strings.Contains(err.Error(), "There are no deployments") {
		t.Error("listHosts failed, but not with expected error message: " + err.Error())
	}

	// Now we verify that with two deployments, we fail as expected
	err = mockDeploymentsForList(t, server, 2)
	if err != nil {
		t.Error("Failed to mock two deployments: " + err.Error())
	}
	err = listHosts(cxt, os.Stdout)
	if err == nil {
		t.Error("listHosts with two deployments succeeded unexpectedly: " + err.Error())
	} else if !strings.Contains(err.Error(), "There are multiple deployments") {
		t.Error("listHosts failed, but not with expected error message: " + err.Error())
	}
}

func mockHostsForList(t *testing.T, server *httptest.Server) error {
	hostList := MockHostsPage{
		Items: []photon.Host{
			{
				Username: "u",
				Password: "p",
				Address:  "testIP",
				Tags:     []string{"CLOUD"},
				ID:       "host-test-id",
				State:    "COMPLETED",
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}

	response, err := json.Marshal(hostList)
	if err != nil {
		return err
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments/0/hosts",
		mocks.CreateResponder(200, string(response[:])))

	hostList = MockHostsPage{
		Items:            []photon.Host{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}
	response, err = json.Marshal(hostList)
	if err != nil {
		return err
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/fake-next-page-link",
		mocks.CreateResponder(200, string(response[:])))
	return nil
}

func mockDeploymentsForList(t *testing.T, server *httptest.Server, numDeployments int) error {
	var deployments []photon.Deployment

	for i := 0; i < numDeployments; i++ {
		deployment := photon.Deployment{
			ID: strconv.Itoa(i),
		}
		deployments = append(deployments, deployment)
	}

	expectedStruct := photon.Deployments{
		Items: deployments,
	}

	response, err := json.Marshal(expectedStruct)
	if err != nil {
		return err
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments",
		mocks.CreateResponder(200, string(response[:])))
	return nil
}

func TestShowHost(t *testing.T) {
	expectedStruct := photon.Host{
		ID:       "506b13eb-f85d-4bad-a29e-e63a1e3eb043",
		Address:  "196.128.1.1",
		Tags:     []string{"CLOUD"},
		State:    "READY",
		Metadata: map[string]string{},
	}

	response, err := json.Marshal(expectedStruct)
	if err != nil {
		log.Fatal("Not expecting error serializing expected response")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/hosts/"+expectedStruct.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{expectedStruct.ID})
	if err != nil {
		log.Fatal("Not expecting arguments parsing to fail")
	}

	cxt := cli.NewContext(nil, set, nil)
	err = showHost(cxt, os.Stdout)
	if err != nil {
		t.Error("Error showing hosts: " + err.Error())
	}
}

func TestSetHostAvailabilityZone(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "SET_AVAILABILITYZONE",
		State:     "QUEUED",
		ID:        "fake-task-id",
		Entity:    photon.Entity{ID: "fake-host-id"},
	}
	completedTask := &photon.Task{
		Operation: "SET_AVAILABILITYZONE",
		State:     "COMPLETED",
		ID:        "fake-task-id",
		Entity:    photon.Entity{ID: "fake-host-id"},
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
		"POST",
		server.URL+"/hosts"+"/fake-host-id"+"/set_availability_zone",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskResponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake-host-id", "fake-availability-zone-id"})
	cxt := cli.NewContext(nil, set, nil)

	err = setHostAvailabilityZone(cxt, os.Stdout)
	if err != nil {
		t.Error("Error listing hosts: " + err.Error())
	}
}

func TestHostTasks(t *testing.T) {
	taskList := MockTasksPage{
		Items: []photon.Task{
			{
				Operation: "CREATE_HOST",
				State:     "COMPLETED",
				ID:        "1",
				Entity:    photon.Entity{ID: "1", Kind: "host"},
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
		server.URL+"/hosts/1/tasks",
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
	err = set.Parse([]string{"1"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = getHostTasks(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error retrieving tenant tasks")
	}
}

func TestHostGetVMs(t *testing.T) {
	vmListStruct := photon.VMs{
		Items: []photon.VM{
			{
				Name:          "fake_vm_name",
				ID:            "fake_vm_ID",
				Flavor:        "fake_vm_flavor_name",
				State:         "STOPPED",
				SourceImageID: "fake_image_ID",
				Host:          "fake_host_ip",
				Datastore:     "fake_datastore_ID",
				AttachedDisks: []photon.AttachedDisk{
					{
						Name:       "d1",
						Kind:       "ephemeral-disk",
						Flavor:     "fake_ephemeral_flavor_ID",
						CapacityGB: 0,
						BootDisk:   true,
					},
				},
			},
		},
	}

	response, err := json.Marshal(vmListStruct)
	if err != nil {
		t.Error("Not expecting error serializing host list")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/hosts/1/vms",
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"1"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = listHostVMs(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting deployment list hosts to fail")
	}
}
