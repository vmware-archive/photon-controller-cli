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
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"log"
	"net/url"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"

	cf "github.com/vmware/photon-controller-cli/photon/cli/configuration"
)

// Create a cli.command object for command "target"
// Subcommands: set;    Usage: target set <url>
//              login;  Usage: target login <token>
//              logout; Usage: target logout
//              show;   Usage: target show
func GetTargetCommand() cli.Command {
	command := cli.Command{
		Name:  "target",
		Usage: "options for target",
		Subcommands: []cli.Command{
			{
				Name:  "set",
				Usage: "Set API target endpoint",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "nocertcheck, c",
						Usage: "flag to avoid validating server cert",
					},
				},
				Action: func(c *cli.Context) {
					err := setEndpoint(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "show",
				Usage: "Show current target endpoint",
				Action: func(c *cli.Context) {
					err := showEndpoint(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "login",
				Usage: "Allow user to login with a token",
				Action: func(c *cli.Context) {
					err := login(c.Args())
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "logout",
				Usage: "Allow user to logout",
				Action: func(c *cli.Context) {
					err := logout(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Read config from config file, change target and then write back to file
// Also check if the target is reachable securely
func setEndpoint(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "target set <url>")
	if err != nil {
		return err
	}
	endpoint := c.Args()[0]
	noCertCheck := c.Bool("nocertcheck")

	config, err := cf.LoadConfig()
	if err != nil {
		return err
	}

	config.CloudTarget = endpoint
	config.IgnoreCertificate = noCertCheck

	err = cf.SaveConfig(config)
	if err != nil {
		return err
	}

	fmt.Printf("API target set to '%s'\n", endpoint)

	err = clearConfigTenant("")
	if err != nil {
		return err
	}
	//If https endpoint, establish trust with the server
	//
	u, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	//
	//u.Scheme == https -> Server endpoint needs https
	//noCertCheck == false -> User wants server cert validation
	//bTrusted = true -> Server cert is trusted

	if u.Scheme == "https" && noCertCheck == false {
		//check if we already trust the server
		bTrusted, _ := isServerTrusted(u.Host)
		if !bTrusted {
			if c.GlobalIsSet("non-interactive") {
				fmt.Printf("Could not establish trust with server : %s.\nYou could either skip server certificate validation or accept the server certificate in interactive mode\n", u.Host)
				return nil
			}
			cert, err := getServerCert(u.Host)
			if err != nil {
				fmt.Printf("Could not establish trust with server : %s\n", u.Host)
				return err
			}
			trustSrvCrt := ""
			if cert != nil {
				fmt.Printf("Certificate (with below fingerprint) presented by server (%s) isn't trusted.\nMD5 = %X\nSHA1  = %X\n",
					u.Host,
					md5.Sum(cert.Raw),
					sha1.Sum(cert.Raw))
				//Get the user input on whether to trust the certificate
				trustSrvCrt, err = askForInput("Do you trust this certificate for future communication? (yes/no): ", trustSrvCrt)
			}
			if err == nil && cert != nil && trustSrvCrt == "yes" {
				err = cf.AddCertToLocalStore(cert)
				if err == nil {
					fmt.Printf("\nSaved your preference for future communicaition with %s\n", u.Host)
				}
			}
		}
	}
	return err
}

// Shows set endpoint
func showEndpoint(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "target show")
	if err != nil {
		return err
	}
	config, err := cf.LoadConfig()
	if err != nil {
		return err
	}

	if len(config.CloudTarget) == 0 {
		fmt.Printf("No API target set\n")
	} else {
		fmt.Printf("Current API target is '%s'\n", config.CloudTarget)
	}
	return nil
}

// Store token in the config file
func login(args []string) error {
	err := checkArgNum(args, 1, "target login <token>")
	if err != nil {
		return err
	}
	token := args[0]

	config, err := cf.LoadConfig()
	if err != nil {
		return err
	}

	config.Token = token

	err = cf.SaveConfig(config)
	if err != nil {
		return err
	}

	fmt.Println("Token stored in config file")

	return nil
}

// Remove token from the config file
func logout(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "target logout")
	if err != nil {
		return err
	}
	config, err := cf.LoadConfig()
	if err != nil {
		return err
	}

	config.Token = ""

	err = cf.SaveConfig(config)
	if err != nil {
		return err
	}

	fmt.Println("Token removed from config file")

	return nil
}
