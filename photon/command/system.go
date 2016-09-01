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
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/manifest"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

// Create a cli.command object for command "system"
// Subcommand: status; Usage: system status
// Subcommand: status; Usage: system deploy <dc_map>
func GetSystemCommand() cli.Command {
	command := cli.Command{
		Name:  "system",
		Usage: "options for system operations",
		Subcommands: []cli.Command{
			{
				Name:  "status",
				Usage: "display system status",
				Action: func(c *cli.Context) {
					err := getStatus(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:  "deploy",
				Usage: "Deploy Photon using DC Map",
				Action: func(c *cli.Context) {
					err := deploy(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "addHosts",
				Usage: "Add multiple hosts",
				Action: func(c *cli.Context) {
					err := addHosts(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "destroy",
				Usage: "destroy Photon deployment",
				Action: func(c *cli.Context) {
					err := destroy(c)
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
						Action: func(c *cli.Context) {
							err := deploymentMigrationPrepareDeprecated(c)
							if err != nil {
								log.Fatal("Error: ", err)
							}
						},
					},
					{
						Name:  "finalize",
						Usage: "finalizes the migration",
						Action: func(c *cli.Context) {
							err := deploymentMigrationFinalizeDeprecated(c)
							if err != nil {
								log.Fatal("Error: ", err)
							}
						},
					},
					{
						Name:  "status",
						Usage: "shows the status of the current migration",
						Action: func(c *cli.Context) {
							err := showMigrationStatusDeprecated(c)
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

// Get endpoint in config file and its status
func getStatus(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "system status")
	if err != nil {
		return err
	}
	client.Esxclient, err = client.GetClient(false)
	if err != nil {
		return err
	}

	status, err := client.Esxclient.Status.Get()
	if err != nil {
		return err
	}

	err = printStatus(status)
	if err != nil {
		return err
	}

	return nil
}

// Print out overall status and status of the four components
func printStatus(status *photon.Status) error {
	fmt.Printf("Overall status: %s\n\n", status.Status)
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 4, 4, 2, ' ', 0)
	fmt.Fprintf(w, "Component\tStatus\n")
	for i := 0; i < len(status.Components); i++ {
		fmt.Fprintf(w, "%s\t%s\n", status.Components[i].Component, status.Components[i].Status)
	}
	err := w.Flush()
	if err != nil {
		return err
	}
	return nil
}

// Deploy Photon Controller based on DC_map
func deploy(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "system deploy <file>")
	if err != nil {
		return err
	}
	file := c.Args().First()
	dcMap, err := manifest.LoadInstallation(file)
	if err != nil {
		return err
	}

	client.Esxclient, err = client.GetClient(false)
	if err != nil {
		return err
	}

	deploymentID, err := createDeploymentFromDcMap(dcMap)
	if err != nil {
		return err
	}

	// Create Hosts
	err = createHostsFromDcMap(dcMap, deploymentID)
	if err != nil {
		return err
	}

	// Deploy
	err = doDeploy(dcMap, deploymentID)
	if err != nil {
		return err
	}

	return nil
}

// Add most hosts in batch mode
func addHosts(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "system addHosts <file>")
	if err != nil {
		return err
	}
	file := c.Args().First()
	dcMap, err := manifest.LoadInstallation(file)
	if err != nil {
		return err
	}

	client.Esxclient, err = client.GetClient(false)
	if err != nil {
		return err
	}

	deployments, err := client.Esxclient.Deployments.GetAll()
	deploymentID := deployments.Items[0].ID

	// Create Hosts
	err = createHostsInBatch(dcMap, deploymentID)
	if err != nil {
		return err
	}

	return nil
}

// Destroy a Photon Controller deployment
func destroy(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "system destroy")
	if err != nil {
		return err
	}

	client.Esxclient, err = client.GetClient(false)
	if err != nil {
		return err
	}

	deployments, err := client.Esxclient.Deployments.GetAll()

	// Destroy deployment
	for _, deployment := range deployments.Items {
		err = doDestroy(deployment.ID)
		if err != nil {
			return err
		}
	}

	// Delete hosts
	for _, deployment := range deployments.Items {
		hosts, err := client.Esxclient.Deployments.GetHosts(deployment.ID)
		if err != nil {
			return err
		}

		for _, host := range hosts.Items {
			deleteTask, err := client.Esxclient.Hosts.Delete(host.ID)
			if err != nil {
				return err
			}

			deleteTask, err = pollTask(deleteTask.ID)
			if err != nil {
				return err
			}
			fmt.Printf("Host has been deleted: ID = %s\n", deleteTask.Entity.ID)
		}
	}

	// Delete availability-zones
	zones, err := client.Esxclient.AvailabilityZones.GetAll()
	if err != nil {
		return err
	}
	for _, zone := range zones.Items {
		deleteTask, err := client.Esxclient.AvailabilityZones.Delete(zone.ID)
		if err != nil {
			return err
		}

		deleteTask, err = pollTask(deleteTask.ID)
		if err != nil {
			return err
		}
		fmt.Printf("AvailabilityZone has been deleted: ID = %s\n", deleteTask.Entity.ID)
	}

	// Delete deployment doc
	for _, deployment := range deployments.Items {
		deleteTask, err := client.Esxclient.Deployments.Delete(deployment.ID)
		if err != nil {
			return err
		}

		task, err := pollTask(deleteTask.ID)
		if err != nil {
			return err
		}
		fmt.Printf("Deleted deployment %s\n", task.Entity.ID)
	}

	return nil
}

// Starts the recurring copy state of source system into destination
func deploymentMigrationPrepareDeprecated(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "system migration prepare <old_management_endpoint>")
	if err != nil {
		return err
	}
	sourceAddress := c.Args().First()
	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}
	deployments, err := client.Esxclient.Deployments.GetAll()
	if err != nil {
		return err
	}
	initializeMigrationSpec := photon.InitializeMigrationOperation{}
	initializeMigrationSpec.SourceLoadBalancerAddress = sourceAddress

	// Initialize deployment migration
	for _, deployment := range deployments.Items {
		initializeMigrate, err := client.Esxclient.Deployments.InitializeDeploymentMigration(&initializeMigrationSpec, deployment.ID)
		if err != nil {
			return err
		}
		_, err = pollTask(initializeMigrate.ID)
		if err != nil {
			return err
		}
		fmt.Printf("Deployment '%s' migration started [source management endpoint: '%s'].\n", deployment.ID, sourceAddress)

		return nil
	}

	return nil
}

// Finishes the copy state of source system into destination and makes this system the active one
func deploymentMigrationFinalizeDeprecated(c *cli.Context) error {
	fmt.Printf("'%d'", len(c.Args()))
	err := checkArgNum(c.Args(), 1, "system migration finalize <old_management_endpoint>")
	if err != nil {
		return err
	}
	sourceAddress := c.Args().First()
	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}
	deployments, err := client.Esxclient.Deployments.GetAll()
	if err != nil {
		return err
	}
	finalizeMigrationSpec := photon.FinalizeMigrationOperation{}
	finalizeMigrationSpec.SourceLoadBalancerAddress = sourceAddress

	// Finalize deployment migration
	for _, deployment := range deployments.Items {
		finalizeMigrate, err := client.Esxclient.Deployments.FinalizeDeploymentMigration(&finalizeMigrationSpec, deployment.ID)
		if err != nil {
			return err
		}
		_, err = pollTask(finalizeMigrate.ID)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

// displays the migration status
func showMigrationStatusDeprecated(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "migration status")
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

	for _, deployment := range deployments.Items {
		if deployment.Migration != nil {
			migration := deployment.Migration
			if c.GlobalIsSet("non-interactive") {
				fmt.Printf("%d\t%d\t%d\t%d\t%d\n", migration.CompletedDataMigrationCycles, migration.DataMigrationCycleProgress,
					migration.DataMigrationCycleSize, migration.VibsUploaded, migration.VibsUploading+migration.VibsUploaded)
			} else {
				fmt.Printf("  Migration status:\n")
				fmt.Printf("    Completed data migration cycles:          %d\n", migration.CompletedDataMigrationCycles)
				fmt.Printf("    Current data migration cycles progress:   %d / %d\n", migration.DataMigrationCycleProgress,
					migration.DataMigrationCycleSize)
				fmt.Printf("    VIB upload progress:                      %d / %d\n", migration.VibsUploaded, migration.VibsUploading+migration.VibsUploaded)
			}
		}
		return nil
	}
	return nil
}

func createDeploymentFromDcMap(dcMap *manifest.Installation) (deploymentID string, err error) {
	err = validateDeploymentArguments(
		dcMap.Deployment.ImageDatastores, dcMap.Deployment.AuthEnabled,
		dcMap.Deployment.AuthTenant, dcMap.Deployment.AuthUsername, dcMap.Deployment.AuthPassword,
		dcMap.Deployment.AuthSecurityGroups, dcMap.Deployment.SdnEnabled,
		dcMap.Deployment.NetworkManagerAddress, dcMap.Deployment.NetworkManagerUsername,
		dcMap.Deployment.NetworkManagerPassword,
		dcMap.Deployment.StatsEnabled, dcMap.Deployment.StatsStoreEndpoint,
		dcMap.Deployment.StatsPort)
	if err != nil {
		return "", err
	}

	lbEnabledString := dcMap.Deployment.LoadBalancerEnabled
	lbEnabled := true
	if len(lbEnabledString) > 0 {
		lbEnabled, err = strconv.ParseBool(lbEnabledString)
		if err != nil {
			return "", err
		}
	}

	authInfo := &photon.AuthInfo{
		Enabled:        dcMap.Deployment.AuthEnabled,
		Tenant:         dcMap.Deployment.AuthTenant,
		Username:       dcMap.Deployment.AuthUsername,
		Password:       dcMap.Deployment.AuthPassword,
		SecurityGroups: dcMap.Deployment.AuthSecurityGroups,
	}
	networkConfiguration := &photon.NetworkConfigurationCreateSpec{
		Enabled:         dcMap.Deployment.SdnEnabled,
		Address:         dcMap.Deployment.NetworkManagerAddress,
		Username:        dcMap.Deployment.NetworkManagerUsername,
		Password:        dcMap.Deployment.NetworkManagerPassword,
		NetworkZoneId:   dcMap.Deployment.NetworkZoneId,
		TopRouterId:     dcMap.Deployment.NetworkTopRouterId,
		IpRange:         dcMap.Deployment.NetworkIpRange,
		FloatingIpRange: dcMap.Deployment.NetworkFloatingIpRange,
		DhcpServers:     dcMap.Deployment.NetworkDhcpServers,
	}
	statsInfo := &photon.StatsInfo{
		Enabled:       dcMap.Deployment.StatsEnabled,
		StoreEndpoint: dcMap.Deployment.StatsStoreEndpoint,
		StorePort:     dcMap.Deployment.StatsPort,
	}

	deploymentSpec := &photon.DeploymentCreateSpec{
		Auth:                 authInfo,
		NetworkConfiguration: networkConfiguration,
		ImageDatastores:      dcMap.Deployment.ImageDatastores,
		NTPEndpoint:          dcMap.Deployment.NTPEndpoint,
		SyslogEndpoint:       dcMap.Deployment.SyslogEndpoint,
		Stats:                statsInfo,
		UseImageDatastoreForVms: dcMap.Deployment.UseImageDatastoreForVms,
		LoadBalancerEnabled:     lbEnabled,
	}

	createDeploymentTask, err := client.Esxclient.Deployments.Create(deploymentSpec)
	if err != nil {
		return "", err
	}

	task, err := pollTask(createDeploymentTask.ID)
	if err != nil {
		return "", err
	}
	fmt.Printf("Created deployment %s\n", task.Entity.ID)
	return task.Entity.ID, nil
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func createAvailabilityZonesFromDcMap(dcMap *manifest.Installation) (map[string]string, error) {
	availabilityZoneNameToIdMap := make(map[string]string)
	for _, host := range dcMap.Hosts {
		if len(host.AvailabilityZone) > 0 {
			if _, present := availabilityZoneNameToIdMap[host.AvailabilityZone]; !present {
				availabilityZoneSpec := &photon.AvailabilityZoneCreateSpec{
					Name: host.AvailabilityZone,
				}

				createAvailabilityZoneTask, err := client.Esxclient.AvailabilityZones.Create(availabilityZoneSpec)
				if err != nil {
					return nil, err
				}

				task, err := pollTask(createAvailabilityZoneTask.ID)
				if err != nil {
					return nil, err
				}
				availabilityZoneNameToIdMap[host.AvailabilityZone] = task.Entity.ID
			}
		}
	}
	return availabilityZoneNameToIdMap, nil
}

func createHostsFromDcMap(dcMap *manifest.Installation, deploymentID string) error {
	hostSpecs, err := createHostSpecs(dcMap)
	if err != nil {
		return err
	}

	for _, spec := range hostSpecs {
		createHostTask, err := client.Esxclient.Hosts.Create(&spec, deploymentID)
		if err != nil {
			return err
		}

		task, err := pollTask(createHostTask.ID)
		if err != nil {
			return err
		}
		fmt.Printf("Host with ip '%s' created: ID = %s\n", spec.Address, task.Entity.ID)
	}

	return nil
}

func createHostsInBatch(dcMap *manifest.Installation, deploymentID string) error {
	hostSpecs, err := createHostSpecs(dcMap)
	if err != nil {
		return err
	}

	createTaskMap := make(map[string]*photon.Task)
	var creationErrors []error
	var pollErrors []error
	for _, spec := range hostSpecs {
		createHostTask, err := client.Esxclient.Hosts.Create(&spec, deploymentID)
		if err != nil {
			creationErrors = append(creationErrors, err)
			fmt.Printf("Creation of Host document with ip '%s' failed: with err '%s'\n",
				spec.Address, err)
		} else {
			createTaskMap[spec.Address] = createHostTask
		}
	}

	for address, createTask := range createTaskMap {
		task, err := pollTask(createTask.ID)
		if err != nil {
			pollErrors = append(pollErrors, err)
			fmt.Printf("Creation of Host with ip '%s' failed: ID = %s with err '%s'\n\n",
				address, task.ID, err)
		} else {
			fmt.Printf("Host with ip '%s' created: ID = %s\n\n", address, task.Entity.ID)
		}
	}
	return nil
}

func createHostSpecs(dcMap *manifest.Installation) ([]photon.HostCreateSpec, error) {
	availabilityZoneNameToIdMap, err := createAvailabilityZonesFromDcMap(dcMap)
	if err != nil {
		return nil, err
	}
	var hostSpecs []photon.HostCreateSpec
	var managementNetworkIps []string
	for _, host := range dcMap.Hosts {
		hostIps, err := parseIpRanges(host.IpRanges)
		if err != nil {
			return nil, err
		}
		if hostIps == nil || len(hostIps) == 0 {
			return nil, errors.New("Host IP Address missing in DC Map")
		}

		if host.Metadata != nil {
			if managementVmIps, exists := host.Metadata["MANAGEMENT_VM_IPS"]; exists {
				managementNetworkIps, err = parseIpRanges(managementVmIps)
				if err != nil {
					return nil, err
				}
			}
		}

		for i := 0; i < len(hostIps); i++ {
			hostIp := hostIps[i]
			metaData := host.Metadata
			if host.Metadata != nil {
				if _, exists := host.Metadata["MANAGEMENT_VM_IPS"]; exists {
					metaData = make(map[string]string)
					for key, value := range host.Metadata {
						metaData[key] = value
					}
					delete(metaData, "MANAGEMENT_VM_IPS")
					if i < len(managementNetworkIps) {
						metaData["MANAGEMENT_NETWORK_IP"] = managementNetworkIps[i]
					}
				}
			}
			hostSpec := photon.HostCreateSpec{
				Username:         host.Username,
				Password:         host.Password,
				Address:          hostIp,
				AvailabilityZone: availabilityZoneNameToIdMap[host.AvailabilityZone],
				Tags:             host.Tags,
				Metadata:         metaData,
			}
			hostSpecs = append(hostSpecs, hostSpec)
		}
	}

	return hostSpecs, nil
}

func parseIpRanges(ipRanges string) ([]string, error) {
	var ipList []string
	for _, ipRange := range regexp.MustCompile(`\s*,\s*`).Split(ipRanges, -1) {
		ips := regexp.MustCompile(`\s*-\s*`).Split(ipRange, -1)
		if len(ips) == 1 {
			ip := net.ParseIP(ips[0]).To4()
			if ip == nil {
				return nil, errors.New("Bad IP Address defined in DC Map")
			}
			ipList = append(ipList, ips[0])
		} else if len(ips) == 2 {
			ip0 := net.ParseIP(ips[0]).To4()
			ip1 := net.ParseIP(ips[1]).To4()
			if ip0 == nil || ip1 == nil {
				return nil, errors.New("Bad IP Address defined in DC Map")
			}
			for ip := ip0; bytes.Compare(ip, ip1) <= 0; inc(ip) {
				ipList = append(ipList, ip.String())
			}
		} else {
			return nil, errors.New("Bad Address Range defined in DC Map")
		}
	}
	return ipList, nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func doDeploy(installSpec *manifest.Installation, deploymentID string) error {
	var desiredState string
	if installSpec.Deployment.ResumeSystem {
		desiredState = "READY"
	} else {
		desiredState = "PAUSED"
	}
	deploymentDeployOperation := &photon.DeploymentDeployOperation{
		DesiredState: desiredState,
	}
	deployTask, err := client.Esxclient.Deployments.Deploy(deploymentID, deploymentDeployOperation)
	if err != nil {
		return err
	}

	_, err = pollTaskWithTimeout(client.Esxclient, deployTask.ID, 120*time.Minute)
	if err != nil {
		return err
	}

	fmt.Printf("Deployment '%s' is complete.\n", deploymentID)
	return nil
}

func doDestroy(deploymentID string) error {
	destroyTask, err := client.Esxclient.Deployments.Destroy(deploymentID)
	if err != nil {
		return err
	}

	_, err = pollTask(destroyTask.ID)
	if err != nil {
		return err
	}
	fmt.Printf("Deployment '%s' is destroyed.\n", deploymentID)

	return nil
}
