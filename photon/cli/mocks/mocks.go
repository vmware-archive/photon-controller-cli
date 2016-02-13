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
