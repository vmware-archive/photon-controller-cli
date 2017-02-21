// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package command

import (
	"log"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-cli/photon/client"
)

// Creates a cli.Command for router
// Subcommands: delete; Usage: router delete <id>
func GetRoutersCommand() cli.Command {
	command := cli.Command{
		Name:  "router",
		Usage: "options for router",
		Subcommands: []cli.Command{
			{
				Name:        "delete",
				Usage:       "Delete router with specified id",
				ArgsUsage:   "<router-id>",
				Description: "Delete a router.",
				Action: func(c *cli.Context) {
					err := deleteRouter(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
		},
	}
	return command
}

// Sends a delete router task to client based on the cli.Context
// Returns an error if one occurred
func deleteRouter(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deleteTask, err := client.Photonclient.Routers.Delete(id)
	if err != nil {
		return err
	}
	_, err = waitOnTaskOperation(deleteTask.ID, c)
	if err != nil {
		return err
	}

	return nil
}
