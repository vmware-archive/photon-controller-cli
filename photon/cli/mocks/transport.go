package mocks

import (
	"errors"
	"net/http"
)

// MockTransport implements http.RoundTripper.
// The implementation doesn't make the call, instead defering to the registered list of responders.
// Return Error if no responder is found when FailNoResponder = true
type MockTransport struct {
	FailNoResponder bool
	responders      map[string]Responder
}

// The global default RoundTripper for all http requests.
var DefaultMockTransport = &MockTransport{}

// RoundTrip is required to implement http.MockTransport.
func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	key := req.Method + " " + req.URL.String()

	// scan through the responders and find one that matches our key
	for k, r := range m.responders {
		if k != key {
			continue
		}
		return r(req)
	}

	if m.FailNoResponder {
		return nil, errors.New("no responder found")
	}

	// fallback to the default roundtripper
	return http.DefaultTransport.RoundTrip(req)
}

// Activate replaces the `Transport` on the `http.DefaultClient` with our `DefaultMockTransport`.
func Activate(failNoResponder bool) {
	DefaultMockTransport.FailNoResponder = failNoResponder
	http.DefaultClient.Transport = DefaultMockTransport
}

// Deactivate replaces our `DefaultMockTransport` with the `http.DefaultTransport`.
func Deactivate() {
	http.DefaultClient.Transport = http.DefaultTransport
}
