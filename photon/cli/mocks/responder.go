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
	"bytes"
	"io/ioutil"
	"net/http"
)

// Responders are callbacks that receive http request and return a mocked response.
type Responder func(*http.Request) (*http.Response, error)

// Create a responder with protocol version "HTTP/1.0" and a mocked response in JSON
func CreateResponder(status int, response string) Responder {
	return Responder(func(req *http.Request) (*http.Response, error) {
		resp := &http.Response{
			StatusCode:    status,
			ProtoMajor:    1,
			ProtoMinor:    0,
			Body:          ioutil.NopCloser(bytes.NewBufferString(response)),
			ContentLength: int64(len(response)),
			Request:       req,
		}

		resp.Header = make(map[string][]string)
		resp.Header.Add("Content-Type", "application/json")

		return resp, nil
	})
}

// Add a responder to `DefaultMockTransport`.
func RegisterResponder(method, url string, responder Responder) {
	DefaultMockTransport.RegisterResponder(method, url, responder)
}

// Add a new responder, associated with a given HTTP method and URL.
// When a request matches, the responder will be called and the response returned.
func (m *MockTransport) RegisterResponder(method, url string, responder Responder) {
	if m.responders == nil {
		m.responders = make(map[string]Responder)
	}
	m.responders[method+" "+url] = responder
}
