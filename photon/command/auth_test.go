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
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/configuration"
	"github.com/vmware/photon-controller-cli/photon/mocks"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

// This access token has to two JSON-encoded chunks plus a signature:
// {
//   "alg": "RS256"
// }
// {
//   "sub": "ec-admin@esxcloud",
//   "aud": [
//     "ec-admin@esxcloud",
//     "rs_esxcloud"
//   ],
//   "scope": "openid offline_access rs_esxcloud at_groups",
//   "iss": "https:\/\/10.146.64.238\/openidconnect\/esxcloud",
//   "groups": [
//     "esxcloud\\ESXCloudAdmins",
//     "esxcloud\\Everyone"
//   ],
//   "token_class": "access_token",
//   "token_type": "Bearer",
//   "exp": 1461803127,
//   "iat": 1461795927,
//   "tenant": "esxcloud",
//   "jti": "CfPby7BAlaOI3Uj_TEq_UJOJmYXJiVOYuCYAXPw2l2U"
// }
var accessToken = "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJlYy1hZG1pbkBlc3hjbG91ZCIsImF1ZCI6WyJlYy1hZG1pbkB" +
	"lc3hjbG91ZCIsInJzX2VzeGNsb3VkIl0sInNjb3BlIjoib3BlbmlkIG9mZmxpbmVfYWNjZXNzIHJzX2VzeGNsb3VkIGF0X" +
	"2dyb3VwcyIsImlzcyI6Imh0dHBzOlwvXC8xMC4xNDYuNjQuMjM4XC9vcGVuaWRjb25uZWN0XC9lc3hjbG91ZCIsImdyb3V" +
	"wcyI6WyJlc3hjbG91ZFxcRVNYQ2xvdWRBZG1pbnMiLCJlc3hjbG91ZFxcRXZlcnlvbmUiXSwidG9rZW5fY2xhc3MiOiJhY" +
	"2Nlc3NfdG9rZW4iLCJ0b2tlbl90eXBlIjoiQmVhcmVyIiwiZXhwIjoxNDYxODAzMTI3LCJpYXQiOjE0NjE3OTU5MjcsInR" +
	"lbmFudCI6ImVzeGNsb3VkIiwianRpIjoiQ2ZQYnk3QkFsYU9JM1VqX1RFcV9VSk9KbVlYSmlWT1l1Q1lBWFB3MmwyVSJ9." +
	"QOpb-8L8if1kEHPEQvsGe_Z8v_gdlPDpjWcu8LxMnAxZELQx6YBn7UM2MO83Qgo-0bqu2ysbcSpjz0mP4pf48z_DyKlMCa" +
	"B6ViStwavIx7lM1TENrt5nURpjqxlzQY0CxjyYIWxoYQIUbn7c5MXe-vt-OTXAg8bGkwphltj7xUak90mQlZGSBrHFCT_Q" +
	"PGwxRTNsRwWq45tF7LgKr49L4z5PnkLQ3LpC8jI7x1SUFBiYcJgi76pGNlD4qihpmKhGJK0WpspEAvXhtsGwBVavGxeXzL" +
	"-PBTYz7Zs1EjD4Isar-91pq-HeTVfhd_KBBqktaQq0WO48Vu0KtHHRv_Us90-Qs53gsY0CnrxHV8qyNR27LyaIMWhG24hq" +
	"TyBsZVgT-gzs9_-QdLqtkXNgr4Oiqoy9Gi8LAmARGFCgTXOS7uPqZ6_ut71WPhwwoUIuXVUG8vvuRD6_UIIGXyPjBM0sfg" +
	"X5rMeo45bYO51mNjqAysz7FBwMetkZUqKg6pxWmTmO_xnH5D55I1P2zd_VBo5be-hr7jjTqqDAGkGMU0PM8IajpnWe24wu" +
	"lPzQqRr5-HlQx50B0nwhYFJVCd_3KW6qCw-MmfGB-1aX-GVG2wa_vUKzc4gDDn65-z0rP_gYtrB9q8oNR-hPY4v18DQEdY" +
	"bsuoJoqriXk1A0zkeoX13kFXY"

