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
	"path"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for images
// Subcommands: create; Usage: image create <path> [<options>]
//              delete; Usage: image delete <id>
//              list;   Usage: image list
//              show;   Usage: image show <id>
//              tasks;  Usage: image tasks <id> [<options>]
//              iam show;  Usage: image iam show <id> [<options>]
//              iam add; Usage: image iam add <id> [<options>]
//              iam remove; Usage: image iam remove <id> [<options>]
func GetImagesCommand() cli.Command {
	command := cli.Command{
		Name:  "image",
		Usage: "options for image",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new image",
				ArgsUsage: "<image-filename>",
				Description: "Upload a new image to Photon Controller.\n" +
					"   If the image replication is EAGER, it will be distributed to all allowed datastores on all ESXi hosts\n" +
					"   If the image replication is ON_DEMAND, it will be distributed to all image datastores",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Image name",
					},
					cli.StringFlag{
						Name:  "image_replication, i",
						Usage: "Image replication type (EAGER or ON_DEMAND)",
					},
				},
				Action: func(c *cli.Context) {
					err := createImage(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "delete",
				Usage:     "Delete an image",
				ArgsUsage: "<image-id>",
				Description: "Delete an image. All copies will be deleted.\n" +
					"   Please note that if the image is in use by one or more VMs, it will not be deleted until\n" +
					"   all VMs that use it are deleted",
				Action: func(c *cli.Context) {
					err := deleteImage(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "list",
				Usage:     "List images",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Image name",
					},
				},
				Action: func(c *cli.Context) {
					err := listImages(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show an image given it's ID",
				ArgsUsage: "<image-id>",
				Action: func(c *cli.Context) {
					err := showImage(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "tasks",
				Usage:     "Show image tasks",
				ArgsUsage: "<image-id>",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "state, s",
						Usage: "Filter by task state",
					},
				},
				Action: func(c *cli.Context) {
					err := getImageTasks(c)
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
						Usage:     "Show the IAM policy associated with an image",
						ArgsUsage: "<image-id>",
						Action: func(c *cli.Context) {
							err := getIam(c)
							if err != nil {
								log.Fatal("Error: ", err)
							}
						},
					},
					{
						Name:      "add",
						Usage:     "Grant a role to a user or group on an image",
						ArgsUsage: "<image-id>",
						Description: "Grant a role to a user or group on an image. \n\n" +
							"   Example: \n" +
							"   photon image iam add <image-id> -p user1@photon.local -r contributor\n" +
							"   photon image iam add <image-id> -p photon.local\\group1 -r viewer",
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
							err := modifyIamPolicy(c, os.Stdout, "ADD")
							if err != nil {
								log.Fatal("Error: ", err)
							}
						},
					},
					{
						Name:      "remove",
						Usage:     "Remove a role from a user or group on an image",
						ArgsUsage: "<image-id>",
						Description: "Remove a role from a user or group on an image. \n\n" +
							"   Example: \n" +
							"   photon image iam remove <image-id> -p user1@photon.local -r contributor \n" +
							"   photon image iam remove <image-id> -p photon.local\\group1 -r viewer",
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
							err := modifyIamPolicy(c, os.Stdout, "REMOVE")
							if err != nil {
								log.Fatal("Error: ", err)
							}
						},
					},
				},
			},
		},
	}
	return command
}

// Create an image
func createImage(c *cli.Context, w io.Writer) error {
	if len(c.Args()) > 1 {
		return fmt.Errorf("Unknown argument: %v", c.Args()[1:])
	}
	filePath := c.Args().First()
	name := c.String("name")
	replicationType := c.String("image_replication")

	if !c.GlobalIsSet("non-interactive") {
		var err error
		filePath, err = askForInput("Image path: ", filePath)
		if err != nil {
			return err
		}
	}

	if len(filePath) == 0 {
		return fmt.Errorf("Please provide image path")
	}

	filePath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	_, err = os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("No such image file at that path")
	}

	if !c.GlobalIsSet("non-interactive") {
		defaultName := path.Base(filePath)
		name, err = askForInput("Image name (default: "+defaultName+"): ", name)
		if err != nil {
			return err
		}

		if len(name) == 0 {
			name = defaultName
		}

		defaultReplication := "EAGER"
		replicationType, err = askForInput("Image replication type (default: "+defaultReplication+"): ", replicationType)
		if err != nil {
			return err
		}
		if len(replicationType) == 0 {
			replicationType = defaultReplication
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	options := &photon.ImageCreateOptions{
		ReplicationType: replicationType,
	}
	if len(replicationType) == 0 {
		options = nil
	}

	uploadTask, err := client.Photonclient.Images.Create(file, name, options)
	if err != nil {
		return err
	}

	imageID, err := waitOnTaskOperation(uploadTask.ID, c)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		image, err := client.Photonclient.Images.Get(imageID)
		if err != nil {
			return err
		}
		utils.FormatObject(image, w, c)
	}

	return nil
}

