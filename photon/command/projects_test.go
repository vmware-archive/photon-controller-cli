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

	"github.com/urfave/cli"
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
				ResourceQuota: photon.Quota{
					QuotaLineItems: photon.QuotaSpec{
						"vm.test1": {Limit: 100, Usage: 0, Unit: "B"},
						"vm.cpu":   {Limit: 100, Usage: 0, Unit: "COUNT"},
					},
				},
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
		server.URL+rootUrl+"/tenants",
		mocks.CreateResponder(200, string(tenantResponse[:])))
	mocks.RegisterResponder(
		"POST",
		server.URL+rootUrl+"/tenants/"+"fake_tenant_ID"+"/projects",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
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
	err = set.Parse([]string{"fake_project_name"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	set.String("tenant", "fake_tenant_name", "tenant name")
	set.String("limits", "vm.test1 1 B", "project limits")
	set.String("security-groups", "fake_security_group", "security groups")
	set.String("default-router-private-ip-cidr", "192.168.0.0/24", "default router private ip cidr")
	cxt := cli.NewContext(nil, set, globalCtx)

	err = createProject(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error creating project: " + err.Error())
	}
}

func TestCreateProjectUsingPercentageOfTenantQuota(t *testing.T) {
	tenantStruct := photon.Tenants{
		Items: []photon.Tenant{
			{
				Name: "fake_tenant_name",
				ID:   "fake_tenant_ID",
				ResourceQuota: photon.Quota{
					QuotaLineItems: photon.QuotaSpec{
						"vm.test1": {Limit: 100, Usage: 0, Unit: "B"},
						"vm.cpu":   {Limit: 100, Usage: 0, Unit: "COUNT"},
					},
				},
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

	// mock the response quota
	mockQuota := photon.Quota{
		QuotaLineItems: map[string]photon.QuotaStatusLineItem{
			"vm.test1": {Unit: "COUNT", Limit: 100, Usage: 0},
			"vm.cpu":   {Unit: "GB", Limit: 100, Usage: 0},
		},
	}
	quotaResponse, err := json.Marshal(mockQuota)

	server = mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tenants",
		mocks.CreateResponder(200, string(tenantResponse[:])))
	mocks.RegisterResponder(
		"POST",
		server.URL+rootUrl+"/tenants/"+"fake_tenant_ID"+"/projects",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tenants/"+"fake_tenant_ID"+"/quota",
		mocks.CreateResponder(200, string(quotaResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
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
	err = set.Parse([]string{"fake_project_name"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	set.String("tenant", "fake_tenant_name", "tenant name")
	set.String("percent", "50", "percentage of tenant limits")
	set.String("security-groups", "fake_security_group", "security groups")
	set.String("default-router-private-ip-cidr", "192.168.0.0/24", "default router private ip cidr")
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
		ResourceQuota: photon.Quota{
			QuotaLineItems: photon.QuotaSpec{
				"vm.test1": {Limit: 1, Usage: 0, Unit: "B"},
				"vm.cpu":   {Limit: 10, Usage: 0, Unit: "COUNT"},
			},
		},
	}
	response, err := json.Marshal(projectStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected project")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/projects/"+"fake_project_ID",
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
				ResourceQuota: photon.Quota{
					QuotaLineItems: photon.QuotaSpec{
						"vm.test1": {Limit: 1, Usage: 0, Unit: "B"},
					},
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
		server.URL+rootUrl+"/tenants/"+"fake_tenant_ID"+"/projects",
		mocks.CreateResponder(200, string(listResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tenants/"+"fake_tenant_ID"+"/projects?name="+"fake_project_name",
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
				ResourceQuota: photon.Quota{
					QuotaLineItems: photon.QuotaSpec{
						"vm.test1": {Limit: 1, Usage: 0, Unit: "B"},
					},
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
		server.URL+rootUrl+"/tenants/"+"fake_tenant_ID"+"/projects",
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
		server.URL+rootUrl+"/projects/"+"fake_project_ID"+"/tasks",
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
		server.URL+rootUrl+"/projects/"+"fake_project_ID",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
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
		server.URL+rootUrl+"/projects/"+projectId+"/set_security_groups",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+taskId,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

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
