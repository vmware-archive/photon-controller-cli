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

func TestListDeployment(t *testing.T) {
	set := flag.NewFlagSet("test", 0)
	err := set.Parse([]string{""})
	cxt := cli.NewContext(nil, set, nil)
	err = listDeployments(cxt, os.Stdout)
	// No responder from mock server for list tenant set yet
	if err == nil {
		t.Error("Expecting an error listing deployments")
	}
}

func TestGetDeployment(t *testing.T) {
	auth := &photon.AuthInfo{}
	stats := &photon.StatsInfo{
		Enabled: false,
	}
	getStruct := photon.Deployment{
		ImageDatastores: []string{"testname"},
		ID:              "1",
		Auth:            auth,
		State:           "COMPLETED",
		Stats:           stats,
	}

	response, err := json.Marshal(getStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments/"+getStruct.ID,
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments/1/vms",
		mocks.CreateResponder(200, string(response[:])))

	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{getStruct.ID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = showDeployment(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting get deployment to fail")
	}
}

func TestListDeploymentHosts(t *testing.T) {
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
		server.URL+"/deployments/1/hosts",
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
	err = listDeploymentHosts(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting deployment list hosts to fail")
	}
}

func TestListDeploymentVms(t *testing.T) {
	server := mocks.NewTestServer()
	defer server.Close()

	vmList := MockVMsPage{
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
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}

	response, err := json.Marshal(vmList)
	if err != nil {
		t.Error("Not expecting error serializing vm list")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/deployments/1/vms",
		mocks.CreateResponder(200, string(response[:])))

	vmList = MockVMsPage{
		Items:            []photon.VM{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}

	response, err = json.Marshal(vmList)
	if err != nil {
		t.Error("Not expecting error serializing vm list")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/fake-next-page-link",
		mocks.CreateResponder(200, string(response[:])))

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"1"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = listDeploymentVms(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting deployment list hosts to fail")
	}
}

func TestUpdateImageDatastores(t *testing.T) {
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
		server.URL+"/deployments/"+deploymentId+"/set_image_datastores",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
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

	err = updateImageDatastores(ctx)
	if err != nil {
		t.Error(err)
	}
}

func TestSyncHostsConfig(t *testing.T) {
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
		server.URL+"/deployments/"+queuedTask.Entity.ID+"/sync_hosts_config",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
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

	err = syncHostsConfig(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting syncHostsConfig to fail")
	}
}

func TestPauseSystem(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "PAUSE_SYSTEM",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask := &photon.Task{
		Operation: "PAUSE_SYSTEM",
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
		server.URL+"/deployments/"+queuedTask.Entity.ID+"/pause_system",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
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

	err = pauseSystem(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting pauseSystem to fail")
	}
}

func TestPauseBackgroundTasks(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "PAUSE_BACKGROUND_TASKS",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask := &photon.Task{
		Operation: "PAUSE_BACKGROUND_TASKS",
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
		server.URL+"/deployments/"+queuedTask.Entity.ID+"/pause_background_tasks",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
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

	err = pauseBackgroundTasks(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting pauseBackgroundTasks to fail")
	}
}

func TestResumeSystem(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "RESUME_SYSTEM",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask := &photon.Task{
		Operation: "RESUME_SYSTEM",
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
		server.URL+"/deployments/"+queuedTask.Entity.ID+"/resume_system",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
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

	err = resumeSystem(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting resumeSystem to fail")
	}
}

func TestSetDeploymentSecurityGroups(t *testing.T) {
	deploymentId := "deployment1"
	queuedTask := &photon.Task{
		Operation: "UPDATE_DEPLOYMENT_SECURITY_GROUPS",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: deploymentId},
	}
	completedTask := &photon.Task{
		Operation: "UPDATE_DEPLOYMENT_SECURITY_GROUPS",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: deploymentId},
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
		server.URL+"/deployments/"+deploymentId+"/set_security_groups",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskResponse[:])))
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
	err = set.Parse([]string{deploymentId, "tenant\admingroup"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, globalCtx)

	err = setDeploymentSecurityGroups(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting setDeploymentSecurityGroups to fail")
	}
}

func TestConfigureNsx(t *testing.T) {
	deploymentId := "deployment1"
	queuedTask := &photon.Task{
		Operation: "CONFIGURE_NSX",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: deploymentId},
	}
	completedTask := &photon.Task{
		Operation: "CONFIGURE_NSX",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: deploymentId},
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
		server.URL+"/deployments/"+deploymentId+"/configure_nsx",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskResponse[:])))

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
	err = set.Parse([]string{deploymentId})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set.String("nsx-address", "nsxAddress", "IP address of NSX")
	set.String("nsx-username", "nsxUsername", "NSX username")
	set.String("nsx-password", "nsxPassword", "NSX password")
	set.String("dhcp-server-private-address", "dhcpServerPrivateAddress", "Private IP address of DHCP server")
	set.String("dhcp-server-public-address", "dhcpServerPublicAddress", "Public IP address of DHCP server")
	set.String("private-ip-root-cidr", "privateIpRootCidr", "Root CIDR of the private IP pool")
	set.String("floating-ip-root-range-start", "floatingIpRootRangeStart",
		"Start of the root range of the floating IP pool")
	set.String("floating-ip-root-range-end", "floatingIpRootRangeEnd", "End of the root range of the floating IP pool")
	set.String("t0-router-id", "t0RouterId", "ID of the T0-Router")
	set.String("edge-cluster-id", "edgeClusterId", "ID of the Edge cluster")
	set.String("overlay-transport-zone-id", "overlayTransportZoneId", "ID of the OVERLAY transport zone")
	set.String("tunnel-ip-pool-id", "tunnelIpPoolId", "ID of the tunnel IP pool")
	set.String("host-uplink-pnic", "hostUplinkPnic", "Name of the host uplink pnic")

	cxt := cli.NewContext(nil, set, globalCtx)
	err = configureNsx(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting deployment configure-nsx to fail")
	}
}

func TestEnableClusterType(t *testing.T) {
	deploymentId := "deployment1"
	queuedTask := &photon.Task{
		Operation: "CONFIGURE_CLUSTER",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: deploymentId},
	}
	completedTask := &photon.Task{
		Operation: "CONFIGURE_CLUSTER",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: deploymentId},
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
		server.URL+"/deployments/"+deploymentId+"/enable_cluster_type",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskResponse[:])))

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
	err = set.Parse([]string{deploymentId})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set.String("type", "SWARM", "Cluster type")
	set.String("image-id", "abcd", "image id")

	cxt := cli.NewContext(nil, set, globalCtx)
	err = enableClusterType(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting deployment list hosts to fail")
	}
}

func TestDisableClusterType(t *testing.T) {
	deploymentId := "deployment1"
	queuedTask := &photon.Task{
		Operation: "DELETE_CLUSTER_CONFIGURATION",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: deploymentId},
	}
	completedTask := &photon.Task{
		Operation: "DELETE_CLUSTER_CONFIGURATION",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: deploymentId},
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
		server.URL+"/deployments/"+queuedTask.Entity.ID+"/disable_cluster_type",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskResponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", false, "doc")
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{queuedTask.Entity.ID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set.String("type", "SWARM", "Cluster type")
	cxt := cli.NewContext(nil, set, globalCtx)

	err = disableClusterType(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting pauseBackgroundTasks to fail")
	}
}
