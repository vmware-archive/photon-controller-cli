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

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"

	"os"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/client"
)

// Create a cli.command object for command "auth"
func GetAuthCommand() cli.Command {
	command := cli.Command{
		Name:  "auth",
		Usage: "options for auth",
		Subcommands: []cli.Command{
			{
				Name:  "show",
				Usage: "display auth info",
				Action: func(c *cli.Context) {
					err := show(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
		},
	}
	return command
}

// Get auth info
func show(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "auth show")
	if err != nil {
		return err
	}
	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	auth, err := client.Esxclient.Auth.Get()
	if err != nil {
		return err
	}

	err = printAuthInfo(auth, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Print out auth info
func printAuthInfo(auth *photon.AuthInfo, isScripting bool) error {
	if isScripting {
		fmt.Printf("%t\t%s\t%d\n", auth.Enabled, auth.Endpoint, auth.Port)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "Enabled\tEndpoint\tPort\n")
		fmt.Fprintf(w, "%t\t%s\t%d\n", auth.Enabled, auth.Endpoint, auth.Port)
		err := w.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}
