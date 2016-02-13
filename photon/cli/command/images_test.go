package command

import (
	"encoding/json"
	"flag"
	"net/http"
	"testing"

	"github.com/vmware/photon-controller-cli/photon/cli/client"
	"github.com/vmware/photon-controller-cli/photon/cli/mocks"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
)

type MockImagesPage struct {
	Items            []photon.Image `json:"items"`
	NextPageLink     string         `json:"nextPageLink"`
	PreviousPageLink string         `json:"previousPageLink"`
}

func TestCreateDeleteImage(t *testing.T) {
	queuedTask := &photon.Task{
		Operation: "CREATE_IMAGE",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask := &photon.Task{
		Operation: "CREATE_IMAGE",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "1"},
	}

	response, err := json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}
	taskresponse, err := json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected createTask")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"POST",
		server.URL+"/images",
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))
	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	set.String("name", "n", "testname")
	set.String("image_replication", "EAGER", "image replication")
	err = set.Parse([]string{"../../../testdata/tty_tiny.ova"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)

	err = createImage(cxt)
	if err != nil {
		t.Error("Not expecting error creating image: " + err.Error())
	}

	queuedTask = &photon.Task{
		Operation: "DELETE_IMAGE",
		State:     "QUEUED",
		Entity:    photon.Entity{ID: "1"},
	}
	completedTask = &photon.Task{
		Operation: "DELETE_IMAGE",
		State:     "COMPLETED",
		Entity:    photon.Entity{ID: "1"},
	}

	response, err = json.Marshal(queuedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected queuedTask")
	}
	taskresponse, err = json.Marshal(completedTask)
	if err != nil {
		t.Error("Not expecting error serializaing expected deleteTask")
	}

	mocks.RegisterResponder(
		"DELETE",
		server.URL+"/images/"+queuedTask.Entity.ID,
		mocks.CreateResponder(200, string(response[:])))
	mocks.RegisterResponder(
		"GET",
		server.URL+"/tasks/"+queuedTask.ID,
		mocks.CreateResponder(200, string(taskresponse[:])))

	globalSet := flag.NewFlagSet("global", 0)
	globalSet.Bool("non-interactive", true, "doc")
	err = globalSet.Parse([]string{"--non-interactive"})
	if err != nil {
		t.Error("Not expectin arguments parsing to fail")
	}
	globalCtx := cli.NewContext(nil, globalSet, nil)
	set = flag.NewFlagSet("test", 0)
	err = set.Parse([]string{queuedTask.Entity.ID})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}

	cxt = cli.NewContext(nil, set, globalCtx)
	err = deleteImage(cxt)
	if err != nil {
		t.Error("Not expecting error deleting image: " + err.Error())
	}
}

func TestListImage(t *testing.T) {
	expectedImageList := MockImagesPage{
		Items: []photon.Image{
			photon.Image{
				Name:            "test",
				Size:            10,
				State:           "COMPLETED",
				ID:              "1",
				ReplicationType: "EAGER",
				Settings: []photon.ImageSetting{
					photon.ImageSetting{
						Name:         "test-setting",
						DefaultValue: "test-default-value",
					},
				},
			},
		},
		NextPageLink:     "fake-next-page-link",
		PreviousPageLink: "",
	}

	response, err := json.Marshal(expectedImageList)
	if err != nil {
		t.Error("Not expecting error serializaing expected response")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/images",
		mocks.CreateResponder(200, string(response[:])))

	expectedImageList = MockImagesPage{
		Items:            []photon.Image{},
		NextPageLink:     "",
		PreviousPageLink: "",
	}

	response, err = json.Marshal(expectedImageList)
	if err != nil {
		t.Error("Not expecting error serializaing expected response")
	}

	mocks.RegisterResponder(
		"GET",
		server.URL+"fake-next-page-link",
		mocks.CreateResponder(200, string(response[:])))

	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = listImages(cxt)
	if err != nil {
		t.Error("Not expecting an error listing images ", err)
	}
}

func TestImageTasks(t *testing.T) {
	taskList := MockTasksPage{
		Items: []photon.Task{
			photon.Task{
				Operation: "CREATE_IMAGE",
				State:     "COMPLETED",
				ID:        "1",
				Entity:    photon.Entity{ID: "1", Kind: "image"},
			},
		},
		NextPageLink:     "/fake-next-page-link",
		PreviousPageLink: "",
	}

	response, err := json.Marshal(taskList)
	if err != nil {
		t.Error("Not expecting error serializaing expected status")
	}

	server := mocks.NewTestServer()
	mocks.RegisterResponder(
		"GET",
		server.URL+"/images/1/tasks",
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

	defer server.Close()

	mocks.Activate(true)
	httpClient := &http.Client{Transport: mocks.DefaultMockTransport}
	client.Esxclient = photon.NewTestClient(server.URL, "", nil, httpClient)

	set := flag.NewFlagSet("test", 0)
	err = set.Parse([]string{"1"})
	if err != nil {
		t.Error("Not expecting arguments parsing to fail")
	}
	cxt := cli.NewContext(nil, set, nil)
	err = getImageTasks(cxt)
	if err != nil {
		t.Error("Not expecting error retrieving tenant tasks")
	}
}
