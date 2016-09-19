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
	"io"
	"log"
	"os"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"errors"
	"github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

const (
	PHYSICAL         = "PHYSICAL"
	SOFTWARE_DEFINED = "SOFTWARE_DEFINED"
	NOT_AVAILABLE    = "NOT_AVAILABLE"
)

// Creates a cli.Command for networks
// Subcommands: create; Usage: network create [<options>]
//              delete; Usage: network delete <id>
//              list;   Usage: network list
//              show;   Usage: network show <id>
//              set-default; Usage: network setDefault <id>
func GetNetworksCommand() cli.Command {
	command := cli.Command{
		Name:  "network",
		Usage: "options for network",
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create a new network",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Network name",
					},
					cli.StringFlag{
						Name:  "description, d",
						Usage: "Description of network",
					},
					cli.StringFlag{
						Name:  "portgroups, p",
						Usage: "PortGroups associated with network (only for physical network)",
					},
					cli.StringFlag{
						Name: "routingType, r",
						Usage: "Routing type for network (only for software-defined network). Supported values are: " +
							"'ROUTED' and 'ISOLATED'",
					},
					cli.StringFlag{
						Name:  "size, s",
						Usage: "Size of the private IP addresses (only for software-defined network)",
					},
					cli.StringFlag{
						Name:  "staticIpSize, f",
						Usage: "Size of the reserved static IP addresses (only for software-defined network)",
					},
					cli.StringFlag{
						Name:  "projectId, i",
						Usage: "ID of the project that network belongs to (only for software-defined network)",
					},
				},
				Action: func(c *cli.Context) {
					sdnEnabled, err := isSoftwareDefinedNetwork(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}

					if sdnEnabled {
						err = createVirtualNetwork(c, os.Stdout)
					} else {
						err = createPhysicalNetwork(c, os.Stdout)
					}
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "delete",
				Usage: "Delete a network",
				Action: func(c *cli.Context) {
					err := deleteNetwork(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "list",
				Usage: "List networks",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Optionally filter by name",
					},
					cli.StringFlag{
						Name: "projectId, i",
						Usage: "ID of the project that networks to be listed belong to (only for software-defined " +
							"network)",
					},
				},
				Action: func(c *cli.Context) {
					sdnEnabled, err := isSoftwareDefinedNetwork(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}

					if sdnEnabled {
						err = listVirtualNetworks(c, os.Stdout)
					} else {
						err = listPhysicalNetworks(c, os.Stdout)
					}
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "show",
				Usage: "Show specified network",
				Action: func(c *cli.Context) {
					sdnEnabled, err := isSoftwareDefinedNetwork(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}

					if sdnEnabled {
						err = showVirtualNetwork(c, os.Stdout)
					} else {
						err = showPhysicalNetwork(c, os.Stdout)
					}
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "set-default",
				Usage: "Set default network",
				Action: func(c *cli.Context) {
					err := setDefaultNetwork(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

func isSoftwareDefinedNetwork(c *cli.Context) (sdnEnabled bool, err error) {
	client.Esxclient, err = client.GetClient(utils.IsNonInteractive(c))
	if err != nil {
		return
	}

	info, err := client.Esxclient.Info.Get()
	if err != nil {
		return
	}

	if info.NetworkType == NOT_AVAILABLE {
		err = errors.New("Network type is missing")
	} else {
		sdnEnabled = (info.NetworkType == SOFTWARE_DEFINED)
	}
	return
}

func deleteNetwork(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "network delete <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(utils.IsNonInteractive(c))
	if err != nil {
		return err
	}

	sdnEnabled, err := isSoftwareDefinedNetwork(c)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	var task *photon.Task
	if sdnEnabled {
		task, err = client.Esxclient.VirtualSubnets.Delete(id)
	} else {
		task, err = client.Esxclient.Subnets.Delete(id)
	}

	if err != nil {
		return err
	}

	if confirmed(c.GlobalIsSet("non-interactive")) {
		_, err = waitOnTaskOperation(task.ID, c)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("OK. Canceled")
	}
	return nil
}

func setDefaultNetwork(c *cli.Context, w io.Writer) error {
	err := checkArgNum(c.Args(), 1, "network set-default <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(utils.IsNonInteractive(c))
	if err != nil {
		return err
	}

	sdnEnabled, err := isSoftwareDefinedNetwork(c)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	var task *photon.Task
	if sdnEnabled {
		task, err = client.Esxclient.VirtualSubnets.SetDefault(id)
	} else {
		task, err = client.Esxclient.Subnets.SetDefault(id)
	}

	if err != nil {
		return err
	}

	if confirmed(utils.IsNonInteractive(c)) {
		id, err := waitOnTaskOperation(task.ID, c)
		if err != nil {
			return err
		}

		if utils.NeedsFormatting(c) {
			network, err := client.Esxclient.Subnets.Get(id)
			if err != nil {
				return err
			}
			utils.FormatObject(network, w, c)
		}
	} else {
		fmt.Println("OK. Canceled")
	}
	return nil
}
