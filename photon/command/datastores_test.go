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
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

func TestListDatastores(t *testing.T) {
	server := mocks.NewTestServer()
	defer server.Close()

	err := mockDatastoresForList(t, server)
	if err != nil {
		t.Error("Failed to mock datastores: " + err.Error())
	}

	mocks.Activate(true)

	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	globalFlags := flag.NewFlagSet("global-flags", flag.ContinueOnError)
	globalFlags.String("output", "json", "output")
	err = globalFlags.Parse([]string{"--output=json"})
	if err != nil {
		t.Error(err)
	}
	globalCtx := cli.NewContext(nil, globalFlags, nil)
	commandFlags := flag.NewFlagSet("command-flags", flag.ContinueOnError)
	err = commandFlags.Parse([]string{})
	if err != nil {
		t.Error(err)
	}
	cxt := cli.NewContext(nil, commandFlags, globalCtx)
	var output bytes.Buffer
	err = listDatastores(cxt, &output)
	if err != nil {
		t.Error("listDatastores failed unexpectedly: " + err.Error())
	}

	// Verify we printed a list of datastores starting with a bracket
	err = checkRegExp(`^\s*\[`, output)
	if err != nil {
		t.Errorf("List datastores didn't produce a JSON list that starts with a bracket (list): %s", err)
	}
	// and end with a bracket
	err = checkRegExp(`\]\s*$`, output)
	if err != nil {
		t.Errorf("List datastores didn't produce JSON that ended in a bracket (list): %s", err)
	}
	// And spot check that we have the "id" field
	err = checkRegExp(`\"id\":\s*\".*\"`, output)
	if err != nil {
		t.Errorf("List datastores didn't produce a JSON field named 'id': %s", err)
	}
}

func mockDatastoresForList(t *testing.T, server *httptest.Server) error {
	datastoresList := photon.Datastores{
		Items: []photon.Datastore{
			{
				Kind: "datastore",
				Type: "LOCAL_VMFS",
				Tags: []string{"LOCAL_VMFS"},
				ID:   "1234",
			},
		},
	}

	response, err := json.Marshal(datastoresList)
	if err != nil {
		return err
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/infrastructure"+"/datastores",
		mocks.CreateResponder(200, string(response[:])))

	return nil
}

func TestShowDatastore(t *testing.T) {
	expectedStruct := photon.Datastore{
		Kind: "datastore",
		Type: "LOCAL_VMFS",
		Tags: []string{"LOCAL_VMFS"},
		ID:   "1234",
	}

	response, err := json.Marshal(expectedStruct)
	if err != nil {
		log.Fatal("Not expecting error serializing expected response")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+rootUrl+"/infrastructure"+"/datastores/"+expectedStruct.ID,
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
	globalCtx := cli.NewContext(nil, globalFlags, nil)
	commandFlags := flag.NewFlagSet("command-flags", flag.ContinueOnError)
	err = commandFlags.Parse([]string{expectedStruct.ID})
	if err != nil {
		log.Fatal("Not expecting arguments parsing to fail")
	}

	cxt := cli.NewContext(nil, commandFlags, globalCtx)
	var output bytes.Buffer
	err = showDatastore(cxt, &output)
	if err != nil {
		t.Error("Error showing datastore: " + err.Error())
	}

	// Verify we printed a datastore starting with a brace
	err = checkRegExp(`^\s*\{`, output)
	if err != nil {
		t.Errorf("Show datastore didn't produce a JSON object that starts with a brace: %s", err)
	}
	// and end with a brace
	err = checkRegExp(`\}\s*$`, output)
	if err != nil {
		t.Errorf("Show datastore didn't produce a JSON object that ended in a bracket: %s", err)
	}
	// And spot check that we have the "id" field
	err = checkRegExp(`\"id\":\s*\".*\"`, output)
	if err != nil {
		t.Errorf("Show datastore didn't produce a JSON field named 'id': %s", err)
	}
}
