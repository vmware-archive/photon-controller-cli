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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"syscall"
	"testing"
	"time"

	"github.com/vmware/photon-controller-cli/photon/client"
	cf "github.com/vmware/photon-controller-cli/photon/configuration"
	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
)

func TestGetStatus(t *testing.T) {
	// test GetStatus when config file doesn't exist
	err := cf.RemoveConfigFile()
	if err != nil {
		t.Error("Not expecting error removing config file")
	}
	set := flag.NewFlagSet("test", 0)
	cxt := cli.NewContext(nil, set, nil)
	err = getStatus(cxt)
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
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	err = getStatus(cxt)
	if err != nil {
		t.Error("Not expecting error getting status of mock client")
	}
}

func TestDeploy(t *testing.T) {
	f, err := ioutil.TempFile("", "tempDcMap")
	if err != nil {
		t.Error("Fail to create temperory Dc_Map")
	}

	dcmap := `---
deployment:
  resume_system: true
  image_datastores: datastore1
  syslog_endpoint: 10.146.64.230
  stats_store_endpoint: 10.146.64.111
  ntp_endpoint: 10.20.144.1
  use_image_datastore_for_vms: true
  auth_enabled: false

hosts:
  - address_ranges: 10.146.38.91
    username: root
    password: Password!
    usage_tags:
    - CLOUD
    - MGMT
    metadata:
       ALLOWED_DATASTORES: "datastore1"
       ALLOWED_NETWORKS: "VM VLAN"
       MANAGEMENT_DATASTORE: datastore1
       MANAGEMENT_NETWORK_DNS_SERVER : 10.142.17.124
       MANAGEMENT_NETWORK_GATEWAY: 10.146.71.253
       MANAGEMENT_VM_IPS: 10.146.65.10
       MANAGEMENT_NETWORK_NETMASK: 255.255.248.0
       MANAGEMENT_PORTGROUP: "VM VLAN"
  - address_ranges: 10.146.38.92-10.146.38.93,10.146.38.94
    username: root
    password: Password!
    availability_zone: Zone1
    usage_tags:
    - CLOUD
    - MGMT
    metadata:
      ALLOWED_DATASTORES: "datastore1"
      ALLOWED_NETWORKS: "VM VLAN"
      MANAGEMENT_DATASTORE: datastore1
      MANAGEMENT_NETWORK_DNS_SERVER : 10.142.17.124
      MANAGEMENT_NETWORK_GATEWAY: 10.146.71.253
      MANAGEMENT_VM_IPS: 10.146.65.11-10.146.65.12
      MANAGEMENT_NETWORK_NETMASK: 255.255.248.0
      MANAGEMENT_PORTGROUP: "VM VLAN"
`
	defer func() {
		err = syscall.Unlink(f.Name())
		if err != nil {
			t.Error("Failed to unlink test dc_map file.")
		}
	}()

	err = ioutil.WriteFile(f.Name(), []byte(dcmap), 0644)
	if err != nil {
		t.Error("Failed to create test dc_map file.")
	}

	set := flag.NewFlagSet("test", 0)
	cxt := cli.NewContext(nil, set, nil)
	err = set.Parse([]string{f.Name()})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	queuedTaskId, response, taskresponse, err := createTaskResponses("CREATE_DEPLOYMENT", "deployment-ID")
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}

	availZoneQueuedTaskId, availZoneResponse, availZoneTaskResponse, err := createTaskResponses(
		"CREATE_AVAILABILITY_ZONE", "availability-zone-ID")
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}

	hostQueuedTaskId, hostResponse, hostTaskResponse, err := createTaskResponses("CREATE_HOST", "deployment-ID")
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}

	deployQueuedTaskId, deployResponse, deployTaskResponse, err := createTaskResponses(
		"PEFORM_DEPLOYMENT", "deployment-ID")
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}

	_, _, resumeTaskResponse, err := createTaskResponses("RESUME_SYSTEM", "deployment-ID")
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}

	server := mocks.NewTestServer()
	defer server.Close()

	newServer := mocks.NewTestServerWithBody(resumeTaskResponse)
	defer newServer.Close()

	testServerUrl, err := url.Parse(newServer.URL)
	if err != nil {
		t.Error("Not expecting error when parsing server url")
	}

	deployment := &photon.Deployment{
		ID: "deployment-ID",
		Auth: &photon.AuthInfo{
			Enabled: false,
		},
		LoadBalancerAddress: testServerUrl.Host,
	}

	deploymentResponse, err := json.Marshal(deployment)
	if err != nil {
		t.Error("Not expecting error serializing expected deployment")
	}

	mocks.RegisterResponder(
		"POST",
		server.URL+"/deployments",
		mocks.CreateResponder(200, response))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTaskId,
		mocks.CreateResponder(200, taskresponse))
	mocks.RegisterResponder(
		"POST",
		server.URL+"/availabilityzones",
		mocks.CreateResponder(200, availZoneResponse))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+availZoneQueuedTaskId,
		mocks.CreateResponder(200, availZoneTaskResponse))
	mocks.RegisterResponder(
		"POST",
		server.URL+"/deployments/deployment-ID/hosts",
		mocks.CreateResponder(200, hostResponse))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+hostQueuedTaskId,
		mocks.CreateResponder(200, hostTaskResponse))
	mocks.RegisterResponder(
		"POST",
		server.URL+"/deployments/deployment-ID/deploy",
		mocks.CreateResponder(200, deployResponse))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+deployQueuedTaskId,
		mocks.CreateResponder(200, deployTaskResponse))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments/deployment-ID",
		mocks.CreateResponder(200, string(deploymentResponse[:])))

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	err = deploy(cxt)
	if err != nil {
		t.Error(err)
	}
}

