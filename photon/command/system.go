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
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/manifest"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

// Create a cli.command object for command "system"
// Subcommand: status; Usage: system status
func GetSystemCommand() cli.Command {
	command := cli.Command{
		Name:  "system",
		Usage: "options for system operations",
		Subcommands: []cli.Command{
			{
				Name:      "status",
				Usage:     "Display system status",
				ArgsUsage: " ",
				Action: func(c *cli.Context) {
					err := getStatus(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "add-hosts",
				Usage:     "Add multiple hosts",
				ArgsUsage: "<host-file>",
				Action: func(c *cli.Context) {
					err := addHosts(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Hidden:      true,
				Name:        "addHosts",
				Usage:       "Add multiple hosts",
				ArgsUsage:   "<host-file>",
				Description: "Deprecated, use add-hosts instead",
				Action: func(c *cli.Context) {
					err := addHosts(c)
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
func getStatus(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	status, err := client.Photonclient.Status.Get()
	if err != nil {
		return err
	}

	if !utils.NeedsFormatting(c) {
		err = printStatus(status)
	} else {
		utils.FormatObject(status, w, c)
	}
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

// Add most hosts in batch mode
func addHosts(c *cli.Context) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	file := c.Args().First()
	dcMap, err := manifest.LoadInstallation(file)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deployments, err := client.Photonclient.Deployments.GetAll()
	deploymentID := deployments.Items[0].ID

	// Create Hosts
	err = createHostsInBatch(dcMap, deploymentID)
	if err != nil {
		return err
	}

	return nil
}

// Starts the recurring copy state of source system into destination
func deploymentMigrationPrepareDeprecated(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	sourceAddress := c.Args().First()
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}
	deployments, err := client.Photonclient.Deployments.GetAll()
	if err != nil {
		return err
	}
	initializeMigrationSpec := photon.InitializeMigrationOperation{}
	initializeMigrationSpec.SourceNodeGroupReference = sourceAddress

	// Initialize deployment migration
	for _, deployment := range deployments.Items {
		initializeMigrate, err := client.Photonclient.Deployments.InitializeDeploymentMigration(&initializeMigrationSpec, deployment.ID)
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
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	sourceAddress := c.Args().First()
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}
	deployments, err := client.Photonclient.Deployments.GetAll()
	if err != nil {
		return err
	}
	finalizeMigrationSpec := photon.FinalizeMigrationOperation{}
	finalizeMigrationSpec.SourceNodeGroupReference = sourceAddress

	// Finalize deployment migration
	for _, deployment := range deployments.Items {
		finalizeMigrate, err := client.Photonclient.Deployments.FinalizeDeploymentMigration(&finalizeMigrationSpec, deployment.ID)
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

				createAvailabilityZoneTask, err := client.Photonclient.AvailabilityZones.Create(availabilityZoneSpec)
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

func createHostsInBatch(dcMap *manifest.Installation, deploymentID string) error {
	hostSpecs, err := createHostSpecs(dcMap)
	if err != nil {
		return err
	}

	createTaskMap := make(map[string]*photon.Task)
	var creationErrors []error
	var pollErrors []error
	for _, spec := range hostSpecs {
		createHostTask, err := client.Photonclient.Hosts.Create(&spec, deploymentID)
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
