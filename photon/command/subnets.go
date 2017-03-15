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
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for subnet
// Subcommands: create;  Usage: subnet create [<options>]
//              delete;  Usage: subnet delete <id>
//              list;    Usage: subnet list <router-id> [<options>]
//              show;    Usage: subnet show <id>
//              update;  Usage: subnet update <id> [<options>]
func GetSubnetsCommand() cli.Command {
	command := cli.Command{
		Name:  "subnet",
		Usage: "options for subnet",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new subnet",
				ArgsUsage: " ",
				Description: "Create a new subnet on the router. \n" +
					"   Example: \n" +
					"   photon subnet create -n subnet-1 -d test-subnet -i 192.168.0.0/16 -r 5f8cap789  \\ \n",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Name of the subnet",
					},
					cli.StringFlag{
						Name:  "description, d",
						Usage: "Description of the subnet",
					},
					cli.StringFlag{
						Name:  "privateIpCidr, i",
						Usage: "The private IP range of subnet in CIDR format, e.g.: 192.168.0.0/16",
					},
					cli.StringFlag{
						Name:  "router, r",
						Usage: "The id of the router on which subnet is to be created",
					},
				},
				Action: func(c *cli.Context) {
					err := createSubnet(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:        "delete",
				Usage:       "Delete subnet with specified id",
				ArgsUsage:   "<subnet-id>",
				Description: "Delete the specified subnet. Example: photon subnet delete 4f9caq234",
				Action: func(c *cli.Context) {
					err := deleteSubnet(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "list",
				Usage:     "List all subnets on a given router in a project",
				ArgsUsage: "<router-id>",
				Description: "List all subnets for the specificed router.\n" +
					"   Example: \n" +
					"   photon subnet list 4f9caq234 \\ \n",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name",
						Usage: "subnet name",
					},
				},
				Action: func(c *cli.Context) {
					err := listSubnets(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show subnet info with specified id",
				ArgsUsage: "<subnet-id>",
				Description: "List the subnet's name, description and private IP range. \n\n" +
					"  Example: photon subnet show 4f9caq234",
				Action: func(c *cli.Context) {
					err := showSubnet(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "update",
				Usage:     "Update subnet",
				ArgsUsage: "<subnet-id>",
				Description: "Update an existing subnet given its id. \n" +
					"   Currently only the subnet name can be updated \n" +
					"   Example: \n" +
					"   photon subnet update -n new-subnet 4f9caq234",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Subnet name",
					},
				},
				Action: func(c *cli.Context) {
					err := updateSubnet(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Sends a create Subnet task to client based on the cli.Context
// Returns an error if one occurred
func createSubnet(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	name := c.String("name")
	description := c.String("description")
	privateIpCidr := c.String("privateIpCidr")
	routerID := c.String("router")

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	router, err := client.Photonclient.Routers.Get(routerID)
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") && !utils.NeedsFormatting(c) {
		name, err = askForInput("Subnet name: ", name)
		if err != nil {
			return err
		}
		description, err = askForInput("Subnet description: ", description)
		if err != nil {
			return err
		}
		privateIpCidr, err = askForInput("Subnet privateIpCidr: ", privateIpCidr)
		if err != nil {
			return err
		}
	}

	if len(name) == 0 || len(description) == 0 || len(privateIpCidr) == 0 {
		return fmt.Errorf("Please provide name, description and privateIpCidr")
	}

	subnetSpec := photon.SubnetCreateSpec{}
	subnetSpec.Name = name
	subnetSpec.Description = description
	subnetSpec.PrivateIpCidr = privateIpCidr
	if !c.GlobalIsSet("non-interactive") && !utils.NeedsFormatting(c) {
		fmt.Printf("\nCreating Subnet: '%s', Description: '%s', PrivateIpCidr: '%s'\n\n",
			subnetSpec.Name, subnetSpec.Description, subnetSpec.PrivateIpCidr)
	}

	if confirmed(c) {
		createTask, err := client.Photonclient.Routers.CreateSubnet(router.ID, &subnetSpec)
		if err != nil {
			return err
		}
		subnetID, err := waitOnTaskOperation(createTask.ID, c)
		if err != nil {
			return err
		}

		err = formatHelper(c, w, client.Photonclient, subnetID)

		return err

	} else {
		fmt.Println("OK. Canceled")
	}

	return nil
}

// Sends a delete subnet task to client based on the cli.Context
// Returns an error if one occurred
func deleteSubnet(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deleteTask, err := client.Photonclient.Subnets.Delete(id)
	if err != nil {
		return err
	}
	_, err = waitOnTaskOperation(deleteTask.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Retrieves a list of subnets, returns an error if one occurred
func listSubnets(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}

	routerId := c.Args().First()
	name := c.String("name")
	options := &photon.SubnetGetOptions{
		Name: name,
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	subnetList, err := client.Photonclient.Routers.GetSubnets(routerId, options)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		utils.FormatObjects(subnetList, w, c)
		return nil
	}

	if c.GlobalIsSet("non-interactive") {
		for _, subnet := range subnetList.Items {
			fmt.Printf("%s\t%s\t%s\t%s\n", subnet.ID, subnet.Name, subnet.Kind, subnet.PrivateIpCidr)
		}
	} else if utils.NeedsFormatting(c) {
		utils.FormatObjects(subnetList, w, c)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\tKind\tDescription\tPrivateIpCidr\tReservedIps\tState\n")
		for _, subnet := range subnetList.Items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", subnet.ID, subnet.Name, subnet.Kind,
				subnet.Description, subnet.PrivateIpCidr, subnet.ReservedIps, subnet.State)
		}
		err := w.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}

// Show subnet info with the specified subnet ID, returns an error if one occurred
func showSubnet(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	subnet, err := client.Photonclient.Subnets.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		fmt.Printf("%s\t%s\t%s\t%s\n",
			subnet.ID, subnet.Name, subnet.Description, subnet.PrivateIpCidr)
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(subnet, w, c)
	} else {
		fmt.Println("Subnet ID: ", subnet.ID)
		fmt.Println("  name:                 ", subnet.Name)
		fmt.Println("  description:          ", subnet.Description)
		fmt.Println("  privateIpCidr:        ", subnet.PrivateIpCidr)
	}

	return nil
}

// Update subnet info with the specified subnet ID,
// currently only the name of subnet can be updated.
// Returns an error if one occurred
func updateSubnet(c *cli.Context, w io.Writer) error {
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

	updateSubnetSpec := photon.SubnetUpdateSpec{}
	updateSubnetSpec.SubnetName = name

	updateSubnetTask, err := client.Photonclient.Subnets.Update(id, &updateSubnetSpec)
	if err != nil {
		return err
	}

	id, err = waitOnTaskOperation(updateSubnetTask.ID, c)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		router, err := client.Photonclient.Subnets.Get(id)
		if err != nil {
			return err
		}
		utils.FormatObject(router, w, c)
	}

	return nil
}
