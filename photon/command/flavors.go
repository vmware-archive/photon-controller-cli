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
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for flavor
// Subcommands: create; Usage: flavor create [<options>]
//              delete; Usage: flavor delete <id>
//              show;   Usage: flavor show <id>
//              list;   Usage: flavor list [<options>]
//              tasks;  Usage: flavor tasks <id> [<options>]
func GetFlavorsCommand() cli.Command {
	command := cli.Command{
		Name:  "flavor",
		Usage: "options for flavor",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a flavor",
				ArgsUsage: " ",
				Description: "This creates a new flavor. Only system administrators can create flavors.\n" +
					"   A flavor is defined by a set of costs. Each cost has a type (e.g. vm.memory),\n" +
					"   a numnber (e.g. 1) and a unit (GB, MB, KB, B, or COUNT). VM flavors must specify at\n" +
					"   least two costs: vm.memory and vm.cpu.\n\n" +
					"   Example VM flavor command:\n" +
					"      photon flavor create --name f1 --kind vm --cost 'vm.memory 1 GB, vm.cpu 1 COUNT'\n" +
					"   Example disk flavor:\n" +
					"      photon flavor create --name f1 --kind persistent-disk --cost 'persistent-disk 1 COUNT'\n",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Flavor name",
					},
					cli.StringFlag{
						Name:  "kind, k",
						Usage: "Flavor kind: persistent-disk, ephemeral-disk, or vm",
					},
					cli.StringFlag{
						Name:  "cost, c",
						Usage: "Comma-separated costs. Each cost is \"type number unit\"",
					},
				},
				Action: func(c *cli.Context) {
					err := createFlavor(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:        "delete",
				Usage:       "Deletes a flavor",
				ArgsUsage:   "<flavor-id>",
				Description: "Deletes a flavor. You must be a system administrator to delete a flavor.",
				Action: func(c *cli.Context) {
					err := deleteFlavor(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "list",
				Usage:     "Lists all flavors",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Filter by flavor name",
					},
					cli.StringFlag{
						Name:  "kind, k",
						Usage: "Filter by flavor kind",
					},
				},
				Action: func(c *cli.Context) {
					err := listFlavors(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show flavor info",
				ArgsUsage: "<flavor-id>",
				Action: func(c *cli.Context) {
					err := showFlavor(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "tasks",
				Usage:     "Show all tasks for given flavor",
				ArgsUsage: "<flavor-id>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "state, s",
						Usage: "Filter by task state (QUEUED, STARTED, ERROR, or COMPLETED)",
					},
				},
				Action: func(c *cli.Context) {
					err := getFlavorTasks(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Sends a create flavor task to client
func createFlavor(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	name := c.String("name")
	kind := c.String("kind")
	cost := c.String("cost")

	costList, err := parseLimitsListFromFlag(cost)
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") {
		name, err = askForInput("Flavor name: ", name)
		if err != nil {
			return err
		}
		kind, err = askForInput("Flavor kind (persistent-disk, ephemeral-disk, or vm): ", kind)
		if err != nil {
			return err
		}
		costList, err = askForLimitList(costList)
		if err != nil {
			return err
		}
	}

	if len(name) == 0 {
		return fmt.Errorf("Please provide flavor name")
	}

	if len(kind) == 0 || (kind != "persistent-disk" && kind != "ephemeral-disk" && kind != "vm") {
		return fmt.Errorf("Please provide flavor kind: persistent-disk, ephemeral-disk, or vm")
	}

	createSpec := &photon.FlavorCreateSpec{
		Name: name,
		Kind: kind,
		Cost: costList,
	}

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("Creating flavor: '%s', Kind: '%s'\n\n", name, kind)
		fmt.Printf("Please make sure limits below are correct: \n")
		for i, l := range costList {
			fmt.Printf("%d: %s, %g, %s\n", i+1, l.Key, l.Value, l.Unit)
		}
	}

	if confirmed(c.GlobalIsSet("non-interactive")) {
		var err error
		client.Esxclient, err = client.GetClient(c)
		if err != nil {
			return err
		}

		createTask, err := client.Esxclient.Flavors.Create(createSpec)
		if err != nil {
			return err
		}
		flavorId, err := waitOnTaskOperation(createTask.ID, c)
		if err != nil {
			return err
		}
		if utils.NeedsFormatting(c) {
			flavor, err := client.Esxclient.Flavors.Get(flavorId)
			if err != nil {
				return err
			}
			utils.FormatObject(flavor, w, c)
		}
	} else {
		fmt.Println("OK. Canceled")
	}

	return nil
}

// Sends a delete flavor task to the client
func deleteFlavor(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deleteTask, err := client.Esxclient.Flavors.Delete(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(deleteTask.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Retrieves a list of flavors
func listFlavors(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	name := c.String("name")
	kind := c.String("kind")
	options := &photon.FlavorGetOptions{
		Name: name,
		Kind: kind,
	}

	flavors, err := client.Esxclient.Flavors.GetAll(options)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		for _, flavor := range flavors.Items {
			costs := quotaLineItemListToString(flavor.Cost)
			fmt.Printf("%s\t%s\t%s\t%s\n", flavor.ID, flavor.Name, flavor.Kind, costs)
		}
	} else if utils.NeedsFormatting(c) {
		utils.FormatObjects(flavors.Items, w, c)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\tKind\tCost\n")
		for _, flavor := range flavors.Items {
			printQuotaList(w, flavor.Cost, flavor.ID, flavor.Name, flavor.Kind)
		}
		err = w.Flush()
		if err != nil {
			return err
		}
		fmt.Printf("Total: %d\n", len(flavors.Items))
	}

	return nil
}

// Retrieves information about a flavor
func showFlavor(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	flavor, err := client.Esxclient.Flavors.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		costs := quotaLineItemListToString(flavor.Cost)
		fmt.Printf("%s\t%s\t%s\t%s\t%s\n", flavor.ID, flavor.Name, flavor.Kind, costs, flavor.State)
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(flavor, w, c)
	} else {
		costList := []string{}
		for _, cost := range flavor.Cost {
			costList = append(costList, fmt.Sprintf("%s %g %s", cost.Key, cost.Value, cost.Unit))
		}
		fmt.Printf("Flavor ID: %s\n", flavor.ID)
		fmt.Printf("  Name:  %s\n", flavor.Name)
		fmt.Printf("  Kind:  %s\n", flavor.Kind)
		fmt.Printf("  Cost:  %s\n", costList)
		fmt.Printf("  State: %s\n", flavor.State)
	}

	return nil
}

// Retrieves tasks from specified flavor
func getFlavorTasks(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	state := c.String("state")
	options := &photon.TaskGetOptions{
		State: state,
	}

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	taskList, err := client.Esxclient.Flavors.GetTasks(id, options)
	if err != nil {
		return err
	}

	err = printTaskList(taskList.Items, c)
	if err != nil {
		return err
	}
	return nil
}
