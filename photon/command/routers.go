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

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for router
// Subcommands: create;  Usage: router create [<options>]
//              delete;  Usage: router delete <id>
//              show;    Usage: router show <id>
//              update;  Usage: router update <id> [<options>]
func GetRoutersCommand() cli.Command {
	command := cli.Command{
		Name:  "router",
		Usage: "options for router",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new router",
				ArgsUsage: " ",
				Description: "Create a new router within a project. Subnets can be created under this router. \n" +
					"   The private IP range of router will be sub-divided into smaller CIDRs for each subnet \n" +
					"   created under this router \n\n" +
					"   Example: \n" +
					"   photon router create -n router-1 -i 192.168.0.0/16 -t cloud-dev -p cloud-dev-staging \\ \n",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Router name",
					},
					cli.StringFlag{
						Name:  "privateIpCidr, i",
						Usage: "The private IP range of router in CIDR format, e.g.: 192.168.0.0/16",
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
					err := createRouter(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:        "delete",
				Usage:       "Delete router with specified id",
				ArgsUsage:   "<router-id>",
				Description: "Delete the specified router. Example: photon router delete 4f9caq234",
				Action: func(c *cli.Context) {
					err := deleteRouter(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show router info with specified id",
				ArgsUsage: "<router-id>",
				Description: "List the router's name and private IP range. \n\n" +
					"  Example: photon router show 4f9caq234",
				Action: func(c *cli.Context) {
					err := showRouter(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "update",
				Usage:     "Update router",
				ArgsUsage: "<router-id>",
				Description: "Update an existing router given its id. \n" +
					"   Currently only the router name can be updated \n" +
					"   Example: \n" +
					"   photon router update -n new-router 4f9caq234",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Router name",
					},
				},
				Action: func(c *cli.Context) {
					err := updateRouter(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Sends a create Router task to client based on the cli.Context
// Returns an error if one occurred
func createRouter(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	name := c.String("name")
	privateIpCidr := c.String("privateIpCidr")
	tenantName := c.String("tenant")
	projectName := c.String("project")

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

	if !c.GlobalIsSet("non-interactive") && !utils.NeedsFormatting(c) {
		name, err = askForInput("Router name: ", name)
		if err != nil {
			return err
		}
		privateIpCidr, err = askForInput("Router privateIpCidr: ", privateIpCidr)
		if err != nil {
			return err
		}
	}

	if len(name) == 0 || len(privateIpCidr) == 0 {
		return fmt.Errorf("Please provide name and privateIpCidr")
	}

	routerSpec := photon.RouterCreateSpec{}
	routerSpec.Name = name
	routerSpec.PrivateIpCidr = privateIpCidr
	if !c.GlobalIsSet("non-interactive") && !utils.NeedsFormatting(c) {
		fmt.Printf("\nCreating Router: %s(%s)\n", routerSpec.Name, routerSpec.PrivateIpCidr)
	}

	if confirmed(c) {
		createTask, err := client.Photonclient.Projects.CreateRouter(project.ID, &routerSpec)
		if err != nil {
			return err
		}
		routerID, err := waitOnTaskOperation(createTask.ID, c)
		if err != nil {
			return err
		}

		err = formatHelper(c, w, client.Photonclient, routerID)

		return err

	} else {
		fmt.Println("OK. Canceled")
	}

	return nil
}

// Sends a delete router task to client based on the cli.Context
// Returns an error if one occurred
func deleteRouter(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deleteTask, err := client.Photonclient.Routers.Delete(id)
	if err != nil {
		return err
	}
	_, err = waitOnTaskOperation(deleteTask.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Show router info with the specified router ID, returns an error if one occurred
func showRouter(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	router, err := client.Photonclient.Routers.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		fmt.Printf("%s\t%s\t%s\n", router.ID, router.Name, router.PrivateIpCidr)
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(router, w, c)
	} else {
		fmt.Println("Router ID: ", router.ID)
		fmt.Println("  name:                 ", router.Name)
		fmt.Println("  privateIpCidr:        ", router.PrivateIpCidr)
	}

	return nil
}

func updateRouter(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()
	name := c.String("name")

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	updateRouterSpec := photon.RouterUpdateSpec{}
	updateRouterSpec.RouterName = name

	updateRouterTask, err := client.Photonclient.Routers.UpdateRouter(id, &updateRouterSpec)
	if err != nil {
		return err
	}

	id, err = waitOnTaskOperation(updateRouterTask.ID, c)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		router, err := client.Photonclient.Routers.Get(id)
		if err != nil {
			return err
		}
		utils.FormatObject(router, w, c)
	}

	return nil
}
