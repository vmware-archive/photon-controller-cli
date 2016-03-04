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
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
	"github.com/vmware/photon-controller-cli/photon/cli/client"
)

// Creates a cli.Command for deployments
// Subcommands: create;     Usage: deployment create [<options>]
//              delete;     Usage: deployment delete <id>
//              list;       Usage: deployment list
//              show;       Usage: deployment show <id>
//              list-hosts; Usage: deployment list-hosts <id>
//              list-vms;   Usage: deployment list-vms <id>
//              prepare-deployment-migration;   Usage: deployment prepare migration <sourceDeploymentAddress> <id>
//              finalize-deployment-migration;  Usage: deployment finalize migration <sourceDeploymentAddress> <id>
//              pause_system;                   Usage: deployment pause_system <id>
//              pause_background_tasks;         Usage: deployment pause_background_tasks <id>
//              resume_system;                  Usage: deployment resume_system <id>
func GetDeploymentsCommand() cli.Command {
	command := cli.Command{
		Name:  "deployment",
		Usage: "options for deployment",
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create a new deployment",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "image_datastores, i",
						Usage: "Comma-separated image Datastore Names",
					},
					cli.StringFlag{
						Name:  "oauth_endpoint, o",
						Usage: "Oauth Endpoint",
					},
					cli.StringFlag{
						Name:  "oauth_tenant, t",
						Usage: "Oauth Tenant name",
					},
					cli.StringFlag{
						Name:  "oauth_username, u",
						Usage: "Oauth Username",
					},
					cli.StringFlag{
						Name:  "oauth_password, p",
						Usage: "Oauth Password",
					},
					cli.StringFlag{
						Name:  "oauth_port, r",
						Usage: "Oauth Port",
					},
					cli.StringFlag{
						Name:  "oauth_security_groups, g",
						Usage: "Oauth Security Groups",
					},
					cli.StringFlag{
						Name:  "syslog_endpoint, s",
						Usage: "Syslog Endpoint",
					},
					cli.BoolFlag{
						Name:  "enable_stats, d",
						Usage: "Enable Stats",
					},
					cli.StringFlag{
						Name:  "stats_store_endpoint, e",
						Usage: "Stats Store Endpoint",
					},
					cli.IntFlag{
						Name:  "stats_store_port, f",
						Usage: "Stats Store Port",
					},
					cli.StringFlag{
						Name:  "ntp_endpoint, n",
						Usage: "Ntp Endpoint",
					},
					cli.BoolFlag{
						Name:  "use_image_datastore_for_vms, v",
						Usage: "Use image Datastore for VMs",
					},
					cli.BoolFlag{
						Name:  "enable_auth, a",
						Usage: "Enable authentication/authorization for deployment",
					},
					cli.BoolFlag{
						Name:  "enable_loadbalancer, l",
						Usage: "Enable Load balancer",
					},
				},
				Action: func(c *cli.Context) {
					err := createDeployment(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "delete",
				Usage: "Delete a deployment by id",
				Action: func(c *cli.Context) {
					err := deleteDeployment(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "list",
				Usage: "Lists all the deployments",
				Action: func(c *cli.Context) {
					err := listDeployments(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "show",
				Usage: "Show deployment info",
				Action: func(c *cli.Context) {
					err := showDeployment(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "list-hosts",
				Usage: "Lists all the hosts associated with the deployment",
				Action: func(c *cli.Context) {
					err := listDeploymentHosts(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "list-vms",
				Usage: "Lists all the vms associated with the deployment",
				Action: func(c *cli.Context) {
					err := listDeploymentVms(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "update-image-datastores",
				Usage: "Updates the list of image datastores",
				Action: func(c *cli.Context) {
					err := updateImageDatastores(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "pause_system",
				Usage: "Pause system under the deployment",
				Action: func(c *cli.Context) {
					err := pauseSystem(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "pause_background_tasks",
				Usage: "Pause system's background tasks under the deployment",
				Action: func(c *cli.Context) {
					err := pauseBackgroundTasks(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "resume_system",
				Usage: "Resume system under the deployment",
				Action: func(c *cli.Context) {
					err := resumeSystem(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Sends a create deployment task to client
func createDeployment(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "deployment create [<options>]")
	if err != nil {
		return err
	}

	imageDatastoreNames := c.String("image_datastores")
	oauthEndpoint := c.String("oauth_endpoint")
	oauthTenant := c.String("oauth_tenant")
	oauthUsername := c.String("oauth_username")
	oauthPassword := c.String("oauth_password")
	oauthPort := c.Int("oauth_port")
	oauthSecurityGroups := c.String("oauth_security_groups")
	syslogEndpoint := c.String("syslog_endpoint")
	statsStoreEndpoint := c.String("stats_store_endpoint")
	enableStats := c.Bool("enable_stats")
	statsStorePort := c.Int("stats_store_port")
	ntpEndpoint := c.String("ntp_endpoint")
	useDatastoreVMs := c.Bool("use_image_datastore_for_vms")
	enableAuth := c.Bool("enable_auth")
	enableLoadBalancer := true
	if c.IsSet("enable_loadbalancer") {
		enableLoadBalancer = c.Bool("enable_loadbalancer")
	}

	if !c.GlobalIsSet("non-interactive") {
		var err error
		imageDatastoreNames, err =
			askForInput("Comma-separated image datastore names: ", imageDatastoreNames)
		if err != nil {
			return err
		}
		oauthEndpoint, err = askForInput("OAuth Endpoint: ", oauthEndpoint)
		if err != nil {
			return err
		}
		oauthTenant, err = askForInput("OAuth Tenant: ", oauthTenant)
		if err != nil {
			return err
		}
		oauthUsername, err = askForInput("OAuth Username: ", oauthUsername)
		if err != nil {
			return err
		}
		oauthPassword, err = askForInput("OAuth Password: ", oauthPassword)
		if err != nil {
			return err
		}
		if !c.IsSet("oauth_port") {
			port, err := askForInput("OAuth Port: ", "")
			if err != nil {
				return err
			}
			oauthPort, err = strconv.Atoi(port)
			if err != nil {
				return err
			}
		}
		oauthSecurityGroups, err = askForInput("Comma-separated oauth security group names: ", oauthSecurityGroups)
		if err != nil {
			return err
		}
		syslogEndpoint, err = askForInput("Syslog Endpoint: ", syslogEndpoint)
		if err != nil {
			return err
		}
		statsStoreEndpoint, err = askForInput("Stats Store Endpoint: ", statsStoreEndpoint)
		if err != nil {
			return err
		}
		statsStorePortString, err := askForInput("Stats Store Port: ", "")
		statsStorePort, err = strconv.Atoi(statsStorePortString)
		if err != nil {
			return err
		}
		ntpEndpoint, err = askForInput("Ntp Endpoint: ", ntpEndpoint)
		if err != nil {
			return err
		}
	}
	err = validate_deployment_arguments(imageDatastoreNames, enableAuth, oauthEndpoint, oauthPort,
		oauthTenant, oauthUsername, oauthPassword, oauthSecurityGroups, enableStats, statsStoreEndpoint, statsStorePort)
	if err != nil {
		return err
	}

	imageDatastoreList := []string{}
	imageDatastoreList = regexp.MustCompile(`\s*,\s*`).Split(imageDatastoreNames, -1)

	oauthSecurityGroupList := []string{}
	if oauthSecurityGroups != "" {
		oauthSecurityGroupList = regexp.MustCompile(`\s*,\s*`).Split(oauthSecurityGroups, -1)
	}

	authInfo := &photon.AuthInfo{
		Enabled:        enableAuth,
		Tenant:         oauthTenant,
		Endpoint:       oauthEndpoint,
		Username:       oauthUsername,
		Password:       oauthPassword,
		Port:           oauthPort,
		SecurityGroups: oauthSecurityGroupList,
	}

	statsInfo := &photon.StatsInfo{
		Enabled:       enableStats,
		StoreEndpoint: statsStoreEndpoint,
		StorePort:     statsStorePort,
	}

	deploymentSpec := &photon.DeploymentCreateSpec{
		Auth:            authInfo,
		ImageDatastores: imageDatastoreList,
		NTPEndpoint:     ntpEndpoint,
		SyslogEndpoint:  syslogEndpoint,
		Stats:           statsInfo,
		UseImageDatastoreForVms: useDatastoreVMs,
		LoadBalancerEnabled:     enableLoadBalancer,
	}

	if len(ntpEndpoint) == 0 {
		deploymentSpec.NTPEndpoint = nil
	}

	if len(syslogEndpoint) == 0 {
		deploymentSpec.SyslogEndpoint = nil
	}

	if confirmed(c.GlobalIsSet("non-interactive")) {
		client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
		if err != nil {
			return err
		}

		createTask, err := client.Esxclient.Deployments.Create(deploymentSpec)
		if err != nil {
			return err
		}
		if c.GlobalIsSet("non-interactive") {
			task, err := client.Esxclient.Tasks.Wait(createTask.ID)
			if err != nil {
				return nil
			}
			fmt.Printf("%s\n", task.Entity.ID)
		} else {
			task, err := pollTask(createTask.ID)
			if err != nil {
				return err
			}
			fmt.Printf("Created deployment %s\n", task.Entity.ID)
		}

		return nil
	} else {
		fmt.Println("OK, canceled")
		return nil
	}
}

// Sends a delete deployment task to client
func deleteDeployment(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "deployment delete <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	deleteTask, err := client.Esxclient.Deployments.Delete(id)
	if err != nil {
		return err
	}

	err = waitOnTaskOperation(deleteTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Retrieves a list of deployments
func listDeployments(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "deployment list")
	if err != nil {
		return err
	}

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	deployments, err := client.Esxclient.Deployments.GetAll()
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		fmt.Println(len(deployments.Items))
		for _, deployment := range deployments.Items {
			fmt.Printf("%s\n", deployment.ID)
		}
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\n")
		for _, deployment := range deployments.Items {
			fmt.Fprintf(w, "%s\n", deployment.ID)
		}
		err = w.Flush()
		if err != nil {
			return err
		}
		fmt.Printf("\nTotal: %d\n", len(deployments.Items))
	}

	return nil
}

// Retrieves information about a deployment
func showDeployment(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "deployment show <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	deployment, err := client.Esxclient.Deployments.Get(id)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		imageDataStores := getCommaSeparatedStringFromStringArray(deployment.ImageDatastores)
		securityGroups := getCommaSeparatedStringFromStringArray(deployment.Auth.SecurityGroups)

		fmt.Printf("%s\t%s\t%s\t%t\t%s\t%s\t%t\n", deployment.ID, deployment.State,
			imageDataStores, deployment.UseImageDatastoreForVms, deployment.SyslogEndpoint,
			deployment.NTPEndpoint, deployment.LoadBalancerEnabled)

		fmt.Printf("%t\t%s\t%s\t%s\t%s\t%d\t%s\n", deployment.Auth.Enabled, deployment.Auth.Username,
			deployment.Auth.Password, deployment.Auth.Endpoint, deployment.Auth.Tenant, deployment.Auth.Port, securityGroups)

	} else {
		syslogEndpoint := deployment.SyslogEndpoint
		if len(deployment.SyslogEndpoint) == 0 {
			syslogEndpoint = "-"
		}
		ntpEndpoint := deployment.NTPEndpoint
		if len(deployment.NTPEndpoint) == 0 {
			ntpEndpoint = "-"
		}
		authUsername := deployment.Auth.Username
		if len(deployment.Auth.Username) == 0 {
			authUsername = "-"
		}
		authPassword := deployment.Auth.Password
		if len(deployment.Auth.Tenant) == 0 {
			authPassword = "-"
		}
		authEndpoint := deployment.Auth.Endpoint
		if len(deployment.Auth.Endpoint) == 0 {
			authEndpoint = "-"
		}
		authTenant := deployment.Auth.Tenant
		if len(deployment.Auth.Tenant) == 0 {
			authTenant = "-"
		}
		fmt.Printf("Deployment ID: %s\n", deployment.ID)
		fmt.Printf("  State:                       %s\n", deployment.State)
		fmt.Printf("  Image Datastores:            %s\n", deployment.ImageDatastores)
		fmt.Printf("  Use image datastore for vms: %t\n\n", deployment.UseImageDatastoreForVms)
		fmt.Printf("  Syslog Endpoint:             %s\n", syslogEndpoint)
		fmt.Printf("  Ntp Endpoint:                %s\n", ntpEndpoint)
		fmt.Printf("  Auth:\n")
		fmt.Printf("    Enabled:                %t\n", deployment.Auth.Enabled)
		fmt.Printf("    UserName:               %s\n", authUsername)
		fmt.Printf("    Password:               %s\n", authPassword)
		fmt.Printf("    Endpoint:               %s\n", authEndpoint)
		fmt.Printf("    Tenant:                 %s\n", authTenant)
		fmt.Printf("    Port:                   %d\n", deployment.Auth.Port)
		fmt.Printf("    Securitygroups:         %v\n", deployment.Auth.SecurityGroups)
	}

	if deployment.Stats != nil {
		stats := deployment.Stats
		if c.GlobalIsSet("non-interactive") {
			fmt.Printf("%t\t%s\t%d\n", stats.Enabled, stats.StoreEndpoint, stats.StorePort)
		} else {
			statsStoreEndpoint := deployment.Stats.StoreEndpoint
			if len(stats.StoreEndpoint) == 0 {
				statsStoreEndpoint = "-"
			}
			fmt.Printf("  Stats:\n")
			fmt.Printf("    Enabled:               %t\n", stats.Enabled)
			fmt.Printf("    Store Endpoint:        %s\n", statsStoreEndpoint)
			fmt.Printf("    Store Port:            %d\n", stats.StorePort)
		}
	} else {
		if c.GlobalIsSet("non-interactive") {
			fmt.Printf("\n")
		}
	}

	if deployment.Migration != nil {
		migration := deployment.Migration
		if c.GlobalIsSet("non-interactive") {
			fmt.Printf("%d\t%d\t%d\t%d\t%d\n", migration.CompletedDataMigrationCycles, migration.DataMigrationCycleProgress,
				migration.DataMigrationCycleSize, migration.VibsUploaded, migration.VibsUploading + migration.VibsUploaded)
		} else {
			fmt.Printf("  Migration status:\n")
			fmt.Printf("    Completed data migration cycles:          %d\n", migration.CompletedDataMigrationCycles)
			fmt.Printf("    Current data migration cycles progress:   %d / %d\n", migration.DataMigrationCycleProgress,
				migration.DataMigrationCycleSize)
			fmt.Printf("    VIB upload progress:                      %d / %d\n", migration.VibsUploaded, migration.VibsUploading + migration.VibsUploaded)
		}
	} else {
		if c.GlobalIsSet("non-interactive") {
			fmt.Printf("\n")
		}
	}

	if deployment.ClusterConfigurations != nil {
		if c.GlobalIsSet("non-interactive") {
			clusterConfigurations := []string{}
			for _, c := range deployment.ClusterConfigurations {
				clusterConfigurations = append(clusterConfigurations, fmt.Sprintf("%s\t%s", c.Type, c.ImageID))
			}
			scriptClusterConfigurations := strings.Join(clusterConfigurations, ",")
			fmt.Println(len(deployment.ClusterConfigurations))
			fmt.Printf("%s\n", scriptClusterConfigurations)
		} else {
			fmt.Println("  Cluster Configurations:")
			for i, c := range deployment.ClusterConfigurations {
				fmt.Printf("    ClusterConfiguration %d:\n", i+1)
				fmt.Println("      Kind:     ", c.Kind)
				fmt.Println("      Type:     ", c.Type)
				fmt.Println("      ImageID:  ", c.ImageID)
			}
		}
	} else {
		if c.GlobalIsSet("non-interactive") {
			fmt.Printf("\n")
		}
	}

	return nil
}

// Lists all the hosts associated with the deployment
func listDeploymentHosts(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "deployment list-hosts <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	hosts, err := client.Esxclient.Deployments.GetHosts(id)
	if err != nil {
		return err
	}

	err = printHostList(hosts.Items, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Lists all the hosts associated with the deployment
func listDeploymentVms(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "deployment list-vms <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	vms, err := client.Esxclient.Deployments.GetVms(id)
	if err != nil {
		return err
	}

	err = printVMList(vms.Items, c.GlobalIsSet("non-interactive"), false)
	if err != nil {
		return err
	}

	return nil
}

// Update the image datastores using the information carried in cli.Context.
func updateImageDatastores(c *cli.Context) error {
	err := checkArgNum(c.Args(), 2, "deployment update-image-datastores <id> <comma separated image datastores>")
	if err != nil {
		return err
	}

	id := c.Args().First()
	imageDataStores := &photon.ImageDatastores{
		Items: regexp.MustCompile(`\s*,\s*`).Split(c.Args()[1], -1),
	}

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	task, err := client.Esxclient.Deployments.UpdateImageDatastores(id, imageDataStores)
	if err != nil {
		return err
	}

	fmt.Printf("Image datastores of deployment %s is finished\n", task.Entity.ID)
	return nil
}

// Sends a pause system task to client
func pauseSystem(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "deployment pause_system <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	pauseSystemTask, err := client.Esxclient.Deployments.PauseSystem(id)
	if err != nil {
		return err
	}

	err = waitOnTaskOperation(pauseSystemTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Sends a pause background task to client
func pauseBackgroundTasks(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "deployment pause_background_tasks <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	pauseBackgroundTask, err := client.Esxclient.Deployments.PauseBackgroundTasks(id)
	if err != nil {
		return err
	}

	err = waitOnTaskOperation(pauseBackgroundTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Sends a resume system task to client
func resumeSystem(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "deployment resume_system <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	resumeSystemTask, err := client.Esxclient.Deployments.ResumeSystem(id)
	if err != nil {
		return err
	}

	err = waitOnTaskOperation(resumeSystemTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}
