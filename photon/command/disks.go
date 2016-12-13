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
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for disk
// Subcommands: create; Usage: disk create [<options>]
//              delete; Usage: disk delete <id>
//              show;   Usage: disk show <id>
//              list;   Usage: disk list [<options>]
//              tasks;  Usage: disk tasks <id> [<options>]
func GetDiskCommand() cli.Command {
	command := cli.Command{
		Name: "disk",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new disk",
				ArgsUsage: " ",
				Description: "This creates a new disk on a datastore selected by Photon Controller.\n" +
					"   If you wish to create the disk on the same datastore as a given VM, specify the\n" +
					"   affinities flag. If you wish to create the disk on a particularly type of datastore\n" +
					"   specify the appropriate flavor.",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "disk name",
					},
					cli.StringFlag{
						Name:  "flavor, f",
						Usage: "disk flavor",
					},
					cli.IntFlag{
						Name:  "capacityGB, c",
						Usage: "disk capacity",
					},
					cli.StringFlag{
						Name:  "affinities, a",
						Usage: "Specify affinity to a VM as: vm:VM-ID",
					},
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name",
					},
					cli.StringFlag{
						Name:  "tags, s",
						Usage: "tags for the disk",
					},
					cli.StringFlag{
						Name:  "project, p",
						Usage: "Project name",
					},
				},
				Action: func(c *cli.Context) {
					err := createDisk(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "delete",
				Usage:     "Delete disk with specified ID",
				ArgsUsage: "<disk-id>",
				Action: func(c *cli.Context) {
					err := deleteDisk(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show disk info with specified ID",
				ArgsUsage: "<disk-id>",
				Action: func(c *cli.Context) {
					err := showDisk(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "list",
				Usage:     "List all disks",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name",
					},
					cli.StringFlag{
						Name:  "project, p",
						Usage: "Project name",
					},
					cli.BoolFlag{
						Name:  "summary, s",
						Usage: "Summary view",
					},
					cli.StringFlag{
						Name:  "name, n",
						Usage: "disk name",
					},
				},
				Action: func(c *cli.Context) {
					err := listDisks(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "tasks",
				Usage:     "List all tasks related to a specific disk",
				ArgsUsage: "<disk-id>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "state, s",
						Usage: "specify task state for filtering (QUEUED, STARTED, ERROR, or COMPLETED)",
					},
				},
				Action: func(c *cli.Context) {
					err := getDiskTasks(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Sends a create disk task to client based on the cli.Context
// Returns an error if one occurred
func createDisk(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	name := c.String("name")
	flavor := c.String("flavor")
	capacityGB := c.Int("capacityGB")
	affinities := c.String("affinities")
	tenantName := c.String("tenant")
	projectName := c.String("project")
	tags := c.String("tags")

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	tenant, err := verifyTenant(tenantName)
	if err != nil {
		return err
	}
	project, err := verifyProject(tenant.ID, projectName)
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") {
		name, err = askForInput("Disk name: ", name)
		if err != nil {
			return err
		}
		flavor, err = askForInput("Disk Flavor: ", flavor)
		if err != nil {
			return err
		}
		if !c.IsSet("capacityGB") {
			capacity, err := askForInput("Disk capacity in GB: ", "")
			if err != nil {
				return err
			}
			capacityGB, err = strconv.Atoi(capacity)
			if err != nil {
				return err
			}
		}
	}

	if len(name) == 0 || len(flavor) == 0 {
		return fmt.Errorf("please provide disk name and flavor")
	}

	affinitiesList, err := parseAffinitiesListFromFlag(affinities)
	if err != nil {
		return err
	}

	diskSpec := photon.DiskCreateSpec{}
	diskSpec.Name = name
	diskSpec.Flavor = flavor
	diskSpec.CapacityGB = capacityGB
	diskSpec.Kind = "persistent-disk"
	diskSpec.Affinities = affinitiesList
	diskSpec.Tags = regexp.MustCompile(`\s*,\s*`).Split(tags, -1)

	if !c.GlobalIsSet("non-interactive") && !utils.NeedsFormatting(c) {
		fmt.Printf("\nCreating disk: %s (%s)\n", diskSpec.Name, diskSpec.Flavor)
		fmt.Printf("Tenant: %s, project: %s\n", tenant.Name, project.Name)
	}

	if confirmed(c) {
		createTask, err := client.Photonclient.Projects.CreateDisk(project.ID, &diskSpec)
		if err != nil {
			return err
		}

		diskID, err := waitOnTaskOperation(createTask.ID, c)
		if err != nil {
			return err
		}

		if utils.NeedsFormatting(c) {
			disk, err := client.Photonclient.Disks.Get(diskID)
			if err != nil {
				return err
			}
			utils.FormatObject(disk, w, c)
		}

	} else {
		fmt.Println("OK. Canceled")
	}

	return nil
}

// Sends a delete disk task to client based on the cli.Context
// Returns an error if one occurred
func deleteDisk(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deleteTask, err := client.Photonclient.Disks.Delete(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(deleteTask.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Sends a show disk task to client based on the cli.Context
// Returns an error if one occurred
func showDisk(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	disk, err := client.Photonclient.Disks.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		tag := strings.Trim(fmt.Sprint(disk.Tags), "[]")
		scriptTag := strings.Replace(tag, " ", ",", -1)
		vms := strings.Trim(fmt.Sprint(disk.VMs), "[]")
		scriptVMs := strings.Replace(vms, " ", ",", -1)
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\n", disk.ID, disk.Name,
			disk.State, disk.Kind, disk.Flavor, disk.CapacityGB, disk.Datastore, scriptTag, scriptVMs)
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(disk, w, c)
	} else {
		fmt.Println("Disk ID: ", disk.ID)
		fmt.Println("  Name:       ", disk.Name)
		fmt.Println("  Kind:       ", disk.Kind)
		fmt.Println("  Flavor:     ", disk.Flavor)
		fmt.Println("  CapacityGB: ", disk.CapacityGB)
		fmt.Println("  State:      ", disk.State)
		fmt.Println("  Datastore:  ", disk.Datastore)
		fmt.Println("  Tags:       ", disk.Tags)
		fmt.Println("  VMs:        ", disk.VMs)
	}

	return nil
}

// Retrieves a list of disk, returns an error if one occurred
func listDisks(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	tenantName := c.String("tenant")
	projectName := c.String("project")
	summaryView := c.IsSet("summary")

	name := c.String("name")
	options := &photon.DiskGetOptions{
		Name: name,
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	tenant, err := verifyTenant(tenantName)
	if err != nil {
		return err
	}
	project, err := verifyProject(tenant.ID, projectName)
	if err != nil {
		return err
	}

	diskList, err := client.Photonclient.Projects.GetDisks(project.ID, options)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		utils.FormatObjects(diskList, w, c)
		return nil
	}

	stateCount := make(map[string]int)
	for _, disk := range diskList.Items {
		stateCount[disk.State]++
	}

	if c.GlobalIsSet("non-interactive") {
		if !summaryView {
			for _, disk := range diskList.Items {
				fmt.Printf("%s\t%s\t%s\n", disk.ID, disk.Name, disk.State)
			}
		}
	} else if !utils.NeedsFormatting(c) {
		if !summaryView {
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 4, 4, 2, ' ', 0)
			fmt.Fprintf(w, "ID\tName\tState\n")
			for _, disk := range diskList.Items {
				fmt.Fprintf(w, "%s\t%s\t%s\n", disk.ID, disk.Name, disk.State)
			}
			err := w.Flush()
			if err != nil {
				return err
			}
		}
		fmt.Printf("\nTotal: %d\n", len(diskList.Items))
		for key, value := range stateCount {
			fmt.Printf("%s: %d\n", key, value)
		}
	}

	return nil
}

// Retrieves tasks for disk
func getDiskTasks(c *cli.Context, w io.Writer) error {
	err := checkArgNum(c.Args(), 1, "disk tasks <id> [<options>]")
	if err != nil {
		return err
	}
	id := c.Args().First()
	state := c.String("state")

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	options := &photon.TaskGetOptions{
		State: state,
	}

	taskList, err := client.Photonclient.Disks.GetTasks(id, options)
	if err != nil {
		return err
	}

	if !utils.NeedsFormatting(c) {
		err = printTaskList(taskList.Items, c)
		if err != nil {
			return err
		}
	} else {
		utils.FormatObjects(taskList, w, c)
	}

	return nil
}
