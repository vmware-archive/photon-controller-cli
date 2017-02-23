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
	"strings"
	"testing"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

type MockServicesPage struct {
	Items            []photon.Service `json:"items"`
	NextPageLink     string           `json:"nextPageLink"`
	PreviousPageLink string           `json:"previousPageLink"`
}

func TestReadSSHKey(t *testing.T) {
	testPath := "../../testdata/TestKey.pub"
	content, err := readSSHKey(testPath)
	if err != nil {
		t.Error("ReadSSHKey function failed" + err.Error())
	}
	expected := "validSSH Part2 user@somewhere.com"
	if strings.Compare(content, expected) != 0 {
		t.Error("expected SSHkey :" + expected + " actual SSHKey read:" + content)
	}

}

func TestReadCACert(t *testing.T) {
	testPath := "../../testdata/TestCA.crt"
	content, err := readCACert(testPath)
	if err != nil {
		t.Error("ReadCACert function failed" + err.Error())
	}
	expected := "-----BEGIN CERTIFICATE-----\nMIIFmzCCA4OgAwIBAgIJAIAZmLcInJMeMA0GCSqGSIb3DQEBCwUAMGQxCzAJBgNV\n-----END CERTIFICATE-----"
	if strings.Compare(content, expected) != 0 {
		t.Error("expected CACert :" + expected + " actual CACert read:" + content)
	}
}

