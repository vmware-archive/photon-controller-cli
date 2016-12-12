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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/configuration"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-cli/photon/utils"
	"github.com/vmware/photon-controller-go-sdk/photon"
	"github.com/vmware/photon-controller-go-sdk/photon/lightwave"
)

// Create a cli.command object for command "auth"
func GetAuthCommand() cli.Command {
	command := cli.Command{
		Name:  "auth",
		Usage: "options for auth",
		Subcommands: []cli.Command{
			{
				Name:      "show",
				Usage:     "Display auth info",
				ArgsUsage: " ",
				Description: "Show information about the authentication service (Lightwave) used by the \n" +
					"   current Photon Controller target.",
				Action: func(c *cli.Context) {
					err := show(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "show-login-token",
				Usage:     "Show login token",
				ArgsUsage: " ",
				Description: "Show information about the current token being used to authenticate with \n" +
					"   Photon Controller. The token is created by doing 'photon target login' \n" +
					"   Using the --detail flag will print the decoded token to stdout.",
				Action: func(c *cli.Context) {
					err := showLoginToken(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "get-api-tokens",
				Usage:     "Retrieve access and refresh tokens",
				ArgsUsage: " ",
				Description: "Retrieve a token you can use with the API. You will get both the API token and" +
					"   API refresh token, which allows you to refresh the API token when it \n" +
					"   expires. Using the --detail flag will print the decoded token to stdout. ",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "username, u",
						Usage: "username, if this is provided a password needs to be provided as well",
					},
					cli.StringFlag{
						Name:  "password, p",
						Usage: "password, if this is provided a username needs to be provided as well",
					},
				},
				Action: func(c *cli.Context) {
					err := getApiTokens(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Get auth info
func show(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	auth, err := client.Photonclient.Auth.Get()
	if err != nil {
		return err
	}

	if !utils.NeedsFormatting(c) {
		err = printAuthInfo(auth, c.GlobalIsSet("non-interactive"))
		if err != nil {
			return err
		}
	} else {
		utils.FormatObject(auth, w, c)
	}

	return nil
}

func showLoginToken(c *cli.Context) error {
	return showLoginTokenWriter(c, os.Stdout, nil)
}

// Handles show-login-token, which shows the current login token, if any
func showLoginTokenWriter(c *cli.Context, w io.Writer, config *configuration.Configuration) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	if config == nil {
		config, err = configuration.LoadConfig()
		if err != nil {
			return err
		}
	}

	if config.Token == "" {
		err = fmt.Errorf("No login token available")
		return err
	}
	if c.GlobalIsSet("detail") {
		dumpTokenDetailsRaw(w, "Login Access Token", config.Token)
	} else if c.GlobalIsSet("non-interactive") {
		fmt.Fprintf(w, "%s\n", config.Token)
	} else if utils.NeedsFormatting(c) {
		mytoken := photon.TokenOptions{AccessToken: config.Token}
		utils.FormatObject(mytoken, w, c)
	} else {
		// General mode
		dumpTokenDetails(w, "Login Access Token", config.Token)
	}
	return nil
}

func getApiTokens(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	username := c.String("username")
	password := c.String("password")

	if !c.GlobalIsSet("non-interactive") && !utils.NeedsFormatting(c) {
		username, err = askForInput("User name (username@tenant): ", username)
		if err != nil {
			return err
		}
		if len(password) == 0 {
			fmt.Printf("Password: ")
			bytePassword, err := terminal.ReadPassword(0)
			if err != nil {
				return err
			}
			password = string(bytePassword)
			fmt.Printf("\n")
		}
	}

	if len(username) == 0 || len(password) == 0 {
		return fmt.Errorf("Please provide username/password")
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	tokens, err := client.Photonclient.Auth.GetTokensByPassword(username, password)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("detail") {
		dumpTokenDetailsRaw(os.Stdout, "API Access Token", tokens.AccessToken)
		dumpTokenDetailsRaw(os.Stdout, "API Refresh Token", tokens.RefreshToken)
	} else if c.GlobalIsSet("non-interactive") {
		fmt.Printf("%s\t%s", tokens.AccessToken, tokens.RefreshToken)
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(tokens, w, c)
	} else {
		// General mode
		dumpTokenDetails(os.Stdout, "API Access Token", tokens.AccessToken)
		dumpTokenDetails(os.Stdout, "API Refresh Token", tokens.RefreshToken)
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

// A JSON web token is a set of Base64 encoded strings separated by a period (.)
// When decoded, it will either be JSON text or a signature
// Here we decode the strings into a single token structure and print the most
// useful fields. We do not print the signature.
func dumpTokenDetails(w io.Writer, name string, encodedToken string) {
	jwtToken := lightwave.ParseTokenDetails(encodedToken)

	fmt.Fprintf(w, "%s:\n", name)
	fmt.Fprintf(w, "\tSubject: %s\n", jwtToken.Subject)
	fmt.Fprintf(w, "\tGroups: ")
	if jwtToken.Groups == nil {
		fmt.Fprintf(w, "<none>\n")
	} else {
		fmt.Fprintf(w, "%s\n", strings.Join(jwtToken.Groups, ", "))
	}
	fmt.Fprintf(w, "\tIssued: %s\n", timestampToString(jwtToken.IssuedAt*1000))
	fmt.Fprintf(w, "\tExpires: %s\n", timestampToString(jwtToken.Expires*1000))
	fmt.Fprintf(w, "\tToken: %s\n", encodedToken)
}

// A JSON web token is a set of Base64 encoded strings separated by a period (.)
// When decoded, it will either be JSON text or a signature
// Here we print the full JSON text. We do not print the signature.
func dumpTokenDetailsRaw(w io.Writer, name string, encodedToken string) {
	jsonStrings, err := lightwave.ParseRawTokenDetails(encodedToken)
	if err != nil {
		fmt.Fprintf(w, "<unparseable>\n")
	}

	fmt.Fprintf(w, "%s:\n", name)
	for _, jsonString := range jsonStrings {
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, []byte(jsonString), "", "  ")
		if err == nil {
			fmt.Fprintf(w, "%s\n", string(prettyJSON.Bytes()))
		}
	}
	fmt.Fprintf(w, "Token: %s\n", encodedToken)
}
