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
	"os"
	"testing"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

type MockHostsPage struct {
	Items            []photon.Host `json:"items"`
	NextPageLink     string        `json:"nextPageLink"`
	PreviousPageLink string        `json:"previousPageLink"`
}

func TestListInfrastructureHosts(t *testing.T) {
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
		t.Error("Not expecting error serializing host list")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/infrastructure/hosts",
		mocks.CreateResponder(200, string(response[:])))

	hostList = MockHostsPage{
		Items:            []photon.Host{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}
	response, err = json.Marshal(hostList)
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
	err = listInfrastructureHosts(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting deployment list hosts to fail")
	}
}

func TestUpdateInfrastructureImageDatastores(t *testing.T) {
	deploymentId := "deployment1"
	queuedTask := &photon.Task{
		Operation: "UPDATE_IMAGE_DATASTORES",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask := photon.Task{
		ID:        "task1",
		Operation: "UPDATE_IMAGE_DATASTORES",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: deploymentId},
	}
	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error when serializing tasks")
	}
	taskResponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error during serializing expected completedTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+rootUrl+"/infrastructure/image-datastores",
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
	err = set.Parse([]string{deploymentId})
	if err != nil {
		t.Error(err)
	}
	set.String("datastores", "ds1,ds2", "blob")
	ctx := cli.NewContext(nil, set, nil)

	err = updateInfrastructureImageDatastores(ctx)
	if err != nil {
		t.Error(err)
	}
}

func TestSyncInfrastructureHostsConfig(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "SYNC_HOSTS_CONFIG",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask := &photon.Task{
		Operation: "SYNC_HOSTS_CONFIG",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "1"},
	}

	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error during serializing expected queuedTask")
	}
	taskResponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error during serializing expected completedTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+rootUrl+"/infrastructure/sync-hosts-config",
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
	err = set.Parse([]string{queuedTask.Entity.ID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = syncInfrastructureHostsConfig(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting syncHostsConfig to fail")
	}
}
