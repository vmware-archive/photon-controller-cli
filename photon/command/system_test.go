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
	cf "github.com/vmware/photon-controller-cli/photon/configuration"
	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

func TestGetStatus(t *testing.T) {
	// test GetStatus when config file doesn't exist
	err := cf.RemoveConfigFile()
	if err != nil {
		t.Error("Not expecting error removing config file")
	}
	set := flag.NewFlagSet("test", 0)
	cxt := cli.NewContext(nil, set, nil)
	err = getStatus(cxt, os.Stdout)
	if err == nil {
		t.Error("Expected to receive error trying to get status when config file does not exist")
	}

	// test GetStatus with mock client and mock server
	expectedStruct := photon.Status{
		Status: "READY",
		Components: []photon.Component{
			{Component: "chairman", Message: "", Status: "READY"},
			{Component: "housekeeper", Message: "", Status: "READY"},
		},
	}
	response, err := json.Marshal(expectedStruct)
	if err != nil {
		t.Error("Not expecting error serializing expected status")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/status",
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	err = getStatus(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error getting status of mock client")
	}
}

func TestInitializeMigrateDeployment(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "INITIALIZE_MIGRATE_DEPLOYMENT",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask := &photon.Task{
		Operation: "INITIALIZE_MIGRATE_DEPLOYMENT",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "1"},
	}
	deployment := photon.Deployment{
		ID: "1",
	}
	deployments := &photon.Deployments{
		Items: []photon.Deployment{deployment},
	}
	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}
	taskresponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}
	deploymensResponse, err := json.Marshal(deployments)
	if err != nil {
		t.Error("Not expecting error serializing deployment list")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments",
		mocks.CreateResponder(200, string(deploymensResponse[:])),
	)
	mocks.RegisterResponder(
		"POST",
		server.URL+"/deployments/1/initialize_migration",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"0.0.0.0:9000"})
	if err != nil {
		t.Error("Not expecting argument parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = deploymentMigrationPrepareDeprecated(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting initialize Deployment to fail")
	}
}

func TestFinalizeeMigrateDeployment(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "FINALIZE_MIGRATE_DEPLOYMENT",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask := &photon.Task{
		Operation: "FINALIZE_MIGRATE_DEPLOYMENT",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "1"},
	}
	deployment := photon.Deployment{
		ID: "1",
	}
	deployments := &photon.Deployments{
		Items: []photon.Deployment{deployment},
	}
	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}
	taskresponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}
	deploymensResponse, err := json.Marshal(deployments)
	if err != nil {
		t.Error("Not expecting error serializing deployment list")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments",
		mocks.CreateResponder(200, string(deploymensResponse[:])),
	)
	mocks.RegisterResponder(
		"POST",
		server.URL+"/deployments/1/finalize_migration",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"0.0.0.0:9000"})
	if err != nil {
		t.Error("Not expecting argument parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = deploymentMigrationFinalizeDeprecated(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting initialize Deployment to fail")
	}
}
