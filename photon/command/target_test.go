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
	"net/http/httptest"
	"testing"

	"github.com/vmware/photon-controller-cli/photon/client"
	cf "github.com/vmware/photon-controller-cli/photon/configuration"
	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

func TestSetEndpoint(t *testing.T) {
	var endpoint string

	configOri, err := cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	// test SetEndpoint when config file does not exist
	err = cf.RemoveConfigFile()
	if err != nil {
		t.Error("Not expecting error removing config file")
	}

	endpoint = "endpoint"
	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"endpoint"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set.Bool("nocertcheck", true, "")
	cxt := cli.NewContext(nil, set, nil)
	err = setEndpoint(cxt)
	if err != nil {
		t.Error("Not expecting error when setting endpoint")
	}

	configRead, err := cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	if configRead.CloudTarget != endpoint {
		t.Error("Endpoint read from file not match what's written to file")
	}

	// test SetEndpoint when overwriting existing endpoint
	configExpected := &cf.Configuration{
		CloudTarget:       "test-setendpoint",
		Token:             "test-setendpoint",
		IgnoreCertificate: true,
	}
	err = cf.SaveConfig(configExpected)
	if err != nil {
		t.Error("Not expecting error when saving config file")
	}

	endpoint = "endpoint-overwrite"
	set = flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"endpoint-overwrite"})
	if err != nil {
		t.Error("Not expecting arguments parsign to fail")
	}
	set.Bool("nocertcheck", true, "")
	cxt = cli.NewContext(nil, set, nil)
	err = setEndpoint(cxt)
	if err != nil {
		t.Error("Not expecting error when overwritting endpoint in file")
	}

	configRead, err = cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	if configRead.CloudTarget != endpoint {
		t.Error("Endpoint read from file not match what's written to file")
	}

	configRead.CloudTarget = configExpected.CloudTarget
	if *configRead != *configExpected {
		t.Error("Other configurations changed when setting only cloudtarget")
	}

	err = cf.SaveConfig(configOri)
	if err != nil {
		t.Error("Not expecting error when saving config file")
	}
}

func TestTargetShow(t *testing.T) {
	configOri, err := cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	configRead, err := cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	set := flag.NewFlagSet("test", 0)
	cxt := cli.NewContext(nil, set, nil)
	err = showEndpoint(cxt)
	if err != nil {
		t.Error("Not expecting error showing endpoint")
	}

	if configRead.CloudTarget != configOri.CloudTarget {
		t.Error("Endpoint should not have changed from show endpoint")
	}

	err = cf.SaveConfig(configOri)
	if err != nil {
		t.Error("Not expecting error when saving config file")
	}
}

func TestTargetInfo(t *testing.T) {
	server := mocks.NewTestServer()
	defer server.Close()

	// We first test that with exactly one deployment, we work as expected.
	// This is the expected case in a real installation
	err := mockInfo(t, server)
	if err != nil {
		t.Error("Failed to mock info: " + err.Error())
	}
	mocks.Activate(true)

	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, nil, httpClient)

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
	err = showInfo(cxt, &output)
	if err != nil {
		t.Error("showInfo ailed unexpectedly: " + err.Error())
	}

	// Verify we have the fields we expect
	err = checkRegExp(`\"baseVersion\":\s*\".*\"`, output)
	if err != nil {
		t.Errorf("Show info produce a JSON field named 'baseVersion': %s", err)
	}
	err = checkRegExp(`\"fullVersion\":\s*\".*\"`, output)
	if err != nil {
		t.Errorf("Show info produce a JSON field named 'fullVersion': %s", err)
	}
	err = checkRegExp(`\"gitCommitHash\":\s*\".*\"`, output)
	if err != nil {
		t.Errorf("Show info produce a JSON field named 'gitCommitHash': %s", err)
	}
	err = checkRegExp(`\"networkType\":\s*\".*\"`, output)
	if err != nil {
		t.Errorf("Show info produce a JSON field named 'networkType': %s", err)
	}
}

func TestLogin(t *testing.T) {
	var token string

	configOri, err := cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	// test Login when config file does not exist
	err = cf.RemoveConfigFile()
	if err != nil {
		t.Error("Not expecting error removing config file")
	}

	token = "token"
	set := flag.NewFlagSet("test", 0)
	set.String("access_token", token, "")
	cxt := cli.NewContext(nil, set, nil)
	err = login(cxt)
	if err != nil {
		t.Error("Not expecting error when logging in")
	}

	configRead, err := cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	if configRead.Token != token {
		t.Error("Token read from file not match what's written to file")
	}

	// test Login when overwriting existing endpoint
	configExpected := &cf.Configuration{
		CloudTarget: "test-login",
		Token:       "test-login",
	}
	err = cf.SaveConfig(configExpected)
	if err != nil {
		t.Error("Not expecting error when saving config file")
	}

	token = "token-overwrite"
	set = flag.NewFlagSet("test", 0)
	set.String("access_token", token, "")
	cxt = cli.NewContext(nil, set, nil)
	err = login(cxt)
	if err != nil {
		t.Error("Not expecting error when overwritting token in file")
	}

	configRead, err = cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	if configRead.Token != token {
		t.Error("Token read from file not match what's written to file")
	}

	configRead.Token = configExpected.Token
	if *configRead != *configExpected {
		t.Error("Other configurations changed when setting only token")
	}

	err = cf.SaveConfig(configOri)
	if err != nil {
		t.Error("Not expecting error when saving config file")
	}
}

func TestLogout(t *testing.T) {
	configOri, err := cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	// test Logout when config file does not exist
	err = cf.RemoveConfigFile()
	if err != nil {
		t.Error("Not expecting error removing config file")
	}
	set := flag.NewFlagSet("test", 0)
	cxt := cli.NewContext(nil, set, nil)
	err = logout(cxt)
	if err != nil {
		t.Error("Not expecting error when logging out")
	}

	configRead, err := cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	if configRead.Token != "" {
		t.Error("Token expected to be empty after logout")
	}

	// test Logout when config file exists
	configExpected := &cf.Configuration{
		CloudTarget: "test-logout",
		Token:       "test-logout",
	}
	err = cf.SaveConfig(configExpected)
	if err != nil {
		t.Error("Not expecting error when saving config file")
	}

	err = logout(cxt)
	if err != nil {
		t.Error("Not expecting error when logging out")
	}

	configRead, err = cf.LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file")
	}

	if configRead.Token != "" {
		t.Error("Token expected to be empty after logout")
	}

	configRead.Token = configExpected.Token
	if *configRead != *configExpected {
		t.Error("Other configurations changed when removing only token")
	}

	err = cf.SaveConfig(configOri)
	if err != nil {
		t.Error("Not expecting error when saving config file")
	}
}

func mockInfo(t *testing.T, server *httptest.Server) error {

	info := photon.Info{
		BaseVersion:   "1.1.0",
		FullVersion:   "1.1.0-12345abcde",
		GitCommitHash: "12345abcde",
		NetworkType:   "PHYSICAL",
	}

	response, err := json.Marshal(info)
	if err != nil {
		return err
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/info",
		mocks.CreateResponder(200, string(response[:])))
	return nil
}
