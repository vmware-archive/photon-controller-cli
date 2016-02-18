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

	"github.com/vmware/photon-controller-cli/photon/cli/client"
	"github.com/vmware/photon-controller-cli/photon/cli/mocks"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
)

type MockDisksPage struct {
	Items            []photon.PersistentDisk `json:"items"`
	NextPageLink     string                  `json:"nextPageLink"`
	PreviousPageLink string                  `json:"previousPageLink"`
}

func TestCreateDisk(t *testing.T) {
	tenantStruct := photon.Tenants{
		Items: []photon.Tenant{
			photon.Tenant{
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
			photon.ProjectCompact{
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
		Operation: "CREATE_DISK",
		State:     "QUEUED",
		ID:        "fake-disk-task-ID",
		Entity:    photon.Entity{ID: "fake_disk_ID"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "CREATE_DISK",
		State:     "COMPLETED",
		ID:        "fake-disk-task-IDS",
		Entity:    photon.Entity{ID: "fake_disk_ID"},
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
		server.URL+"/projects/"+"fake_project_ID"+"/disks",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", true, "doc")
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set := flag.NewFlagSet("test", 0)
	set.String("name", "fake_disk_name", "disk name")
	set.String("flavor", "fake_flavor_name", "disk flavor")
	set.Int("capacityGB", 1, "disk capacity")
	set.String("affinities", "vm:fake_vm_id", "affinities")
	set.String("tenant", "fake_tenant_name", "tenant name")
	set.String("project", "fake_project_name", "project name")
	set.String("tags", "fake_disk_tag1, fake_disk_tag2", "Tags for disk")
	cxt := cli.NewContext(nil, set, globalCtx)

	err = createDisk(cxt)
	if err != nil {
		t.Error("Not expecting error creating project: " + err.Error())
	}
}

func TestShowDisk(t *testing.T) {
	diskStruct := photon.PersistentDisk{
		Name:       "fake_disk_name",
		ID:         "fake_disk_ID",
		Flavor:     "fake_flavor_name",
		Kind:       "persistent-disk",
		CapacityGB: 1,
		State:      "DETACHED",
		Datastore:  "fake_datastore_ID",
		Tags:       []string{"fake_disk_tag1", "fake_disk_tag2"},
		VMs:        []string{"fake_vm_id"},
	}
	response, err := json.Marshal(diskStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected disk")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/disks/"+"fake_disk_ID",
		mocks.CreateResponder(200, string(response[:])))

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_disk_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = showDisk(cxt)
	if err != nil {
		t.Error("Not expecting error showing disk: " + err.Error())
	}
}

func TestListDisks(t *testing.T) {
	diskList := MockDisksPage{
		Items: []photon.PersistentDisk{
			photon.PersistentDisk{
				Name:       "fake_disk_name",
				ID:         "fake_disk_ID",
				Flavor:     "fake_flavor_name",
				Kind:       "persistent-disk",
				CapacityGB: 1,
				State:      "DETACHED",
				Datastore:  "fake_datastore_ID",
			},
		},
		NextPageLink:     "fake-next-page-link",
		PreviousPageLink: "",
	}
	listResponse, err := json.Marshal(diskList)
	if err != nil {
		t.Error("Not expecting error serializaing expected disksList")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/projects/"+"fake_project_ID"+"/disks",
		mocks.CreateResponder(200, string(listResponse[:])))

	diskList = MockDisksPage{
		Items:            []photon.PersistentDisk{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}
	listResponse, err = json.Marshal(diskList)
	if err != nil {
		t.Error("Not expecting error serializaing expected disksList")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"fake-next-page-link",
		mocks.CreateResponder(200, string(listResponse[:])))

	set := flag.NewFlagSet("test", 0)
	set.String("tenant", "fake_tenant_name", "tenant name")
	set.String("project", "fake_project_name", "project name")
	cxt := cli.NewContext(nil, set, nil)

	err = listDisks(cxt)
	if err != nil {
		t.Error("Not expecting error listing disks: " + err.Error())
	}
}

func TestListDiskTasks(t *testing.T) {
	taskList := MockTasksPage{
		Items: []photon.Task{
			photon.Task{
				Operation: "CREATE_DISK",
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

	mocks.RegisterResponder(
		"GET",
		server.URL+"/disks/"+"fake_disk_ID"+"/tasks",
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
	err = set.Parse([]string{"fake_disk_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = getDiskTasks(cxt)
	if err != nil {
		t.Error("Not expecting error showing disk tasks: " + err.Error())
	}
}

func TestDeleteDisk(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "DELETE_DISK",
		State:     "QUEUED",
		ID:        "fake-disk-task-ID",
		Entity:    photon.Entity{ID: "fake_disk_ID"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "DELETE_DISK",
		State:     "COMPLETED",
		ID:        "fake-disk-task-ID",
		Entity:    photon.Entity{ID: "fake_disk_ID"},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/disks/"+"fake_disk_ID",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_disk_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	cxt := cli.NewContext(nil, set, nil)
	err = deleteDisk(cxt)
	if err != nil {
		t.Error("Not expecting error deleting disk: " + err.Error())
	}
}
