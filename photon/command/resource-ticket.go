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

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for resource-ticket
// Subcommands: create; Usage: resource-ticket create [<options>]
//              show;   Usage: resource-ticket show <name> [<options>]
//              list;   Usage: resource-ticket list [<options>]
//              tasks;  Usage: resource-ticket tasks <name> [<options>]
func GetResourceTicketCommand() cli.Command {
	command := cli.Command{
		Name:  "resource-ticket",
		Usage: "options for resource-ticket",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new resource-ticket",
				ArgsUsage: " ",
				Description: "Create a new resource ticket for a tenant.\n" +
					"   Only system administrators can create new resource tickets.\n" +
					"   A flavor is defined by a set of maximum resource costs. Each usage has a type (e.g. vm.memory),\n" +
					"   a numnber (e.g. 1) and a unit (GB, MB, KB, B, or COUNT). You must specify at least one cost\n" +
					"   Example ticket with 1000 GB of RAM and 500 vCPUs:\n" +
					"      photon resource-ticket create --name ticket1 --tenant tenant1 --limits 'vm.memory 1000 GB, vm.cpu 500 COUNT'",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Resource-ticket name",
					},
					cli.StringFlag{
						Name:  "limits, l",
						Usage: "Resource-ticket limits(key value unit)",
					},
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name for resource-ticket",
					},
				},
				Action: func(c *cli.Context) {
					err := createResourceTicket(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show resource-ticket info with specified name",
				ArgsUsage: "<resource-ticket-name>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name for resource-ticket",
					},
				},
				Action: func(c *cli.Context) {
					err := showResourceTicket(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "list",
				Usage:     "List all resource tickets",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name for resource-ticket",
					},
				},
				Action: func(c *cli.Context) {
					err := listResourceTickets(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "tasks",
				Usage:     "List all tasks related to the resource ticket",
				ArgsUsage: "<resource-ticket-name>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "state, s",
						Usage: "specify task state for filtering",
					},
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name for resource-ticket",
					},
				},
				Action: func(c *cli.Context) {
					err := getResourceTicketTasks(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
		},
	}
	return command
}

// Sends a create resource-ticket task to client based on the cli.Context
// Returns an error if one occurred
func createResourceTicket(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	tenantName := c.String("tenant")
	name := c.String("name")
	limits := c.String("limits")

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	tenant, err := verifyTenant(tenantName)
	if err != nil {
		return err
	}

	limitsList, err := parseLimitsListFromFlag(limits)
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") {
		name, err = askForInput("Resource ticket name: ", name)
		if err != nil {
			return err
		}
		limitsList, err = askForLimitList(limitsList)
		if err != nil {
			return err
		}
	}

	rtSpec := photon.ResourceTicketCreateSpec{}
	rtSpec.Name = name
	rtSpec.Limits = limitsList

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("\nTenant name: %s\n", tenant.Name)
		fmt.Printf("Creating resource ticket name: %s\n\n", name)
		fmt.Println("Please make sure limits below are correct:")
		for i, l := range limitsList {
			fmt.Printf("%d: %s, %g, %s\n", i+1, l.Key, l.Value, l.Unit)
		}
	}

	if confirmed(c) {
		createTask, err := client.Photonclient.Tenants.CreateResourceTicket(tenant.ID, &rtSpec)
		if err != nil {
			return err
		}

		_, err = waitOnTaskOperation(createTask.ID, c)
		if err != nil {
			return err
		}

		if utils.NeedsFormatting(c) {
			options := &photon.ResourceTicketGetOptions{
				Name: name,
			}
			tickets, err := client.Photonclient.Tenants.GetResourceTickets(tenant.ID, options)
			if err != nil {
				return err
			}
			utils.FormatObject(tickets.Items[0], w, c)
		}
	} else {
		fmt.Println("OK. Canceled")
	}

	return nil
}

// Sends a show resource-ticket task to client based on the cli.Context
// Returns an error if one occurred
func showResourceTicket(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	name := c.Args().First()
	tenantName := c.String("tenant")

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	tenant, err := verifyTenant(tenantName)
	if err != nil {
		return err
	}

	rt, err := findResourceTicket(tenant.ID, name)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		usage := quotaLineItemListToString(rt.Usage)
		limits := quotaLineItemListToString(rt.Limits)
		fmt.Printf("%s\t%s\t%s\t%s\n", rt.Name, rt.ID, limits, usage)
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(rt, w, c)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\tLimit\tUsage\n")
		for i := 0; i < len(rt.Limits); i++ {
			if i == 0 {
				fmt.Fprintf(w, "%s\t%s\t%s %g %s\t%s %g %s\n", rt.ID, rt.Name,
					rt.Limits[i].Key, rt.Limits[i].Value, rt.Limits[i].Unit,
					rt.Usage[i].Key, rt.Usage[i].Value, rt.Usage[i].Unit)
			} else {
				fmt.Fprintf(w, "\t\t%s %g %s\t%s %g %s\n",
					rt.Limits[i].Key, rt.Limits[i].Value, rt.Limits[i].Unit,
					rt.Usage[i].Key, rt.Usage[i].Value, rt.Usage[i].Unit)
			}
		}
		err := w.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}

// Retrieves a list of resource tickets, returns an error if one occurred
func listResourceTickets(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	tenantName := c.String("tenant")

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	tenant, err := verifyTenant(tenantName)
	if err != nil {
		return err
	}

	tickets, err := client.Photonclient.Tenants.GetResourceTickets(tenant.ID, nil)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		for _, t := range tickets.Items {
			limits := quotaLineItemListToString(t.Limits)
			fmt.Printf("%s\t%s\t%s\n", t.ID, t.Name, limits)
		}
	} else if utils.NeedsFormatting(c) {
		utils.FormatObjects(tickets.Items, w, c)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\tLimit\n")
		for _, t := range tickets.Items {
			printQuotaList(w, t.Limits, t.ID, t.Name)
		}
		err := w.Flush()
		if err != nil {
			return err
		}
		fmt.Printf("\nTotal resource tickets: %d\n", len(tickets.Items))
	}
	return nil
}

// Retrieves tasks for resource ticket
func getResourceTicketTasks(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	name := c.Args().First()
	tenantName := c.String("tenant")
	state := c.String("state")

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	tenant, err := verifyTenant(tenantName)
	if err != nil {
		return err
	}

	rt, err := findResourceTicket(tenant.ID, name)
	if err != nil {
		return err
	}

	options := &photon.TaskGetOptions{
		State: state,
	}

	taskList, err := client.Photonclient.ResourceTickets.GetTasks(rt.ID, options)
	if err != nil {
		return err
	}

	err = printTaskList(taskList.Items, c)
	if err != nil {
		return err
	}

	return nil
}
