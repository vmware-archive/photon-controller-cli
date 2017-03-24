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
	"fmt"
	"io"
	"log"
	"os"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for zone
// Subcommands: create; Usage: zone create [<options>]
//              delete; Usage: zone delete <id>
//              list;   Usage: zone list
//              show;   Usage: zone show <id>
//              tasks;  Usage: zone tasks <id> [<options>]
func GetZonesCommand() cli.Command {
	command := cli.Command{
		Name:  "zone",
		Usage: "options for zone",
		Subcommands: []cli.Command{
			{
				Name:        "create",
				Usage:       "Create a new zone",
				ArgsUsage:   " ",
				Description: "This creates a new zone. Only a system adminstrator can create one.",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Zone name",
					},
				},
				Action: func(c *cli.Context) {
					err := createZone(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "delete",
				Usage: "Delete an zone",
				Description: "This deletess an existing availablity zone given its id.\n" +
					"   Only a system adminstrator can delete the zone.",
				ArgsUsage: "<zone-id>",
				Action: func(c *cli.Context) {
					err := deleteZone(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "list",
				Usage:     "List all zones",
				ArgsUsage: " ",
				Action: func(c *cli.Context) {
					err := listZones(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show specified zone",
				ArgsUsage: "<zone-id>",
				Action: func(c *cli.Context) {
					err := showZone(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "tasks",
				Usage:     "Show zone tasks",
				ArgsUsage: "<zone-id>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "state, s",
						Usage: "Filter by task sate",
					},
				},
				Action: func(c *cli.Context) {
					err := getZoneTasks(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Sends a create zone task to client based on the cli.Context
// Returns an error if one occurred
func createZone(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	name := c.String("name")

	if !c.GlobalIsSet("non-interactive") {
		var err error
		name, err = askForInput("Zone name: ", name)
		if err != nil {
			return err
		}
	}

	if len(name) == 0 {
		return fmt.Errorf("Please provide zone name")
	}

	azSpec := &photon.ZoneCreateSpec{
		Name: name,
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	createTask, err := client.Photonclient.Zones.Create(azSpec)
	if err != nil {
		return err
	}

	id, err := waitOnTaskOperation(createTask.ID, c)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		zone, err := client.Photonclient.Zones.Get(id)
		if err != nil {
			return err
		}
		utils.FormatObject(zone, w, c)
	}

	return nil
}

// Retrieves zone against specified id.
func showZone(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	zone, err := client.Photonclient.Zones.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		fmt.Printf("%s\t%s\t%s\t%s\n", zone.ID, zone.Name, zone.Kind, zone.State)
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(zone, w, c)
	} else {
		fmt.Printf("  Zone ID:     %s\n", zone.ID)
		fmt.Printf("  Name:        %s\n", zone.Name)
		fmt.Printf("  Kind:        %s\n", zone.Kind)
		fmt.Printf("  State:       %s\n", zone.State)
	}

	return nil
}

// Retrieves a list of zones, returns an error if one occurred
func listZones(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	zones, err := client.Photonclient.Zones.GetAll()
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		for _, zone := range zones.Items {
			fmt.Printf("%s\t%s\n", zone.ID, zone.Name)
		}
	} else if utils.NeedsFormatting(c) {
		utils.FormatObjects(zones.Items, w, c)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\n")
		for _, zone := range zones.Items {
			fmt.Fprintf(w, "%s\t%s\n", zone.ID, zone.Name)
		}
		err = w.Flush()
		if err != nil {
			return err
		}
		fmt.Printf("\nTotal: %d\n", len(zones.Items))
	}

	return nil
}

// Sends a delete zone task to client based on the cli.Context
// Returns an error if one occurred
func deleteZone(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deleteTask, err := client.Photonclient.Zones.Delete(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(deleteTask.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Retrieves tasks from specified zone
func getZoneTasks(c *cli.Context, w io.Writer) error {
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

	taskList, err := client.Photonclient.Zones.GetTasks(id, options)
	if err != nil {
		return err
	}

	err = printTaskList(taskList.Items, c)
	if err != nil {
		return err
	}
	return nil
}
