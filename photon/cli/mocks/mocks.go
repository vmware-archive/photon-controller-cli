// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package mocks

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

// Start a new Server with status OK, caller should call Close when finished
func NewTestServer() (server *httptest.Server) {
	statusCode := 200
	body := ""
	server = httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(statusCode)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, body)
		}))
	return
}
