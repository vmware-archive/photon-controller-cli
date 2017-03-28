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
	"errors"
	"log"
	"os"

	"github.com/vmware/photon-controller-cli/photon/client"

	"github.com/urfave/cli"
)

const (
	PHYSICAL         = "PHYSICAL"
	SOFTWARE_DEFINED = "SOFTWARE_DEFINED"
	NOT_AVAILABLE    = "NOT_AVAILABLE"
)

// Creates a cli.Command for networks
// Subcommands: create; Usage: network create [<options>]
//              list;   Usage: network list
func GetNetworksCommand() cli.Command {
	command := cli.Command{
		Name:  "network",
		Usage: "options for network",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new network",
				ArgsUsage: " ",
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
					err := createPhysicalNetwork(c, os.Stdout)
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
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return
	}

	info, err := client.Photonclient.Info.Get()
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
