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
	"strings"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/client"
	cf "github.com/vmware/photon-controller-cli/photon/configuration"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for project
// Subcommands: create; Usage: project create [<options>]
//              delete; Usage: project delete <id>
//              show;   Usage: project show <id>
//              set;    Usage: project set <name>
//              get;    Usage: project get
//              list;   Usage: project list [<options>]
//              tasks;  Usage: project tasks <id> [<options>]
//              quota;  Usage: project quota <operation> <name> [<options>]
func GetProjectsCommand() cli.Command {
	command := cli.Command{
		Name:  "project",
		Usage: "options for project",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new project",
				ArgsUsage: " ",
				Description: "Create a new project within a tenant and assigns it some or all of a resource ticket.\n" +
					"   Only system administrators can create new projects.\n" +
					"   If default-router-private-ip-cidr option is omitted,\n" +
					"   it will use 192.168.0.0/16 as default router's private IP CIDR.",
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
					cli.StringFlag{
						Name:  "default-router-private-ip-cidr, c",
						Usage: "Private IP range of the default router in CIDR format. Default value: 192.168.0.0/16",
					},
				},
				Action: func(c *cli.Context) {
					err := createProject(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:        "delete",
				Usage:       "Delete project with specified id",
				ArgsUsage:   "<project-id>",
				Description: "Delete a project. You must be a system administrator to delete a project.",
				Action: func(c *cli.Context) {
					err := deleteProject(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show project info with specified id",
				ArgsUsage: "<project-id>",
				Action: func(c *cli.Context) {
					err := showProject(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "get",
				Usage:     "Show default project.",
				ArgsUsage: " ",
				Description: "Show default project in use for photon CLI commands. Most command allow you to either\n" +
					"   use this default or specify a specific project to use.",
				Action: func(c *cli.Context) {
					err := getProject(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "set",
				Usage:     "Set default project",
				ArgsUsage: "<project-name>",
				Description: "Set the default project that will be used for all photon CLI commands that need a project.\n" +
					"   Most commands allow you to override the default.",
				Action: func(c *cli.Context) {
					err := setProject(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "list",
				Usage:     "List all projects",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name for project",
					},
				},
				Action: func(c *cli.Context) {
					err := listProjects(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "tasks",
				Usage:     "List all tasks related to a given project",
				ArgsUsage: "<project-id>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "state, s",
						Usage: "specify task state for filtering",
					},
					cli.StringFlag{
						Name:  "kind, k",
						Usage: "specify task kind for filtering",
					},
				},
				Action: func(c *cli.Context) {
					err := getProjectTasks(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "set-security-groups",
				Usage:     "Set security groups for a project",
				ArgsUsage: "<project-id> <comma separated list of groups>",
				Description: "Set the list of Lightwave groups that can use this project. This may only be\n" +
					"   be set by a member of the project. Be cautious--you can remove your own access if you specify\n" +
					"   the wrong set of groups.\n\n" +
					"   A security group specifies both the Lightwave domain and Lightwave group.\n" +
					"   For example, a security group may be photon.vmware.com\\group-1\n\n" +
					"   Example: photon project 3f78619d-20b1-4b86-a7a6-5a9f09e59ef6 set-security-groups 'photon.vmware.com\\group-1,photon.vmware.com\\group-2'",
				Action: func(c *cli.Context) {
					err := setSecurityGroupsForProject(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Hidden:      true,
				Name:        "set_security_groups",
				Usage:       "Set security groups for a project",
				ArgsUsage:   "<project-id> <comma separated list of groups>",
				Description: "Deprecated, use set-security-groups instead",
				Action: func(c *cli.Context) {
					err := setSecurityGroupsForProject(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "iam",
				Usage: "options for identity and access management",
				Subcommands: []cli.Command{
					{
						Name:      "show",
						Usage:     "Show the IAM policy associated with a project",
						ArgsUsage: "<project-id>",
						Action: func(c *cli.Context) {
							err := getProjectIam(c)
							if err != nil {
								log.Fatal("Error: ", err)
							}
						},
					},
					{
						Name:      "add",
						Usage:     "Grant a role to a user or group on a project",
						ArgsUsage: "<project-id>",
						Description: "Grant a role to a user or group on a project. \n\n" +
							"   Example: \n" +
							"   photon project iam add <project-id> -p user1@photon.local -r contributor\n" +
							"   photon project iam add <project-id> -p photon.local\\group1 -r viewer",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "principal, p",
								Usage: "User or group",
							},
							cli.StringFlag{
								Name:  "role, r",
								Usage: "'owner', 'contributor' and 'viewer'",
							},
						},
						Action: func(c *cli.Context) {
							err := modifyProjectIamPolicy(c, os.Stdout, "ADD")
							if err != nil {
								log.Fatal("Error: ", err)
							}
						},
					},
					{
						Name:      "remove",
						Usage:     "Remove a role from a user or group on a project",
						ArgsUsage: "<project-id>",
						Description: "Remove a role from a user or group on a project. \n\n" +
							"   Example: \n" +
							"   photon project iam remove <project-id> -p user1@photon.local -r contributor \n" +
							"   photon project iam remove <project-id> -p photon.local\\group1 -r viewer",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "principal, p",
								Usage: "User or group",
							},
							cli.StringFlag{
								Name:  "role, r",
								Usage: "'owner', 'contributor' and 'viewer'. Or use '*' to remove all existing roles.",
							},
						},
						Action: func(c *cli.Context) {
							err := modifyProjectIamPolicy(c, os.Stdout, "REMOVE")
							if err != nil {
								log.Fatal("Error: ", err)
							}
						},
					},
				},
			},
			// Load Project Quota related logic from separated file.
			getProjectQuotaCommand(),
		},
	}
	return command
}

// Sends a create project task to client based on the cli.Context
// Returns an error if one occurred
func createProject(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	tenantName := c.String("tenant")
	rtName := c.String("resource-ticket")
	name := c.String("name")
	limits := c.String("limits")
	percent := c.Float64("percent")
	securityGroups := c.String("security-groups")
	defaultRouterPrivateIpCidr := c.String("default-router-private-ip-cidr")

	if len(defaultRouterPrivateIpCidr) == 0 {
		defaultRouterPrivateIpCidr = "192.168.0.0/16"
	}

	client.Photonclient, err = client.GetClient(c)
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
	projectSpec.DefaultRouterPrivateIpCidr = defaultRouterPrivateIpCidr

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("\nTenant name: %s\n", tenant.Name)
		fmt.Printf("Resource ticket name: %s\n", rtName)
		fmt.Printf("Creating project name: %s\n\n", name)
		fmt.Println("Please make sure limits below are correct:")
		for i, l := range limitsList {
			fmt.Printf("%d: %s, %g, %s\n", i+1, l.Key, l.Value, l.Unit)
		}
	}
	if confirmed(c) {
		if len(securityGroups) > 0 {
			projectSpec.SecurityGroups = regexp.MustCompile(`\s*,\s*`).Split(securityGroups, -1)
		}
		createTask, err := client.Photonclient.Tenants.CreateProject(tenant.ID, &projectSpec)
		if err != nil {
			return err
		}

		id, err := waitOnTaskOperation(createTask.ID, c)
		if err != nil {
			return err
		}

		if utils.NeedsFormatting(c) {
			project, err := client.Photonclient.Projects.Get(id)
			if err != nil {
				return err
			}
			utils.FormatObject(project, w, c)
		}
	} else {
		fmt.Println("OK. Canceled")
	}

	return nil
}

// Sends a delete project task to client based on the cli.Context
// Returns an error if one occurred
func deleteProject(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deleteTask, err := client.Photonclient.Projects.Delete(id)
	if err != nil {
		return err
	}
	_, err = waitOnTaskOperation(deleteTask.ID, c)
	if err != nil {
		return err
	}

	err = clearConfigProject(id)
	if err != nil {
		return err
	}

	return nil
}

// Show project info with the specified project id, returns an error if one occurred
func showProject(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	project, err := client.Photonclient.Projects.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		securityGroups := []string{}
		for _, s := range project.SecurityGroups {
			securityGroups = append(securityGroups, fmt.Sprintf("%s:%t", s.Name, s.Inherited))
		}
		scriptSecurityGroups := strings.Join(securityGroups, ",")
		limits := quotaLineItemListToString(project.ResourceTicket.Limits)
		usages := quotaLineItemListToString(project.ResourceTicket.Usage)

		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", project.ID, project.Name, project.ResourceTicket.TenantTicketID,
			project.ResourceTicket.TenantTicketName, limits, usages, scriptSecurityGroups)
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(project, w, c)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "Project ID: %s\n", project.ID)
		fmt.Fprintf(w, "  Name: %s\n", project.Name)
		fmt.Fprintf(w, "  TenantTicketID: %s\n", project.ResourceTicket.TenantTicketID)
		fmt.Fprintf(w, "    TenantTicketName: %s\n", project.ResourceTicket.TenantTicketName)
		fmt.Fprintf(w, "    Limits:\n")
		for _, l := range project.ResourceTicket.Limits {
			fmt.Fprintf(w, "      %s\t%g\t%s\n", l.Key, l.Value, l.Unit)
		}
		fmt.Fprintf(w, "    Usage:\n")
		for _, u := range project.ResourceTicket.Usage {
			fmt.Fprintf(w, "      %s\t%g\t%s\n", u.Key, u.Value, u.Unit)
		}
		if len(project.SecurityGroups) != 0 {
			fmt.Fprintf(w, "  SecurityGroups:\n")
			for _, s := range project.SecurityGroups {
				fmt.Fprintf(w, "    %s\t%t\n", s.Name, s.Inherited)
			}
		}
		err = w.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}

// Sends a get project task to client based on the config file
// Returns an error if one occurred
func getProject(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
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
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(project, w, c)
	} else {
		fmt.Printf("Current project is '%s' with ID %s\n", project.ID, project.Name)
	}
	return nil
}

// Set project name and id to config file
// Returns an error if one occurred
func setProject(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	name := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
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
func listProjects(c *cli.Context, w io.Writer) error {
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

	projects, err := client.Photonclient.Tenants.GetProjects(tenant.ID, nil)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		for _, t := range projects.Items {
			limits := quotaLineItemListToString(t.ResourceTicket.Limits)
			usage := quotaLineItemListToString(t.ResourceTicket.Usage)
			fmt.Printf("%s\t%s\t%s\t%s\n", t.ID, t.Name, limits, usage)
		}
	} else if utils.NeedsFormatting(c) {
		utils.FormatObjects(projects.Items, w, c)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\tLimit\tUsage\n")
		for _, t := range projects.Items {
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
		fmt.Printf("\nTotal projects: %d\n", len(projects.Items))
	}
	return nil
}

// Retrieves tasks for project
func getProjectTasks(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()
	state := c.String("state")
	kind := c.String("kind")

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	options := &photon.TaskGetOptions{
		State: state,
		Kind:  kind,
	}

	taskList, err := client.Photonclient.Projects.GetTasks(id, options)
	if err != nil {
		return err
	}

	err = printTaskList(taskList.Items, c)
	if err != nil {
		return err
	}

	return nil
}

// Set security groups for a project
func setSecurityGroupsForProject(c *cli.Context) error {
	err := checkArgCount(c, 2)
	if err != nil {
		return err
	}
	id := c.Args().First()
	securityGroups := &photon.SecurityGroupsSpec{
		Items: regexp.MustCompile(`\s*,\s*`).Split(c.Args()[1], -1),
	}
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Photonclient.Projects.SetSecurityGroups(id, securityGroups)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Retrieves IAM Policy for specified project
func getProjectIam(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	policy, err := client.Photonclient.Projects.GetIam(id)
	if err != nil {
		return err
	}

	err = printIamPolicy(*policy, c)
	if err != nil {
		return err
	}

	return nil
}

// Grant or remove a role from a principal on the specified project
func modifyProjectIamPolicy(c *cli.Context, w io.Writer, action string) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	projectID := c.Args()[0]
	principal := c.String("principal")
	role := c.String("role")

	if !c.GlobalIsSet("non-interactive") {
		var err error
		principal, err = askForInput("Principal: ", principal)
		if err != nil {
			return err
		}
	}

	if len(principal) == 0 {
		return fmt.Errorf("Please provide principal")
	}

	if !c.GlobalIsSet("non-interactive") {
		var err error
		role, err = askForInput("Role: ", role)
		if err != nil {
			return err
		}
	}

	if len(role) == 0 {
		return fmt.Errorf("Please provide role")
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	var delta photon.PolicyDelta
	delta = photon.PolicyDelta{Principal: principal, Action: action, Role: role}
	task, err := client.Photonclient.Projects.ModifyIam(projectID, &delta)

	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		policy, err := client.Photonclient.Projects.GetIam(projectID)
		if err != nil {
			return err
		}
		utils.FormatObject(policy, w, c)
	}

	return nil
}
