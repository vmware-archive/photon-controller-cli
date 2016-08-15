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
	"testing"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

type MockVMsPage struct {
	Items            []photon.VM `json:"items"`
	NextPageLink     string      `json:"nextPageLink"`
	PreviousPageLink string      `json:"previousPageLink"`
}

func TestCreateDeleteVM(t *testing.T) {
	tenantStruct := photon.Tenants{
		Items: []photon.Tenant{
			{
				Name: "fake_tenant_name",
				ID:   "fake_tenant_ID",
			},
		},
	}
	tenantResponse, err := json.Marshal(tenantStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected tenantStruct")
	}

	projectStruct := photon.ProjectList{
		Items: []photon.ProjectCompact{
			{
				Name: "fake_project_name",
				ID:   "fake_project_ID",
			},
		},
	}
	projectResponse, err := json.Marshal(projectStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected projectStruct")
	}

	queuedTask := &photon.Task{
		Operation: "CREATE_VM",
		State:     "QUEUED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "CREATE_VM",
		State:     "COMPLETED",
		ID:        "fake-vm-task-IDS",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
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
		"GET",
		server.URL+"/tenants/"+"fake_tenant_ID"+"/projects?name="+"fake_project_name",
		mocks.CreateResponder(200, string(projectResponse[:])))
	mocks.RegisterResponder(
		"POST",
		server.URL+"/projects/"+"fake_project_ID"+"/vms",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", true, "doc")
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set := flag.NewFlagSet("test", 0)
	set.String("name", "fake_vm_name", "VM name")
	set.String("flavor", "fake_vm_flavor_name", "VM flavor")
	set.String("image", "fake_image_ID", "VM image")
	set.String("disks", "d1 fake_ephemeral_flavor_name boot=true", "VM disks")
	set.String("tenant", "fake_tenant_name", "tenant name")
	set.String("project", "fake_project_name", "project name")
	set.String("environment", "vm:fake_environment", "environment")
	set.String("network", "networkid1", "VM Network")
	cxt := cli.NewContext(nil, set, globalCtx)

	err = createVM(cxt)
	if err != nil {
		t.Error("Not expecting error creating VM: " + err.Error())
	}

	queuedTask = &photon.Task{
		Operation: "DELETE_VM",
		State:     "QUEUED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	taskResponse, err = json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask = &photon.Task{
		Operation: "DELETE_VM",
		State:     "COMPLETED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	response, err = json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/vms/"+"fake_vm_ID",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set = flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	cxt = cli.NewContext(nil, set, nil)
	err = deleteVM(cxt)
	if err != nil {
		t.Error("Not expecting error deleting vm: " + err.Error())
	}
}

func TestShowVM(t *testing.T) {
	vmStruct := photon.VM{
		Name:          "fake_vm_name",
		ID:            "fake_vm_ID",
		Flavor:        "fake_vm_flavor_name",
		State:         "ERROR",
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
	}
	response, err := json.Marshal(vmStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected VM")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/vms/"+"fake_vm_ID",
		mocks.CreateResponder(200, string(response[:])))

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = showVM(cxt)
	if err != nil {
		t.Error("Not expecting error showing VM: " + err.Error())
	}
}

func TestStartVM(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "START_VM",
		State:     "QUEUED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "START_VM",
		State:     "COMPLETED",
		ID:        "fake-vm-task-IDS",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	mocks.RegisterResponder(
		"POST",
		server.URL+"/vms/"+"fake_vm_ID"+"/start",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = startVM(cxt)
	if err != nil {
		t.Error("Not expecting error starting VM: " + err.Error())
	}
}

func TestStopVM(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "STOP_VM",
		State:     "QUEUED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "STOP_VM",
		State:     "COMPLETED",
		ID:        "fake-vm-task-IDS",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	mocks.RegisterResponder(
		"POST",
		server.URL+"/vms/"+"fake_vm_ID"+"/stop",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = stopVM(cxt)
	if err != nil {
		t.Error("Not expecting error stoping VM: " + err.Error())
	}
}

func TestResumeVM(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "RESUME_VM",
		State:     "QUEUED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "RESUME_VM",
		State:     "COMPLETED",
		ID:        "fake-vm-task-IDS",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	mocks.RegisterResponder(
		"POST",
		server.URL+"/vms/"+"fake_vm_ID"+"/resume",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = resumeVM(cxt)
	if err != nil {
		t.Error("Not expecting error resuming VM: " + err.Error())
	}
}

func TestRestartVM(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "RESTART_VM",
		State:     "QUEUED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "RESTART_VM",
		State:     "COMPLETED",
		ID:        "fake-vm-task-IDS",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	mocks.RegisterResponder(
		"POST",
		server.URL+"/vms/"+"fake_vm_ID"+"/restart",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = restartVM(cxt)
	if err != nil {
		t.Error("Not expecting error restarting VM: " + err.Error())
	}
}

func TestSuspendVM(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "SUSPEND_VM",
		State:     "QUEUED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "SUSPEND_VM",
		State:     "COMPLETED",
		ID:        "fake-vm-task-IDS",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	mocks.RegisterResponder(
		"POST",
		server.URL+"/vms/"+"fake_vm_ID"+"/suspend",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = suspendVM(cxt)
	if err != nil {
		t.Error("Not expecting error suspending VM: " + err.Error())
	}
}

func TestAttachDisk(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "ATTACH_DISK",
		State:     "QUEUED",
		ID:        "fake-vm-task-ID",
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "ATTACH_DISK",
		State:     "COMPLETED",
		ID:        "fake-vm-task-IDS",
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	mocks.RegisterResponder(
		"POST",
		server.URL+"/vms/"+"fake_vm_ID"+"/attach_disk",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set.String("disk", "fake_disk_ID", "attach disk")
	cxt := cli.NewContext(nil, set, nil)

	err = attachDisk(cxt)
	if err != nil {
		t.Error("Not expecting error attaching disk: " + err.Error())
	}
}

func TestDetachDisk(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "DETACH_DISK",
		State:     "QUEUED",
		ID:        "fake-vm-task-ID",
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "DETACH_DISK",
		State:     "COMPLETED",
		ID:        "fake-vm-task-IDS",
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	mocks.RegisterResponder(
		"POST",
		server.URL+"/vms/"+"fake_vm_ID"+"/detach_disk",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set.String("disk", "fake_disk_ID", "detach disk")
	cxt := cli.NewContext(nil, set, nil)

	err = detachDisk(cxt)
	if err != nil {
		t.Error("Not expecting error detaching disk: " + err.Error())
	}
}

func TestAttachDetachISO(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "ATTACH_ISO",
		State:     "QUEUED",
		ID:        "fake-vm-task-ID",
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "ATTACH_ISO",
		State:     "COMPLETED",
		ID:        "fake-vm-task-ID",
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+"/vms/"+"fake_vm_ID"+"/attach_iso",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set.String("name", "ttylinux-pc_i486-16.1.iso", "attach iso")
	set.String("path", "../../testdata/ttylinux-pc_i486-16.1.iso", "attach iso")
	cxt := cli.NewContext(nil, set, nil)

	err = attachIso(cxt)
	if err != nil {
		t.Error("Not expecting error attaching iso: " + err.Error())
	}

	queuedTask = &photon.Task{
		Operation: "DETACH_ISO",
		State:     "QUEUED",
		ID:        "fake-vm-task-ID",
	}
	taskResponse, err = json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask = &photon.Task{
		Operation: "DETACH_ISO",
		State:     "COMPLETED",
		ID:        "fake-vm-task-ID",
	}
	response, err = json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	mocks.RegisterResponder(
		"POST",
		server.URL+"/vms/"+"fake_vm_ID"+"/detach_iso",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set = flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt = cli.NewContext(nil, set, nil)
	err = detachIso(cxt)
	if err != nil {
		t.Error("Not expecting error detaching iso: " + err.Error())
	}
}

func TestListVMs(t *testing.T) {
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

	vmList = MockVMsPage{
		Items:            []photon.VM{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}

	var nextVmPageResponse []byte
	nextVmPageResponse, err = json.Marshal(vmList)
	if err != nil {
		t.Error("Not expecting error serializaing expected vmList")
	}

	expectedStruct := photon.Tenants{
		Items: []photon.Tenant{
			{
				Name: "fake_tenant_name",
				ID:   "fake_tenant_ID",
			},
		},
	}

	response, err := json.Marshal(expectedStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected tenants")
	}

	projectListStruct := photon.ProjectList{
		Items: []photon.ProjectCompact{
			{
				Name: "fake_project_name",
				ID:   "fake_project_ID",
				ResourceTicket: photon.ProjectTicket{
					Limits: []photon.QuotaLineItem{{Key: "vm.test1", Value: 1, Unit: "B"}},
					Usage:  []photon.QuotaLineItem{{Key: "vm.test1", Value: 0, Unit: "B"}},
				},
			},
		},
	}
	listProjectResponse, err := json.Marshal(projectListStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected projectLists")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants/"+"fake_tenant_ID"+"/projects?name="+"fake_project_name",
		mocks.CreateResponder(200, string(listProjectResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/projects/"+"fake_project_ID"+"/vms",
		mocks.CreateResponder(200, string(listResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/fake-next-page-link",
		mocks.CreateResponder(200, string(nextVmPageResponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	set.String("tenant", "fake_tenant_name", "tenant name")
	set.String("project", "fake_project_name", "project name")
	cxt := cli.NewContext(nil, set, nil)

	err = listVMs(cxt)
	if err != nil {
		t.Error("Not expecting error listing VMs: " + err.Error())
	}
}

func TestFindVMsByName(t *testing.T) {
	vmName := "fake_vm_name"

	tenantStruct := photon.Tenants{
		Items: []photon.Tenant{
			{
				Name: "fake_tenant_name",
				ID:   "fake_tenant_ID",
			},
		},
	}
	tenantResponse, err := json.Marshal(tenantStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected tenants")
	}

	projectListStruct := photon.ProjectList{
		Items: []photon.ProjectCompact{
			{
				Name: "fake_project_name",
				ID:   "fake_project_ID",
				ResourceTicket: photon.ProjectTicket{
					Limits: []photon.QuotaLineItem{{Key: "vm.test1", Value: 1, Unit: "B"}},
					Usage:  []photon.QuotaLineItem{{Key: "vm.test1", Value: 0, Unit: "B"}},
				},
			},
		},
	}
	listProjectResponse, err := json.Marshal(projectListStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected projectLists")
	}

	vmList := MockVMsPage{
		Items: []photon.VM{
			{
				Name:          vmName,
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
	listVmResponse, err := json.Marshal(vmList)
	if err != nil {
		t.Error("Not expecting error serializaing expected vmList")
	}

	vmList = MockVMsPage{
		Items:            []photon.VM{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}
	nextVmPageResponse, err := json.Marshal(vmList)
	if err != nil {
		t.Error("Not expecting error serializaing expected vmList")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants",
		mocks.CreateResponder(200, string(tenantResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants/"+"fake_tenant_ID"+"/projects?name="+"fake_project_name",
		mocks.CreateResponder(200, string(listProjectResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/projects/"+"fake_project_ID"+"/vms?name="+vmName,
		mocks.CreateResponder(200, string(listVmResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/fake-next-page-link",
		mocks.CreateResponder(200, string(nextVmPageResponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	set.String("tenant", "fake_tenant_name", "tenant name")
	set.String("project", "fake_project_name", "project name")
	set.String("name", vmName, "VM name")
	cxt := cli.NewContext(nil, set, nil)

	err = listVMs(cxt)
	if err != nil {
		t.Error("Not expecting error listing VMs by name: " + err.Error())
	}
}

func TestListVMTasks(t *testing.T) {
	taskList := MockTasksPage{
		Items: []photon.Task{
			{
				Operation: "CREATE_VM",
				State:     "COMPLETED",
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}
	response, err := json.Marshal(taskList)
	if err != nil {
		t.Error("Not expecting error serializaing expected taskLists")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/vms/"+"fake_vm_ID"+"/tasks",
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
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = getVMTasks(cxt)
	if err != nil {
		t.Error("Not expecting error showing VM tasks: " + err.Error())
	}
}

func TestSetVMMetadata(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "SET_METADATA",
		State:     "QUEUED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "SET_METADATA",
		State:     "COMPLETED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+"/vms/"+"fake_vm_ID"+"/set_metadata",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set.String("metadata", "{\"a\":\"b\", \"c\":\"d\"}", "vm metadata")

	cxt := cli.NewContext(nil, set, nil)

	err = setVMMetadata(cxt)
	if err != nil {
		t.Error("Not expecting error setting vm metadata: " + err.Error())
	}
}

func TestVMNetworks(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "GET_NETWORKS",
		State:     "COMPLETED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	networkMap := make(map[string]interface{})
	networkMap["network"] = "VMmgmtNetwork"
	networkMap["macAddress"] = "00:0c:29:7a:b4:d5"
	networkMap["ipAddress"] = "10.144.121.12"
	networkMap["netmask"] = "255.255.252.0"
	networkMap["isConnected"] = "true"
	networkConnectionMap := make(map[string]interface{})
	networkConnectionMap["networkConnections"] = []interface{}{networkMap}

	completedTask := &photon.Task{
		Operation:          "GET_NETWORKS",
		State:              "COMPLETED",
		ResourceProperties: networkConnectionMap,
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/vms/"+"fake_vm_ID"+"/subnets",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = listVMNetworks(cxt)
	if err != nil {
		t.Error("Not expecting error getting vm networks: " + err.Error())
	}
}

func TestSetVMTag(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "ADD_TAG",
		State:     "QUEUED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}
	completedTask := &photon.Task{
		Operation: "ADD_TAG",
		State:     "COMPLETED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+"/vms/"+"fake_vm_ID"+"/tags",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set.String("tag", "namespace:predicate=value", "vm tag")

	cxt := cli.NewContext(nil, set, nil)

	err = setVMTag(cxt)
	if err != nil {
		t.Error("Not expecting error setting vm tag: " + err.Error())
	}
}

func TestGetVMMksTicket(t *testing.T) {
	completedTask := &photon.Task{
		Operation: "GET_MKS_TICKET",
		State:     "COMPLETED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	mksMap := make(map[string]interface{})
	mksMap["ticket"] = "ticket-id"

	mksTask := &photon.Task{
		State:              "COMPLETED",
		ID:                 "fake-vm-task-ID",
		Entity:             photon.Entity{ID: "fake_vm_ID"},
		ResourceProperties: mksMap,
	}

	mksresponse, err := json.Marshal(mksTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/vms/"+"fake_vm_ID"+"/mks_ticket",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+completedTask.ID,
		mocks.CreateResponder(200, string(mksresponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = getVMMksTicket(cxt)
	if err != nil {
		t.Error("Not expecting error getting vm mks ticket: " + err.Error())
	}
}

func TestCreateVMImage(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "CREATE_VM_IMAGE",
		State:     "QUEUED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}
	completedTask := &photon.Task{
		Operation: "CREATE_VM_IMAGE",
		State:     "COMPLETED",
		ID:        "fake-vm-task-ID",
		Entity:    photon.Entity{ID: "fake_vm_ID"},
	}

	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}
	taskResponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+"/vms/"+"fake_vm_ID"+"/create_image",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskResponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", true, "doc")
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_vm_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, globalCtx)

	err = createVmImage(cxt)
	if err != nil {
		t.Error("Not expecting error creating VM image: " + err.Error())
	}
}
