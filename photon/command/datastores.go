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
	"strings"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/urfave/cli"
)

// Creates a cli.Command for datastore
// Subcommands: show;                  Usage: datastore show <id>
//              list;                  Usage: datastore list
func GetDatastoresCommand() cli.Command {
	command := cli.Command{
		Name:  "datastore",
		Usage: "options for datastore",
		Subcommands: []cli.Command{
			{
				Name:      "list",
				Usage:     "List all the datastores known by Photon Controller",
				ArgsUsage: " ",
				Action: func(c *cli.Context) {
					err := listDatastores(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "show",
				Usage: "Show information about the datastore with the given id",
				Action: func(c *cli.Context) {
					err := showDatastore(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// listDatastores will list all datastores known to Photon Controller
func listDatastores(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	datastores, err := client.Photonclient.Datastores.GetAll()
	if err != nil {
		return err
	}

	err = printDatastoreList(datastores.Items, w, c)
	if err != nil {
		return err
	}
	return nil
}

// showDatastore will show detailed information about a single datastore
func showDatastore(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	datastore, err := client.Photonclient.Datastores.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		tag := strings.Trim(fmt.Sprint(datastore.Tags), "[]")
		scriptTag := strings.Replace(tag, " ", ",", -1)
		fmt.Printf("%s\t%s\t%s\n", datastore.ID, datastore.Type, scriptTag)
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(datastore, w, c)
	} else {
		fmt.Println("Datastore ID: ", datastore.ID)
		fmt.Println("  Type:       ", datastore.Type)
		fmt.Println("  Tags:       ", datastore.Tags)
	}

	return nil
}