// This access token has two JSON-encoded chunks plus a signature:
// {
//   "alg": "RS256"
// }
// {
//   "sub": "ec-admin@esxcloud",
//   "aud": "ec-admin@esxcloud",
//   "scope": "openid offline_access rs_esxcloud at_groups",
//   "iss": "https:\/\/10.146.64.238\/openidconnect\/esxcloud",
//   "token_class": "refresh_token",
//   "token_type": "Bearer",
//   "exp": 1461817527,
//   "iat": 1461795927,
//   "tenant": "esxcloud",
//   "jti": "yE2jPR59JYn-c3y5zirI1KwZGQ8HUi7o2TEwLfnC5OI"
// }
var refreshToken = "eyJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJlYy1hZG1pbkBlc3hjbG91ZCIsImF1ZCI6ImVjLWFkbWluQG" +
	"VzeGNsb3VkIiwic2NvcGUiOiJvcGVuaWQgb2ZmbGluZV9hY2Nlc3MgcnNfZXN4Y2xvdWQgYXRfZ3JvdXBzIiwiaXNzIjoi" +
	"aHR0cHM6XC9cLzEwLjE0Ni42NC4yMzhcL29wZW5pZGNvbm5lY3RcL2VzeGNsb3VkIiwidG9rZW5fY2xhc3MiOiJyZWZyZX" +
	"NoX3Rva2VuIiwidG9rZW5fdHlwZSI6IkJlYXJlciIsImV4cCI6MTQ2MTgxNzUyNywiaWF0IjoxNDYxNzk1OTI3LCJ0ZW5h" +
	"bnQiOiJlc3hjbG91ZCIsImp0aSI6InlFMmpQUjU5SlluLWMzeTV6aXJJMUt3WkdROEhVaTdvMlRFd0xmbkM1T0kifQ.Ran" +
	"j-Wfqt2-uX5WUvXYSZhBwUrz6mrtntKApWrxXbPK-LMsD9HuQLJVz9XJRrcmcPbeKOgnGxZU-GOPNbwFYNLHKTUmrg4Pwj" +
	"t81xMQFygIUNgblrjK7SnsmmqnPA5t09nPrLD3usA3pc00nFQg0ml1K16zHVhx4l6Ppd6nzajOD25poIKYypli2kGGpWSd" +
	"nIz1Jnb4ipzngDnGHl8LCAUtuotCz4rK6qXetk8dKoQAwIb5l5SqCxR0q6cqPLIup-q4McEq5y8-7irviCp-VVF-y-2Lr3" +
	"5inhZTcVv4B47TB-1qwe35HKWbCDg4T02PDSQslu54wfKpOfgpcRSkgsxxPEkcI8dTTjKBfdaXceLSLb44Xzzw3uaSG6Mq" +
	"aLbwXSw0c3vrpAhgq3ZqSkLn59D7g1PnFfjc9tXnZejdjYAS1yKBRHMfw_rUey9-dd1iTPBGivk1gNLRRV0iNZAX5zttQB" +
	"KAuraFJFXnY9_elRvp-tfwQlBHpVVzaSrxKQg9Bm-HMmjxnRsA9i-uOm4tGtPh6tmF7Z-4rcj_2imwQSgv8ydp8Qk55a1Y" +
	"SvfvqAwOaq3XNsHQKSgt7Mxxtz72CZxnci4LHWTytUHQOJtd6pYa667A6Dqa0QgH11WOOvgZE_G2oLpTq2-BHPlxl6Csom" +
	"pBXzdiTYWQ-pe0nPmAZY0I"

const expectedSubject = "ec-admin@esxcloud"
const expectedAdminGroup = `esxcloud\\ESXCloudAdmins`
const expectedEveryoneGroup = `esxcloud\\Everyone`
const expectedDate = `2016-04-27`

