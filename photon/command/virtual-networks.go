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
	size := c.Int("size")
	staticIpSize := c.Int("staticIpSize")
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
	}

	if len(name) == 0 {
		return fmt.Errorf("Please provide network name")
	}
	if routingType != ROUTED && routingType != ISOLATED {
		return fmt.Errorf("Please choose the correct routing type for network (ROUTED or ISOLATED)")
	}
	if size <= 0 {
		return fmt.Errorf("Network size must be greater than 0")
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
