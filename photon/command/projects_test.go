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
	"testing"

	"github.com/vmware/photon-controller-cli/photon/client"
	cf "github.com/vmware/photon-controller-cli/photon/configuration"
	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

type MockProjectsPage struct {
	Items            []photon.ProjectCompact `json:"items"`
	NextPageLink     string                  `json:"nextPageLink"`
	PreviousPageLink string                  `json:"previousPageLink"`
}

func TestCreateProject(t *testing.T) {
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

	queuedTask := &photon.Task{
		Operation: "CREATE_PROJECT",
		State:     "QUEUED",
		ID:        "fake-project-task-id",
		Entity:    photon.Entity{ID: "fake_project_ID"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "CREATE_PROJECT",
		State:     "COMPLETED",
		ID:        "fake-project-task-id",
		Entity:    photon.Entity{ID: "fake_project_ID"},
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
		server.URL+"/tenants/"+"fake_tenant_ID"+"/projects",
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
	set.String("name", "fake_project_name", "project name")
	set.String("resource-ticket", "fake_rt_Name", "rt name")
	set.String("tenant", "fake_tenant_name", "tenant name")
	set.String("limits", "vm.test1 1 B", "project limits")
	set.String("security-groups", "fake_security_group", "security groups")
	cxt := cli.NewContext(nil, set, globalCtx)

	err = createProject(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error creating project: " + err.Error())
	}
}

func TestShowProject(t *testing.T) {
	projectStruct := photon.ProjectCompact{
		Name: "fake_project_name",
		ID:   "fake_project_ID",
		ResourceTicket: photon.ProjectTicket{
			TenantTicketID:   "fake_ticket_id",
			TenantTicketName: "fake_ticket_name",
			Limits:           []photon.QuotaLineItem{{Key: "vm.test1", Value: 1, Unit: "B"}},
			Usage:            []photon.QuotaLineItem{{Key: "vm.test1", Value: 0, Unit: "B"}},
		},
	}
	response, err := json.Marshal(projectStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected project")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/projects/"+"fake_project_ID",
		mocks.CreateResponder(200, string(response[:])))

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_project_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = showProject(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error showing project: " + err.Error())
	}
}

func TestSetGetProject(t *testing.T) {
	configOri, err := cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
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
	listResponse, err := json.Marshal(projectListStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected projectLists")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants/"+"fake_tenant_ID"+"/projects",
		mocks.CreateResponder(200, string(listResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants/"+"fake_tenant_ID"+"/projects?name="+"fake_project_name",
		mocks.CreateResponder(200, string(listResponse[:])))

	config := &cf.Configuration{
		Tenant: &cf.TenantConfiguration{Name: "fake_tenant_name", ID: "fake_tenant_ID"},
	}
	err = cf.SaveConfig(config)
	if err != nil {
		t.Error("Not expecting error when saving config file")
	}

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_project_name"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = setProject(cxt)
	if err != nil {
		t.Error("Not expecting error setting project: " + err.Error())
	}

	set = flag.NewFlagSet("test", 0)
	cxt = cli.NewContext(nil, set, nil)

	err = getProject(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error showing project: " + err.Error())
	}

	err = cf.SaveConfig(configOri)
	if err != nil {
		t.Error("Not expecting error when saving config file")
	}
}

func TestListProjects(t *testing.T) {
	projectList := MockProjectsPage{
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
		NextPageLink:     "fake-next-page-link",
		PreviousPageLink: "",
	}
	listResponse, err := json.Marshal(projectList)
	if err != nil {
		t.Error("Not expecting error serializaing expected projectList")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants/"+"fake_tenant_ID"+"/projects",
		mocks.CreateResponder(200, string(listResponse[:])))

	projectList = MockProjectsPage{
		Items:            []photon.ProjectCompact{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}
	listResponse, err = json.Marshal(projectList)
	if err != nil {
		t.Error("Not expecting error serializaing expected projectList")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"fake-next-page-link",
		mocks.CreateResponder(200, string(listResponse[:])))

	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalFlags := flag.NewFlagSet("global-flags", flag.ContinueOnError)
	globalFlags.String("output", "json", "output")
	err = globalFlags.Parse([]string{"--output=json"})
	if err != nil {
		t.Error(err)
	}
	globalCxt := cli.NewContext(nil, globalFlags, nil)
	commandFlags := flag.NewFlagSet("command-flags", flag.ContinueOnError)
	commandFlags.String("tenant", "fake_tenant_name", "tenant name")
	err = commandFlags.Parse([]string{})
	if err != nil {
		t.Error(err)
	}
	cxt := cli.NewContext(nil, commandFlags, globalCxt)

	var output bytes.Buffer

	err = listProjects(cxt, &output)
	if err != nil {
		t.Error("Not expecting error listing projects: " + err.Error())
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

func TestListProjectTasks(t *testing.T) {
	taskList := MockTasksPage{
		Items: []photon.Task{
			{
				Operation: "CREATE_PROJECT",
				State:     "COMPLETED",
				ID:        "fake_project_task_id",
				Entity:    photon.Entity{ID: "fake_project_ID", Kind: "project"},
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
		server.URL+"/projects/"+"fake_project_ID"+"/tasks",
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
	err = set.Parse([]string{"fake_project_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = getProjectTasks(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error showing project tasks: " + err.Error())
	}
}

func TestDeleteProject(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "DELETE_PROJECT",
		State:     "QUEUED",
		ID:        "fake-project-task-id",
		Entity:    photon.Entity{ID: "fake_project_ID"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "DELETE_PROJECT",
		State:     "COMPLETED",
		ID:        "fake-project-task-id",
		Entity:    photon.Entity{ID: "fake_project_ID"},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/projects/"+"fake_project_ID",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_project_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	cxt := cli.NewContext(nil, set, nil)
	err = deleteProject(cxt)
	if err != nil {
		t.Error("Not expecting error deleting project: " + err.Error())
	}
}

func TestSetSecurityGroupsForProject(t *testing.T) {
	taskId := "task1"
	projectId := "project1"
	completedTask := photon.Task{
		ID:        taskId,
		Operation: "PUSH_TENANT_SECURITY_GROUPS",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: projectId},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected status")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+"/projects/"+projectId+"/set_security_groups",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+taskId,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{projectId, "sg1"})
	if err != nil {
		t.Error(err)
	}
	cxt := cli.NewContext(nil, set, nil)
	err = setSecurityGroupsForProject(cxt)
	if err != nil {
		t.Error(err)
	}
}
