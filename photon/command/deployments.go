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
//              list-hosts; Usage: deployment list-hosts [<id>]
//              list-vms;   Usage: deployment list-vms [<id>]

//              update-image-datastores;        Usage: deployment update-image-datastores [<id> <options>]
//              sync-hosts-config;              Usage: deployment sync-hosts-config [<id>]

//              migration prepare;              Usage: deployment prepare migration [<id> <options>]
//              migration finalize;             Usage: deployment finalize migration [<id> <options>]
//              migration status;               Usage: deployment finalize migration [<id>]

//              enable-service-type, enable-cluster-type;            Usage: deployment enable-service-type [<id> <options>]
//              disable-service-type, disable-cluster-type;           Usage: deployment disable-service-type [<id> <options>]

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
				Name:      "configure-nsx",
				Usage:     "Configure NSX for deployment",
				ArgsUsage: " ",
				Description: "Configure NSX for the deployment. This is a one-time operatino and may not be repeated\n" +
					"If you deploy Photon Controller with the installer, you should not need to run this command.\n" +
					"If you deploy Photon Controller with ovftool, you probably need to run this command.",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "nsx-address",
						Usage: "IP address of NSX",
					},
					cli.StringFlag{
						Name:  "nsx-username",
						Usage: "NSX username",
					},
					cli.StringFlag{
						Name:  "nsx-password",
						Usage: "NSX password",
					},
					cli.StringFlag{
						Name:  "private-ip-root-cidr",
						Usage: "Root CIDR of the private IP pool",
					},
					cli.StringFlag{
						Name:  "floating-ip-root-range-start",
						Usage: "Start of the root range of the floating IP pool",
					},
					cli.StringFlag{
						Name:  "floating-ip-root-range-end",
						Usage: "End of the root range of the floating IP pool",
					},
					cli.StringFlag{
						Name:  "t0-router-id",
						Usage: "ID of the T0-Router",
					},
					cli.StringFlag{
						Name:  "edge-cluster-id",
						Usage: "ID of the Edge cluster",
					},
					cli.StringFlag{
						Name:  "overlay-transport-zone-id",
						Usage: "ID of the OVERLAY transport zone",
					},
					cli.StringFlag{
						Name:  "tunnel-ip-pool-id",
						Usage: "ID of the tunnel IP pool",
					},
					cli.StringFlag{
						Name:  "host-uplink-pnic",
						Usage: "Name of the host uplink pnic",
					},
					cli.IntFlag{
						Name:  "host-uplink-vlan-id",
						Usage: "VLAN ID of the host uplink",
					},
				},
				Action: func(c *cli.Context) {
					err := configureNsx(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "enable-service-type",
				Aliases:   []string{"enable-cluster-type"},
				Usage:     "Enable service type for deployment",
				ArgsUsage: " ",
				Description: "Enable a service type (e.g. Kubernetes) and specify the image to be used\n" +
					"   when creating the service.\n" +
					"   Requires system administrator access.",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "type, k",
						Usage: "Service type (accepted values are KUBERNETES or HARBOR)",
					},
					cli.StringFlag{
						Name:  "image-id, i",
						Usage: "ID of the service image",
					},
				},
				Action: func(c *cli.Context) {
					err := enableServiceType(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "disable-service-type",
				Aliases:   []string{"disable-cluster-type"},
				Usage:     "Disable service type for deployment",
				ArgsUsage: " ",
				Description: "Disable a service type (e.g. Kubernetes). Users will no longer be able\n" +
					"   to deploy services of that type, but existing services will be unaffected.\n" +
					"   Requires system administrator access.",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "type, k",
						Usage: "Service type (accepted values are KUBERNETES or HARBOR)",
					},
				},
				Action: func(c *cli.Context) {
					err := disableServiceType(c)
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

//Configure NSX for the specified deployment id
func configureNsx(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	nsxAddress := c.String("nsx-address")
	nsxUsername := c.String("nsx-username")
	nsxPassword := c.String("nsx-password")
	privateIpRootCidr := c.String("private-ip-root-cidr")
	floatingIpRootRangeStart := c.String("floating-ip-root-range-start")
	floatingIpRootRangeEnd := c.String("floating-ip-root-range-end")
	t0RouterId := c.String("t0-router-id")
	edgeClusterId := c.String("edge-cluster-id")
	overlayTransportZoneId := c.String("overlay-transport-zone-id")
	tunnelIpPoolId := c.String("tunnel-ip-pool-id")
	hostUplinkPnic := c.String("host-uplink-pnic")
	hostUplinkVlanId := c.Int("host-uplink-vlan-id")

	if len(nsxAddress) == 0 {
		return fmt.Errorf("Please provide IP address of NSX")
	}
	if len(nsxUsername) == 0 {
		return fmt.Errorf("Please provide NSX username")
	}
	if len(nsxPassword) == 0 {
		return fmt.Errorf("Please provide NSX password")
	}
	if len(privateIpRootCidr) == 0 {
		return fmt.Errorf("Please provide root CIDR of the private IP pool")
	}
	if len(floatingIpRootRangeStart) == 0 {
		return fmt.Errorf("Please provide the start of the root range of the floating IP pool")
	}
	if len(floatingIpRootRangeEnd) == 0 {
		return fmt.Errorf("Please provide the end of the root range of the floating IP pool")
	}
	if len(t0RouterId) == 0 {
		return fmt.Errorf("Please provide the ID of the T0-Router")
	}
	if len(edgeClusterId) == 0 {
		return fmt.Errorf("Please provide the ID of the Edge cluster")
	}
	if len(overlayTransportZoneId) == 0 {
		return fmt.Errorf("Please provide the ID of the OVERLAY transport zone")
	}
	if len(tunnelIpPoolId) == 0 {
		return fmt.Errorf("Please provide the ID of the tunnel IP pool")
	}
	if len(hostUplinkPnic) == 0 {
		return fmt.Errorf("Please provide name of the host uplink pnic")
	}

	if confirmed(c) {
		client.Photonclient, err = client.GetClient(c)
		if err != nil {
			return err
		}

		nsxConfigSpec := &photon.NsxConfigurationSpec{
			NsxAddress:             nsxAddress,
			NsxUsername:            nsxUsername,
			NsxPassword:            nsxPassword,
			PrivateIpRootCidr:      privateIpRootCidr,
			FloatingIpRootRange:    photon.IpRange{Start: floatingIpRootRangeStart, End: floatingIpRootRangeEnd},
			T0RouterId:             t0RouterId,
			EdgeClusterId:          edgeClusterId,
			OverlayTransportZoneId: overlayTransportZoneId,
			TunnelIpPoolId:         tunnelIpPoolId,
			HostUplinkPnic:         hostUplinkPnic,
			HostUplinkVlanId:       hostUplinkVlanId,
		}

		task, err := client.Photonclient.Deployments.ConfigureNsx(id, nsxConfigSpec)
		if err != nil {
			return err
		}

		_, err = waitOnTaskOperation(task.ID, c)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Cancelled")
	}

	return nil
}

//Enable service type for the specified deployment id
func enableServiceType(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	serviceType := c.String("type")
	imageID := c.String("image-id")

	if !c.GlobalIsSet("non-interactive") {
		var err error
		serviceType, err = askForInput("Service Type: ", serviceType)
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
	if len(serviceType) == 0 {
		return fmt.Errorf("Please provide service type using --type flag")
	}

	if len(imageID) == 0 {
		return fmt.Errorf("Please provide image ID using --image-id flag")
	}

	if confirmed(c) {
		client.Photonclient, err = client.GetClient(c)
		if err != nil {
			return err
		}
		serviceConfigSpec := &photon.ServiceConfigurationSpec{
			Type:    serviceType,
			ImageID: imageID,
		}

		task, err := client.Photonclient.Deployments.EnableServiceType(id, serviceConfigSpec)
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

//Disable service type for the specified deployment id
func disableServiceType(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	serviceType := c.String("type")

	if !c.GlobalIsSet("non-interactive") {
		var err error
		serviceType, err = askForInput("Service Type: ", serviceType)
		if err != nil {
			return err
		}
	}

	if len(id) == 0 {
		return fmt.Errorf("Please provide deployment id")
	}
	if len(serviceType) == 0 {
		return fmt.Errorf("Please provide service type using --type flag")
	}

	if confirmed(c) {
		client.Photonclient, err = client.GetClient(c)
		if err != nil {
			return err
		}

		serviceConfigSpec := &photon.ServiceConfigurationSpec{
			Type: serviceType,
		}

		task, err := client.Photonclient.Deployments.DisableServiceType(id, serviceConfigSpec)
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

	deployment, err := client.Photonclient.System.GetSystemInfo()
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

	deployment, err := client.Photonclient.System.GetSystemInfo()
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

	deployment, err := client.Photonclient.System.GetSystemInfo()
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

func deploymentJsonHelper(c *cli.Context, id string, client *photon.Client) error {
	if utils.NeedsFormatting(c) {
		deployment, err := client.System.GetSystemInfo()
		if err != nil {
			return err
		}

		utils.FormatObject(deployment, os.Stdout, c)
	}
	return nil
}