// Deletes an image by id
func deleteImage(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	if confirmed(c) {
		client.Photonclient, err = client.GetClient(c)
		if err != nil {
			return err
		}

		deleteTask, err := client.Photonclient.Images.Delete(id)
		if err != nil {
			return err
		}

		_, err = waitOnTaskOperation(deleteTask.ID, c)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("OK, canceled")
	}

	return nil
}

// Lists all images
func listImages(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	name := c.String("name")
	options := &photon.ImageGetOptions{
		Name: name,
	}
	images, err := client.Photonclient.Images.GetAll(options)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		for _, image := range images.Items {
			fmt.Printf("%s\t%s\t%s\t%d\t%s\t%s\t%s\n", image.ID, image.Name, image.State, image.Size,
				image.ReplicationType, image.ReplicationProgress, image.SeedingProgress)
		}
	} else if utils.NeedsFormatting(c) {
		utils.FormatObjects(images.Items, w, c)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\tState\tSize(Byte)\tReplication_type\tReplicationProgress\tSeedingProgress\n")
		for _, image := range images.Items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\t%s\n", image.ID, image.Name, image.State, image.Size,
				image.ReplicationType, image.ReplicationProgress, image.SeedingProgress)
		}
		err = w.Flush()
		if err != nil {
			return err
		}
		fmt.Printf("\nTotal: %d\n", len(images.Items))
	}

	return nil
}

// Shows an image based on id
func showImage(c *cli.Context, w io.Writer) error {
	id := c.Args().First()

	if !c.GlobalIsSet("non-interactive") {
		var err error
		id, err = askForInput("Image id: ", id)
		if err != nil {
			return err
		}
	}

	if len(id) == 0 {
		return fmt.Errorf("Please provide image id")
	}

	var err error
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	image, err := client.Photonclient.Images.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		settings := []string{}
		for _, setting := range image.Settings {
			settings = append(settings, fmt.Sprintf("%s:%s", setting.Name, setting.DefaultValue))
		}
		scriptSettings := strings.Join(settings, ",")
		fmt.Printf("%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\n", image.ID, image.Name, image.State, image.Size, image.ReplicationType,
			image.ReplicationProgress, image.SeedingProgress, scriptSettings)

	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(image, w, c)
	} else {
		fmt.Printf("Image ID: %s\n", image.ID)
		fmt.Printf("  Name:                       %s\n", image.Name)
		fmt.Printf("  State:                      %s\n", image.State)
		fmt.Printf("  Size:                       %d Byte(s)\n", image.Size)
		fmt.Printf("  Image Replication Type:     %s\n", image.ReplicationType)
		fmt.Printf("  Image Replication Progress: %s\n", image.ReplicationProgress)
		fmt.Printf("  Image Seeding Progress:     %s\n", image.SeedingProgress)
		fmt.Printf("  Settings: \n")
		for _, setting := range image.Settings {
			fmt.Printf("    %s : %s\n", setting.Name, setting.DefaultValue)
		}
	}

	return nil
}

// Retrieves tasks from specified image
func getImageTasks(c *cli.Context) error {
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

	taskList, err := client.Photonclient.Images.GetTasks(id, options)
	if err != nil {
		return err
	}

	err = printTaskList(taskList.Items, c)
	if err != nil {
		return err
	}

	return nil
}

// Retrieves IAM Policy for specified image
func getIam(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	policy, err := client.Photonclient.Images.GetIam(id)
	if err != nil {
		return err
	}

	err = printIamPolicy(*policy, c)
	if err != nil {
		return err
	}

	return nil
}

// Grant or remove a role from a principal on the specified image
func modifyIamPolicy(c *cli.Context, w io.Writer, action string) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	imageID := c.Args()[0]
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
	task, err := client.Photonclient.Images.ModifyIam(imageID, &delta)

	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		policy, err := client.Photonclient.Images.GetIam(imageID)
		if err != nil {
			return err
		}
		utils.FormatObject(policy, w, c)
	}

	return nil
}