func TestDestroy(t *testing.T) {
	server = mocks.NewTestServer()
	defer server.Close()

	expectedStruct := photon.Deployments{
		Items: []photon.Deployment{
			{
				ImageDatastores: []string{"testname"},
				ID:              "1",
			},
			{
				ImageDatastores: []string{"secondname"},
				ID:              "2",
			},
		},
	}

	response, err := json.Marshal(expectedStruct)
	if err != nil {
		t.Error("Not expecting error serializing expected status")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments",
		mocks.CreateResponder(200, string(response[:])))

	queuedTask := &photon.Task{
		Operation: "DESTROY_DEPLOYMENT",
		State:     "QUEUED",
		ID:        "fake-destroy-task-IDS",
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "DESTROY_DEPLOYMENT",
		State:     "COMPLETED",
		ID:        "fake-destroy-task-IDS",
	}
	response, err = json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completedTask")
	}

	deleteDeploymentqueuedTask1 := &photon.Task{
		Operation: "DELETE_DEPLOYMENT",
		State:     "QUEUED",
		ID:        "fake-delete-task-IDS1",
		Entity:    photon.Entity{ID: "1"},
	}
	deleteDeploymenttaskResponse1, err := json.Marshal(deleteDeploymentqueuedTask1)
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}

	deleteDeploymentcompletedTask1 := &photon.Task{
		Operation: "DELETE_DEPLOYMENT",
		State:     "COMPLETED",
		ID:        "fake-delete-task-ID1",
		Entity:    photon.Entity{ID: "1"},
	}
	deleteDeploymentresponse1, err := json.Marshal(deleteDeploymentcompletedTask1)
	if err != nil {
		t.Error("Not expecting error serializing expected completedTask")
	}
	deleteDeploymentqueuedTask2 := &photon.Task{
		Operation: "DELETE_DEPLOYMENT",
		State:     "QUEUED",
		ID:        "fake-delete-task-ID2",
		Entity:    photon.Entity{ID: "2"},
	}
	deleteDeploymenttaskResponse2, err := json.Marshal(deleteDeploymentqueuedTask2)
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}

	deleteDeploymentcompletedTask2 := &photon.Task{
		Operation: "DELETE_DEPLOYMENT",
		State:     "COMPLETED",
		ID:        "fake-delete-task-ID2",
		Entity:    photon.Entity{ID: "2"},
	}
	deleteDeploymentresponse2, err := json.Marshal(deleteDeploymentcompletedTask2)
	if err != nil {
		t.Error("Not expecting error serializing expected completedTask")
	}

	gethostexpectedStruct := MockHostsPage{
		Items: []photon.Host{
			{
				ID:       "fake-host-id",
				Address:  "196.128.1.1",
				Tags:     []string{"CLOUD", "MGMT"},
				State:    "READY",
				Metadata: map[string]string{"a": "b"},
			},
		},
		NextPageLink:     "/fake-next-hosts-page-link",
		PreviousPageLink: "",
	}
	gethostresponse, err := json.Marshal(gethostexpectedStruct)
	if err != nil {
		t.Error("Not expecting error serializing expected response")
	}
	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments/1/hosts",
		mocks.CreateResponder(200, string(gethostresponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments/2/hosts",
		mocks.CreateResponder(200, string(gethostresponse[:])))
	hostqueuedTask := &photon.Task{
		Operation: "DELETE_HOST",
		State:     "QUEUED",
		ID:        "fake-task-id",
		Entity:    photon.Entity{ID: "fake-host-id"},
	}
	nextHostsPage := MockHostsPage{
		Items:            []photon.Host{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}
	gethostresponse, err = json.Marshal(nextHostsPage)
	if err != nil {
		t.Error("Not expecting error serializing expected nextHostsPage")
	}
	mocks.RegisterResponder(
		"GET",
		server.URL+"/fake-next-hosts-page-link",
		mocks.CreateResponder(200, string(gethostresponse[:])))
	hostcompletedTask := &photon.Task{
		Operation: "DELETE_HOST",
		State:     "COMPLETED",
		ID:        "fake-task-id",
		Entity:    photon.Entity{ID: "fake-host-id"},
	}
	hostresponse, err := json.Marshal(hostqueuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected createTask")
	}
	hosttaskresponse, err := json.Marshal(hostcompletedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected createTask")
	}
	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/hosts/"+hostqueuedTask.Entity.ID,
		mocks.CreateResponder(200, string(hostresponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+hostqueuedTask.ID,
		mocks.CreateResponder(200, string(hosttaskresponse[:])))

	mocks.RegisterResponder(
		"POST",
		server.URL+"/deployments"+"/fake-deployment-id"+"/hosts",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskResponse[:])))

	mocks.RegisterResponder(
		"POST",
		server.URL+"/deployments/1/destroy",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"POST",
		server.URL+"/deployments/2/destroy",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/deployments/1",
		mocks.CreateResponder(200, string(deleteDeploymenttaskResponse1[:])))
	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/deployments/1",
		mocks.CreateResponder(200, string(deleteDeploymenttaskResponse1[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+deleteDeploymentqueuedTask1.ID,
		mocks.CreateResponder(200, string(deleteDeploymentresponse1[:])))
	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/deployments/2",
		mocks.CreateResponder(200, string(deleteDeploymenttaskResponse2[:])))
	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/deployments/2",
		mocks.CreateResponder(200, string(deleteDeploymenttaskResponse2[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+deleteDeploymentqueuedTask2.ID,
		mocks.CreateResponder(200, string(deleteDeploymentresponse2[:])))

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	cxt := cli.NewContext(nil, set, nil)
	err = destroy(cxt)
	if err != nil {
		t.Error(err)
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
		[]photon.Deployment{deployment},
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
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"0.0.0.0:9000"})
	if err != nil {
		t.Error("Not expecting argument parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = deploymentMigrationPrepare(cxt)
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
		[]photon.Deployment{deployment},
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
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"0.0.0.0:9000"})
	if err != nil {
		t.Error("Not expecting argument parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = deploymentMigrationFinalize(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting initialize Deployment to fail")
	}
}

func createTaskResponses(op string, entityId string) (string, string, string, error) {

	task := &photon.Task{
		ID:        fmt.Sprintf("ID-%v", time.Now().Unix()),
		Operation: op,
		State:     "QUEUED",
		Entity:    photon.Entity{ID: entityId},
	}

	queued, err := json.Marshal(task)
	if err != nil {
		return "", "", "", err
	}

	task.State = "COMPLETED"
	completed, err := json.Marshal(task)
	if err != nil {
		return "", "", "", err
	}

	return task.ID, string(queued[:]), string(completed[:]), nil
}