func TestCreateDeleteService(t *testing.T) {
	tenantStruct := photon.Tenants{
		Items: []photon.Tenant{
			{
				Name: "fake_tenant_name",
				ID:   "fake_tenant_id",
			},
		},
	}
	tenantResponse, err := json.Marshal(tenantStruct)
	if err != nil {
		t.Error("Not expecting error serializing expected tenantStruct")
	}

	projectStruct := photon.ProjectList{
		Items: []photon.ProjectCompact{
			{
				Name: "fake_project_name",
				ID:   "fake_project_id",
			},
		},
	}
	projectResponse, err := json.Marshal(projectStruct)
	if err != nil {
		t.Error("Not expecting error serializing expected projectStruct")
	}

	queuedCreationTask := &photon.Task{
		Operation: "CREATE_SERVICE",
		State:     "QUEUED",
		ID:        "fake_create_service_task_id",
		Entity:    photon.Entity{ID: "fake_service_id"},
	}
	queuedCreationTaskResponse, err := json.Marshal(queuedCreationTask)
	if err != nil {
		t.Error("Not expecting error serializing expected queued creation task")
	}

	completedCreationTask := &photon.Task{
		Operation: "CREATE_SERVICE",
		State:     "COMPLETED",
		ID:        "fake_create_service_task_id",
		Entity:    photon.Entity{ID: "fake_service_id"},
	}
	completedCreationTaskResponse, err := json.Marshal(completedCreationTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completed creation task")
	}

	server = mocks.NewTestServer()
	defer server.Close()

	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants",
		mocks.CreateResponder(200, string(tenantResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants/fake_tenant_id/projects?name=fake_project_name",
		mocks.CreateResponder(200, string(projectResponse[:])))
	mocks.RegisterResponder(
		"POST",
		server.URL+"/projects/fake_project_id/services",
		mocks.CreateResponder(200, string(queuedCreationTaskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedCreationTask.ID,
		mocks.CreateResponder(200, string(completedCreationTaskResponse[:])))
	mocks.Activate(true)

	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", true, "doc")
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting argument parsing to fail")
	}

	set := flag.NewFlagSet("test", 0)
	set.String("tenant", "fake_tenant_name", "tenant name")
	set.String("project", "fake_project_name", "project name")
	set.String("name", "fake_service_name", "service name")
	set.String("type", "KUBERNETES", "service type")
	set.String("vm_flavor", "fake_vm_flavor", "vm flavor name")
	set.String("disk_flavor", "fake_disk_flavor", "disk flavor name")
	set.Int("worker_count", 50, "worker count")
	set.String("dns", "1.1.1.1", "VM network DNS")
	set.String("gateway", "1.1.1.2", "VM network gateway")
	set.String("netmask", "0.0.0.255", "VM network netmask")
	set.String("ssh-key", "../../testdata/TestKey.pub", "ssh key")
	ctx := cli.NewContext(nil, set, globalCtx)

	err = createService(ctx, os.Stdout)
	if err != nil {
		t.Error("Not expecting error creating service: " + err.Error())
	}

	queuedDeletionTask := &photon.Task{
		Operation: "DELETE_SERVICE",
		State:     "QUEUED",
		ID:        "fake_delete_service_task_id",
		Entity:    photon.Entity{ID: "fake_service_id"},
	}
	queuedDeletionTaskResponse, err := json.Marshal(queuedDeletionTask)
	if err != nil {
		t.Error("Not expecting error serializing expected queued deletion task")
	}

	completedDeletionTask := &photon.Task{
		Operation: "DELETE_SERVICE",
		State:     "COMPLETED",
		ID:        "fake_delete_service_task_id",
		Entity:    photon.Entity{ID: "fake_service_id"},
	}
	completedDeletionTaskResponse, err := json.Marshal(completedDeletionTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completed deletion task")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/services/fake_service_id",
		mocks.CreateResponder(200, string(queuedDeletionTaskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedDeletionTask.ID,
		mocks.CreateResponder(200, string(completedDeletionTaskResponse[:])))

	set = flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_service_id"})
	if err != nil {
		t.Error("Not expecting argument parsing to fail")
	}

	ctx = cli.NewContext(nil, set, globalCtx)
	err = deleteService(ctx)
	if err != nil {
		t.Error("Not expecting error deleting service: " + err.Error())
	}
}

func TestShowService(t *testing.T) {
	service := &photon.Service{
		Name:        "fake_service_name",
		State:       "ERROR",
		ID:          "fake_service_id",
		Type:        "KUBERNETES",
		WorkerCount: 50,
	}
	serviceResponse, err := json.Marshal(service)
	if err != nil {
		t.Error("Not expecting error serializing expected service")
	}

	vmListStruct := photon.VMs{
		Items: []photon.VM{
			{
				Name:          "fake_vm_name",
				ID:            "fake_vm_id",
				Flavor:        "fake_vm_flavor_name",
				State:         "STOPPED",
				SourceImageID: "fake_image_id",
				Host:          "fake_host_ip",
				Datastore:     "fake_datastore_ID",
				Tags: []string{
					"service:" + service.ID + ":master",
				},
				AttachedDisks: []photon.AttachedDisk{
					{
						Name:       "d1",
						Kind:       "ephemeral-disk",
						Flavor:     "fake_ephemeral_flavor_id",
						CapacityGB: 0,
						BootDisk:   true,
					},
				},
			},
		},
	}
	vmListResponse, err := json.Marshal(vmListStruct)
	if err != nil {
		t.Error("Not expecting error serializing expected vmList")
	}

	queuedNetworkTask := &photon.Task{
		Operation: "GET_NETWORKS",
		State:     "COMPLETED",
		ID:        "fake_get_networks_task_id",
		Entity:    photon.Entity{ID: "fake_vm_id"},
	}
	queuedNetworkTaskResponse, err := json.Marshal(queuedNetworkTask)
	if err != nil {
		t.Error("Not expecting error serializing expected queuedNetworkTask")
	}

	networkMap := make(map[string]interface{})
	networkMap["network"] = "VMmgmtNetwork"
	networkMap["macAddress"] = "00:0c:29:7a:b4:d5"
	networkMap["ipAddress"] = "10.144.121.12"
	networkMap["netmask"] = "255.255.252.0"
	networkMap["isConnected"] = "true"
	networkConnectionMap := make(map[string]interface{})
	networkConnectionMap["networkConnections"] = []interface{}{networkMap}

	completedNetworkTask := &photon.Task{
		Operation:          "GET_NETWORKS",
		State:              "COMPLETED",
		ResourceProperties: networkConnectionMap,
	}
	completedNetworkTaskResponse, err := json.Marshal(completedNetworkTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completedNetworkTask")
	}

	server = mocks.NewTestServer()
	defer server.Close()

	mocks.RegisterResponder(
		"GET",
		server.URL+"/services/"+service.ID,
		mocks.CreateResponder(200, string(serviceResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/services/"+service.ID+"/vms",
		mocks.CreateResponder(200, string(vmListResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/vms/"+"fake_vm_id"+"/networks",
		mocks.CreateResponder(200, string(queuedNetworkTaskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedNetworkTask.ID,
		mocks.CreateResponder(200, string(completedNetworkTaskResponse[:])))
	mocks.Activate(true)

	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_service_id"})
	if err != nil {
		t.Error("Not expecting argument parsing to fail")
	}
	ctx := cli.NewContext(nil, set, nil)

	err = showService(ctx, os.Stdout)
	if err != nil {
		t.Error("Not expecting error showing service: " + err.Error())
	}
}

func TestListServices(t *testing.T) {
	tenantStruct := photon.Tenants{
		Items: []photon.Tenant{
			{
				Name: "fake_tenant_name",
				ID:   "fake_tenant_id",
			},
		},
	}
	tenantResponse, err := json.Marshal(tenantStruct)
	if err != nil {
		t.Error("Not expecting error serializing expected tenantStruct")
	}

	projectStruct := photon.ProjectList{
		Items: []photon.ProjectCompact{
			{
				Name: "fake_project_name",
				ID:   "fake_project_id",
			},
		},
	}
	projectResponse, err := json.Marshal(projectStruct)
	if err != nil {
		t.Error("Not expecting error serializing expected projectStruct")
	}

	firstServicesPage := MockServicesPage{
		Items: []photon.Service{
			{
				Name:        "fake_service_name",
				State:       "READY",
				ID:          "fake_service_id",
				Type:        "KUBERNETES",
				WorkerCount: 50,
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}

	firstServicesPageResponse, err := json.Marshal(firstServicesPage)
	if err != nil {
		t.Error("Not expecting error serializing expected first services page")
	}

	secondServicesPage := MockServicesPage{
		Items:            []photon.Service{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}

	secondServicesPageResponse, err := json.Marshal(secondServicesPage)
	if err != nil {
		t.Error("Not expecting error serializing expected second services page")
	}

	server := mocks.NewTestServer()
	defer server.Close()

	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants",
		mocks.CreateResponder(200, string(tenantResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants/fake_tenant_id/projects?name=fake_project_name",
		mocks.CreateResponder(200, string(projectResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/projects/fake_project_id/services",
		mocks.CreateResponder(200, string(firstServicesPageResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/fake-next-page-link",
		mocks.CreateResponder(200, string(secondServicesPageResponse[:])))
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
	commandFlags.String("tenant", "fake_tenant_name", "tenant name")
	commandFlags.String("project", "fake_project_name", "project name")
	err = commandFlags.Parse([]string{})
	if err != nil {
		t.Error(err)
	}
	ctx := cli.NewContext(nil, commandFlags, globalCxt)
	var output bytes.Buffer

	err = listServices(ctx, &output)
	if err != nil {
		t.Error("Not expecting error listing services: " + err.Error())
	}

	// Verify we printed a list of services starting with a bracket
	err = checkRegExp(`^\s*\[`, output)
	if err != nil {
		t.Errorf("List services didn't produce a JSON list that starts with a bracket (list): %s", err)
	}
	// and end with a bracket (two regular expressions because it's multiline, it's easier)
	err = checkRegExp(`\]\s*$`, output)
	if err != nil {
		t.Errorf("List services didn't produce JSON that ended in a bracket (list): %s", err)
	}
	// And spot check that we have the "id" field
	err = checkRegExp(`\"id\":\s*\".*\"`, output)
	if err != nil {
		t.Errorf("List services didn't produce a JSON field named 'id': %s", err)
	}
}

func TestResizeService(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "RESIZE_SERVICE",
		State:     "QUEUED",
		ID:        "fake_resize_service_task_id",
		Entity:    photon.Entity{ID: "fake_service_id"},
	}
	queuedTaskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected queued task")
	}

	completedTask := &photon.Task{
		Operation: "RESIZE_SERVICE",
		State:     "COMPLETED",
		ID:        "fake_resize_service_task_id",
		Entity:    photon.Entity{ID: "fake_service_id"},
	}
	completedTaskResponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completed task")
	}

	server := mocks.NewTestServer()
	defer server.Close()

	mocks.RegisterResponder(
		"POST",
		server.URL+"/services/fake_service_id/resize",
		mocks.CreateResponder(200, string(queuedTaskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/fake_resize_service_task_id",
		mocks.CreateResponder(200, string(completedTaskResponse[:])))
	mocks.Activate(true)

	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", true, "doc")
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting argument parsing to fail")
	}

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_service_id", "50"})
	if err != nil {
		t.Error("Not expecting argument parsing to fail")
	}
	ctx := cli.NewContext(nil, set, globalCtx)

	err = resizeService(ctx, os.Stdout)
	if err != nil {
		t.Error("Not expecting error resizing service: " + err.Error())
	}
}

func TestListServiceVms(t *testing.T) {
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

	listResponse, err := json.Marshal(vmList)
	if err != nil {
		t.Error("Not expecting error serializaing expected vmList")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/services/fake_service_id/vms",
		mocks.CreateResponder(200, string(listResponse[:])))

	vmList = MockVMsPage{
		Items:            []photon.VM{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}

	listResponse, err = json.Marshal(vmList)
	if err != nil {
		t.Error("Not expecting error serializaing expected vmList")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/fake-next-page-link",
		mocks.CreateResponder(200, string(listResponse[:])))

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
	err = commandFlags.Parse([]string{"fake_service_id"})
	if err != nil {
		t.Error(err)
	}
	ctx := cli.NewContext(nil, commandFlags, globalCxt)
	var output bytes.Buffer

	err = listVms(ctx, &output)
	if err != nil {
		t.Error("Not expecting error listing service VMs: " + err.Error())
	}

	// Verify we printed a list of service vms starting with a bracket
	err = checkRegExp(`^\s*\[`, output)
	if err != nil {
		t.Errorf("List service vms didn't produce a JSON list that starts with a bracket (list): %s", err)
	}
	// and end with a bracket (two regular expressions because it's multiline, it's easier)
	err = checkRegExp(`\]\s*$`, output)
	if err != nil {
		t.Errorf("List service vms didn't produce JSON that ended in a bracket (list): %s", err)
	}
	// And spot check that we have the "id" field
	err = checkRegExp(`\"id\":\s*\".*\"`, output)
	if err != nil {
		t.Errorf("List service vms didn't produce a JSON field named 'id': %s", err)
	}
}

func TestServiceTriggerMaintenance(t *testing.T) {
	// Start mock server
	server := mocks.NewTestServer()
	defer server.Close()

	// Create mock response
	completedTask := &photon.Task{
		Operation: "TRIGGER_SERVICE_MAINTENANCE",
		State:     "COMPLETED",
		ID:        "fake_service_task_id",
		Entity:    photon.Entity{ID: "fake_service_id"},
	}
	completedTaskResponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completed task")
	}

	// Register mock response with mock server
	mocks.RegisterResponder(
		"POST",
		server.URL+"/services/fake_service_id/trigger_maintenance",
		mocks.CreateResponder(200, string(completedTaskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/fake_service_task_id",
		mocks.CreateResponder(200, string(completedTaskResponse[:])))
	mocks.Activate(true)

	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	commandFlags := flag.NewFlagSet("command-flags", flag.ContinueOnError)
	err = commandFlags.Parse([]string{"fake_service_id"})
	if err != nil {
		t.Error(err)
	}
	ctx := cli.NewContext(nil, commandFlags, nil)

	err = triggerMaintenance(ctx)
	if err != nil {
		t.Error("Not expecting error for service trigger maintenance: " + err.Error())
	}
}

func TestChangeVersionService(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "CHANGE_VERSION_SERVICE",
		State:     "QUEUED",
		ID:        "service-id",
		Entity:    photon.Entity{ID: "service-id"},
	}
	queuedTaskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected queued task")
	}

	completedTask := &photon.Task{
		Operation: "CHANGE_VERSION_SERVICE",
		State:     "COMPLETED",
		ID:        "service-id",
		Entity:    photon.Entity{ID: "service-id"},
	}
	completedTaskResponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completed task")
	}

	server := mocks.NewTestServer()
	defer server.Close()

	mocks.RegisterResponder(
		"POST",
		server.URL+"/services/service-id/change_version",
		mocks.CreateResponder(200, string(queuedTaskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/service-id",
		mocks.CreateResponder(200, string(completedTaskResponse[:])))
	mocks.Activate(true)

	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", true, "doc")
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting argument parsing to fail")
	}

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"service-id"})
	set.String("image-id", "test-image-id", "image name")
	if err != nil {
		t.Error("Not expecting argument parsing to fail")
	}
	ctx := cli.NewContext(nil, set, globalCtx)

	err = changeVersion(ctx, os.Stdout)
	if err != nil {
		t.Error("Not expecting error change version service: " + err.Error())
	}
}

func TestServiceCertToFile(t *testing.T) {
	service := &photon.Service{
		Name:        "fake_service_name",
		State:       "ERROR",
		ID:          "fake_service_id",
		Type:        "KUBERNETES",
		WorkerCount: 50,
		ExtendedProperties: map[string]string{
			photon.ExtendedPropertyRegistryCACert: "-----BEGIN CERTIFICATE-----\nMIIFmzCCA4OgAwIBAgIJAIAZmLcInJMeMA0GCSqGSIb3DQEBCwUAMGQxCzAJBgNV\n-----END CERTIFICATE-----",
		},
	}
	serviceResponse, err := json.Marshal(service)
	if err != nil {
		t.Error("Not expecting error serializing expected service")
	}

	server = mocks.NewTestServer()
	defer server.Close()

	mocks.RegisterResponder(
		"GET",
		server.URL+"/services/"+service.ID,
		mocks.CreateResponder(200, string(serviceResponse[:])))

	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	commandFlags := flag.NewFlagSet("command-flags", flag.ContinueOnError)
	err = commandFlags.Parse([]string{"fake_service_id", "test.cert"})
	if err != nil {
		t.Error(err)
	}
	ctx := cli.NewContext(nil, commandFlags, nil)

	err = certToFile(ctx)
	if err != nil {
		t.Error("Not expecting error for service trigger maintenance: " + err.Error())
	}

	content, err := readCACert("test.cert")

	if err != nil {
		t.Error("ReadCACert function failed" + err.Error())
	}
	expected := "-----BEGIN CERTIFICATE-----\nMIIFmzCCA4OgAwIBAgIJAIAZmLcInJMeMA0GCSqGSIb3DQEBCwUAMGQxCzAJBgNV\n-----END CERTIFICATE-----"
	if strings.Compare(content, expected) != 0 {
		t.Error("expected CACert :" + expected + " actual CACert read:" + content)
	}

	err = os.Remove("test.cert")

	if err != nil {
		t.Error("Error deleting test.cert file")
	}
}

func TestValidateHarborPassword(t *testing.T) {
	string1 := "Harbor123"
	if validateHarborPassword(string1) == false {
		t.Error("expected: true and result was false")
	}
	string2 := "1234567"
	if validateHarborPassword(string2) == true {
		t.Error("expected: false and result was true")
	}
	string3 := "abcHHH "
	if validateHarborPassword(string3) == true {
		t.Error("expected: false and result was true")
	}
	string4 := "harbor2134"
	if validateHarborPassword(string4) == true {
		t.Error("expected: false and result was true")
	}
}
