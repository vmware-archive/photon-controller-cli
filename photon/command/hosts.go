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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"syscall"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"

	"golang.org/x/crypto/ssh/terminal"
)

// Creates a cli.Command for host
// Subcommands: create;                Usage: host create [<options>]
//              delete;                Usage: host delete <id>
//              show;                  Usage: host show <id>
//              list;                  Usage: host list
//              list-vms;              Usage: host list-vms <id>
//              set-availability-zone; Usage: host set-availability-zone <id> <availability-zone-id>
//              tasks;                 Usage: host tasks <id> [<options>]
//              provision;             Usage: host provision <id>
//              suspend;               Usage: host suspend <id>
//              resume;                Usage: host resume <id>
//              enter-maintenance;     Usage: host enter-maintenance <id>
//              exit-maintenance;      Usage: host exit-maintenance <id>
func GetHostsCommand() cli.Command {
	command := cli.Command{
		Name:  "host",
		Usage: "options for host",
		Subcommands: []cli.Command{
			{
				Name:        "create",
				Usage:       "Add a new host",
				ArgsUsage:   " ",
				Description: "Add a new host to Photon Controller. You must a system administrator to add a host.",
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
						Hidden: true,
						Name:   "deployment_id, d",
						Usage:  "deployment id to create host",
					},
				},
				Action: func(c *cli.Context) {
					err := createHost(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "delete",
				Usage:     "Delete a host with specified id",
				ArgsUsage: "<id>",
				Description: "Removes a host from management by Photon Controller.\n" +
					"   You must be a system administrator to do this.",
				Action: func(c *cli.Context) {
					err := deleteHost(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "list",
				Usage:     "List all the hosts managed by Photon Controller",
				ArgsUsage: " ",
				Action: func(c *cli.Context) {
					err := listHosts(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "show",
				Usage: "Show host info with specified id",
				Action: func(c *cli.Context) {
					err := showHost(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "list-vms",
				Usage:     "List all the VMs on a given host",
				ArgsUsage: "<host-id>",
				Action: func(c *cli.Context) {
					err := listHostVMs(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:        "set-availability-zone",
				Usage:       "Set a host's availability zone",
				ArgsUsage:   "<host-id> <availability-zone-id>",
				Description: "Set a host's availability zone. You must be a system administrator to do this.",
				Action: func(c *cli.Context) {
					err := setHostAvailabilityZone(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "tasks",
				Usage:     "Show host tasks",
				ArgsUsage: "<host-id>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "state, s",
						Usage: "Filter by task sate",
					},
				},
				Action: func(c *cli.Context) {
					err := getHostTasks(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "provision",
				Usage:     "Provision host with given id",
				ArgsUsage: "<host-id>",
				Description: "Provision a host given its id. You must be a system administrator to do this.\n" +
					"   This will configure photon controller agent and make the host ready.",
				Action: func(c *cli.Context) {
					err := provisionHost(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "suspend",
				Usage:     "Suspend host with given id",
				ArgsUsage: "<host-id>",
				Description: "Suspend a host given its id. You must be a system administrator to do this.\n" +
					"   This is a precursor to entering maintenance mode. No new VMs will be placed on the host\n" +
					"   while it is suspended.",
				Action: func(c *cli.Context) {
					err := suspendHost(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "resume",
				Usage:     "Resume host with specified id",
				ArgsUsage: "<host-id>",
				Description: "Resume a host given its id. You must be a system administrator to do this.\n" +
					"   This will return a host to normal service and new VMs can be placed on this host again.",
				Action: func(c *cli.Context) {
					err := resumeHost(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "enter-maintenance",
				Usage:     "Host with specified id enter maintenance mode",
				ArgsUsage: "<host-id>",
				Description: "Put a host into maintenance mode. You must be a system administrator to do this.\n" +
					"   A host must be suspended and have no VMs placed on it in order to enter maintenance mode.",
				Action: func(c *cli.Context) {
					err := enterMaintenanceMode(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "exit-maintenance",
				Usage:     "Host with specified id exit maintenance mode",
				ArgsUsage: "<host-id>",
				Description: "Resume a host that was in maintenance mode given its id.\n" +
					"   You must be a system administrator to do this.\n" +
					"   This will return a host to normal service and new VMs can be placed on this host again.",
				Action: func(c *cli.Context) {
					err := exitMaintenanceMode(c, os.Stdout)
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
func createHost(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
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

	deploymentID, err = getDeploymentId(c)
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") {
		var err error
		username, err = askForInput("Username: ", username)
		if err != nil {
			return err
		}
		if len(password) == 0 {
			fmt.Printf("Password: ")
			// Casting syscall.Stdin to int because during
			// Windows cross-compilation syscall.Stdin is incorrectly
			// treated as a String.
			bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return err
			}
			password = string(bytePassword)
			fmt.Printf("\n")
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
	}

	hostSpec := photon.HostCreateSpec{}
	hostSpec.Username = username
	hostSpec.Password = password
	hostSpec.Address = address
	hostSpec.Zone = availabilityZone
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

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}
	createTask, err := client.Photonclient.Hosts.Create(&hostSpec, deploymentID)
	if err != nil {
		return err
	}
	id, err := waitOnTaskOperation(createTask.ID, c)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		host, err := client.Photonclient.Hosts.Get(id)
		if err != nil {
			return err
		}
		utils.FormatObject(host, w, c)
	}

	return nil
}

// Sends a delete host task to client based on the cli.Context
// Returns an error if one occurred
func deleteHost(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deleteTask, err := client.Photonclient.Hosts.Delete(id)
	if err != nil {
		return err
	}
	_, err = waitOnTaskOperation(deleteTask.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// List all the hosts in the current deployment
// This uses the same back-end code as "deployments list-hosts", but we look up the
// deployment ID so that users don't have to specify it. In most or all installations,
// there will not be more than one deployment ID.
func listHosts(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	// Find the current deployment
	deployments, err := client.Photonclient.Deployments.GetAll()
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

	hosts, err := client.Photonclient.Deployments.GetHosts(id)
	if err != nil {
		return err
	}

	err = printHostList(hosts.Items, w, c)
	if err != nil {
		return err
	}
	return nil
}

// Show host info with the specified host ID, returns an error if one occurred
func showHost(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	host, err := client.Photonclient.Hosts.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		tag := strings.Trim(fmt.Sprint(host.Tags), "[]")
		scriptTag := strings.Replace(tag, " ", ",", -1)
		metadata := strings.Trim(strings.TrimLeft(fmt.Sprint(host.Metadata), "map"), "[]")
		scriptMetadata := strings.Replace(metadata, " ", ",", -1)
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", host.ID, host.Username, host.Address,
			scriptTag, host.State, scriptMetadata, host.Zone, host.EsxVersion)
	} else if utils.NeedsFormatting(c) {
		host.Password = ""
		utils.FormatObject(host, w, c)
	} else {
		fmt.Println("Host ID: ", host.ID)
		fmt.Println("  Username:          ", host.Username)
		fmt.Println("  IP:                ", host.Address)
		fmt.Println("  Tags:              ", host.Tags)
		fmt.Println("  State:             ", host.State)
		fmt.Println("  Metadata:          ", host.Metadata)
		fmt.Println("  AvailabilityZone:  ", host.Zone)
		fmt.Println("  Version:           ", host.EsxVersion)
	}

	return nil
}

// Set host's availability zone with the specified host ID, returns an error if one occurred
func setHostAvailabilityZone(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 2)
	if err != nil {
		return err
	}
	id := c.Args().First()
	availabilityZoneId := c.Args()[1]

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	setAvailabilityZoneSpec := photon.HostSetAvailabilityZoneOperation{}
	setAvailabilityZoneSpec.AvailabilityZoneId = availabilityZoneId
	setTask, err := client.Photonclient.Hosts.SetAvailabilityZone(id, &setAvailabilityZoneSpec)
	if err != nil {
		return err
	}
	id, err = waitOnTaskOperation(setTask.ID, c)
	if err != nil {
		return err
	}
	if utils.NeedsFormatting(c) {
		host, err := client.Photonclient.Hosts.Get(id)
		if err != nil {
			return err
		}
		utils.FormatObject(host, w, c)
	}

	return nil
}

func getHostTasks(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	state := c.String("state")
	options := &photon.TaskGetOptions{
		State: state,
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	taskList, err := client.Photonclient.Hosts.GetTasks(id, options)
	if err != nil {
		return err
	}

	err = printTaskList(taskList.Items, c)
	if err != nil {
		return err
	}
	return nil
}

func listHostVMs(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	vmList, err := client.Photonclient.Hosts.GetVMs(id)
	if err != nil {
		return err
	}

	err = printVMList(vmList.Items, w, c, false)
	if err != nil {
		return err
	}

	return nil
}

// Sends a provision host task to client based on the cli.Context
// Returns an error if one occurred
func provisionHost(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	resumeTask, err := client.Photonclient.Hosts.Provision(id)
	if err != nil {
		return err
	}
	_, err = waitOnTaskOperation(resumeTask.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Sends a suspend host task to client based on the cli.Context
// Returns an error if one occurred
func suspendHost(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	suspendTask, err := client.Photonclient.Hosts.Suspend(id)
	if err != nil {
		return err
	}
	_, err = waitOnTaskOperation(suspendTask.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Sends a resume host task to client based on the cli.Context
// Returns an error if one occurred
func resumeHost(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	resumeTask, err := client.Photonclient.Hosts.Resume(id)
	if err != nil {
		return err
	}
	_, err = waitOnTaskOperation(resumeTask.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Put host with specified id into maintenance mode
// Returns an error if one occurred
func enterMaintenanceMode(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	enterTask, err := client.Photonclient.Hosts.EnterMaintenanceMode(id)
	if err != nil {
		return err
	}
	_, err = waitOnTaskOperation(enterTask.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Take host with specified id out of maintenance mode
// Returns an error if one occurred
func exitMaintenanceMode(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	exitTask, err := client.Photonclient.Hosts.ExitMaintenanceMode(id)
	if err != nil {
		return err
	}
	_, err = waitOnTaskOperation(exitTask.ID, c)
	if err != nil {
		return err
	}

	return nil
}
