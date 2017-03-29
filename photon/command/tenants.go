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

// Creates a cli.Command for tenant
// Subcommands: create; Usage: tenant create <name> [<options>]
//              delete; Usage: tenant delete <id>
//              show;   Usage: tenant show <id>
//              list;   Usage: tenant list
//              set;    Usage: tenant set <name>
//              get;    Usage: tenant get
//              tasks;  Usage: tenant tasks <id> [<options>]
//              quota;  Usage: tenant quota <operation> <name> [<options>]
func GetTenantsCommand() cli.Command {
	command := cli.Command{
		Name:  "tenant",
		Usage: "options for tenant",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new tenant",
				ArgsUsage: "<tenant-name>",
				Description: "Create a tenant. Only system administrators can create new tenant.\n" +
					"   A quota for the tenant can be defined during tenant creation and " +
					"   it is defined by a set of maximum resource costs. Each usage has a type,\n" +
					"   a numnber (e.g. 1) and a unit (e.g. GB). You must specify at least one cost\n" +
					"   Valid units:  GB, MB, KB, B, or COUNT\n" +
					"   Common costs:\n" +
					"     vm.count:            Total number of VMs (use with COUNT)\n" +
					"     vm.cpu:              Total number of vCPUs for a VM (use with COUNT)\n" +
					"     vm.memory:           Total amount of RAM for a VM (use with GB, MB, KB, or B)\n" +
					"     disk.capacity:       Total disk capacity (use with GB, MB, KB, or B)\n" +
					"     disk.count:          Number of disks (use with COUNT)\n" +
					"     sdn.floatingip.size: Number of floating ip \n" +
					"   Example: set tenant quota with 100 VMs, 1000 GB of RAM and 500 vCPUs:\n" +
					"      photon tenant create tenant1 \\\n" +
					"             --limits 'vm.count 100 COUNT, vm.memory 1000 GB, vm.cpu 500 COUNT'\n",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "security-groups, s",
						Usage: "Comma-separated Lightwave group names, to specify the tenant administrators",
					},
					cli.StringFlag{
						Name:  "limits, l",
						Usage: "Tenant limits (key value unit)",
					},
				},
				Action: func(c *cli.Context) {
					err := createTenant(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "delete",
				Usage:     "Delete a tenant",
				ArgsUsage: "<tenant-id>",
				Action: func(c *cli.Context) {
					err := deleteTenant(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show detailed tenant info with specified id",
				ArgsUsage: "<tenant-id>",
				Action: func(c *cli.Context) {
					err := showTenant(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "list",
				Usage:     "List all tenants",
				ArgsUsage: " ",
				Action: func(c *cli.Context) {
					err := listTenants(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "set",
				Usage:     "Set default tenant",
				ArgsUsage: "<tenant-name>",
				Description: "Set the default project that will be used for all photon CLI commands that need a project.\n" +
					"   Most commands allow you to override the default.",
				Action: func(c *cli.Context) {
					err := setTenant(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "get",
				Usage:     "Get default tenant",
				ArgsUsage: " ",
				Description: "Show default project in use for photon CLI commands. Most command allow you to either\n" +
					"   use this default or specify a specific project to use.",
				Action: func(c *cli.Context) {
					err := getTenant(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "tasks",
				Usage:     "Show tenant tasks",
				ArgsUsage: "<tenant-id>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "state, s",
						Usage: "Filter by task sate",
					},
				},
				Action: func(c *cli.Context) {
					err := getTenantTasks(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "set-security-groups",
				Usage:     "Set security groups for a tenant",
				ArgsUsage: "<tenant-id> <comma separated list of groups>",
				Description: "Set the list of Lightwave groups that can administer this tenant. This may only be\n" +
					"   be set by a member of the tenant. Be cautious--you can remove your own access if you specify\n" +
					"   the wrong set of groups.\n\n" +
					"   A security group specifies both the Lightwave domain and Lightwave group.\n" +
					"   For example, a security group may be photon.vmware.com\\group-1\n\n" +
					"   Example: photon tenant 10323808-7b07-49f7-9e72-b5ee2af768ad set-security-groups 'photon.vmware.com\\group-1,photon.vmware.com\\group-2'",
				Action: func(c *cli.Context) {
					err := setSecurityGroups(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Hidden:      true,
				Name:        "set_security_groups",
				Usage:       "Set security groups for a tenant",
				ArgsUsage:   "<tenant-id> <comma separated list of groups>",
				Description: "Deprecated, use set-security-groups instead",
				Action: func(c *cli.Context) {
					err := setSecurityGroups(c)
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
						Usage:     "Show the IAM policy associated with a tenant",
						ArgsUsage: "<tenant-id>",
						Action: func(c *cli.Context) {
							err := getTenantIam(c)
							if err != nil {
								log.Fatal("Error: ", err)
							}
						},
					},
					{
						Name:      "add",
						Usage:     "Grant a role to a user or group on a tenant",
						ArgsUsage: "<tenant-id>",
						Description: "Grant a role to a user or group on a tenant. \n\n" +
							"   Example: \n" +
							"   photon tenant iam add <tenant-id> -p user1@photon.local -r contributor\n" +
							"   photon tenant iam add <tenant-id> -p photon.local\\group1 -r viewer",
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
							err := modifyTenantIamPolicy(c, os.Stdout, "ADD")
							if err != nil {
								log.Fatal("Error: ", err)
							}
						},
					},
					{
						Name:      "remove",
						Usage:     "Remove a role from a user or group on a tenant",
						ArgsUsage: "<tenant-id>",
						Description: "Remove a role from a user or group on a tenant. \n\n" +
							"   Example: \n" +
							"   photon tenant iam remove <tenant-id> -p user1@photon.local -r contributor \n" +
							"   photon tenant iam remove <tenant-id> -p photon.local\\group1 -r viewer",
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
							err := modifyTenantIamPolicy(c, os.Stdout, "REMOVE")
							if err != nil {
								log.Fatal("Error: ", err)
							}
						},
					},
				},
			},
			// Load Tenant Quota related logic from separated file.
			getTenantQuotaCommand(),
		},
	}
	return command
}

// Sends a create tenant task to client based on the cli.Context
// Returns an error if one occurred
func createTenant(c *cli.Context, w io.Writer) error {
	if len(c.Args()) > 1 {
		return fmt.Errorf("Unknown argument: %v", c.Args()[1:])
	}
	name := c.Args().First()
	securityGroups := c.String("security-groups")
	limits := c.String("limits")

	if !c.GlobalIsSet("non-interactive") {
		var err error
		name, err = askForInput("Tenant name: ", name)
		if err != nil {
			return err
		}
		securityGroups, err =
			askForInput("Comma-separated security group names, or hit enter for no security groups): ",
				securityGroups)
		if err != nil {
			return err
		}
	}

	if len(name) == 0 {
		return fmt.Errorf("Please provide tenant name")
	}
	securityGroupList := []string{}
	if securityGroups != "" {
		securityGroupList = regexp.MustCompile(`\s*,\s*`).Split(securityGroups, -1)
	}

	// Get project quota if present
	quota := photon.Quota{}
	if c.IsSet("limits") {
		limitsList, err := parseLimitsListFromFlag(limits)
		if err != nil {
			return err
		}

		quotaSpec := convertQuotaSpecFromQuotaLineItems(limitsList)
		quota.QuotaLineItems = quotaSpec
	}

	tenantSpec := &photon.TenantCreateSpec{
		Name:           name,
		SecurityGroups: securityGroupList,
		ResourceQuota:  quota,
	}

	var err error
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	createTask, err := client.Photonclient.Tenants.Create(tenantSpec)

	if err != nil {
		return err
	}

	id, err := waitOnTaskOperation(createTask.ID, c)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		tenant, err := client.Photonclient.Tenants.Get(id)
		if err != nil {
			return err
		}
		utils.FormatObject(tenant, w, c)
	}

	return nil
}

// Retrieves a list of tenants, returns an error if one occurred
func listTenants(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	tenants, err := client.Photonclient.Tenants.GetAll()
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		for _, tenant := range tenants.Items {
			quotaString := quotaSpecToString(tenant.ResourceQuota.QuotaLineItems)
			fmt.Printf("%s\t%s\t%s\n", tenant.ID, tenant.Name, quotaString)
		}
	} else if utils.NeedsFormatting(c) {
		utils.FormatObjects(tenants.Items, w, c)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\n")
		for _, tenant := range tenants.Items {
			fmt.Fprintf(w, "%s\t%s\n", tenant.ID, tenant.Name)
			fmt.Fprintf(w, "    Limits:\n")
			for k, l := range tenant.ResourceQuota.QuotaLineItems {
				fmt.Fprintf(w, "      %s\t%g\t%s\n", k, l.Limit, l.Unit)
			}
			fmt.Fprintf(w, "    Usage:\n")
			for k, u := range tenant.ResourceQuota.QuotaLineItems {
				fmt.Fprintf(w, "      %s\t%g\t%s\n", k, u.Usage, u.Unit)
			}
		}
		err = w.Flush()
		if err != nil {
			return err
		}
		fmt.Printf("\nTotal: %d\n", len(tenants.Items))
	}

	return nil
}

// Sends a delete tenant task to client based on the cli.Context
// Returns an error if one occurred
func deleteTenant(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deleteTask, err := client.Photonclient.Tenants.Delete(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(deleteTask.ID, c)
	if err != nil {
		return err
	}

	err = clearConfigTenant(id)
	if err != nil {
		return err
	}

	return nil
}

func showTenant(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	tenant, err := client.Photonclient.Tenants.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		securityGroups := []string{}
		for _, s := range tenant.SecurityGroups {
			securityGroups = append(securityGroups, fmt.Sprintf("%s:%t", s.Name, s.Inherited))
		}
		scriptSecurityGroups := strings.Join(securityGroups, ",")
		quotaString := quotaSpecToString(tenant.ResourceQuota.QuotaLineItems)
		fmt.Printf("%s\t%s\t%s\t%s\n", tenant.ID, tenant.Name, scriptSecurityGroups, quotaString)
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(tenant, w, c)
	} else {
		fmt.Println("Tenant ID: ", tenant.ID)
		fmt.Println("  Name:              ", tenant.Name)
		for i, s := range tenant.SecurityGroups {
			fmt.Printf("    SecurityGroups %d:\n", i+1)
			fmt.Println("      Name:          ", s.Name)
			fmt.Println("      Inherited:     ", s.Inherited)
		}

		fmt.Fprintf(w, "    Limits:\n")
		for k, l := range tenant.ResourceQuota.QuotaLineItems {
			fmt.Fprintf(w, "      %s\t%g\t%s\n", k, l.Limit, l.Unit)
		}
		fmt.Fprintf(w, "    Usage:\n")
		for k, u := range tenant.ResourceQuota.QuotaLineItems {
			fmt.Fprintf(w, "      %s\t%g\t%s\n", k, u.Usage, u.Unit)
		}
	}

	return nil
}

// Overwrites the tenant in the config file
func setTenant(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	name := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	// Ensure tenant exists
	id, err := findTenantID(name)
	if len(id) == 0 || err != nil {
		return err
	}

	config, err := cf.LoadConfig()
	if err != nil {
		return err
	}

	config.Tenant = &cf.TenantConfiguration{Name: name, ID: id}
	err = cf.SaveConfig(config)
	if err != nil {
		return err
	}

	err = clearConfigProject("")
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("Tenant set to '%s'\n", name)
	}
	return nil
}

// Outputs the set tenant otherwise informs user it is not set
func getTenant(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	config, err := cf.LoadConfig()
	if err != nil {
		return err
	}

	tenant := config.Tenant
	if tenant == nil {
		fmt.Printf("No tenant selected\n")
	} else {
		if c.GlobalIsSet("non-interactive") {
			fmt.Printf("%s\t%s\n", tenant.ID, tenant.Name)
		} else if utils.NeedsFormatting(c) {
			utils.FormatObject(tenant, w, c)
		} else {
			fmt.Printf("Current tenant is '%s' with ID %s\n", tenant.Name, tenant.ID)
		}
	}
	return nil
}

// Retrieves tasks from specified tenant
func getTenantTasks(c *cli.Context, w io.Writer) error {
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

	taskList, err := client.Photonclient.Tenants.GetTasks(id, options)
	if err != nil {
		return err
	}

	err = printTaskList(taskList.Items, c)
	if err != nil {
		return err
	}
	return nil
}

// Set security groups for a tenant
func setSecurityGroups(c *cli.Context) error {
	err := checkArgCount(c, 2)
	if err != nil {
		return err
	}
	id := c.Args().First()
	items := []string{}
	if c.Args()[1] != "" {
		items = regexp.MustCompile(`\s*,\s*`).Split(c.Args()[1], -1)
	}
	securityGroups := &photon.SecurityGroupsSpec{
		Items: items,
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Photonclient.Tenants.SetSecurityGroups(id, securityGroups)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Retrieves IAM Policy for specified tenant
func getTenantIam(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	policy, err := client.Photonclient.Tenants.GetIam(id)
	if err != nil {
		return err
	}

	err = printIamPolicy(*policy, c)
	if err != nil {
		return err
	}

	return nil
}

// Grant or remove a role from a principal on the specified tenant
func modifyTenantIamPolicy(c *cli.Context, w io.Writer, action string) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	tenantID := c.Args()[0]
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
	task, err := client.Photonclient.Tenants.ModifyIam(tenantID, &delta)

	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		policy, err := client.Photonclient.Tenants.GetIam(tenantID)
		if err != nil {
			return err
		}
		utils.FormatObject(policy, w, c)
	}

	return nil
}
