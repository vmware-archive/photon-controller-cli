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
	"log"
	"os"
	"regexp"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/client"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for networks
// Subcommands: create; Usage: network create [<options>]
//              delete; Usage: network delete <id>
//              list;   Usage: network list
//              show;   Usage: network show <id>
//              setDefault; Usage: network setDefault <id>
func GetNetworksCommand() cli.Command {
	command := cli.Command{
		Name:  "network",
		Usage: "options for network",
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create a new network",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Network name",
					},
					cli.StringFlag{
						Name:  "description, d",
						Usage: "Description of network",
					},
					cli.StringFlag{
						Name:  "portgroups, p",
						Usage: "PortGroups associated with network",
					},
				},
				Action: func(c *cli.Context) {
					err := createNetwork(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "delete",
				Usage: "Delete a network",
				Action: func(c *cli.Context) {
					err := deleteNetwork(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "list",
				Usage: "List networks",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Optionally filter by name",
					},
				},
				Action: func(c *cli.Context) {
					err := listNetworks(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "show",
				Usage: "Show specified network",
				Action: func(c *cli.Context) {
					err := showNetwork(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "setDefault",
				Usage: "Set default network",
				Action: func(c *cli.Context) {
					err := setDefaultNetwork(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

func createNetwork(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "network create [<options>]")
	if err != nil {
		return err
	}

	name := c.String("name")
	description := c.String("description")
	portGroups := c.String("portgroups")

	if !c.GlobalIsSet("non-interactive") {
		name, err = askForInput("Network name: ", name)
		if err != nil {
			return err
		}
		description, err = askForInput("Description of network: ", description)
		if err != nil {
			return err
		}
		portGroups, err = askForInput("PortGroups of network: ", portGroups)
		if err != nil {
			return err
		}
	}

	if len(name) == 0 {
		return fmt.Errorf("Please provide network name")
	}
	if len(portGroups) == 0 {
		return fmt.Errorf("Please provide portgroups")
	}

	portGroupList := regexp.MustCompile(`\s*,\s*`).Split(portGroups, -1)
	createSpec := &photon.NetworkCreateSpec{
		Name:        name,
		Description: description,
		PortGroups:  portGroupList,
	}

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	task, err := client.Esxclient.Networks.Create(createSpec)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	return nil
}

func deleteNetwork(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "network delete <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	task, err := client.Esxclient.Networks.Delete(id)
	if err != nil {
		return err
	}

	if confirmed(c.GlobalIsSet("non-interactive")) {
		_, err = waitOnTaskOperation(task.ID, c)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("OK. Canceled")
	}
	return nil
}

func listNetworks(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "network list [<options>]")
	if err != nil {
		return err
	}
	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	name := c.String("name")
	options := &photon.NetworkGetOptions{
		Name: name,
	}

	networks, err := client.Esxclient.Networks.GetAll(options)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		for _, network := range networks.Items {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", network.ID, network.Name, network.State, network.PortGroups, network.Description)
		}
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\tState\tPortGroups\tDescriptions\tIsDefault\n")
		for _, network := range networks.Items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%t\n", network.ID, network.Name, network.State, network.PortGroups,
				network.Description, network.IsDefault)
		}
		err = w.Flush()
		if err != nil {
			return err
		}
		fmt.Printf("Total: %d\n", len(networks.Items))
	}

	return nil
}

func showNetwork(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "network show <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	network, err := client.Esxclient.Networks.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		portGroups := getCommaSeparatedStringFromStringArray(network.PortGroups)
		fmt.Printf("%s\t%s\t%s\t%s\t%s\n", network.ID, network.Name, network.State, portGroups, network.Description)
	} else {
		fmt.Printf("Network ID: %s\n", network.ID)
		fmt.Printf("  Name:        %s\n", network.Name)
		fmt.Printf("  State:       %s\n", network.State)
		fmt.Printf("  Description: %s\n", network.Description)
		fmt.Printf("  Port Groups: %s\n", network.PortGroups)
		fmt.Printf("  Is Default: %t\n", network.IsDefault)
	}

	return nil
}

func setDefaultNetwork(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "network setDefault <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	task, err := client.Esxclient.Networks.SetDefault(id)
	if err != nil {
		return err
	}

	if confirmed(c.GlobalIsSet("non-interactive")) {
		_, err = waitOnTaskOperation(task.ID, c)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("OK. Canceled")
	}
	return nil
}
