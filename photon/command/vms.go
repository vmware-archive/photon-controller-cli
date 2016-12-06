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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/vmware/photon-controller-cli/photon/client"

	"regexp"

	"github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/photon/utils"
	"github.com/vmware/photon-controller-go-sdk/photon"
	"io"
)

// Creates a cli.Command for vm
// Subcommands:
//      create;       Usage: vm create [<options>]
//      delete;       Usage: vm delete <id>
//      show;         Usage: vm show <id>
//      list;         Usage: vm list [<options>]
//      tasks;        Usage: vm tasks <id> [<options>]
//      start;        Usage: vm start <id>
//      stop;         Usage: vm stop <id>
//      suspend;      Usage: vm suspend <id>
//      resume;       Usage: vm resume <id>
//      restart;      Usage: vm restart <id>
//      attach-disk;  Usage: vm attach-disk <vm-id> [<options>]
//      detach-disk;  Usage: vm detach-disk <vm-id> [<options>]
//      attach-iso;   Usage: vm attach-iso <id> [<options>]
//      detach-iso;   Usage: vm detach-iso <id> [<options>]
//      set-metadata; Usage: vm set-metadata <id> [<options>]
//      set-tag;      Usage: vm set-tag <id> [<options>]
//      networks;     Usage: vm networks <id>
//      mks-ticket;   Usage: vm mks-ticket <id>
//      create-image; Usage: vm create-image <id> [<options>]
//      aquire-floating-ip; Usage: vm aquare-floating-ip <id> [<options>]
//      release-floating-ip; Usage: vm release-floating-ip <id> [<options>]
func GetVMCommand() cli.Command {
	command := cli.Command{
		Name:  "vm",
		Usage: "options for vm",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new VM",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "VM name",
					},
					cli.StringFlag{
						Name:  "flavor, f",
						Usage: "VM flavor",
					},
					cli.StringFlag{
						Name:  "image, i",
						Usage: "Image ID",
					},
					cli.StringFlag{
						Name:  "disks, d",
						Usage: "VM disks",
					},
					cli.StringFlag{
						Name:  "environment, e",
						Usage: "VM environment({key:value})",
					},
					cli.StringFlag{
						Name:  "affinities, a",
						Usage: "VM Locality(kind id)",
					},
					cli.StringFlag{
						Name:  "networks, w",
						Usage: "VM Networks(id1, id2)",
					},
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name",
					},
					cli.StringFlag{
						Name:  "project, p",
						Usage: "Project name",
					},
				},
				Action: func(c *cli.Context) {
					err := createVM(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "delete",
				Usage:     "Delete VM with specified ID",
				ArgsUsage: "<vm-id>",
				Action: func(c *cli.Context) {
					err := deleteVM(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show VM info with specified ID",
				ArgsUsage: "<vm-id>",
				Action: func(c *cli.Context) {
					err := showVM(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "list",
				Usage:     "List all VMs",
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
						Usage: "VM name",
					},
				},
				Action: func(c *cli.Context) {
					err := listVMs(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "tasks",
				Usage:     "List all tasks related to a VM",
				ArgsUsage: "<vm-id>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "state, s",
						Usage: "specify task state for filtering",
					},
				},
				Action: func(c *cli.Context) {
					err := getVMTasks(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "start",
				Usage:     "Start a VM",
				ArgsUsage: "<vm-id>",
				Action: func(c *cli.Context) {
					err := startVM(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "stop",
				Usage:     "Stop a VM",
				ArgsUsage: "<vm-id>",
				Action: func(c *cli.Context) {
					err := stopVM(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "suspend",
				Usage:     "Suspend a VM",
				ArgsUsage: "<vm-id>",
				Action: func(c *cli.Context) {
					err := suspendVM(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "resume",
				Usage:     "Resume a VM",
				ArgsUsage: "<vm-id>",
				Action: func(c *cli.Context) {
					err := resumeVM(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "restart",
				Usage:     "Restart a VM",
				ArgsUsage: "<vm-id>",
				Action: func(c *cli.Context) {
					err := restartVM(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "attach-disk",
				Usage:     "Attach a disk to a VM",
				ArgsUsage: "<vm-id>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "disk, d",
						Usage: "Disk ID",
					},
				},
				Action: func(c *cli.Context) {
					err := attachDisk(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "detach-disk",
				Usage:     "Detach a disk from a VM",
				ArgsUsage: "<vm-id>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "disk, d",
						Usage: "Disk ID",
					},
				},
				Action: func(c *cli.Context) {
					err := detachDisk(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "attach-iso",
				Usage:     "Attach an ISO to a VM",
				ArgsUsage: "<vm-id>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "path, p",
						Usage: "ISO path",
					},
					cli.StringFlag{
						Name:  "name, n",
						Usage: "ISO name",
					},
				},
				Action: func(c *cli.Context) {
					err := attachIso(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "detach-iso",
				Usage:     "Detach an ISO from a VM",
				ArgsUsage: "<vm-id>",
				Action: func(c *cli.Context) {
					err := detachIso(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "set-metadata",
				Usage:     "Set VM's metadata",
				ArgsUsage: "<vm-id>",
				Description: "This allows you to set the metadata for VM. The metadata is a set of key/value\n" +
					"   pairs associated with a VM. Each key and value must be a string\n" +
					"   Example:\n" +
					"      photon vm set-metadata 8e138ee8-e135-43b2-8453-ebe4cea29f48 -m '{\"k1\":\"v1\", \"k2\":\"v2\"}'",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "metadata, m",
						Usage: "The metadata: a JSON string representing a map of string keys with string values",
					},
				},
				Action: func(c *cli.Context) {
					err := setVMMetadata(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "set-tag",
				Usage:     "Set a tag on a VM",
				ArgsUsage: "<vm-id>",
				Description: "Set a single tag on a VM. A VM can have multiple tags: this will add a single tag.\n" +
					"   Example:\n" +
					"      photon vm set-tag 8e138ee8-e135-43b2-8453-ebe4cea29f48 --tag foo",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "tag, t",
						Usage: "tag (arbitary text)",
					},
				},
				Action: func(c *cli.Context) {
					err := setVMTag(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "networks",
				Usage:     "Show the networks a VM is attached to",
				ArgsUsage: "<vm-id>",
				Action: func(c *cli.Context) {
					err := listVMNetworks(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "mks-ticket",
				Usage:     "Get VM MKS ticket for a VM",
				ArgsUsage: "<vm-id>",
				Action: func(c *cli.Context) {
					err := getVMMksTicket(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "create-image",
				Usage:     "Create an image by cloning VM",
				ArgsUsage: "<vm-id>",
				Description: "Create an image by cloning the VM. The image will be available within Photon Controller.\n" +
					"   as if you had done a 'photon image create' command",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Image name",
					},
					cli.StringFlag{
						Name:  "image_replication, i",
						Usage: "Image replication type",
					},
				},
				Action: func(c *cli.Context) {
					err := createVmImage(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "acquire-floating-ip",
				Usage:     "Acquire a floating IP",
				ArgsUsage: "<vm-id>",
				Description: "Acquire a floating IP and assign it to the given VM. This can be used only when using\n" +
					"   virtual networking. This will allow you to access a VM from outside of its virtual network",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "network_id, i",
						Usage: "Network ID",
					},
				},
				Action: func(c *cli.Context) {
					err := acquireFloatingIp(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:        "release-floating-ip",
				Usage:       "Release floating IP",
				ArgsUsage:   "<vm-id>",
				Description: "Release the floating IP associated with the given VM",
				Action: func(c *cli.Context) {
					err := releaseFloatingIp(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
		},
	}
	return command
}

// Sends a create VM task to client based on the cli.Context
// Returns an error if one occurred
func createVM(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	name := c.String("name")
	flavor := c.String("flavor")
	imageID := c.String("image")
	disks := c.String("disks")
	environment := c.String("environment")
	affinities := c.String("affinities")
	tenantName := c.String("tenant")
	projectName := c.String("project")
	networks := c.String("networks")

	client.Esxclient, err = client.GetClient(c)
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

	disksList, err := parseDisksListFromFlag(disks)
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") && !utils.NeedsFormatting(c) {
		name, err = askForInput("VM name: ", name)
		if err != nil {
			return err
		}
		flavor, err = askForInput("VM Flavor: ", flavor)
		if err != nil {
			return err
		}
		imageID, err = askForInput("Image ID: ", imageID)
		if err != nil {
			return err
		}
		disksList, err = askForVMDiskList(disksList)
		if err != nil {
			return err
		}
	}

	if len(name) == 0 || len(flavor) == 0 || len(imageID) == 0 {
		return fmt.Errorf("Please provide name, flavor and image")
	}

	var environmentMap map[string]string
	if len(environment) != 0 {
		environmentMap, err = parseMapFromFlag(environment)
		if err != nil {
			return err
		}
	}

	affinitiesList, err := parseAffinitiesListFromFlag(affinities)
	if err != nil {
		return err
	}

	var networkList []string
	if len(networks) > 0 {
		networkList = regexp.MustCompile(`\s*,\s*`).Split(networks, -1)
	}

	vmSpec := photon.VmCreateSpec{}
	vmSpec.Name = name
	vmSpec.Flavor = flavor
	vmSpec.SourceImageID = imageID
	vmSpec.AttachedDisks = disksList
	vmSpec.Affinities = affinitiesList
	vmSpec.Environment = environmentMap
	vmSpec.Subnets = networkList

	if !c.GlobalIsSet("non-interactive") && !utils.NeedsFormatting(c) {
		fmt.Printf("\nCreating VM: %s(%s)\n", vmSpec.Name, vmSpec.Flavor)
		fmt.Printf("Source image ID: %s\n\n", vmSpec.SourceImageID)
		fmt.Println("Please make sure disks below are correct:")
		for i, disk := range disksList {
			if disk.BootDisk {
				fmt.Printf("%d: %s, %s, %s\n", i+1, disk.Name, disk.Flavor, "boot")
			} else {
				fmt.Printf("%d: %s, %s, %d GB, %s\n", i+1, disk.Name, disk.Flavor, disk.CapacityGB, "non-boot")
			}
		}
	}

	if confirmed(c) {
		createTask, err := client.Esxclient.Projects.CreateVM(project.ID, &vmSpec)
		if err != nil {
			return err
		}
		vmID, err := waitOnTaskOperation(createTask.ID, c)
		if err != nil {
			return err
		}

		err = formatHelper(c, w, client.Esxclient, vmID)

		return err

	} else {
		fmt.Println("OK. Canceled")
	}

	return nil
}

// Sends a delete VM task to client based on the cli.Context
// Returns an error if one occurred
func deleteVM(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deleteTask, err := client.Esxclient.VMs.Delete(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(deleteTask.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Sends a show VM task to client based on the cli.Context
// Returns an error if one occurred
func showVM(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	vm, err := client.Esxclient.VMs.Get(id)
	if err != nil {
		return err
	}
	var networks []interface{}
	if vm.State != "ERROR" {
		networks, err = getVMNetworks(id, c)
		if err != nil {
			return err
		}
	}

	if c.GlobalIsSet("non-interactive") {
		tag := strings.Trim(fmt.Sprint(vm.Tags), "[]")
		scriptTag := strings.Replace(tag, " ", ",", -1)
		metadata := strings.Trim(strings.TrimLeft(fmt.Sprint(vm.Metadata), "map"), "[]")
		scriptMetadata := strings.Replace(metadata, " ", ",", -1)
		disks := []string{}
		for _, d := range vm.AttachedDisks {
			disks = append(disks, fmt.Sprintf("%s\t%s\t%s\t%s\t%d\t%t", d.ID, d.Name, d.Kind, d.Flavor, d.CapacityGB, d.BootDisk))
		}
		scriptDisks := strings.Join(disks, ",")
		iso := []string{}
		for _, i := range vm.AttachedISOs {
			iso = append(iso, fmt.Sprintf("%s\t%s\t%s\t%d", i.ID, i.Name, i.Kind, i.Size))
		}
		scriptIso := strings.Join(iso, ",")
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", vm.ID, vm.Name, vm.State, vm.Flavor,
			vm.SourceImageID, vm.Host, vm.Datastore, scriptMetadata, scriptTag, vm.FloatingIp)
		fmt.Printf("%s\n", scriptDisks)
		fmt.Printf("%s\n", scriptIso)

		err = printVMNetworks(networks, c.GlobalIsSet("non-interactive"))
		if err != nil {
			return err
		}
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(vm, w, c)
	} else {
		fmt.Println("VM ID: ", vm.ID)
		fmt.Println("  Name:        ", vm.Name)
		fmt.Println("  State:       ", vm.State)
		fmt.Println("  Floating IP: ", vm.FloatingIp)
		fmt.Println("  Flavor:      ", vm.Flavor)
		fmt.Println("  Source Image:", vm.SourceImageID)
		fmt.Println("  Host:        ", vm.Host)
		fmt.Println("  Datastore:   ", vm.Datastore)
		fmt.Println("  Metadata:    ", vm.Metadata)
		fmt.Println("  Disks:       ")
		for i, d := range vm.AttachedDisks {
			fmt.Printf("    Disk %d:\n", i+1)
			fmt.Println("      ID:       ", d.ID)
			fmt.Println("      Name:     ", d.Name)
			fmt.Println("      Kind:     ", d.Kind)
			fmt.Println("      Flavor:   ", d.Flavor)
			fmt.Println("      Capacity: ", d.CapacityGB)
			fmt.Println("      Boot:     ", d.BootDisk)
		}
		for i, iso := range vm.AttachedISOs {
			fmt.Printf("    ISO %d:\n", i+1)
			fmt.Println("      Name: ", iso.Name)
			fmt.Println("      Size: ", iso.Size)
		}
		for i, nt := range networks {
			network := nt.(map[string]interface{})
			fmt.Printf("    Networks: %d\n", i+1)
			networkName := ""
			ipAddr := ""
			if val, ok := network["network"]; ok && val != nil {
				networkName = val.(string)
			}
			if val, ok := network["ipAddress"]; ok && val != nil {
				ipAddr = val.(string)
			}
			fmt.Println("      Name:       ", networkName)
			fmt.Println("      IP Address: ", ipAddr)
		}
		for i, tag := range vm.Tags {
			fmt.Printf("    Tag %d:\n", i+1)
			fmt.Println("      Tag Info:     ", tag)
		}
	}

	return nil
}

// Retrieves a list of VMs, returns an error if one occurred
func listVMs(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	tenantName := c.String("tenant")
	projectName := c.String("project")
	summaryView := c.IsSet("summary")

	name := c.String("name")
	options := &photon.VmGetOptions{
		Name: name,
	}

	client.Esxclient, err = client.GetClient(c)
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

	vmList, err := client.Esxclient.Projects.GetVMs(project.ID, options)
	if err != nil {
		return err
	}

	if !utils.NeedsFormatting(c) {
		err = printVMList(vmList.Items, os.Stdout, c, summaryView)
	} else {
		utils.FormatObjects(vmList, w, c)
	}
	if err != nil {
		return err
	}

	return nil
}

// Retrieves tasks for VM
func getVMTasks(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()
	state := c.String("state")

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	options := &photon.TaskGetOptions{
		State: state,
	}

	taskList, err := client.Esxclient.VMs.GetTasks(id, options)
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

func startVM(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	opTask, err := client.Esxclient.VMs.Start(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(opTask.ID, c)
	if err != nil {
		return err
	}

	err = formatHelper(c, w, client.Esxclient, id)

	return err
}

func stopVM(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	opTask, err := client.Esxclient.VMs.Stop(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(opTask.ID, c)
	if err != nil {
		return err
	}

	err = formatHelper(c, w, client.Esxclient, id)

	return err
}

func suspendVM(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	opTask, err := client.Esxclient.VMs.Suspend(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(opTask.ID, c)
	if err != nil {
		return err
	}

	err = formatHelper(c, w, client.Esxclient, id)

	return err
}

func resumeVM(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	opTask, err := client.Esxclient.VMs.Resume(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(opTask.ID, c)
	if err != nil {
		return err
	}

	err = formatHelper(c, w, client.Esxclient, id)

	return err
}

func restartVM(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	opTask, err := client.Esxclient.VMs.Restart(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(opTask.ID, c)
	if err != nil {
		return err
	}

	err = formatHelper(c, w, client.Esxclient, id)

	return err
}

func attachDisk(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()
	diskID := c.String("disk")

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	operation := &photon.VmDiskOperation{
		DiskID: diskID,
	}

	task, err := client.Esxclient.VMs.AttachDisk(id, operation)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = formatHelper(c, w, client.Esxclient, id)

	return err
}

func detachDisk(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()
	diskID := c.String("disk")

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	operation := &photon.VmDiskOperation{
		DiskID: diskID,
	}

	task, err := client.Esxclient.VMs.DetachDisk(id, operation)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = formatHelper(c, w, client.Esxclient, id)

	return err
}

func attachIso(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()
	path := c.String("path")
	name := c.String("name")

	if !c.GlobalIsSet("non-interactive") {
		path, err = askForInput("Iso path: ", path)
		if err != nil {
			return err
		}
		name, err = askForInput("ISO name (default: "+filepath.Base(path)+"): ", name)
		if err != nil {
			return err
		}
	}

	if len(path) == 0 {
		return fmt.Errorf("Please provide iso path")
	}
	if len(name) == 0 {
		name = filepath.Base(path)
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Esxclient.VMs.AttachISO(id, file, name)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = formatHelper(c, w, client.Esxclient, id)

	return err
}

func detachIso(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Esxclient.VMs.DetachISO(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = formatHelper(c, w, client.Esxclient, id)

	return err
}

func setVMMetadata(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	metadata := c.String("metadata")
	vmMetadata := &photon.VmMetadata{}

	if len(metadata) == 0 {
		return fmt.Errorf("Please provide metadata")
	} else {
		var data map[string]string
		err := json.Unmarshal([]byte(metadata), &data)
		if err != nil {
			return err
		}
		vmMetadata.Metadata = data
	}

	task, err := client.Esxclient.VMs.SetMetadata(id, vmMetadata)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = formatHelper(c, w, client.Esxclient, id)

	return err
}

func listVMNetworks(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	networks, err := getVMNetworks(id, c)
	if err != nil {
		return err
	}

	if !utils.NeedsFormatting(c) {
		err = printVMNetworks(networks, c.GlobalIsSet("non-interactive"))
		if err != nil {
			return err
		}
	} else {
		utils.FormatObjects(networks, w, c)
	}
	return nil
}

func setVMTag(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()

	tag := c.String("tag")
	vmTag := &photon.VmTag{}

	if len(tag) == 0 {
		return fmt.Errorf("Please input a tag")
	}
	vmTag.Tag = tag

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Esxclient.VMs.SetTag(id, vmTag)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = formatHelper(c, w, client.Esxclient, id)

	return err
}

func getVMMksTicket(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Esxclient.VMs.GetMKSTicket(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		task, err := client.Esxclient.Tasks.Wait(task.ID)
		if err != nil {
			return err
		}
		mksTicket := task.ResourceProperties.(map[string]interface{})
		fmt.Printf("%s\t%v\n", task.Entity.ID, mksTicket["ticket"])
	} else if utils.NeedsFormatting(c) {
		task, err := client.Esxclient.Tasks.Wait(task.ID)
		if err != nil {
			return err
		}
		mksTicket := task.ResourceProperties.(map[string]interface{})
		utils.FormatObject(mksTicket, w, c)
	} else {
		task, err = pollTask(task.ID)
		if err != nil {
			return err
		}
		mksTicket := task.ResourceProperties.(map[string]interface{})
		fmt.Printf("VM ID: %s \nMks ticket ID is %v\n", task.Entity.ID, mksTicket["ticket"])
	}
	return nil
}

func createVmImage(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()

	name := c.String("name")
	replicationType := c.String("image_replication")

	if !c.GlobalIsSet("non-interactive") && !utils.NeedsFormatting(c) {
		defaultName := "image-from-vm-" + id
		defaultReplication := "EAGER"

		name, err = askForInput("Image name (default: "+defaultName+"): ", name)
		if err != nil {
			return err
		}

		replicationType, err = askForInput("Image replication type (default: "+defaultReplication+"): ", replicationType)
		if err != nil {
			return err
		}

		if len(name) == 0 {
			name = defaultName
		}
		if len(replicationType) == 0 {
			replicationType = defaultReplication
		}
	}

	options := &photon.ImageCreateSpec{
		Name:            name,
		ReplicationType: replicationType,
	}

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Esxclient.VMs.CreateImage(id, options)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = formatHelper(c, w, client.Esxclient, id)

	return err
}

func acquireFloatingIp(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()
	networkId := c.String("network_id")

	options := &photon.VmFloatingIpSpec{
		NetworkId: networkId,
	}

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Esxclient.VMs.AcquireFloatingIp(id, options)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = formatHelper(c, w, client.Esxclient, id)

	return err
}

func releaseFloatingIp(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Esxclient.VMs.ReleaseFloatingIp(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = formatHelper(c, w, client.Esxclient, id)

	return err
}

func formatHelper(c *cli.Context, w io.Writer, client *photon.Client, id string) error {
	if utils.NeedsFormatting(c) {
		vm, err := client.VMs.Get(id)
		if err != nil {
			return err
		}
		utils.FormatObject(vm, w, c)
	}
	return nil
}
