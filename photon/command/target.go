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
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"syscall"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon/lightwave"

	"github.com/vmware/photon-controller-cli/photon/client"
	cf "github.com/vmware/photon-controller-cli/photon/configuration"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"golang.org/x/crypto/ssh/terminal"
	"runtime"
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
				Name:      "set",
				Usage:     "Set API target endpoint",
				ArgsUsage: "<endpoint>",
				Description: "Sets the endpoint for Photon Controller that is used for all Photon CLI commands.\n" +
					"   This is saved persistently in the CLI configuration file (typically ~/.photon-cli/.photon-config)\n" +
					"   This should be the full URL (including port) of Photon Controller. Most installations use port 443.\n" +
					"   Example:\n" +
					"      photon target set https://192.0.2.42:443",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "nocertcheck, c",
						Usage: "flag to avoid validating the server's certificate",
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
				Name:      "info",
				Usage:     "Display information about the Photon Controller that is the current target",
				ArgsUsage: " ",
				Action: func(c *cli.Context) {
					err := showInfo(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name: "login",
				Usage: "Allow user to login with a access token, refresh token, username/password" +
					" or using Windows logged in credentials on Windows OS",
				ArgsUsage: " ",
				Description: "The typical usage is to provide a username and password.\n" +
					"   If you do not provide any arguments, you will be prompted for the username and password.\n" +
					"   On Windows you can login using logged in Windows credentials.\n" +
					"   Logging in will result in you receiving a token, which is stored in the CLI configuration file\n" +
					"   in ~/.photon-cli/.photon-config. You can see it with the 'photon auth show-login-token'.",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "access_token, t",
						Usage: "oauth access token that grants access",
					},
					cli.StringFlag{
						Name:  "username, u",
						Usage: "username, if this is provided a password needs to be provided as well",
					},
					cli.StringFlag{
						Name:  "password, p",
						Usage: "password, if this is provided a username needs to be provided as well",
					},
					cli.BoolFlag{
						Name: "windows, w",
						Usage: "flag to use logged in Windows credentials to authenticate. Can be " +
							"used only on Windows OS",
					},
				},
				Action: func(c *cli.Context) {
					err := login(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "logout",
				Usage:     "Remove the token created by the login command. Future requests will require you log in again.",
				ArgsUsage: " ",
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
	err := checkArgCount(c, 1)
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

	err = configureServerCerts(endpoint, noCertCheck, c)
	if err != nil {
		return err
	}

	fmt.Printf("API target set to '%s'\n", endpoint)

	err = clearConfigTenant("")
	if err != nil {
		return err
	}

	return err
}

// Shows set endpoint
func showEndpoint(c *cli.Context) error {
	err := checkArgCount(c, 0)
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

// Shows information about Photon Controller: version, etc.
func showInfo(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	info, err := client.Photonclient.System.GetSystemInfo()

	if utils.NeedsFormatting(c) {
		utils.FormatObject(info, w, c)
	} else {
		fmt.Printf("Version: '%s'\n", info.FullVersion)
		fmt.Printf("Network Type: '%s'\n", info.NetworkType)
	}

	return nil
}

// Store token in the config file
func login(c *cli.Context) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	username := c.String("username")
	password := c.String("password")
	token := c.String("access_token")
	windows := c.Bool("windows")

	if windows {
		if runtime.GOOS != "windows" {
			fmt.Println("--windows flag is only available on Windows OS")
		} else if len(token) != 0 || len(username) != 0 || len(password) != 0 {
			fmt.Println("You cannot use --windows flag with other options")
		} else {
			return loginUsingWindowsCredentials(c)
		}
		return nil
	}

	if !c.GlobalIsSet("non-interactive") && len(token) == 0 {
		username, err = askForInput("User name (username@tenant): ", username)
		if err != nil {
			return err
		}
		if len(password) == 0 {
			fmt.Printf("Password: ")
			// Casting syscall.Stdin to int because during
			// Windows cross-compilation syscall.Stdin is incorrectly
			// treated as a String.
			bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				return err
			}
			password = string(bytePassword)
			fmt.Printf("\n")
		}
	}

	if len(token) == 0 && (len(username) == 0 || len(password) == 0) {
		return fmt.Errorf("Please provide either a token or username/password")
	}

	config, err := cf.LoadConfig()
	if err != nil {
		return err
	}

	if len(token) > 0 {
		config.Token = token

	} else {
		client.Photonclient, err = client.GetClient(c)
		if err != nil {
			return err
		}

		options, err := client.Photonclient.Auth.GetTokensByPassword(username, password)
		if err != nil {
			return err
		}

		config.Token = options.AccessToken
		config.RefreshToken = options.RefreshToken
	}

	err = cf.SaveConfig(config)
	if err != nil {
		return err
	}

	fmt.Println("Login successful")

	return nil
}

