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
	"testing"

	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

func TestDeleteRouter(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "DELETE_ROUTER",
		State:     "QUEUED",
		ID:        "fake-router-task-id",
		Entity:    photon.Entity{ID: "fake_router_ID"},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "DELETE_ROUTER",
		State:     "COMPLETED",
		ID:        "fake-router-task-id",
		Entity:    photon.Entity{ID: "fake_router_ID"},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializing expected completedTask")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/routers/"+"fake_router_ID",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"fake_router_ID"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	cxt := cli.NewContext(nil, set, nil)
	err = deleteRouter(cxt)
	if err != nil {
		t.Error("Not expecting error deleting router: " + err.Error())
	}
}
