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
	"errors"
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

const (
	PHYSICAL         = "PHYSICAL"
	SOFTWARE_DEFINED = "SOFTWARE_DEFINED"
	NOT_AVAILABLE    = "NOT_AVAILABLE"
)

func isSoftwareDefinedNetwork(c *cli.Context) (sdnEnabled bool, err error) {
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return
	}

	info, err := client.Photonclient.System.GetSystemInfo()
	if err != nil {
		return
	}

	if info.NetworkType == NOT_AVAILABLE {
		err = errors.New("Network type is missing")
	} else {
		sdnEnabled = (info.NetworkType == SOFTWARE_DEFINED)
	}
	return
}

// Creates a cli.Command for network
// Subcommands: create;  Usage: network create [<options>]
//              delete;  Usage: network delete <id>
//              list;    Usage: network list [<options>]
//              show;    Usage: network show <id>
//              update;  Usage: network update <id> [<options>]
func GetNetworksCommand() cli.Command {
	command := cli.Command{
		Name:  "network",
		Usage: "options for network",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new network",
				ArgsUsage: " ",
				Description: "Create a new network within a project. Subnets can be created in this network. \n" +
					"   The private IP range of network will be sub-divided into smaller CIDRs for each subnet \n" +
					"   created under this network \n\n" +
					"   Example: \n" +
					"   photon network create -n network-1 -i 192.168.0.0/16 -t cloud-dev -p cloud-dev-staging \n",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Network name",
					},
					cli.StringFlag{
						Name:  "privateIpCidr, i",
						Usage: "The private IP range of network in CIDR format, e.g.: 192.168.0.0/16",
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
					err := createNetwork(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:        "delete",
				Usage:       "Delete network with specified id",
				ArgsUsage:   "<network-id>",
				Description: "Delete the specified network. Example: photon network delete 4f9caq234",
				Action: func(c *cli.Context) {
					err := deleteNetwork(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "list",
				Usage:     "List all networks in a project",
				ArgsUsage: " ",
				Description: "List all networks in a project. It will show networks id, name, kind and\n" +
					"   private IP range. If tenant and project names are not mentioned, it will list networks\n" +
					"   for current tenant and project.\n" +
					"   Example: \n" +
					"   photon network list -t tenanat_1 -p project_1 \\ \n",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name",
					},
					cli.StringFlag{
						Name:  "project, p",
						Usage: "Project name",
					},
					cli.StringFlag{
						Name:  "name, n",
						Usage: "network name",
					},
				},
				Action: func(c *cli.Context) {
					err := listNetworks(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show network info with specified id",
				ArgsUsage: "<network-id>",
				Description: "List the network's name and private IP range. \n\n" +
					"  Example: photon network show 4f9caq234",
				Action: func(c *cli.Context) {
					err := showNetwork(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "update",
				Usage:     "Update network",
				ArgsUsage: "<network-id>",
				Description: "Update an existing network given its id. \n" +
					"   Currently only the network name can be updated \n" +
					"   Example: \n" +
					"   photon network update -n new-network 4f9caq234",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Network name",
					},
				},
				Action: func(c *cli.Context) {
					err := updateNetwork(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Sends a create Network task to client based on the cli.Context
// Returns an error if one occurred
func createNetwork(c *cli.Context, w io.Writer) error {
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
		name, err = askForInput("Network name: ", name)
		if err != nil {
			return err
		}
		privateIpCidr, err = askForInput("Network privateIpCidr: ", privateIpCidr)
		if err != nil {
			return err
		}
	}

	if len(name) == 0 || len(privateIpCidr) == 0 {
		return fmt.Errorf("Please provide name and privateIpCidr")
	}

	networkSpec := photon.NetworkCreateSpec{}
	networkSpec.Name = name
	networkSpec.PrivateIpCidr = privateIpCidr
	if !c.GlobalIsSet("non-interactive") && !utils.NeedsFormatting(c) {
		fmt.Printf("\nCreating Network: %s(%s)\n", networkSpec.Name, networkSpec.PrivateIpCidr)
	}

	if confirmed(c) {
		createTask, err := client.Photonclient.Projects.CreateNetwork(project.ID, &networkSpec)
		if err != nil {
			return err
		}
		networkID, err := waitOnTaskOperation(createTask.ID, c)
		if err != nil {
			return err
		}

		err = formatHelper(c, w, client.Photonclient, networkID)

		return err

	} else {
		fmt.Println("OK. Canceled")
	}

	return nil
}

// Sends a delete network task to client based on the cli.Context
// Returns an error if one occurred
func deleteNetwork(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deleteTask, err := client.Photonclient.Networks.Delete(id)
	if err != nil {
		return err
	}
	_, err = waitOnTaskOperation(deleteTask.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Retrieves a list of networks, returns an error if one occurred
func listNetworks(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	tenantName := c.String("tenant")
	projectName := c.String("project")

	name := c.String("name")
	options := &photon.NetworkGetOptions{
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

	networkList, err := client.Photonclient.Projects.GetNetworks(project.ID, options)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		utils.FormatObjects(networkList, w, c)
		return nil
	}

	if c.GlobalIsSet("non-interactive") {
		for _, network := range networkList.Items {
			fmt.Printf("%s\t%s\t%s\t%s\t%t\n", network.ID, network.Name, network.Kind, network.PrivateIpCidr,
				network.IsDefault)
		}
	} else if !utils.NeedsFormatting(c) {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\tKind\tPrivateIpCidr\tIsDefault\n")
		for _, network := range networkList.Items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%t\n", network.ID, network.Name, network.Kind,
				network.PrivateIpCidr, network.IsDefault)
		}
		err := w.Flush()
		if err != nil {
			return err
		}
	}

	return nil
}

// Show network info with the specified network ID, returns an error if one occurred
func showNetwork(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	network, err := client.Photonclient.Networks.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		fmt.Printf("%s\t%s\t%s\t%t\n", network.ID, network.Name, network.PrivateIpCidr, network.IsDefault)
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(network, w, c)
	} else {
		fmt.Println("Network ID: ", network.ID)
		fmt.Println("  name:                 ", network.Name)
		fmt.Println("  privateIpCidr:        ", network.PrivateIpCidr)
		fmt.Println("  isDefault:            ", network.IsDefault)
	}

	return nil
}

func updateNetwork(c *cli.Context, w io.Writer) error {
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

	updateNetworkSpec := photon.NetworkUpdateSpec{}
	updateNetworkSpec.NetworkName = name

	updateNetworkTask, err := client.Photonclient.Networks.UpdateNetwork(id, &updateNetworkSpec)
	if err != nil {
		return err
	}

	id, err = waitOnTaskOperation(updateNetworkTask.ID, c)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		network, err := client.Photonclient.Networks.Get(id)
		if err != nil {
			return err
		}
		utils.FormatObject(network, w, c)
	}

	return nil
}
