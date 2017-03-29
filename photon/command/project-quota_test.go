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

// create mock quota instance
func mockProjectQuota() photon.Quota {
	mockQuota := photon.Quota{
		QuotaLineItems: map[string]photon.QuotaStatusLineItem{
			"vmCpu":        {Unit: "COUNT", Limit: 100, Usage: 0},
			"vmMemory":     {Unit: "GB", Limit: 180, Usage: 0},
			"diskCapacity": {Unit: "GB", Limit: 1000, Usage: 0},
		},
	}
	return mockQuota
}

// Test Project Quota can be retrieved.
func TestGetProjectQuota(t *testing.T) {
	projectName := "fake_project_Name"
	projectID := "fake_project_ID"

	// response for the project query
	projectStruct := photon.ProjectCompact{
		Name: projectName,
		ID:   projectID,
	}
	projectResponse, err := json.Marshal(projectStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected projectStruct")
	}

	// response for the quota query
	mockQuota := mockProjectQuota()
	getResponse, err := json.Marshal(mockQuota)
	if err != nil {
		t.Error("Not expecting error serializaing expected quota")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/projects/"+projectID,
		mocks.CreateResponder(200, string(projectResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/projects/"+projectID+"/quota",
		mocks.CreateResponder(200, string(getResponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{projectID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = getProjectQuota(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error showing quota: " + err.Error())
	}
}

// Test ProjectQuota can be set (overwrite).
func TestSetProjectQuota(t *testing.T) {
	projectName := "fake_project_Name"
	projectID := "fake_project_ID"

	// response for the project query
	projectStruct := photon.ProjectCompact{
		Name: projectName,
		ID:   projectID,
	}
	projectResponse, err := json.Marshal(projectStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected projectStruct")
	}

	queuedTask := &photon.Task{
		Operation: "MODIFY_QUOTA",
		State:     "QUEUED",
		ID:        "fake_task_Id",
		Entity:    photon.Entity{ID: projectID},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "MODIFY_QUOTA",
		State:     "COMPLETED",
		ID:        "fake_task_Id",
		Entity:    photon.Entity{ID: projectID},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	// mock the response quota
	mockQuota := photon.Quota{
		QuotaLineItems: map[string]photon.QuotaStatusLineItem{
			"vmCpu":    {Unit: "COUNT", Limit: 10, Usage: 0},
			"vmMemory": {Unit: "GB", Limit: 20, Usage: 0},
		},
	}
	getResponse, err := json.Marshal(mockQuota)

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/projects/"+projectID,
		mocks.CreateResponder(200, string(projectResponse[:])))
	mocks.RegisterResponder(
		"PUT",
		server.URL+rootUrl+"/projects/"+projectID+"/quota",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/projects/"+projectID+"/quota",
		mocks.CreateResponder(200, string(getResponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", true, "doc")
	globalSet.String("output", "json", "doc")
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	set := flag.NewFlagSet("test", 0)
	set.String("limits", "vmCpu 10 COUNT, vmMemory 20 GB", "quota limits")
	err = set.Parse([]string{projectID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	cxt := cli.NewContext(nil, set, globalCtx)

	err = setProjectQuota(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error creating resource ticket: " + err.Error())
	}
}

// Test ProjectQuota can be partially updated.
func TestUpdateProjectQuota(t *testing.T) {
	projectName := "fake_project_Name"
	projectID := "fake_project_ID"

	// response for the project query
	projectStruct := photon.ProjectCompact{
		Name: projectName,
		ID:   projectID,
	}
	projectResponse, err := json.Marshal(projectStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected projectStruct")
	}

	queuedTask := &photon.Task{
		Operation: "MODIFY_QUOTA",
		State:     "QUEUED",
		ID:        "fake_task_Id",
		Entity:    photon.Entity{ID: projectID},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "MODIFY_QUOTA",
		State:     "COMPLETED",
		ID:        "fake_task_Id",
		Entity:    photon.Entity{ID: projectID},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	// mock the response quota
	mockQuota := photon.Quota{
		QuotaLineItems: map[string]photon.QuotaStatusLineItem{
			"vmCpu":    {Unit: "COUNT", Limit: 100, Usage: 0},
			"vmMemory": {Unit: "GB", Limit: 200, Usage: 0},
		},
	}
	getResponse, err := json.Marshal(mockQuota)

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/projects/"+projectID,
		mocks.CreateResponder(200, string(projectResponse[:])))
	mocks.RegisterResponder(
		"PATCH",
		server.URL+rootUrl+"/projects/"+projectID+"/quota",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/projects/"+projectID+"/quota",
		mocks.CreateResponder(200, string(getResponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", true, "doc")
	globalSet.String("output", "json", "doc")
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	set := flag.NewFlagSet("test", 0)
	set.String("limits", "vmCpu 100 COUNT, vmMemory 200 GB", "quota limits")
	err = set.Parse([]string{projectID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, globalCtx)

	err = updateProjectQuota(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error creating resource ticket: " + err.Error())
	}
}

// Test quota line items in ProjectQuota can be excluded.
func TestExcludeProjectQuota(t *testing.T) {
	projectName := "fake_project_Name"
	projectID := "fake_project_ID"

	// response for the project query
	projectStruct := photon.ProjectCompact{
		Name: projectName,
		ID:   projectID,
	}
	projectResponse, err := json.Marshal(projectStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected projectStruct")
	}

	queuedTask := &photon.Task{
		Operation: "MODIFY_QUOTA",
		State:     "QUEUED",
		ID:        "fake_task_Id",
		Entity:    photon.Entity{ID: projectID},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "MODIFY_QUOTA",
		State:     "COMPLETED",
		ID:        "fake_task_Id",
		Entity:    photon.Entity{ID: projectID},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	// mock the response quota
	mockQuota := photon.Quota{
		QuotaLineItems: map[string]photon.QuotaStatusLineItem{
			"vmCpu":    {Unit: "COUNT", Limit: 10, Usage: 0},
			"vmMemory": {Unit: "GB", Limit: 20, Usage: 0},
		},
	}
	getResponse, err := json.Marshal(mockQuota)

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/projects/"+projectID,
		mocks.CreateResponder(200, string(projectResponse[:])))
	mocks.RegisterResponder(
		"DELETE",
		server.URL+rootUrl+"/projects/"+projectID+"/quota",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/projects/"+projectID+"/quota",
		mocks.CreateResponder(200, string(getResponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", true, "doc")
	globalSet.String("output", "json", "doc")
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	set := flag.NewFlagSet("test", 0)
	set.String("limits", "diskCapacity 200 GB", "quota limits")
	err = set.Parse([]string{projectID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	cxt := cli.NewContext(nil, set, globalCtx)
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	err = excludeProjectQuota(cxt, os.Stdout)
	if err != nil {
		t.Error("Not expecting error creating resource ticket: " + err.Error())
	}
}
