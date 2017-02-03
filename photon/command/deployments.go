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
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

type VM_NetworkIPs struct {
	vm  photon.VM
	ips string
}

type ipsSorter []VM_NetworkIPs

func (ip ipsSorter) Len() int           { return len(ip) }
func (ip ipsSorter) Swap(i, j int)      { ip[i], ip[j] = ip[j], ip[i] }
func (ip ipsSorter) Less(i, j int) bool { return ip[i].ips < ip[j].ips }

// Creates a cli.Command for deployments
// Subcommands:
//              list;       Usage: deployment list
//              show;       Usage: deployment show [<id>]
//              list-hosts; Usage: deployment list-hosts [<id>]
//              list-vms;   Usage: deployment list-vms [<id>]

//              update-image-datastores;        Usage: deployment update-image-datastores [<id> <options>]
//              sync-hosts-config;              Usage: deployment sync-hosts-config [<id>]

//              pause;                          Usage: deployment pause [<id>]
//              pause-background-tasks;         Usage: deployment pause-background-tasks [<id>]
//              resume;                         Usage: deployment resume [<id>]
//              set-security-groups             Usage: deployment set-security-groups [<id>] comma-separated-groups

//              migration prepare;              Usage: deployment prepare migration [<id> <options>]
//              migration finalize;             Usage: deployment finalize migration [<id> <options>]
//              migration status;               Usage: deployment finalize migration [<id>]

//              enable-cluster-type;            Usage: deployment enable-cluster-type [<id> <options>]
//              disable-cluster-type;           Usage: deployment disable-cluster-type [<id> <options>]