func TestShow(t *testing.T) {
	authInfo := photon.AuthInfo{}
	response, err := json.Marshal(authInfo)
	if err != nil {
		t.Error("Not expecting error serializing expected status")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/auth",
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Photonclient = photon.NewTestClient(server.URL, nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	cxt := cli.NewContext(nil, set, nil)
	err = show(cxt)
	if err != nil {
		t.Error(err)
	}
}

/*
 * This does spot checking (not 100% thorough, since the format will change)
 * on the output of "show-login-token"
 */
func TestShowLoginToken(t *testing.T) {
	config := configuration.Configuration{
		Token: accessToken,
	}

	globalFlags := flag.NewFlagSet("global-flags", flag.ContinueOnError)
	err := globalFlags.Parse([]string{})
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
	err = showLoginTokenWriter(cxt, &output, &config)

	// Verify it didn't fail
	if err != nil {
		t.Error(err)
	}

	// Verify we printed the subject
	err = checkRegExp(`Subject:\s+`+expectedSubject, output)
	if err != nil {
		t.Error(err)
	}

	// Verify we printed the admin group
	err = checkRegExp(`Groups:.*`+expectedAdminGroup, output)
	if err != nil {
		t.Error(err)
	}

	// Verify we printed the everyone group
	err = checkRegExp(`Groups:.*`+expectedEveryoneGroup, output)
	if err != nil {
		t.Error(err)
	}

	// Verify we printed the date for Issued and Expires
	err = checkRegExp(`Issued:.*`+expectedDate, output)
	if err != nil {
		t.Error(err)
	}
	err = checkRegExp(`Expires:.*`+expectedDate, output)
	if err != nil {
		t.Error(err)
	}

	// Verify we printed the token itself
	err = checkRegExp(`Token:\s+`+accessToken, output)
	if err != nil {
		t.Error(err)
	}
}

/*
 * This does spot checking (not 100% thorough, just enough to verify the basics)
 * on the output of "show-login-token --raw"
 */
func TestShowLoginTokenRaw(t *testing.T) {
	config := configuration.Configuration{
		Token: accessToken,
	}

	globalFlags := flag.NewFlagSet("global-flags", flag.ContinueOnError)
	err := globalFlags.Parse([]string{})
	if err != nil {
		t.Error(err)
	}
	globalCtx := cli.NewContext(nil, globalFlags, nil)
	commandFlags := flag.NewFlagSet("command-flags", flag.ContinueOnError)
	commandFlags.Bool("raw", true, "raw")
	err = commandFlags.Parse([]string{"--raw"})
	if err != nil {
		t.Error(err)
	}
	cxt := cli.NewContext(nil, commandFlags, globalCtx)

	var output bytes.Buffer
	err = showLoginTokenWriter(cxt, &output, &config)

	// Verify it didn't fail
	if err != nil {
		t.Error(err)
	}

	// Verify we printed the subject
	err = checkRegExp(`"sub":.*`+expectedSubject, output)
	if err != nil {
		t.Error(err)
	}

	// Verify we printed the token class (which isn't printed when it's not raw)
	err = checkRegExp(`"token_class":\s+\"access_token\"`, output)
	if err != nil {
		t.Error(err)
	}

	// Verify we printed the token type (which isn't printed when it's not raw)
	err = checkRegExp(`"token_type":\s+\"Bearer\"`, output)
	if err != nil {
		t.Error(err)
	}
}

/*
 * This validates the output of "show-login-token --non-interactive"
 */
func TestShowLoginTokenNonInteractive(t *testing.T) {
	config := configuration.Configuration{
		Token: accessToken,
	}

	globalFlags := flag.NewFlagSet("global-flags", flag.ContinueOnError)
	globalFlags.Bool("non-interactive", true, "non-interactive")
	err := globalFlags.Parse([]string{"--non-interactive"})
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
	err = showLoginTokenWriter(cxt, &output, &config)

	// Verify it didn't fail
	if err != nil {
		t.Error(err)
	}

	// Verify we printed the token and only the token (well, with a newline)
	outputString := output.String()
	if outputString != accessToken+"\n" {
		t.Errorf("Expected just access token, found '%s'", outputString)
	}
}

func checkRegExp(pattern string, output bytes.Buffer) error {
	matched, err := regexp.MatchString(pattern, output.String())
	if !matched {
		return fmt.Errorf("Expected %s, but not found", pattern)
	}
	return err
}
