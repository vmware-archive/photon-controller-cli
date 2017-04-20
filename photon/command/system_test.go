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
			{Component: "PHOTON_CONTROLLER", Message: "", Status: "READY"},
		},
	}
	response, err := json.Marshal(expectedStruct)
	if err != nil {
		t.Error("Not expecting error serializing expected status")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/system/status",
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

func TestGetSystemInfo(t *testing.T) {
	auth := &photon.AuthInfo{}
	stats := &photon.StatsInfo{
		Enabled: false,
	}
	getStruct := photon.SystemInfo{
		ImageDatastores: []string{"testname"},
		Auth:            auth,
		State:           "COMPLETED",
		Stats:           stats,
	}

	response, err := json.Marshal(getStruct)
	if err != nil {
		t.Error("Not expecting error serializing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/system/info",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/system/vms",
		mocks.CreateResponder(200, string(response[:])))

	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	cxt := cli.NewContext(nil, set, nil)

	err = showSystemInfo(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting get system info to fail")
	}
}

func TestPauseSystemCommand(t *testing.T) {
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
		server.URL+rootUrl+"/system/pause",
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

	err = PauseSystem(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting pauseSystem to fail")
	}
}

func TestPauseBackgroundTasksCommand(t *testing.T) {
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
		server.URL+rootUrl+"/system/pause-background-tasks",
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

	err = PauseBackgroundTasks(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting pauseBackgroundTasks to fail")
	}
}

func TestResumeSystemCommand(t *testing.T) {
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
		server.URL+rootUrl+"/system/resume",
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

	err = ResumeSystem(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting resumeSystem to fail")
	}
}

func TestSystemSecurityGroups(t *testing.T) {
	deploymentId := "default"

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
		server.URL+rootUrl+"/system/set-security-groups",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
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
	err = set.Parse([]string{"tenant\\admingroup"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, globalCtx)

	err = setSystemSecurityGroups(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting TestSystemSecurityGroups to fail")
	}
}

func TestListSystemVms(t *testing.T) {
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
		server.URL+rootUrl+"/system/vms",
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
	err = listSystemVms(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting deployment list hosts to fail")
	}
}

func TestConfigureNSX(t *testing.T) {
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
		server.URL+rootUrl+"/system/configure-nsx",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
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
	set.String("dns-server-addresses", "dnsServerAddress1,dnsServerAddress2", "Comma separated DNS server "+
		"addresses")

	cxt := cli.NewContext(nil, set, globalCtx)
	err = configureNSX(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting deployment configure-nsx to fail")
	}
}

func TestEnableSystemServiceType(t *testing.T) {
	deploymentId := "deployment1"
	queuedTask := &photon.Task{
		Operation: "CONFIGURE_SERVICE",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: deploymentId},
	}
	completedTask := &photon.Task{
		Operation: "CONFIGURE_SERVICE",
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
		server.URL+rootUrl+"/system/enable-service-type",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
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
	set.String("type", "SWARM", "Service type")
	set.String("image-id", "abcd", "image id")

	cxt := cli.NewContext(nil, set, globalCtx)
	err = enableSystemServiceType(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting deployment list hosts to fail")
	}
}

func TestDisableSystemServiceType(t *testing.T) {
	deploymentId := "deployment1"
	queuedTask := &photon.Task{
		Operation: "DELETE_SERVICE_CONFIGURATION",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: deploymentId},
	}
	completedTask := &photon.Task{
		Operation: "DELETE_SERVICE_CONFIGURATION",
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
		server.URL+rootUrl+"/system/disable-service-type",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
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
	set.String("type", "SWARM", "Service type")
	cxt := cli.NewContext(nil, set, globalCtx)

	err = disableSystemServiceType(cxt)
	if err != nil {
		t.Error(err)
		t.Error("Not expecting pauseBackgroundTasks to fail")
	}
}