func GetDeploymentsCommand() cli.Command {
	command := cli.Command{
		Name:  "deployment",
		Usage: "options for deployment",
		Subcommands: []cli.Command{
			{
				Name:        "list",
				Usage:       "Lists all the deployments",
				ArgsUsage:   " ",
				Description: "[Deprecated] List the current deployment.",
				Action: func(c *cli.Context) {
					err := listDeployments(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show deployment info",
				ArgsUsage: " ",
				Description: "Show detailed information about the current deployment.\n" +
					"   Requires system administrator access,",
				Action: func(c *cli.Context) {
					err := showDeployment(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "list-hosts",
				Usage:     "Lists all ESXi hosts",
				ArgsUsage: " ",
				Description: "List information about all ESXi hosts used in the deployment.\n" +
					"   For each host, the ID, the current state, the IP, and the type (MGMT and/or CLOUD)\n" +
					"   Requires system administrator access.",
				Action: func(c *cli.Context) {
					err := listDeploymentHosts(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "list-vms",
				Usage:     "Lists all VMs",
				ArgsUsage: " ",
				Description: "List all VMs associated with all tenants and projects.\n" +
					"   Requires system administrator access.",
				Action: func(c *cli.Context) {
					err := listDeploymentVms(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "enable-cluster-type",
				Usage:     "Enable cluster type for deployment",
				ArgsUsage: " ",
				Description: "Enable a cluster type (e.g. Kubernetes) and specify the image to be used\n" +
					"   when creating the cluster.\n" +
					"   Requires system administrator access.",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "type, k",
						Usage: "Cluster type (accepted values are KUBERNETES or HARBOR)",
					},
					cli.StringFlag{
						Name:  "image-id, i",
						Usage: "ID of the cluster image",
					},
				},
				Action: func(c *cli.Context) {
					err := enableClusterType(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "disable-cluster-type",
				Usage:     "Disable cluster type for deployment",
				ArgsUsage: " ",
				Description: "Disable a cluster type (e.g. Kubernetes). Users will no longer be able\n" +
					"   to deploy clusters of that type, but existing clusters will be unaffected.\n" +
					"   Requires system administrator access.",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "type, k",
						Usage: "Cluster type (accepted values are KUBERNETES or HARBOR)",
					},
				},
				Action: func(c *cli.Context) {
					err := disableClusterType(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "update-image-datastores",
				Usage:     "Updates the list of image datastores",
				ArgsUsage: " ",
				Description: "Update the list of allowed image datastores.\n" +
					"   Requires system administrator access.",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "datastores, d",
						Usage: "Comma separated name of datastore names",
					},
				},
				Action: func(c *cli.Context) {
					err := updateImageDatastores(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "sync-hosts-config",
				Usage:     "Synchronizes hosts configurations",
				ArgsUsage: " ",
				Action: func(c *cli.Context) {
					err := syncHostsConfig(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "pause",
				Usage:     "Pause system under the deployment",
				ArgsUsage: " ",
				Description: "Pause Photon Controller. All incoming requests that modify the system\n" +
					"   state (other than resume) will be refused. This implies pause-background-states" +
					"   Requires system administrator access.",
				Action: func(c *cli.Context) {
					err := pauseSystem(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "pause-background-tasks",
				Usage:     "Pause background tasks",
				ArgsUsage: " ",
				Description: "Pause all background tasks in Photon Controller, such as image replication." +
					"   Incoming requests from users will continue to work\n" +
					"   Requires system administrator access.",
				Action: func(c *cli.Context) {
					err := pauseBackgroundTasks(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "resume",
				Usage:     "Resume system under the deployment",
				ArgsUsage: " ",
				Description: "Resume Photon Controller after it has been paused.\n" +
					"   Requires system administrator access.",
				Action: func(c *cli.Context) {
					err := resumeSystem(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "set-security-groups",
				Usage:     "Set security groups for a deployment",
				ArgsUsage: "<comma separate list of security groups>",
				Description: "Provide the list of Lightwave groups that contain the people who are\n" +
					"   allowed to be system administrators. Be careful: providing the wrong group could remove\n" +
					"   your access.\n\n" +
					"   A security group specifies both the Lightwave domain and Lightwave group.\n" +
					"   For example, a security group may be photon.vmware.com\\group-1\n\n" +
					"   Example: photon deployment set-security-groups 'photon.vmware.com\\group-1,photon.vmware.com\\group-2'\n\n" +
					"   Requires system administrator access.",
				Action: func(c *cli.Context) {
					err := setDeploymentSecurityGroups(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "migration",
				Usage: "migrates state and hosts between photon controller deployments",
				Subcommands: []cli.Command{
					{
						Name:  "prepare",
						Usage: "initializes the migration",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "endpoint, e",
								Usage: "API endpoint of the old management plane",
							},
						},
						Action: func(c *cli.Context) {
							err := deploymentMigrationPrepare(c)
							if err != nil {
								log.Fatal("Error: ", err)
							}
						},
					},
					{
						Name:  "finalize",
						Usage: "finalizes the migration",
						Flags: []cli.Flag{
							cli.StringFlag{
								Name:  "endpoint, e",
								Usage: "API endpoint of the old management plane",
							},
						},
						Action: func(c *cli.Context) {
							err := deploymentMigrationFinalize(c)
							if err != nil {
								log.Fatal("Error: ", err)
							}
						},
					},
					{
						Name:  "status",
						Usage: "shows the status of the current migration",
						Action: func(c *cli.Context) {
							err := showMigrationStatus(c)
							if err != nil {
								log.Fatal("Error: ", err)
							}
						},
					},
				},
			},
		},
	}
	return command
}

// Retrieves a list of deployments
func listDeployments(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deployments, err := client.Photonclient.Deployments.GetAll()
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		utils.FormatObjects(deployments, w, c)
	} else if c.GlobalIsSet("non-interactive") {
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
func showDeployment(c *cli.Context, w io.Writer) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deployment, err := client.Photonclient.Deployments.Get(id)
	if err != nil {
		return err
	}

	vms, err := client.Photonclient.Deployments.GetVms(id)
	if err != nil {
		return err
	}

	var data []VM_NetworkIPs

	for _, vm := range vms.Items {
		networks, err := getVMNetworks(vm.ID, c)
		if err != nil {
			return err
		}
		ipAddr := "N/A"
		for _, nt := range networks {
			network := nt.(map[string]interface{})
			if len(network) != 0 && network["network"] != nil {
				if val, ok := network["ipAddress"]; ok && val != nil {
					ipAddr = val.(string)
					break
				}
			}

		}
		data = append(data, VM_NetworkIPs{vm, ipAddr})
	}
	if utils.NeedsFormatting(c) {
		utils.FormatObject(deployment, w, c)
	} else if c.GlobalIsSet("non-interactive") {
		imageDataStores := getCommaSeparatedStringFromStringArray(deployment.ImageDatastores)
		securityGroups := getCommaSeparatedStringFromStringArray(deployment.Auth.SecurityGroups)

		fmt.Printf("%s\t%s\t%s\t%t\t%s\t%s\t%t\t%s\n", deployment.ID, deployment.State,
			imageDataStores, deployment.UseImageDatastoreForVms, deployment.SyslogEndpoint,
			deployment.NTPEndpoint, deployment.LoadBalancerEnabled,
			deployment.LoadBalancerAddress)

		fmt.Printf("%t\t%s\t%s\t%d\t%s\n", deployment.Auth.Enabled, deployment.Auth.Endpoint,
			deployment.Auth.Tenant, deployment.Auth.Port, securityGroups)

	} else {
		syslogEndpoint := deployment.SyslogEndpoint
		if len(deployment.SyslogEndpoint) == 0 {
			syslogEndpoint = "-"
		}
		ntpEndpoint := deployment.NTPEndpoint
		if len(deployment.NTPEndpoint) == 0 {
			ntpEndpoint = "-"
		}

		fmt.Printf("\n")
		fmt.Printf("Deployment ID: %s\n", deployment.ID)
		fmt.Printf("  State:                       %s\n", deployment.State)
		fmt.Printf("\n  Image Datastores:            %s\n", deployment.ImageDatastores)
		fmt.Printf("  Use image datastore for vms: %t\n", deployment.UseImageDatastoreForVms)
		fmt.Printf("\n  Syslog Endpoint:             %s\n", syslogEndpoint)
		fmt.Printf("  Ntp Endpoint:                %s\n", ntpEndpoint)
		fmt.Printf("\n  LoadBalancer:\n")
		fmt.Printf("    Enabled:                   %t\n", deployment.LoadBalancerEnabled)
		if deployment.LoadBalancerEnabled {
			fmt.Printf("    Address:                   %s\n", deployment.LoadBalancerAddress)
		}

		fmt.Printf("\n  Auth:\n")
		fmt.Printf("    Enabled:                   %t\n", deployment.Auth.Enabled)
		if deployment.Auth.Enabled {
			fmt.Printf("    Endpoint:                  %s\n", deployment.Auth.Endpoint)
			fmt.Printf("    Tenant:                    %s\n", deployment.Auth.Tenant)
			fmt.Printf("    Port:                      %d\n", deployment.Auth.Port)
			fmt.Printf("    Securitygroups:            %v\n", deployment.Auth.SecurityGroups)
		}
	}

	if deployment.Stats != nil {
		stats := deployment.Stats
		if c.GlobalIsSet("non-interactive") {
			fmt.Printf("%t\t%s\t%d\n", stats.Enabled, stats.StoreEndpoint, stats.StorePort)
		} else if !utils.NeedsFormatting(c) {

			fmt.Printf("\n  Stats:\n")
			fmt.Printf("    Enabled:               %t\n", stats.Enabled)
			if stats.Enabled {
				fmt.Printf("    Store Endpoint:        %s\n", stats.StoreEndpoint)
				fmt.Printf("    Store Port:            %d\n", stats.StorePort)
			}
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
				migration.DataMigrationCycleSize, migration.VibsUploaded, migration.VibsUploading+migration.VibsUploaded)
		} else if !utils.NeedsFormatting(c) {
			fmt.Printf("\n  Migration status:\n")
			fmt.Printf("    Completed data migration cycles:          %d\n", migration.CompletedDataMigrationCycles)
			fmt.Printf("    Current data migration cycles progress:   %d / %d\n", migration.DataMigrationCycleProgress,
				migration.DataMigrationCycleSize)
			fmt.Printf("    VIB upload progress:                      %d / %d\n", migration.VibsUploaded, migration.VibsUploading+migration.VibsUploaded)
		}
	} else {
		if c.GlobalIsSet("non-interactive") {
			fmt.Printf("\n")
		}
	}

	if deployment.ServiceConfigurations != nil && len(deployment.ServiceConfigurations) != 0 {
		if c.GlobalIsSet("non-interactive") {
			clusterConfigurations := []string{}
			for _, c := range deployment.ServiceConfigurations {
				clusterConfigurations = append(clusterConfigurations, fmt.Sprintf("%s\t%s", c.Type, c.ImageID))
			}
			scriptClusterConfigurations := strings.Join(clusterConfigurations, ",")
			fmt.Printf("%s\n", scriptClusterConfigurations)
		} else if !utils.NeedsFormatting(c) {
			fmt.Println("\n  Cluster Configurations:")
			for i, c := range deployment.ServiceConfigurations {
				fmt.Printf("    ClusterConfiguration %d:\n", i+1)
				fmt.Println("      Kind:     ", c.Kind)
				fmt.Println("      Type:     ", c.Type)
				fmt.Println("      ImageID:  ", c.ImageID)
			}
		}
	} else {
		if c.GlobalIsSet("non-interactive") {
			fmt.Printf("\n")
		} else if !utils.NeedsFormatting(c) {
			fmt.Println("\n  Cluster Configurations:")
			fmt.Printf("    No cluster is supported")
		}
	}

	if !utils.NeedsFormatting(c) {
		err = displayDeploymentSummary(data, c.GlobalIsSet("non-interactive"))
		if err != nil {
			return err
		}
	}

	return nil
}

// Lists all the hosts associated with the deployment
func listDeploymentHosts(c *cli.Context, w io.Writer) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	hosts, err := client.Photonclient.Deployments.GetHosts(id)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		utils.FormatObjects(hosts, w, c)
	} else {
		err = printHostList(hosts.Items, os.Stdout, c)
		if err != nil {
			return err
		}
	}

	return nil
}

// Lists all the hosts associated with the deployment
func listDeploymentVms(c *cli.Context, w io.Writer) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	vms, err := client.Photonclient.Deployments.GetVms(id)
	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		utils.FormatObjects(vms, w, c)
	} else {
		err = printVMList(vms.Items, os.Stdout, c, false)
		if err != nil {
			return err
		}
	}

	return nil
}

// Update the image datastores using the information carried in cli.Context.
func updateImageDatastores(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	datastores := c.String("datastores")

	if !c.GlobalIsSet("non-interactive") {
		var err error
		datastores, err = askForInput("Datastores: ", datastores)
		if err != nil {
			return err
		}
	}

	if len(datastores) == 0 {
		return fmt.Errorf("Please provide datastores using --datastores flag")
	}

	imageDataStores := &photon.ImageDatastores{
		Items: regexp.MustCompile(`\s*,\s*`).Split(datastores, -1),
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Photonclient.Deployments.SetImageDatastores(id, imageDataStores)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = deploymentJsonHelper(c, id, client.Photonclient)
	if err != nil {
		return err
	}

	return nil
}

// Synchronizes hosts configurations
func syncHostsConfig(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Photonclient.Deployments.SyncHostsConfig(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = deploymentJsonHelper(c, id, client.Photonclient)
	if err != nil {
		return err
	}

	return nil
}

// Sends a pause system task to client
func pauseSystem(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	pauseSystemTask, err := client.Photonclient.Deployments.PauseSystem(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(pauseSystemTask.ID, c)
	if err != nil {
		return err
	}

	err = deploymentJsonHelper(c, id, client.Photonclient)
	if err != nil {
		return err
	}

	return nil
}

// Sends a pause background task to client
func pauseBackgroundTasks(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	pauseBackgroundTask, err := client.Photonclient.Deployments.PauseBackgroundTasks(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(pauseBackgroundTask.ID, c)
	if err != nil {
		return err
	}

	err = deploymentJsonHelper(c, id, client.Photonclient)
	if err != nil {
		return err
	}

	return nil
}

// Sends a resume system task to client
func resumeSystem(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	resumeSystemTask, err := client.Photonclient.Deployments.ResumeSystem(id)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(resumeSystemTask.ID, c)
	if err != nil {
		return err
	}

	err = deploymentJsonHelper(c, id, client.Photonclient)
	if err != nil {
		return err
	}

	return nil
}

// Set security groups for a deployment
func setDeploymentSecurityGroups(c *cli.Context) error {
	var err error
	var deploymentId string
	var groups string

	// We have two cases:
	// Case 1: arguments are: id groups
	// Case 2: arguments are: groups
	if len(c.Args()) == 2 {
		deploymentId = c.Args()[0]
		groups = c.Args()[1]
	} else if len(c.Args()) == 1 {
		deploymentId, err = getDefaultDeploymentId(c)
		if err != nil {
			return err
		}
		groups = c.Args()[0]
	} else {
		return fmt.Errorf("Usage: deployments set-security-groups [id] groups")
	}

	items := regexp.MustCompile(`\s*,\s*`).Split(groups, -1)
	securityGroups := &photon.SecurityGroupsSpec{
		Items: items,
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Photonclient.Deployments.SetSecurityGroups(deploymentId, securityGroups)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = deploymentJsonHelper(c, deploymentId, client.Photonclient)
	if err != nil {
		return err
	}

	return nil
}

//Enable cluster type for the specified deployment id
func enableClusterType(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	clusterType := c.String("type")
	imageID := c.String("image-id")

	if !c.GlobalIsSet("non-interactive") {
		var err error
		clusterType, err = askForInput("Cluster Type: ", clusterType)
		if err != nil {
			return err
		}
		imageID, err = askForInput("Image ID: ", imageID)
		if err != nil {
			return err
		}
	}

	if len(id) == 0 {
		return fmt.Errorf("Please provide deployment id")
	}
	if len(clusterType) == 0 {
		return fmt.Errorf("Please provide cluster type using --type flag")
	}

	if len(imageID) == 0 {
		return fmt.Errorf("Please provide image ID using --image-id flag")
	}

	if confirmed(c) {
		client.Photonclient, err = client.GetClient(c)
		if err != nil {
			return err
		}
		clusterConfigSpec := &photon.ServiceConfigurationSpec{
			Type:    clusterType,
			ImageID: imageID,
		}

		task, err := client.Photonclient.Deployments.EnableServiceType(id, clusterConfigSpec)
		if err != nil {
			return err
		}

		_, err = waitOnTaskOperation(task.ID, c)
		if err != nil {
			return err
		}

		err = deploymentJsonHelper(c, id, client.Photonclient)
		if err != nil {
			return err
		}

	} else {
		fmt.Println("Cancelled")
	}
	return nil
}

//Disable cluster type for the specified deployment id
func disableClusterType(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	clusterType := c.String("type")

	if !c.GlobalIsSet("non-interactive") {
		var err error
		clusterType, err = askForInput("Cluster Type: ", clusterType)
		if err != nil {
			return err
		}
	}

	if len(id) == 0 {
		return fmt.Errorf("Please provide deployment id")
	}
	if len(clusterType) == 0 {
		return fmt.Errorf("Please provide cluster type using --type flag")
	}

	if confirmed(c) {
		client.Photonclient, err = client.GetClient(c)
		if err != nil {
			return err
		}

		clusterConfigSpec := &photon.ServiceConfigurationSpec{
			Type: clusterType,
		}

		task, err := client.Photonclient.Deployments.DisableServiceType(id, clusterConfigSpec)
		if err != nil {
			return err
		}

		_, err = waitOnTaskOperation(task.ID, c)
		if err != nil {
			return err
		}

		err = deploymentJsonHelper(c, id, client.Photonclient)
		if err != nil {
			return err
		}

	} else {
		fmt.Println("Cancelled")
	}
	return nil
}

// Starts the recurring copy state of source system into destination
func deploymentMigrationPrepare(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	sourceAddress := c.String("endpoint")
	if len(sourceAddress) == 0 {
		return fmt.Errorf("Please provide the API endpoint of the old control plane")
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deployment, err := client.Photonclient.Deployments.Get(id)
	if err != nil {
		return err
	}
	initializeMigrationSpec := photon.InitializeMigrationOperation{}
	initializeMigrationSpec.SourceNodeGroupReference = sourceAddress

	// Initialize deployment migration
	initializeMigrate, err := client.Photonclient.Deployments.InitializeDeploymentMigration(&initializeMigrationSpec, deployment.ID)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(initializeMigrate.ID, c)
	if err != nil {
		return err
	}

	if !utils.NeedsFormatting(c) {
		fmt.Printf("Deployment '%s' migration started [source management endpoint: '%s'].\n", deployment.ID, sourceAddress)
	}

	err = deploymentJsonHelper(c, id, client.Photonclient)
	if err != nil {
		return err
	}

	return nil
}

// Finishes the copy state of source system into destination and makes this system the active one
func deploymentMigrationFinalize(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	sourceAddress := c.String("endpoint")
	if len(sourceAddress) == 0 {
		return fmt.Errorf("Please provide the API endpoint of the old control plane")
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deployment, err := client.Photonclient.Deployments.Get(id)
	if err != nil {
		return err
	}
	finalizeMigrationSpec := photon.FinalizeMigrationOperation{}
	finalizeMigrationSpec.SourceNodeGroupReference = sourceAddress

	// Finalize deployment migration
	finalizeMigrate, err := client.Photonclient.Deployments.FinalizeDeploymentMigration(&finalizeMigrationSpec, deployment.ID)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(finalizeMigrate.ID, c)
	if err != nil {
		return err
	}

	err = deploymentJsonHelper(c, id, client.Photonclient)
	if err != nil {
		return err
	}

	return nil
}

// displays the migration status
func showMigrationStatus(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deployment, err := client.Photonclient.Deployments.Get(id)
	if err != nil {
		return err
	}

	if deployment.Migration == nil {
		fmt.Print("No migration information available")
		return nil
	}

	migration := deployment.Migration
	if c.GlobalIsSet("non-interactive") {
		fmt.Printf("%d\t%d\t%d\t%d\t%d\n", migration.CompletedDataMigrationCycles, migration.DataMigrationCycleProgress,
			migration.DataMigrationCycleSize, migration.VibsUploaded, migration.VibsUploading+migration.VibsUploaded)
	} else if !utils.NeedsFormatting(c) {
		fmt.Printf("  Migration status:\n")
		fmt.Printf("    Completed data migration cycles:          %d\n", migration.CompletedDataMigrationCycles)
		fmt.Printf("    Current data migration cycles progress:   %d / %d\n", migration.DataMigrationCycleProgress,
			migration.DataMigrationCycleSize)
		fmt.Printf("    VIB upload progress:                      %d / %d\n", migration.VibsUploaded, migration.VibsUploading+migration.VibsUploaded)
	} else {
		err = deploymentJsonHelper(c, id, client.Photonclient)
		if err != nil {
			return err
		}
	}

	return nil
}

// Retrieves the deployment id from the first command line argument or if it was not provided attempts to
// find it by using the "list" API. The "automatic" retrieval assumes that there is only one deployment object present.
func getDeploymentId(c *cli.Context) (id string, err error) {
	if len(c.Args()) > 1 {
		err = fmt.Errorf("Unknown arguments: %v.", c.Args()[1:])
		return
	}

	if len(c.Args()) == 1 {
		id = c.Args().First()
		return
	}

	return getDefaultDeploymentId(c)
}

// If there is exactly one deployment, return its id, otherwise return an error
func getDefaultDeploymentId(c *cli.Context) (id string, err error) {
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return
	}

	deployments, err := client.Photonclient.Deployments.GetAll()
	if err != nil {
		return
	}

	if len(deployments.Items) != 1 {
		err = fmt.Errorf(
			"We were unable to determine the deployment 'id'." +
				"Please make sure a deployment exists and provide the deployment 'id' argument.")
		return
	}

	id = deployments.Items[0].ID
	return
}

func displayDeploymentSummary(data []VM_NetworkIPs, isScripting bool) error {
	deployment_info := make(map[string]map[string][]string)
	for _, d := range data {
		for k, v := range d.vm.Metadata {
			if strings.HasPrefix(k, "CONTAINER_") {
				if _, ok := deployment_info[v]; ok {
					deployment_info[v]["port"] = append(deployment_info[v]["port"], getPort(k))
					deployment_info[v]["ips"] = append(deployment_info[v]["ips"], d.ips)

				} else {
					deployment_info[v] = map[string][]string{"port": []string{getPort(k)}, "ips": []string{d.ips}}
				}
			}
		}
	}
	var keys []string
	for k := range deployment_info {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if isScripting {
		for _, job := range keys {
			ips := removeDuplicates(deployment_info[job]["ips"])
			sort.Strings(ips)
			ports := removeDuplicates(deployment_info[job]["port"])
			sort.Strings(ports)
			fmt.Printf("%s\t%s\t%s\n", job, getCommaSeparatedStringFromStringArray(ips), getCommaSeparatedStringFromStringArray(ports))
		}
		fmt.Printf("\n")
		for _, vmIPs := range data {
			fmt.Printf("%s\t%s\t%s\t%s\n", vmIPs.ips, vmIPs.vm.Host, vmIPs.vm.ID, vmIPs.vm.Name)
		}
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "\n\n")
		fmt.Fprintf(w, "  Job\tVM IP(s)\tPorts\n")
		for _, job := range keys {
			ips := removeDuplicates(deployment_info[job]["ips"])
			sort.Strings(ips)
			scriptIPs := strings.Replace((strings.Trim(fmt.Sprint(ips), "[]")), " ", ", ", -1)
			ports := removeDuplicates(deployment_info[job]["port"])
			sort.Strings(ports)
			scriptPorts := strings.Replace(strings.Trim(fmt.Sprint(ports), "[]"), " ", ", ", -1)
			fmt.Fprintf(w, "  %s\t%s\t%s\n", job, scriptIPs, scriptPorts)
		}

		fmt.Fprintf(w, "\n\n")
		fmt.Fprintf(w, "  VM IP\tHost IP\tVM ID\tVM Name\n")

		sort.Sort(ipsSorter(data))
		for _, vmIPs := range data {
			fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", vmIPs.ips, vmIPs.vm.Host, vmIPs.vm.ID, vmIPs.vm.Name)
		}

		err := w.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}

func getPort(container_port string) string {
	return strings.TrimPrefix(container_port, "CONTAINER_")
}

func removeDuplicates(a []string) []string {
	result := []string{}
	seen := map[string]string{}
	for _, val := range a {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = val
		}
	}
	return result
}

func validateDeploymentArguments(imageDatastoreNames []string, enableAuth bool, authEndpoint string, authPort int, oauthTenant string,
	oauthUsername string, oauthPassword string, oauthSecurityGroups []string, enableVirtualNetwork bool, networkManagerAddress string,
	networkManagerUsername string, networkManagerPassword string, enableStats bool, statsStoreEndpoint string, statsStorePort int) error {
	if len(imageDatastoreNames) == 0 {
		return fmt.Errorf("Image datastore names cannot be nil.")
	}
	if enableAuth {
		if oauthTenant == "" {
			return fmt.Errorf("OAuth tenant cannot be nil when auth is enabled.")
		}
		if oauthUsername == "" {
			return fmt.Errorf("OAuth username cannot be nil when auth is enabled.")
		}
		if oauthPassword == "" {
			return fmt.Errorf("OAuth password cannot be nil when auth is enabled.")
		}
		if len(oauthSecurityGroups) == 0 {
			return fmt.Errorf("OAuth security groups cannot be nil when auth is enabled.")
		}
	}
	if enableVirtualNetwork {
		if networkManagerAddress == "" {
			return fmt.Errorf("Network manager address cannot be nil when virtual network is enabled.")
		}
		if networkManagerUsername == "" {
			return fmt.Errorf("Network manager username cannot be nil when virtual network is enabled.")
		}
		if networkManagerPassword == "" {
			return fmt.Errorf("Network manager password cannot be nil when virtual network is enabled.")
		}
	}
	if enableStats {
		if statsStoreEndpoint == "" {
			return fmt.Errorf("Stats store endpoint cannot be nil when stats is enabled.")
		}
		if statsStorePort == 0 {
			return fmt.Errorf("Stats store port cannot be nil when stats is enabled.")
		}
	}
	return nil
}

func deploymentJsonHelper(c *cli.Context, id string, client *photon.Client) error {
	if utils.NeedsFormatting(c) {
		deployment, err := client.Deployments.Get(id)
		if err != nil {
			return err
		}

		utils.FormatObject(deployment, os.Stdout, c)
	}
	return nil
}
