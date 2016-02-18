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
	"os"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/cli/client"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
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
				Name:  "create",
				Usage: "Create a new resource-ticket",
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
					err := createResourceTicket(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:  "show",
				Usage: "Show resource-ticket info with specified name",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name for resource-ticket",
					},
				},
				Action: func(c *cli.Context) {
					err := showResourceTicket(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:  "list",
				Usage: "List all resource tickets",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name for resource-ticket",
					},
				},
				Action: func(c *cli.Context) {
					err := listResourceTickets(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:  "tasks",
				Usage: "List all tasks related to the resource ticket",
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
					err := getResourceTicketTasks(c)
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
func createResourceTicket(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "resource-ticket create [<options>]")
	if err != nil {
		return err
	}
	tenantName := c.String("tenant")
	name := c.String("name")
	limits := c.String("limits")

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
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

	if confirmed(c.GlobalIsSet("non-interactive")) {
		createTask, err := client.Esxclient.Tenants.CreateResourceTicket(tenant.ID, &rtSpec)
		if err != nil {
			return err
		}
		if c.GlobalIsSet("non-interactive") {
			createTask, err = client.Esxclient.Tasks.Wait(createTask.ID)
			if err != nil {
				return err
			}
			fmt.Println(createTask.Entity.ID)
		} else {
			createTask, err = pollTask(createTask.ID)
			if err != nil {
				return err
			}
			fmt.Printf("Resource ticket created: ID = %s\n", createTask.Entity.ID)
		}
	} else {
		fmt.Println("OK. Canceled")
	}

	return nil
}

// Sends a show resource-ticket task to client based on the cli.Context
// Returns an error if one occurred
func showResourceTicket(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "resource-ticket show <name> [<options>]")
	if err != nil {
		return err
	}
	name := c.Args().First()
	tenantName := c.String("tenant")

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
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
func listResourceTickets(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "resource-ticket list [<options>]")
	if err != nil {
		return err
	}
	tenantName := c.String("tenant")

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	tenant, err := verifyTenant(tenantName)
	if err != nil {
		return err
	}

	tickets, err := client.Esxclient.Tenants.GetResourceTickets(tenant.ID, nil)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		fmt.Println(len(tickets.Items))
		for _, t := range tickets.Items {
			limits := quotaLineItemListToString(t.Limits)
			fmt.Printf("%s\t%s\t%s\n", t.ID, t.Name, limits)
		}
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
func getResourceTicketTasks(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "resource-ticket tasks <name> [<options>]")
	if err != nil {
		return err
	}
	name := c.Args().First()
	tenantName := c.String("tenant")
	state := c.String("state")

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
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

	taskList, err := client.Esxclient.ResourceTickets.GetTasks(rt.ID, options)
	if err != nil {
		return err
	}

	err = printTaskList(taskList.Items, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}
