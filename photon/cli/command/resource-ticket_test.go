package command

import (
	"encoding/json"
	"flag"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmware/photon-controller-cli/photon/cli/client"
	"github.com/vmware/photon-controller-cli/photon/cli/mocks"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
)

var tenantName string
var tenantID string
var rtName string
var rtID string
var server *httptest.Server

type MockResourceTicketsPage struct {
	Items            []photon.ResourceTicket `json:"items"`
	NextPageLink     string                  `json:"nextPageLink"`
	PreviousPageLink string                  `json:"previousPageLink"`
}

func TestCreateResourceTicket(t *testing.T) {
	tenantName = "fake_tenant_Name"
	tenantID = "fake_tenant_ID"
	rtName = "fake_rt_Name"
	rtID = "fake_rt_ID"

	tenantStruct := photon.Tenants{
		Items: []photon.Tenant{
			photon.Tenant{
				Name: tenantName,
				ID:   tenantID,
			},
		},
	}
	tenantResponse, err := json.Marshal(tenantStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected tenantStruct")
	}

	queuedTask := &photon.Task{
		Operation: "CREATE_RESOURCE_TICKET",
		State:     "QUEUED",
		ID:        "fake-rt-task-id",
		Entity:    photon.Entity{ID: rtID},
	}
	taskResponse, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}

	completedTask := &photon.Task{
		Operation: "CREATE_RESOURCE_TICKET",
		State:     "COMPLETED",
		ID:        "fake-rt-task-id",
		Entity:    photon.Entity{ID: rtID},
	}
	response, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected completedTask")
	}

	server = mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants",
		mocks.CreateResponder(200, string(tenantResponse[:])))
	mocks.RegisterResponder(
		"POST",
		server.URL+"/tenants/"+tenantID+"/resource-tickets",
		mocks.CreateResponder(200, string(taskResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(response[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	globalSet := flag.NewFlagSet("test", 0)
	globalSet.Bool("non-interactive", true, "doc")
	globalCtx := cli.NewContext(nil, globalSet, nil)
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	set := flag.NewFlagSet("test", 0)
	set.String("name", "name", "rt name")
	set.String("tenant", tenantName, "tenant name")
	set.String("limits", "vm.test1 1 B, vm.test2 1 GB", "rt limits")
	cxt := cli.NewContext(nil, set, globalCtx)

	err = createResourceTicket(cxt)
	if err != nil {
		t.Error("Not expecting error creating resource ticket: " + err.Error())
	}
}

func TestShowResourceTicket(t *testing.T) {
	rtListStruct := photon.ResourceList{
		Items: []photon.ResourceTicket{
			photon.ResourceTicket{
				Name:   rtName,
				ID:     rtID,
				Limits: []photon.QuotaLineItem{photon.QuotaLineItem{Key: "k", Value: 1, Unit: "B"}},
				Usage:  []photon.QuotaLineItem{photon.QuotaLineItem{Key: "k", Value: 0, Unit: "B"}},
			},
		},
	}
	listResponse, err := json.Marshal(rtListStruct)
	if err != nil {
		t.Error("Not expecting error serializaing expected rtLists")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants/"+tenantID+"/resource-tickets",
		mocks.CreateResponder(200, string(listResponse[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants/"+tenantID+"/resource-tickets?name="+rtName,
		mocks.CreateResponder(200, string(listResponse[:])))

	set := flag.NewFlagSet("test", 0)
	set.String("tenant", tenantName, "tenant name")
	err = set.Parse([]string{rtName})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = showResourceTicket(cxt)
	if err != nil {
		t.Error("Not expecting error showing resource ticket: " + err.Error())
	}
}

func TestListResourceTickets(t *testing.T) {
	rtList := MockResourceTicketsPage{
		Items: []photon.ResourceTicket{
			photon.ResourceTicket{
				Name:   rtName,
				ID:     rtID,
				Limits: []photon.QuotaLineItem{photon.QuotaLineItem{Key: "k", Value: 1, Unit: "B"}},
				Usage:  []photon.QuotaLineItem{photon.QuotaLineItem{Key: "k", Value: 0, Unit: "B"}},
			},
		},
		NextPageLink:     "fake-next-page-link",
		PreviousPageLink: "",
	}
	listResponse, err := json.Marshal(rtList)
	if err != nil {
		t.Error("Not expecting error serializaing expected rtLists")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/tenants/"+tenantID+"/resource-tickets",
		mocks.CreateResponder(200, string(listResponse[:])))

	rtList = MockResourceTicketsPage{
		Items:            []photon.ResourceTicket{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}
	listResponse, err = json.Marshal(rtList)
	if err != nil {
		t.Error("Not expecting error serializaing expected rtLists")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"fake-next-page-link",
		mocks.CreateResponder(200, string(listResponse[:])))

	set := flag.NewFlagSet("test", 0)
	set.String("tenant", tenantName, "tenant name")
	cxt := cli.NewContext(nil, set, nil)

	err = listResourceTickets(cxt)
	if err != nil {
		t.Error("Not expecting error listing resource ticket: " + err.Error())
	}
}

func TestListResourceTicketTasks(t *testing.T) {
	taskList := MockTasksPage{
		Items: []photon.Task{
			photon.Task{
				Operation: "CREATE_RESOURCE_TICKET",
				State:     "COMPLETED",
				ID:        "fake-rt-task-id",
				Entity:    photon.Entity{ID: rtID},
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}
	response, err := json.Marshal(taskList)
	if err != nil {
		t.Error("Not expecting error serializaing expected taskLists")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"/resource-tickets/"+rtID+"/tasks",
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

	set := flag.NewFlagSet("test", 0)
	set.String("tenant", tenantName, "tenant name")
	err = set.Parse([]string{rtName})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = getResourceTicketTasks(cxt)
	if err != nil {
		t.Error("Not expecting error showing resource ticket tasks: " + err.Error())
	}
}
