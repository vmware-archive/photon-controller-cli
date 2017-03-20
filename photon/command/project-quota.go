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

// Creates a cli.Command for project quota
// Subcommands: set;      Usage: project-quota set     <project-id> <options>
//              show;     Usage: project-quota get     <project-id>
//              update;   Usage: project-quota update  <project-id> <options>
//              exclude;  Usage: project-quota exclude <project-id> <options>
func getProjectQuotaCommand() cli.Command {
	quotaCommand := cli.Command{
		Name:  "quota",
		Usage: "options for project quota",
		Subcommands: []cli.Command{
			{
				Name:      "set",
				Usage:     "Set quota for the project with specified limits.",
				ArgsUsage: "<project-id>",
				Description: "Set a quota for project. It will overwrite the whole existing quota.\n" +
					"   Only tenant administrators can operate on project quota.\n" +
					"   A quota is defined by a set of maximum resource costs. Each usage has a type (e.g. vmMemory),\n\n" +
					"   a numnber (e.g. 1) and a unit (e.g. GB). You must specify at least one cost\n" +
					"   Valid units:  GB, MB, KB, B, or COUNT\n" +
					"   Common costs:\n" +
					"     vm.count:       Total number of VMs (use with COUNT)\n" +
					"     vm.cpu:         Total number of vCPUs for a VM (use with COUNT)\n" +
					"     vm.memory:      Total amount of RAM for a VM (use with GB, MB, KB, or B)\n" +
					"     disk.capacity:  Total disk capacity (use with GB, MB, KB, or B)\n" +
					"     disk.count:     Number of disks (use with COUNT)\n" +
					"   Example: set project quota with 100 VMs, 1000 GB of RAM and 500 vCPUs:\n" +
					"      photon project quota set projectid1 \\\n" +
					"             --limits 'vm.count 100 COUNT, vmMemory 1000 GB, vmCpu 500 COUNT'\n",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "limits, l",
						Usage: "Quota limits(key value unit)",
					},
				},
				Action: func(c *cli.Context) {
					err := setProjectQuota(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show quota info for specified project.",
				ArgsUsage: "<project-id>",
				Description: "Show a quota for project.\n" +
					"   Example:\n" +
					"      photon project quota show projectid1 \n",
				Flags: []cli.Flag{},
				Action: func(c *cli.Context) {
					err := getProjectQuota(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "update",
				Usage:     "Update the project quota with the specific limits",
				ArgsUsage: "<project-id>",
				Description: "Update the quota for project. It updates existing quota with the quota items provided.\n" +
					"   Example:\n" +
					"      photon project quota update projectid1 \\\n" +
					"             --limits 'vm.count 10 COUNT'\n",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "limits, l",
						Usage: "Quota limits(key value unit)",
					},
				},
				Action: func(c *cli.Context) {
					err := updateProjectQuota(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "exclude",
				Usage:     "Exclude project quota items with specific quota item keys.",
				ArgsUsage: "<project-id>",
				Description: "Exclude the specified quota items from the project quota.\n" +
					"   Example:\n" +
					"      photon project quota exclude --project projectid1 \\\n" +
					"             --limits 'vm.count 10 COUNT'\n",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "limits, l",
						Usage: "Quota limits(key value unit)",
					},
				},
				Action: func(c *cli.Context) {
					err := excludeProjectQuota(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
		},
	}
	return quotaCommand
}

// Get the Quota info for the specified Project.
// Returns an error if one occurred
func getProjectQuota(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	projectId := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	project, err := client.Photonclient.Projects.Get(projectId)
	if err != nil {
		return err
	}

	quota, err := client.Photonclient.Projects.GetQuota(project.ID)
	if err != nil {
		return err
	}

	for k, v := range quota.QuotaLineItems {
		fmt.Printf("%s\t%g\t%g\t%s\n", k, v.Limit, v.Usage, v.Unit)
	}

	return nil
}

// Set (replace) the whole project quota with the quota line items specified in limits flag.
func setProjectQuota(c *cli.Context, w io.Writer) error {
	err := modifyProjectQuota(c, w, "setQuota")
	return err
}

// Update portion of the project quota with the quota line items specified in limits flag.
func updateProjectQuota(c *cli.Context, w io.Writer) error {
	err := modifyProjectQuota(c, w, "updateQuota")
	return err
}

// Exclude project quota line items from the specific quota line items specified in limits flag.
func excludeProjectQuota(c *cli.Context, w io.Writer) error {
	err := modifyProjectQuota(c, w, "excludeQuota")
	return err
}

// The common function for performing different type of modification on Quota.
func modifyProjectQuota(c *cli.Context, w io.Writer, operation string) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	projectId := c.Args().First()
	limits := c.String("limits")

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	project, err := client.Photonclient.Projects.Get(projectId)
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
		fmt.Printf("\nProject Id: %s\n", project.ID)
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
			createTask, err = client.Photonclient.Projects.SetQuota(project.ID, &quotaSpec)
		case "updateQuota":
			createTask, err = client.Photonclient.Projects.UpdateQuota(project.ID, &quotaSpec)
		case "excludeQuota":
			createTask, err = client.Photonclient.Projects.ExcludeQuota(project.ID, &quotaSpec)
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
			quota, err := client.Photonclient.Projects.GetQuota(project.ID)
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