func loginUsingWindowsCredentials(c *cli.Context) error {
	config, err := cf.LoadConfig()
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	options, err := client.Photonclient.Auth.GetTokensFromWindowsLogInContext()
	if err != nil {
		return err
	}

	config.Token = options.AccessToken
	config.RefreshToken = options.RefreshToken

	err = cf.SaveConfig(config)
	if err != nil {
		return err
	}

	fmt.Println("Login successful")
	return nil
}

// Remove token from the config file
func logout(c *cli.Context) error {
	err := checkArgCount(c, 0)
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

func configureServerCerts(endpoint string, noChertCheck bool, c *cli.Context) (err error) {
	if noChertCheck {
		return
	}

	//
	// If https endpoint, establish trust with the server
	//
	u, err := url.Parse(endpoint)
	if err != nil {
		return
	}

	// u.Scheme == https -> Server endpoint needs https
	// noCertCheck == false -> User wants server cert validation
	// bTrusted = true -> Server cert is trusted
	if u.Scheme == "https" {
		err = setupApiServerCert(u.Host, c.GlobalIsSet("non-interactive"))
		if err != nil {
			return
		}
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return
	}

	authInfo, err := client.Photonclient.System.GetAuthInfo()
	if err != nil {
		return
	}

	port := authInfo.Port
	if port == 0 {
		port = 443
	}

	host := fmt.Sprintf("%s:%v", authInfo.Endpoint, authInfo.Port)
	err = setupLightWaveCerts(host, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return
	}

	return
}

func setupApiServerCert(host string, isNonInterractive bool) (err error) {
	err = verifyServerTrust("API", host, isNonInterractive)
	if err != nil {
		return
	}

	cert, err := getServerCert(host)
	if err != nil {
		fmt.Printf("Could not establish trust with API server : %s\n", host)
		return
	}

	err = processCert(cert, "API", host)
	if err != nil {
		return
	}

	return
}

func setupLightWaveCerts(host string, isNonInterractive bool) (err error) {
	err = verifyServerTrust("Authentication", host, isNonInterractive)
	if err != nil {
		return
	}

	oidcClient := lightwave.NewOIDCClient(fmt.Sprintf("https://%s", host), nil, nil)
	certs, err := oidcClient.GetRootCerts()
	if err != nil {
		return
	}

	for _, cert := range certs {
		err = processCert(cert, "Authentication", host)
		if err != nil {
			return
		}
	}

	return
}

func verifyServerTrust(serverName string, host string, isNonInterractive bool) (err error) {
	//check if we already trust the server
	bTrusted, _ := isServerTrusted(host)
	if bTrusted {
		return
	}

	if isNonInterractive {
		err = fmt.Errorf(
			"Could not establish trust with API server : %s.\nEither skip certificate validation or accept the server certificate in interactive mode\n",
			host)
		return
	}

	return
}

func processCert(cert *x509.Certificate, serverName string, host string) (err error) {
	trustSrvCrt := ""
	if cert != nil {
		fmt.Printf(
			"Certificate (with below fingerprint) presented by %s server (%s) isn't trusted.\nMD5 = %X\nSHA1  = %X\n",
			serverName,
			host,
			md5.Sum(cert.Raw),
			sha1.Sum(cert.Raw))
		//Get the user input on whether to trust the certificate
		trustSrvCrt, err = askForInput("Do you trust this certificate for future communication? (yes/no): ", trustSrvCrt)
	}

	if err == nil && cert != nil && trustSrvCrt == "yes" {
		err = cf.AddCertToLocalStore(cert)
		if err == nil {
			fmt.Printf(
				"Saved your preference for future communication with %s server %s\n", serverName, host)
		}
	}

	return
}
