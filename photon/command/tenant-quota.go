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

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"errors"
	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for tenant quota
// Subcommands: set;      Usage: tenant quota set     <tenant-name> <options>
//              show;     Usage: tenant quota get     <tenant-name>
//              update;   Usage: tenant quota update  <tenant-name> <options>
//              exclude;  Usage: tenant quota exclude <tenant-name> <options>
func getTenantQuotaCommand() cli.Command {
	quotaCommand := cli.Command{
		Name:  "quota",
		Usage: "options for tenant quota",
		Subcommands: []cli.Command{
			{
				Name:      "set",
				Usage:     "Set quota for the tenant with specified limits.",
				ArgsUsage: "<tenant-name>",
				Description: "Set a quota for tenant. It will overwrite the whole existing quota.\n" +
					"   Only system administrators can operate on tenant quota.\n" +
					"   A quota is defined by a set of maximum resource costs. Each usage has a type (e.g. vmMemory),\n\n" +
					"   a numnber (e.g. 1) and a unit (e.g. GB). You must specify at least one cost\n" +
					"   Valid units:  GB, MB, KB, B, or COUNT\n" +
					"   Common costs:\n" +
					"     vm.count:      Total number of VMs (use with COUNT)\n" +
					"     vm.cpu:        Total number of vCPUs for a VM (use with COUNT)\n" +
					"     vm.memory:     Total amount of RAM for a VM (use with GB, MB, KB, or B)\n" +
					"     disk.capacity: Total disk capacity (use with GB, MB, KB, or B)\n" +
					"     disk.count:    Number of disks (use with COUNT)\n" +
					"   Example: set tenant quota with 100 VMs, 1000 GB of RAM and 500 vCPUs:\n" +
					"      photon tenant quota set tenant1 \\\n" +
					"             --limits 'vm.count 100 COUNT, vm.memory 1000 GB, vm.cpu 500 COUNT'\n",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "limits, l",
						Usage: "Quota limits(key value unit)",
					},
				},
				Action: func(c *cli.Context) {
					err := setTenantQuota(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show quota info for specified tenant.",
				ArgsUsage: "<tenant-name>",
				Description: "Show a quota for tenant.\n" +
					"   Example:\n" +
					"      photon tenant quota show tenant1 \n",
				Flags: []cli.Flag{},
				Action: func(c *cli.Context) {
					err := getTenantQuota(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "update",
				Usage:     "Update the tenant quota with the specific limits",
				ArgsUsage: "<tenant-name>",
				Description: "Update the quota for tenant. It updates existing quota with the quota items provided.\n" +
					"   Example:\n" +
					"      photon tenant quota update tenant1 \\\n" +
					"             --limits 'vm.count 10 COUNT'\n",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "limits, l",
						Usage: "Quota limits(key value unit)",
					},
				},
				Action: func(c *cli.Context) {
					err := updateTenantQuota(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "exclude",
				Usage:     "Exclude tenant quota items with specific quota item keys.",
				ArgsUsage: "<tenant-name>",
				Description: "Exclude the specified quota items from the tenant quota.\n" +
					"   Example:\n" +
					"      photon tenant quota exclude tenant1 \\\n" +
					"             --limits 'vm.count 10 COUNT'\n",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "limits, l",
						Usage: "Quota limits(key value unit)",
					},
				},
				Action: func(c *cli.Context) {
					err := excludeTenantQuota(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
		},
	}
	return quotaCommand
}

// Get the Quota info for the specified Tenant.
// Returns an error if one occurred
func getTenantQuota(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	tenantName := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	tenant, err := verifyTenant(tenantName)
	if err != nil {
		return err
	}

	quota, err := client.Photonclient.Tenants.GetQuota(tenant.ID)
	if err != nil {
		return err
	}

	for k, v := range quota.QuotaLineItems {
		fmt.Printf("%s\t%g\t%g\t%s\n", k, v.Limit, v.Usage, v.Unit)
	}

	return nil
}

// Set (replace) the whole tenant quota with the quota line items specified in limits flag.
func setTenantQuota(c *cli.Context, w io.Writer) error {
	err := modifyTenantQuota(c, w, "setQuota")
	return err
}

// Update portion of the tenant quota with the quota line items specified in limits flag.
func updateTenantQuota(c *cli.Context, w io.Writer) error {
	err := modifyTenantQuota(c, w, "updateQuota")
	return err
}

// Exclude tenant quota line items from the specific quota line items specified in limits flag.
func excludeTenantQuota(c *cli.Context, w io.Writer) error {
	err := modifyTenantQuota(c, w, "excludeQuota")
	return err
}

// The common function for performing different type of modification on Quota.
func modifyTenantQuota(c *cli.Context, w io.Writer, operation string) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	tenantName := c.Args().First()
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
		limitsList, err = askForLimitList(limitsList)
		if err != nil {
			return err
		}
	}

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("\nTenant name: %s\n", tenant.Name)
		fmt.Println("Please make sure limits below are correct:")
		for i, l := range limitsList {
			fmt.Printf("%d: %s, %g, %s\n", i+1, l.Key, l.Value, l.Unit)
		}
	}

	if confirmed(c) {
		var createTask *photon.Task
		var err error

		// convert to QuotaSpec
		quotaSpec := convertQuotaSpecFromQuotaLineItems(limitsList)

		switch operation {
		case "setQuota":
			createTask, err = client.Photonclient.Tenants.SetQuota(tenant.ID, &quotaSpec)
		case "updateQuota":
			createTask, err = client.Photonclient.Tenants.UpdateQuota(tenant.ID, &quotaSpec)
		case "excludeQuota":
			createTask, err = client.Photonclient.Tenants.ExcludeQuota(tenant.ID, &quotaSpec)
		default:
			err = errors.New("Unsupported Quota operation: " + operation)
		}

		if err != nil {
			return err
		}

		_, err = waitOnTaskOperation(createTask.ID, c)
		if err != nil {
			return err
		}

		if utils.NeedsFormatting(c) {
			quota, err := client.Photonclient.Tenants.GetQuota(tenant.ID)
			if err != nil {
				return err
			}
			utils.FormatObject(quota, w, c)
		}
	} else {
		fmt.Println("OK. Canceled")
	}

	return nil
}
