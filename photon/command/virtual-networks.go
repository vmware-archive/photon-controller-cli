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
	"github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"
	"github.com/vmware/photon-controller-go-sdk/photon"
	"io"
	"os"
	"strconv"
	"text/tabwriter"
)

const (
	ROUTED   = "ROUTED"
	ISOLATED = "ISOLATED"
)

func createVirtualNetwork(c *cli.Context, w io.Writer) error {
	err := checkArgNum(c.Args(), 0, "network create [<options>]")
	if err != nil {
		return err
	}

	name := c.String("name")
	description := c.String("description")
	routingType := c.String("routingType")
	sizeStr := c.String("size")
	staticIpSizeStr := c.String("staticIpSize")
	projectId := c.String("projectId")

	if !utils.IsNonInteractive(c) {
		name, err = askForInput("Network name: ", name)
		if err != nil {
			return err
		}
		description, err = askForInput("Description of network: ", description)
		if err != nil {
			return err
		}
		routingType, err = askForInput("Routing type of network: ", routingType)
		if err != nil {
			return err
		}
		projectId, err = askForInput("Project ID that network belongs to: ", projectId)
		if err != nil {
			return err
		}
		sizeStr, err = askForInput("Size of IP pool of the network (must be power of 2, at least 8): ", sizeStr)
		if err != nil {
			return err
		}
		staticIpSizeStr, err = askForInput("Size of the static IP pool (must be less than size of IP pool): ",
			staticIpSizeStr)
		if err != nil {
			return err
		}
	}

	if len(name) == 0 {
		return fmt.Errorf("Please provide network name")
	}
	if routingType != ROUTED && routingType != ISOLATED {
		return fmt.Errorf("Please choose the correct routing type for network (ROUTED or ISOLATED)")
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		return err
	}
	if size < 8 {
		return fmt.Errorf("Network size must be at least 8")
	}
	staticIpSize, err := strconv.Atoi(staticIpSizeStr)
	if err != nil {
		return err
	}

	createSpec := &photon.VirtualSubnetCreateSpec{
		Name:                 name,
		Description:          description,
		RoutingType:          routingType,
		Size:                 size,
		ReservedStaticIpSize: staticIpSize,
	}

	client.Esxclient, err = client.GetClient(utils.IsNonInteractive(c))
	if err != nil {
		return err
	}

	task, err := client.Esxclient.VirtualSubnets.Create(projectId, createSpec)
	if err != nil {
		return err
	}

	id, err := waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		network, err := client.Esxclient.Subnets.Get(id)
		if err != nil {
			return err
		}
		utils.FormatObject(network, w, c)
	}

	return nil
}

func listVirtualNetworks(c *cli.Context, w io.Writer) error {
	err := checkArgNum(c.Args(), 0, "network list [<options>]")
	if err != nil {
		return err
	}
	client.Esxclient, err = client.GetClient(utils.IsNonInteractive(c))
	if err != nil {
		return err
	}

	name := c.String("name")
	options := &photon.VirtualSubnetGetOptions{
		Name: name,
	}

	projectId := c.String("projectId")
	if len(projectId) == 0 {
		return fmt.Errorf("Please provide project ID")
	}

	networks, err := client.Esxclient.VirtualSubnets.GetAll(projectId, options)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		for _, network := range networks.Items {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", network.ID, network.Name, network.State,
				network.Description, network.RoutingType, network.IsDefault, network.Cidr, network.LowIpDynamic,
				network.HighIpDynamic, network.LowIpStatic, network.HighIpStatic, network.ReservedIpList)
		}
	} else if utils.NeedsFormatting(c) {
		utils.FormatObjects(networks.Items, w, c)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tName\tState\tDescriptions\tRoutingType\tIsDefault\tCIDR\tLowDynamicIP\tHighDynamicIP"+
			"\tLowStaticIP\tHighStaticIP\tReservedIpList\n")
		for _, network := range networks.Items {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", network.ID, network.Name, network.State,
				network.Description, network.RoutingType, network.IsDefault, network.Cidr, network.LowIpDynamic,
				network.HighIpDynamic, network.LowIpStatic, network.HighIpStatic, network.ReservedIpList)
		}
		err = w.Flush()
		if err != nil {
			return err
		}
		fmt.Printf("Total: %d\n", len(networks.Items))
	}

	return nil
}

func showVirtualNetwork(c *cli.Context, w io.Writer) error {
	err := checkArgNum(c.Args(), 1, "network show <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(utils.IsNonInteractive(c))
	if err != nil {
		return err
	}

	network, err := client.Esxclient.VirtualSubnets.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", network.ID, network.Name, network.State,
			network.Description, network.RoutingType, network.IsDefault, network.Cidr, network.LowIpDynamic,
			network.HighIpDynamic, network.LowIpStatic, network.HighIpStatic, network.ReservedIpList)
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(network, w, c)
	} else {
		fmt.Printf("Network ID: %s\n", network.ID)
		fmt.Printf("  Name:             %s\n", network.Name)
		fmt.Printf("  State:            %s\n", network.State)
		fmt.Printf("  Description:      %s\n", network.Description)
		fmt.Printf("  Routing Type:     %s\n", network.RoutingType)
		fmt.Printf("  Is Default:       %s\n", network.IsDefault)
		fmt.Printf("  CIDR:             %s\n", network.Cidr)
		fmt.Printf("  Start Dynamic IP: %s\n", network.LowIpDynamic)
		fmt.Printf("  End Dynamic IP:   %s\n", network.HighIpDynamic)
		fmt.Printf("  Start Static IP:  %s\n", network.LowIpStatic)
		fmt.Printf("  End Static IP:    %s\n", network.HighIpStatic)
		fmt.Printf("  Reserved IP List: %s\n", network.ReservedIpList)
	}

	return nil
}
