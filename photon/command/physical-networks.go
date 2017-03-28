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
	"regexp"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

func createPhysicalNetwork(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
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
		network, err := client.Photonclient.Subnets.Get(id)
		if err != nil {
			return err
		}
		utils.FormatObject(network, w, c)
	}

	return nil
}
