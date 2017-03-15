// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

type MockRoutersPage struct {
	Items            []photon.Router `json:"items"`
	NextPageLink     string          `json:"nextPageLink"`
	PreviousPageLink string          `json:"previousPageLink"`
}

func TestCreateDeleteRouter(t *testing.T) {
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
		Operation: "CREATE_ROUTER",
		State:     "QUEUED",
		ID:        "fake-router-task-id",
		Entity:    photon.Entity{ID: "fake-router-id"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "CREATE_ROUTER",
		State:     "COMPLETED",
		ID:        "fake-router-task-id",
		Entity:    photon.Entity{ID: "fake-router-id"},
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
		"GET",
		server.URL+rootUrl+"/tenants/"+"fake_tenant_ID"+"/projects?name="+"fake_project_name",
		mocks.CreateResponder(200, string(projectResponse[:])))
	mocks.RegisterResponder(
		"POST",
		server.URL+rootUrl+"/projects/"+"fake_project_ID"+"/routers",
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
	set.String("name", "fake_router_name", "Router name")
	set.String("privateIpCidr", "fake_router_privateIpCidr", "Router privateIpCidr")
	set.String("tenant", "fake_tenant_name", "tenant name")
	set.String("project", "fake_project_name", "project name")
	cxt := cli.NewContext(nil, set, globalCtx)

	err = createRouter(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error creating Router: " + err.Error())
	}

	queuedTask = &photon.Task{
		Operation: "DELETE_ROUTER",
		State:     "QUEUED",
		ID:        "fake-router-task-id",
		Entity:    photon.Entity{ID: "fake-router-id"},
	}
	taskResponse, err = json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}

	completedTask = &photon.Task{
		Operation: "DELETE_ROUTER",
		State:     "COMPLETED",
		ID:        "fake-router-task-id",
		Entity:    photon.Entity{ID: "fake-router-id"},
	}
	response, err = json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completedTask")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+rootUrl+"/routers/"+"fake-router-id",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set = flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake-router-id"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	cxt = cli.NewContext(nil, set, nil)
	err = deleteRouter(cxt)
	if err != nil {
		t.Error("Not expecting error deleting router: " + err.Error())
	}
}

func TestUpdateRouter(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "UPDATE_ROUTER",
		State:     "QUEUED",
		ID:        "fake-task-id",
		Entity:    photon.Entity{ID: "fake-router-id"},
	}
	completedTask := &photon.Task{
		Operation: "UPDATE_ROUTER",
		State:     "COMPLETED",
		ID:        "fake-task-id",
		Entity:    photon.Entity{ID: "fake-router-id"},
	}

	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected createTask")
	}
	taskResponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"PUT",
		server.URL+rootUrl+"/routers"+"/fake-router-id",
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
	err = set.Parse([]string{"fake-router-id"})
	set.String("name", "router-1", "router name")
	cxt := cli.NewContext(nil, set, nil)

	err = updateRouter(cxt, os.Stdout)
	if err != nil {
		t.Error("Error listing routers: " + err.Error())
	}
}

func TestShowRouter(t *testing.T) {
	getStruct := photon.Router{
		Name:          "routername",
		ID:            "1",
		Kind:          "router",
		PrivateIpCidr: "cidr1",
	}

	response, err := json.Marshal(getStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/routers/"+getStruct.ID,
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

	err = showRouter(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting get deployment to fail")
	}
}

func TestListRouters(t *testing.T) {
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

	routerList := MockRoutersPage{
		Items: []photon.Router{
			{
				Name:          "fake_router_name",
				ID:            "fake_router_ID",
				Kind:          "fake_router_kind",
				PrivateIpCidr: "fake_cidr",
			},
		},
		NextPageLink:     "fake-next-page-link",
		PreviousPageLink: "",
	}
	listResponse, err := json.Marshal(routerList)
	if err != nil {
		t.Error("Not expecting error serializaing expected routerList")
	}

	server := mocks.NewTestServer()
	defer server.Close()

	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tenants",
		mocks.CreateResponder(200, string(tenantResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tenants/"+"fake_tenant_ID"+"/projects?name="+"fake_project_name",
		mocks.CreateResponder(200, string(projectResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/projects/"+"fake_project_ID"+"/routers",
		mocks.CreateResponder(200, string(listResponse[:])))

	routerList = MockRoutersPage{
		Items:            []photon.Router{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}
	listResponse, err = json.Marshal(routerList)
	if err != nil {
		t.Error("Not expecting error serializaing expected routerList")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"fake-next-page-link",
		mocks.CreateResponder(200, string(listResponse[:])))

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	set.String("tenant", "fake_tenant_name", "tenant name")
	set.String("project", "fake_project_name", "project name")
	cxt := cli.NewContext(nil, set, nil)

	err = listRouters(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error listing routers: " + err.Error())
	}
}
