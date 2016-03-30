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
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/vmware/photon-controller-cli/photon/client"

	"encoding/json"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for host
// Subcommands: create;                Usage: host create [<options>]
//              delete;                Usage: host delete <id>
//              show;                  Usage: host show <id>
//              list;                  Usage: host list
//              list-vms;              Usage: host list-vms <id>
//              set-availability-zone; Usage: host set-availability-zone <id> <availability-zone-id>
//              tasks;                 Usage: host tasks <id> [<options>]
func GetHostsCommand() cli.Command {
	command := cli.Command{
		Name:  "host",
		Usage: "options for host",
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create a new host",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "username, u",
						Usage: "username to create host",
					},
					cli.StringFlag{
						Name:  "password, p",
						Usage: "password to create host",
					},
					cli.StringFlag{
						Name:  "address, i",
						Usage: "ip address of the host",
					},
					cli.StringFlag{
						Name:  "availability_zone, z",
						Usage: "availability zone of the host",
					},
					cli.StringFlag{
						Name:  "tag, t",
						Usage: "tag for the host",
					},
					cli.StringFlag{
						Name:  "metadata, m",
						Usage: "metadata for the host",
					},
					cli.StringFlag{
						Name:  "deployment_id, d",
						Usage: "deployment id to create host",
					},
				},
				Action: func(c *cli.Context) {
					err := createHost(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "delete",
				Usage: "Delete a host with specified id",
				Action: func(c *cli.Context) {
					err := deleteHost(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "list",
				Usage: "List all the hosts",
				Action: func(c *cli.Context) {
					err := listHosts(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "show",
				Usage: "Show host info with specified id",
				Action: func(c *cli.Context) {
					err := showHost(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "list-vms",
				Usage: "List all the vms on the host",
				Action: func(c *cli.Context) {
					err := listHostVMs(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "set-availability-zone",
				Usage: "Set host's availability zone",
				Action: func(c *cli.Context) {
					err := setHostAvailabilityZone(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "tasks",
				Usage: "Show tenant tasks",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "state, s",
						Usage: "Filter by task sate",
					},
				},
				Action: func(c *cli.Context) {
					err := getHostTasks(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Sends a create host task to client based on the cli.Context
// Returns an error if one occurred
func createHost(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "host create [<options>]")
	if err != nil {
		return err
	}
	username := c.String("username")
	password := c.String("password")
	address := c.String("address")
	availabilityZone := c.String("availability_zone")
	tags := c.String("tag")
	metadata := c.String("metadata")
	deploymentID := c.String("deployment_id")

	if !c.GlobalIsSet("non-interactive") {
		var err error
		username, err = askForInput("Username: ", username)
		if err != nil {
			return err
		}
		password, err = askForInput("Password: ", password)
		if err != nil {
			return err
		}
		address, err = askForInput("Host Address: ", address)
		if err != nil {
			return err
		}
		tags, err = askForInput("Host Tags (Options [CLOUD,MGMT]. ',' separated): ", tags)
		if err != nil {
			return err
		}
		metadata, err = askForInput("Host Metadata ({'key':'value'}. required by host of 'MGMT' tag): ", metadata)
		if err != nil {
			return err
		}
		username, err = askForInput("Deployment ID: ", deploymentID)
		if err != nil {
			return err
		}
	}

	hostSpec := photon.HostCreateSpec{}
	hostSpec.Username = username
	hostSpec.Password = password
	hostSpec.Address = address
	hostSpec.AvailabilityZone = availabilityZone
	hostSpec.Tags = regexp.MustCompile(`\s*,\s*`).Split(tags, -1)

	if len(metadata) == 0 {
		hostSpec.Metadata = map[string]string{}
	} else {
		var data map[string]string
		err := json.Unmarshal([]byte(metadata), &data)
		if err != nil {
			return err
		}
		hostSpec.Metadata = data
	}

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}
	createTask, err := client.Esxclient.Hosts.Create(&hostSpec, deploymentID)
	if err != nil {
		return err
	}
	err = waitOnTaskOperation(createTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Sends a delete host task to client based on the cli.Context
// Returns an error if one occurred
func deleteHost(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "host delete <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	deleteTask, err := client.Esxclient.Hosts.Delete(id)
	if err != nil {
		return err
	}
	err = waitOnTaskOperation(deleteTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// List all the hosts in the current deployment
// This uses the same back-end code as "deployments list-hosts", but we look up the
// deployment ID so that users don't have to specify it. In most or all installations,
// there will not be more than one deployment ID.
func listHosts(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "host list")
	if err != nil {
		return err
	}

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	// Find the current deployment
	deployments, err := client.Esxclient.Deployments.GetAll()
	if err != nil {
		return err
	}
	numDeployments := len(deployments.Items)
	if numDeployments == 0 {
		return fmt.Errorf("There are no deployments, so the hosts cannot be listed.")
	} else if numDeployments > 1 {
		return fmt.Errorf("There are multiple deployments, which normally should not happen. Use deployments list-hosts.")
	}
	id := deployments.Items[0].ID

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	hosts, err := client.Esxclient.Deployments.GetHosts(id)
	if err != nil {
		return err
	}

	err = printHostList(hosts.Items, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}
	return nil
}

// Show host info with the specified host ID, returns an error if one occurred
func showHost(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "host show <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	host, err := client.Esxclient.Hosts.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		tag := strings.Trim(fmt.Sprint(host.Tags), "[]")
		scriptTag := strings.Replace(tag, " ", ",", -1)
		metadata := strings.Trim(strings.TrimLeft(fmt.Sprint(host.Metadata), "map"), "[]")
		scriptMetadata := strings.Replace(metadata, " ", ",", -1)
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", host.ID, host.Username, host.Password, host.Address,
			scriptTag, host.State, scriptMetadata, host.AvailabilityZone, host.EsxVersion)
	} else {
		fmt.Println("Host ID: ", host.ID)
		fmt.Println("  Username:          ", host.Username)
		fmt.Println("  Password:          ", host.Password)
		fmt.Println("  IP:                ", host.Address)
		fmt.Println("  Tags:              ", host.Tags)
		fmt.Println("  State:             ", host.State)
		fmt.Println("  Metadata:          ", host.Metadata)
		fmt.Println("  AvailabilityZone:  ", host.AvailabilityZone)
		fmt.Println("  Version:           ", host.EsxVersion)
	}

	return nil
}

// Set host's availability zone with the specified host ID, returns an error if one occurred
func setHostAvailabilityZone(c *cli.Context) error {
	err := checkArgNum(c.Args(), 2, "host set-availability-zone <id> <availability-zone-id>")
	if err != nil {
		return err
	}
	id := c.Args().First()
	availabilityZoneId := c.Args()[1]

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	setAvailabilityZoneSpec := photon.HostSetAvailabilityZoneOperation{}
	setAvailabilityZoneSpec.AvailabilityZoneId = availabilityZoneId
	setTask, err := client.Esxclient.Hosts.SetAvailabilityZone(id, &setAvailabilityZoneSpec)
	if err != nil {
		return err
	}
	err = waitOnTaskOperation(setTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

func getHostTasks(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "host tasks <id> [<options>]")
	if err != nil {
		return err
	}
	id := c.Args().First()

	state := c.String("state")
	options := &photon.TaskGetOptions{
		State: state,
	}

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	taskList, err := client.Esxclient.Hosts.GetTasks(id, options)
	if err != nil {
		return err
	}

	err = printTaskList(taskList.Items, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}
	return nil
}

func listHostVMs(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "host list-vms <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	vmList, err := client.Esxclient.Hosts.GetVMs(id)
	if err != nil {
		return err
	}

	err = printVMList(vmList.Items, c.GlobalIsSet("non-interactive"), false)
	if err != nil {
		return err
	}

	return nil
}
