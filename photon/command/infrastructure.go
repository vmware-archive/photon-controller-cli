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
	"regexp"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for infrastructure
// Subcommands:
//              list-hosts; Usage: infrastructure list-hosts [<id>]
//              update-image-datastores;        Usage: infrastructure update-image-datastores [<options>]
//              sync-hosts-config;              Usage: infrastructure sync-hosts-config

func GetInfrastructureCommand() cli.Command {
	command := cli.Command{
		Name:  "infrastructure",
		Usage: "options for infrastructure",
		Subcommands: []cli.Command{
			{
				Name:      "list-hosts",
				Usage:     "Lists all ESXi hosts",
				ArgsUsage: " ",
				Description: "List information about all ESXi hosts used in the infrastructure.\n" +
					"   For each host, the ID, the current state, the IP, and the type (MGMT and/or CLOUD)\n" +
					"   Requires system administrator access.",
				Action: func(c *cli.Context) {
					err := listInfrastructureHosts(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "update-image-datastores",
				Usage:     "Updates the list of image datastores",
				ArgsUsage: " ",
				Description: "Update the list of allowed image datastores.\n" +
					"   Requires system administrator access.",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "datastores, d",
						Usage: "Comma separated name of datastore names",
					},
				},
				Action: func(c *cli.Context) {
					err := updateInfrastructureImageDatastores(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "sync-hosts-config",
				Usage:     "Synchronizes hosts configurations",
				ArgsUsage: " ",
				Action: func(c *cli.Context) {
					err := syncInfrastructureHostsConfig(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Lists all the hosts associated with the deployment
func listInfrastructureHosts(c *cli.Context, w io.Writer) error {

	var err error
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	hosts, err := client.Photonclient.InfraHosts.GetHosts()
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		utils.FormatObjects(hosts, w, c)
	} else {
		err = printHostList(hosts.Items, os.Stdout, c)
		if err != nil {
			return err
		}
	}

	return nil
}

// Update the image datastores using the information carried in cli.Context.
func updateInfrastructureImageDatastores(c *cli.Context) error {

	datastores := c.String("datastores")

	if !c.GlobalIsSet("non-interactive") {
		var err error
		datastores, err = askForInput("Datastores: ", datastores)
		if err != nil {
			return err
		}
	}

	if len(datastores) == 0 {
		return fmt.Errorf("Please provide datastores using --datastores flag")
	}

	imageDataStores := &photon.ImageDatastores{
		Items: regexp.MustCompile(`\s*,\s*`).Split(datastores, -1),
	}
	var err error
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Photonclient.Infra.SetImageDatastores(imageDataStores)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = systemInfoJsonHelper(c, client.Photonclient)
	if err != nil {
		return err
	}

	return nil
}

// Synchronizes hosts configurations
func syncInfrastructureHostsConfig(c *cli.Context) error {
	var err error
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Photonclient.Infra.SyncHostsConfig()
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = systemInfoJsonHelper(c, client.Photonclient)
	if err != nil {
		return err
	}

	return nil
}
