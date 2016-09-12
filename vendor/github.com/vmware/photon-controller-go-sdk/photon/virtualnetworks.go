package photon

import (
	"bytes"
	"encoding/json"
)

// Contains pointer to api client.
type VirtualSubnetsAPI struct {
	client *Client
}

// Options used for GetAll API
type VirtualSubnetGetOptions struct {
	Name string `urlParam:"name"`
}

var subnetsUrl = "/subnets"
var projectsUrl = "/projects"

// Create a virtual network
func (api *VirtualSubnetsAPI) Create(projectId string,
	virtualNetworkSpec *VirtualSubnetCreateSpec) (task *Task, err error) {

	body, err := json.Marshal(virtualNetworkSpec)
	if err != nil {
		return
	}

	res, err := api.client.restClient.Post(
		api.client.Endpoint+projectsUrl+"/"+projectId+"/subnets",
		"application/json",
		bytes.NewBuffer(body),
		api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}

	defer res.Body.Close()

	task, err = getTask(getError(res))
	return
}

// Delete a virtual network with the specified ID.
func (api *VirtualSubnetsAPI) Delete(id string) (task *Task, err error) {
	res, err := api.client.restClient.Delete(api.client.Endpoint+subnetsUrl+"/"+id,
		api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}

	defer res.Body.Close()

	task, err = getTask(getError(res))
	return
}

// Get the virtual subnet with the specified id
func (api *VirtualSubnetsAPI) Get(id string) (subnet *VirtualSubnet, err error) {
	res, err := api.client.restClient.Get(api.client.Endpoint+subnetsUrl+"/"+id,
		api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}

	defer res.Body.Close()

	res, err = getError(res)
	if err != nil {
		return
	}

	subnet = new(VirtualSubnet)
	err = json.NewDecoder(res.Body).Decode(subnet)
	return
}

// Return all virtual networks
func (api *VirtualSubnetsAPI) GetAll(projectId string,
	options *VirtualSubnetGetOptions) (subnets *VirtualSubnets, err error) {

	uri := api.client.Endpoint + projectsUrl + "/" + projectId + "/subnets"
	if options != nil {
		uri += getQueryString(options)
	}

	res, err := api.client.restClient.GetList(api.client.Endpoint, uri, api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}

	subnets = &VirtualSubnets{}
	err = json.Unmarshal(res, subnets)

	return
}

func (api *VirtualSubnetsAPI) SetDefault(id string) (task *Task, err error) {
	res, err := api.client.restClient.Post(
		api.client.Endpoint+subnetsUrl+"/"+id+"/set_default",
		"application/json",
		bytes.NewBuffer([]byte("")),
		api.client.options.TokenOptions.AccessToken)
	if err != nil {
		return
	}

	defer res.Body.Close()

	task, err = getTask(getError(res))
	return
}
