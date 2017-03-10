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
	"reflect"
	"testing"

	"github.com/vmware/photon-controller-cli/photon/client"
	cf "github.com/vmware/photon-controller-cli/photon/configuration"
	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

type MockTenantsPage struct {
	Items            []photon.Tenant `json:"items"`
	NextPageLink     string          `json:"nextPageLink"`
	PreviousPageLink string          `json:"previousPageLink"`
}

func TestCreateDeleteTenant(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "CREATE_TENANT",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask := &photon.Task{
		Operation: "CREATE_TENANT",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "1"},
	}
	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}
	taskresponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+rootUrl+"/tenants",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	set.String("security-groups", "a,b,c", "Comma-separated security group names")
	err = set.Parse([]string{"testname"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = createTenant(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting create tenant to fail")
	}

	queuedTask = &photon.Task{
		Operation: "DELETE_TENANT",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask = &photon.Task{
		Operation: "DELETE_TENANT",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "1"},
	}

	response, err = json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}
	taskresponse, err = json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+rootUrl+"/tenants/"+queuedTask.Entity.ID,
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))

	set = flag.NewFlagSet("test", 0)
	err = set.Parse([]string{queuedTask.Entity.ID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt = cli.NewContext(nil, set, nil)
	err = deleteTenant(cxt)
	if err != nil {
		t.Error("Not expecting delete tenant to fail")
	}
}

func TestShowTenant(t *testing.T) {
	tenantStruct := &photon.Tenant{
		Name: "fake_tenant_name",
		ID:   "fake_tenant_ID",
	}
	response, err := json.Marshal(tenantStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected tenant")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tenants"+"/"+"fake_tenant_ID",
		mocks.CreateResponder(200, string(response[:])))

	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_tenant_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = showTenant(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error showing tenant: " + err.Error())
	}
}

func TestListTenant(t *testing.T) {
	expectedTenants := MockTenantsPage{
		Items: []photon.Tenant{
			{
				Name: "testname",
				ID:   "1",
			},
			{
				Name: "secondname",
				ID:   "2",
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}

	response, err := json.Marshal(expectedTenants)
	if err != nil {
		t.Error("Not expecting error serializaing expected tenants")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tenants",
		mocks.CreateResponder(200, string(response[:])))

	expectedTenants = MockTenantsPage{
		Items:            []photon.Tenant{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}

	response, err = json.Marshal(expectedTenants)
	if err != nil {
		t.Error("Not expecting error serializaing expected tenants")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/fake-next-page-link",
		mocks.CreateResponder(200, string(response[:])))

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
	err = commandFlags.Parse([]string{})
	if err != nil {
		t.Error(err)
	}
	cxt := cli.NewContext(nil, commandFlags, globalCxt)
	var output bytes.Buffer
	err = listTenants(cxt, &output)
	if err != nil {
		t.Errorf("Not expecting list tenant to fail: %s", err)
	}

	// Verify we printed a list of tenants starting with a bracket
	err = checkRegExp(`^\s*\[`, output)
	if err != nil {
		t.Errorf("List tenants didn't produce a JSON list that starts with a bracket (list): %s", err)
	}
	// and end with a bracket (two regular expressions because it's multiline, it's easier)
	err = checkRegExp(`\]\s*$`, output)
	if err != nil {
		t.Errorf("List tenants didn't produce JSON that ended in a bracket (list): %s", err)
	}
	// And spot check that we have the "id" field
	err = checkRegExp(`\"id\":\s*\".*\"`, output)
	if err != nil {
		t.Errorf("List tenants didn't produce a JSON field named 'id': %s", err)
	}
}

func TestSetTenant(t *testing.T) {
	configRead, err := cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	initialEndpoint := configRead.CloudTarget

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"errorname"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = setTenant(cxt)
	if err == nil {
		t.Error("Expecting error should not set tenant")
	}

	tenant := &cf.TenantConfiguration{Name: "testname", ID: "1"}
	err = set.Parse([]string{"testname"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt = cli.NewContext(nil, set, nil)

	err = setTenant(cxt)
	if err != nil {
		t.Error("Not expecting setting tenant to fail")
	}

	configRead, err = cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	if !reflect.DeepEqual(configRead.Tenant, tenant) {
		t.Error("Tenant in config does not match what was to be written")
	}

	tenantOverwrite := &cf.TenantConfiguration{Name: "secondname", ID: "2"}
	err = set.Parse([]string{"secondname"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt = cli.NewContext(nil, set, nil)

	err = setTenant(cxt)
	if err != nil {
		t.Error("Not expecting setting tenant to fail")
	}

	configRead, err = cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	if !reflect.DeepEqual(configRead.Tenant, tenantOverwrite) {
		t.Error("Tenant in config does not match what was to be written")
	}

	if configRead.CloudTarget != initialEndpoint {
		t.Error("Cloud target should not have been modified while changing tenant")
	}
}

func TestSetTenantAfterDelete(t *testing.T) {
	configRead, err := cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	tenant := &cf.TenantConfiguration{Name: "testname", ID: "1"}

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"testname"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = setTenant(cxt)
	if err != nil {
		t.Error("Not expecting setting tenant to fail")
	}

	configRead, err = cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	if !reflect.DeepEqual(configRead.Tenant, tenant) {
		t.Error("Tenant in config does not match what was to be written")
	}

	completedTask := &photon.Task{
		Operation: "DELETE_TENANT",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "1"},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"DELETE",
		server.URL+rootUrl+"/tenants/"+completedTask.Entity.ID,
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+completedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	err = set.Parse([]string{"1"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt = cli.NewContext(nil, set, nil)

	err = deleteTenant(cxt)
	if err != nil {
		t.Error("Not expecting delete Tenant to fail", err)
	}

	configRead, err = cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	if configRead.Tenant != nil {
		t.Error("Tenant in config does not match what was to be written")
	}
}

func TestTenantTasks(t *testing.T) {
	taskList := MockTasksPage{
		Items: []photon.Task{
			{
				Operation: "CREATE_TENANT",
				State:     "COMPLETED",
				ID:        "1",
				Entity:    photon.Entity{ID: "1", Kind: "tenant"},
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}

	response, err := json.Marshal(taskList)
	if err != nil {
		t.Error("Not expecting error serializaing expected status")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tenants/1/tasks",
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
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"1"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = getTenantTasks(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error retrieving tenant tasks")
	}
}

func TestSetSecurityGroups(t *testing.T) {
	taskId := "task1"
	tenantId := "tenant1"
	completedTask := photon.Task{
		ID:        taskId,
		Operation: "PUSH_TENANT_SECURITY_GROUPS",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: tenantId},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected status")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+rootUrl+"/tenants/"+tenantId+"/set_security_groups",
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
	err = set.Parse([]string{tenantId, "sg1"})
	if err != nil {
		t.Error(err)
	}
	cxt := cli.NewContext(nil, set, nil)
	err = setSecurityGroups(cxt)
	if err != nil {
		t.Error(err)
	}
}
