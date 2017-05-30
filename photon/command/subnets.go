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
	"regexp"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for subnet
// Subcommands: create;  Usage: subnet create [<options>]
//              delete;  Usage: subnet delete <id>
//              list;    Usage: subnet list [<options>]
//              show;    Usage: subnet show <id>
//              update;  Usage: subnet update <id> [<options>]
//              set-default; Usage: subnet setDefault <id>
func GetSubnetsCommand() cli.Command {
	command := cli.Command{
		Name:  "subnet",
		Usage: "options for subnet",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new subnet",
				ArgsUsage: " ",
				Description: "Create a new subnet. \n\n" +
					"   Example: \n" +
					"    Virtual Subnet:\n" +
					"      photon subnet create -n test -d \"Testing Subnet\" -i 192.168.0.0/16 -r id -s 172.10.0.1\n" +
					"    Physical Subnet:\n" +
					"      photon subnet create -n test -d \"Testing Subnet\" -p port1,port2 \n",
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
					cli.StringFlag{
						Name:  "type, t",
						Usage: "Type of subnet to be created. Types: NAT, NO_NAT or PROVIDER. Default: NAT",
					},
					cli.StringFlag{
						Name:  "portgroups, p",
						Usage: "PortGroups associated with subnet (only for physical subnet)",
					},
					cli.StringFlag{
						Name:  "dns-server-addresses, s",
						Usage: "Comma-separated list of DNS server addresses (Max allowed addresses: 2)",
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
				Usage:     "List all subnets.",
				ArgsUsage: "",
				Description: "List all subnets. If router-id is specified returns subnets only for the specificed router.\n" +
					"   Example: \n" +
					"   photon subnet list -r 4f9caq234 \n",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "name, n",
						Usage: "subnet name",
					},
					cli.StringFlag{
						Name:  "router-id, r",
						Usage: "router id",
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
			{
				Name:      "set-default",
				Usage:     "Set default subnet",
				ArgsUsage: "<subnet-id>",
				Description: "Set the default subnet to be used in the current project when creating" +
					" a VM \n" +
					"   This is not required. When creating a VM you can either specify the \n" +
					"   subnet to use, or rely on the default subnet.",
				Action: func(c *cli.Context) {
					err := setDefaultSubnet(c, os.Stdout)
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
	routerID := c.String("router")

	if len(routerID) == 0 {
		return createPhysicalSubnet(c, w)
	} else {
		return createVirtualSubnet(c, w, routerID)
	}
}

// Creates a virtual subnet under a router
// Returns an error if one occurred
func createVirtualSubnet(c *cli.Context, w io.Writer, routerId string) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	name := c.String("name")
	description := c.String("description")
	privateIpCidr := c.String("privateIpCidr")
	subnetType := c.String("type")
	dnsServerAddresses := c.String("dns-server-addresses")

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	router, err := client.Photonclient.Routers.Get(routerId)
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
		subnetType, err = askForInput("Subnet type (NAT, NO_NAT or PROVIDER. Default is NAT): ", subnetType)
		if err != nil {
			return err
		}
		dnsServerAddresses, err = askForInput("DNS server addresses (comma separated, Max allowed addresses: 2): ", dnsServerAddresses)
		if err != nil {
			return err
		}
	}

	if len(name) == 0 || len(description) == 0 || len(privateIpCidr) == 0 {
		return fmt.Errorf("Please provide name, description and privateIpCidr")
	}

	if len(subnetType) == 0 {
		subnetType = "NAT"
	}

	dnsServerAddressList := []string{}
	if dnsServerAddresses != "" {
		dnsServerAddressList = regexp.MustCompile(`\s*,\s*`).Split(dnsServerAddresses, -1)
	}

	subnetSpec := photon.SubnetCreateSpec{}
	subnetSpec.Name = name
	subnetSpec.Description = description
	subnetSpec.PrivateIpCidr = privateIpCidr
	subnetSpec.Type = subnetType
	subnetSpec.DnsServerAddresses = dnsServerAddressList
	subnetSpec.PortGroups = photon.PortGroups{
		Names: []string{
			"helloWorld",
		},
	}

	if !c.GlobalIsSet("non-interactive") && !utils.NeedsFormatting(c) {
		fmt.Printf("\nCreating Subnet: '%s', Description: '%s', PrivateIpCidr: '%s', DnsServerAddresses: '%s'\n\n",
			subnetSpec.Name, subnetSpec.Description, subnetSpec.PrivateIpCidr, dnsServerAddresses)
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

// Creates a PORT_GROUP type subnet
// Returns an error if one occurred
func createPhysicalSubnet(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	name := c.String("name")
	description := c.String("description")
	portGroups := c.String("portgroups")

	if !c.GlobalIsSet("non-interactive") {
		name, err = askForInput("Subnet name: ", name)
		if err != nil {
			return err
		}
		description, err = askForInput("Subnet Description: ", description)
		if err != nil {
			return err
		}
		portGroups, err = askForInput("Subnet PortGroups: ", portGroups)
		if err != nil {
			return err
		}
	}

	if len(name) == 0 {
		return fmt.Errorf("Please provide subnet name")
	}
	if len(portGroups) == 0 {
		return fmt.Errorf("Please provide portgroups")
	}

	portGroupList := regexp.MustCompile(`\s*,\s*`).Split(portGroups, -1)
	portGroupNames := photon.PortGroups{
		Names: portGroupList,
	}

	createSpec := &photon.SubnetCreateSpec{
		Name:        name,
		Description: description,
		Type:        "PORT_GROUP",
		PortGroups:  portGroupNames,
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Photonclient.Subnets.Create(createSpec)
	if err != nil {
		return err
	}

	id, err := waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		subnet, err := client.Photonclient.Subnets.Get(id)
		if err != nil {
			return err
		}
		utils.FormatObject(subnet, w, c)
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
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	name := c.String("name")
	options := &photon.SubnetGetOptions{
		Name: name,
	}

	routerId := c.String("router-id")

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	var subnetList *photon.Subnets

	if len(routerId) == 0 {
		subnetList, err = client.Photonclient.Subnets.GetAll(options)
	} else {
		subnetList, err = client.Photonclient.Routers.GetSubnets(routerId, options)
	}

	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		utils.FormatObjects(subnetList, w, c)
		return nil
	}

	if c.GlobalIsSet("non-interactive") {
		for _, subnet := range subnetList.Items {
			fmt.Printf("%s\t%s\t%s\t%s\t%t\t%s\t%s\n", subnet.ID, subnet.Name, subnet.Kind, subnet.PrivateIpCidr,
				subnet.IsDefault, subnet.PortGroups.Names, subnet.DnsServerAddresses)
		}
	} else if utils.NeedsFormatting(c) {
		utils.FormatObjects(subnetList, w, c)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\tKind\tDescription\tPrivateIpCidr\tReservedIps\tState\tIsDefault\tPortGroups\tDnsServerAddresses\n")
		for _, subnet := range subnetList.Items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%t\t%s\t%s\n", subnet.ID, subnet.Name, subnet.Kind,
				subnet.Description, subnet.PrivateIpCidr, subnet.ReservedIps, subnet.State,
				subnet.IsDefault, subnet.PortGroups.Names, subnet.DnsServerAddresses)
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
		portGroups := getCommaSeparatedStringFromStringArray(subnet.PortGroups.Names)
		fmt.Printf("%s\t%s\t%s\t%s\t%t\t%s\n",
			subnet.ID, subnet.Name, subnet.Description, subnet.PrivateIpCidr, subnet.IsDefault, portGroups)
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(subnet, w, c)
	} else {
		fmt.Println("Subnet ID: ", subnet.ID)
		fmt.Println("  name:                 ", subnet.Name)
		fmt.Println("  description:          ", subnet.Description)
		fmt.Println("  privateIpCidr:        ", subnet.PrivateIpCidr)
		fmt.Println("  isDefault:            ", subnet.IsDefault)
		fmt.Println("  Port Groups:          ", subnet.PortGroups.Names)
		fmt.Println("  DNS Server Addresses: ", subnet.DnsServerAddresses)
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

func setDefaultSubnet(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	var task *photon.Task
	task, err = client.Photonclient.Subnets.SetDefault(id)

	if err != nil {
		return err
	}

	if confirmed(c) {
		id, err := waitOnTaskOperation(task.ID, c)
		if err != nil {
			return err
		}

		if utils.NeedsFormatting(c) {
			subnet, err := client.Photonclient.Subnets.Get(id)
			if err != nil {
				return err
			}
			utils.FormatObject(subnet, w, c)
		}
	} else {
		fmt.Println("OK. Canceled")
	}
	return nil
}
