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
	"regexp"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/cli/client"
	cf "github.com/vmware/photon-controller-cli/photon/cli/configuration"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for project
// Subcommands:  create; Usage: project create [<options>]
//	             delete; Usage: project delete <id>
//	             set;    Usage: project set <name>
//	             show;   Usage: project show
//	             list;   Usage: project list [<options>]
//	             tasks;  Usage: project tasks <name> [<options>]
func GetProjectsCommand() cli.Command {
	command := cli.Command{
		Name:  "project",
		Usage: "options for project",
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create a new project",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Project name",
					},
					cli.StringFlag{
						Name:  "limits, l",
						Usage: "Project limits(key,value,unit)",
					},
					cli.Float64Flag{
						Name:  "percent, p",
						Usage: "Project limits(percentage of resource-ticket)",
					},
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name for project",
					},
					cli.StringFlag{
						Name:  "resource-ticket, r",
						Usage: "Resource-ticket name for project",
					},
					cli.StringFlag{
						Name:  "security-groups, g",
						Usage: "Security Groups for project",
					},
				},
				Action: func(c *cli.Context) {
					err := createProject(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:  "delete",
				Usage: "Delete project with specified id",
				Action: func(c *cli.Context) {
					err := deleteProject(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:  "show",
				Usage: "Show project in config file",
				Action: func(c *cli.Context) {
					err := showProject(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:  "set",
				Usage: "Set project in config file",
				Action: func(c *cli.Context) {
					err := setProject(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:  "list",
				Usage: "List all projects",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name for project",
					},
				},
				Action: func(c *cli.Context) {
					err := listProjects(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:  "tasks",
				Usage: "List all tasks related to the project",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "state, s",
						Usage: "specify task state for filtering",
					},
					cli.StringFlag{
						Name:  "kind, k",
						Usage: "specify task kind for filtering",
					},
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name for project",
					},
				},
				Action: func(c *cli.Context) {
					err := getProjectTasks(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:  "set_security_groups",
				Usage: "Set security groups for a project",
				Action: func(c *cli.Context) {
					err := setSecurityGroupsForProject(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Sends a create project task to client based on the cli.Context
// Returns an error if one occurred
func createProject(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "project create [<options>]")
	if err != nil {
		return err
	}
	tenantName := c.String("tenant")
	rtName := c.String("resource-ticket")
	name := c.String("name")
	limits := c.String("limits")
	percent := c.Float64("percent")
	securityGroups := c.String("security-groups")

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	tenant, err := verifyTenant(tenantName)
	if err != nil {
		return err
	}

	var limitsList []photon.QuotaLineItem
	if c.IsSet("limits") && c.IsSet("percent") {
		return fmt.Errorf("Error: Can only specify one of '--limits' or '--percent'")
	}
	if c.IsSet("limits") {
		limitsList, err = parseLimitsListFromFlag(limits)
		if err != nil {
			return err
		}
	}
	if c.IsSet("percent") {
		limitsList = []photon.QuotaLineItem{
			{Key: "subdivide.percent", Value: percent, Unit: "COUNT"}}
	}

	if !c.GlobalIsSet("non-interactive") {
		name, err = askForInput("Project name: ", name)
		if err != nil {
			return err
		}
		rtName, err = askForInput("Resource-ticket name: ", rtName)
		if err != nil {
			return err
		}
		limitsList, err = askForLimitList(limitsList)
		if err != nil {
			return err
		}
	}

	projectSpec := photon.ProjectCreateSpec{}
	projectSpec.Name = name
	projectSpec.ResourceTicket = photon.ResourceTicketReservation{Name: rtName, Limits: limitsList}

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("\nTenant name: %s\n", tenant.Name)
		fmt.Printf("Resource ticket name: %s\n", rtName)
		fmt.Printf("Creating project name: %s\n\n", name)
		fmt.Println("Please make sure limits below are correct:")
		for i, l := range limitsList {
			fmt.Printf("%d: %s, %g, %s\n", i+1, l.Key, l.Value, l.Unit)
		}
	}
	if confirmed(c.GlobalIsSet("non-interactive")) {
		if len(securityGroups) > 0 {
			projectSpec.SecurityGroups = regexp.MustCompile(`\s*,\s*`).Split(securityGroups, -1)
		}
		createTask, err := client.Esxclient.Tenants.CreateProject(tenant.ID, &projectSpec)
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
			fmt.Printf("Project created: ID = %s\n", createTask.Entity.ID)
		}
	} else {
		fmt.Println("OK. Canceled")
	}

	return nil
}

// Sends a delete project task to client based on the cli.Context
// Returns an error if one occurred
func deleteProject(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "project delete <project id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	deleteTask, err := client.Esxclient.Projects.Delete(id)
	if err != nil {
		return err
	}
	err = waitOnTaskOperation(deleteTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	err = clearConfigProject(id)
	if err != nil {
		return err
	}

	return nil
}

// Sends a show project task to client based on the config file
// Returns an error if one occurred
func showProject(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "project show")
	if err != nil {
		return err
	}
	config, err := cf.LoadConfig()
	if err != nil {
		return err
	}

	project := config.Project
	if project == nil {
		return fmt.Errorf("Error: No Project selected\n")
	}

	if c.GlobalIsSet("non-interactive") {
		fmt.Printf("%s\t%s\n", project.ID, project.Name)
	} else {
		fmt.Printf("Current project is '%s' with ID %s\n", project.ID, project.Name)
	}
	return nil
}

// Set project name and id to config file
// Returns an error if one occurred
func setProject(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "project set <project name>")
	if err != nil {
		return err
	}
	name := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	config, err := cf.LoadConfig()
	if err != nil {
		return err
	}

	if config == nil || config.Tenant == nil {
		return fmt.Errorf("Error: Set tenant first using 'tenant set <name>' or '-t <name>' option")
	}

	project, err := findProject(config.Tenant.ID, name)
	if err != nil {
		return err
	}

	config.Project = &cf.ProjectConfiguration{Name: project.Name, ID: project.ID}
	err = cf.SaveConfig(config)
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("Project set to '%s'\n", name)
	}

	return nil
}

// Retrieves a list of projects, returns an error if one occurred
func listProjects(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "project list [<options>]")
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

	tickets, err := client.Esxclient.Tenants.GetProjects(tenant.ID, nil)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		for _, t := range tickets.Items {
			limits := quotaLineItemListToString(t.ResourceTicket.Limits)
			usage := quotaLineItemListToString(t.ResourceTicket.Usage)
			fmt.Printf("%s\t%s\t%s\t%s\n", t.ID, t.Name, limits, usage)
		}
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\tLimit\tUsage\n")
		for _, t := range tickets.Items {
			rt := t.ResourceTicket
			for i := 0; i < len(rt.Limits); i++ {
				if i == 0 {
					fmt.Fprintf(w, "%s\t%s\t%s %g %s\t%s %g %s\n", t.ID, t.Name,
						rt.Limits[i].Key, rt.Limits[i].Value, rt.Limits[i].Unit,
						rt.Usage[i].Key, rt.Usage[i].Value, rt.Usage[i].Unit)
				} else {
					fmt.Fprintf(w, "\t\t%s %g %s\t%s %g %s\n",
						rt.Limits[i].Key, rt.Limits[i].Value, rt.Limits[i].Unit,
						rt.Usage[i].Key, rt.Usage[i].Value, rt.Usage[i].Unit)
				}
			}
			for i := len(rt.Limits); i < len(rt.Usage); i++ {
				fmt.Fprintf(w, "\t\t\t%s %g %s\n", rt.Usage[i].Key, rt.Usage[i].Value, rt.Usage[i].Unit)
			}
		}
		err := w.Flush()
		if err != nil {
			return err
		}
		fmt.Printf("\nTotal projects: %d\n", len(tickets.Items))
	}
	return nil
}

// Retrieves tasks for project
func getProjectTasks(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "project tasks <project name> [<options>]")
	if err != nil {
		return err
	}
	name := c.Args().First()
	tenantName := c.String("tenant")
	state := c.String("state")
	kind := c.String("kind")

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	tenant, err := verifyTenant(tenantName)
	if err != nil {
		return err
	}

	project, err := findProject(tenant.ID, name)
	if err != nil {
		return err
	}

	options := &photon.TaskGetOptions{
		State: state,
		Kind:  kind,
	}

	taskList, err := client.Esxclient.Projects.GetTasks(project.ID, options)
	if err != nil {
		return err
	}

	err = printTaskList(taskList.Items, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Set security groups for a project
func setSecurityGroupsForProject(c *cli.Context) error {
	err := checkArgNum(c.Args(), 2, "project set_security_groups <id> <comma-separated security group names>")
	if err != nil {
		return err
	}
	id := c.Args().First()
	securityGroups := &photon.SecurityGroups{
		Items: regexp.MustCompile(`\s*,\s*`).Split(c.Args()[1], -1),
	}
	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	task, err := client.Esxclient.Projects.SetSecurityGroups(id, securityGroups)
	if err != nil {
		return err
	}

	err = waitOnTaskOperation(task.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}
