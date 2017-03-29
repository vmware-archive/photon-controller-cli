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
	"strings"
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
				Name:      "info",
				Usage:     "Show system info",
				ArgsUsage: " ",
				Description: "Show detailed information about the system.\n" +
					"   Requires system administrator access,",
				Action: func(c *cli.Context) {
					err := showSystemInfo(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "pause",
				Usage:     "Pause system",
				ArgsUsage: " ",
				Description: "Pause Photon Controller. All incoming requests that modify the system\n" +
					"   state (other than resume) will be refused. This implies pause-background-states" +
					"   Requires system administrator access.",
				Action: func(c *cli.Context) {
					err := PauseSystem(c)
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
					err := PauseBackgroundTasks(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "resume",
				Usage:     "Resume system",
				ArgsUsage: " ",
				Description: "Resume Photon Controller after it has been paused.\n" +
					"   Requires system administrator access.",
				Action: func(c *cli.Context) {
					err := ResumeSystem(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "set-security-groups",
				Usage:     "Set security groups for the system",
				ArgsUsage: "<comma separate list of security groups>",
				Description: "Provide the list of Lightwave groups that contain the people who are\n" +
					"   allowed to be system administrators. Be careful: providing the wrong group could remove\n" +
					"   your access.\n\n" +
					"   A security group specifies both the Lightwave domain and Lightwave group.\n" +
					"   For example, a security group may be photon.vmware.com\\group-1\n\n" +
					"   Example: photon deployment set-security-groups 'photon.vmware.com\\group-1,photon.vmware.com\\group-2'\n\n" +
					"   Requires system administrator access.",
				Action: func(c *cli.Context) {
					err := setSystemSecurityGroups(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
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

	status, err := client.Photonclient.System.GetSystemStatus()
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

// Retrieves information about a system
func showSystemInfo(c *cli.Context, w io.Writer) error {
	var err error
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	deployment, err := client.Photonclient.System.GetSystemInfo()

	if err != nil {
		return err
	}

	vms, err := client.Photonclient.Deployments.GetVms("default")
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

		fmt.Printf("%s\t%s\t%d\t%s\n", deployment.Auth.Endpoint,
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
		fmt.Printf("    Endpoint:                  %s\n", deployment.Auth.Endpoint)
		fmt.Printf("    Tenant:                    %s\n", deployment.Auth.Tenant)
		fmt.Printf("    Port:                      %d\n", deployment.Auth.Port)
		fmt.Printf("    SecurityGroups:            %v\n", deployment.Auth.SecurityGroups)
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

	if deployment.ServiceConfigurations != nil && len(deployment.ServiceConfigurations) != 0 {
		if c.GlobalIsSet("non-interactive") {
			serviceConfigurations := []string{}
			for _, c := range deployment.ServiceConfigurations {
				serviceConfigurations = append(serviceConfigurations, fmt.Sprintf("%s\t%s", c.Type, c.ImageID))
			}
			scriptServiceConfigurations := strings.Join(serviceConfigurations, ",")
			fmt.Printf("%s\n", scriptServiceConfigurations)
		} else if !utils.NeedsFormatting(c) {
			fmt.Println("\n  Service Configurations:")
			for i, c := range deployment.ServiceConfigurations {
				fmt.Printf("    ServiceConfiguration %d:\n", i+1)
				fmt.Println("      Kind:     ", c.Kind)
				fmt.Println("      Type:     ", c.Type)
				fmt.Println("      ImageID:  ", c.ImageID)
			}
		}
	} else {
		if c.GlobalIsSet("non-interactive") {
			fmt.Printf("\n")
		} else if !utils.NeedsFormatting(c) {
			fmt.Println("\n  Service Configurations:")
			fmt.Printf("    No Service is supported")
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

// Sends a pause system task to client
func PauseSystem(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	pauseSystemTask, err := client.Photonclient.System.PauseSystem()
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
func PauseBackgroundTasks(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	pauseBackgroundTask, err := client.Photonclient.System.PauseBackgroundTasks()
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
func ResumeSystem(c *cli.Context) error {
	id, err := getDeploymentId(c)
	if err != nil {
		return err
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	resumeSystemTask, err := client.Photonclient.System.ResumeSystem()
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

// Set security groups for the system
func setSystemSecurityGroups(c *cli.Context) error {
	var err error
	var groups string

	if len(c.Args()) != 1 {
		return fmt.Errorf("Usage: system set-security-group <groups>")
	}

	groups = c.Args()[0]

	items := regexp.MustCompile(`\s*,\s*`).Split(groups, -1)
	securityGroups := &photon.SecurityGroupsSpec{
		Items: items,
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, err := client.Photonclient.System.SetSecurityGroups(securityGroups)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	err = systemInfoJsonHelper(c, "default", client.Photonclient)
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

	// Create Hosts
	err = createHostsInBatch(dcMap)
	if err != nil {
		return err
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

func createZonesFromDcMap(dcMap *manifest.Installation) (map[string]string, error) {
	zoneNameToIdMap := make(map[string]string)
	for _, host := range dcMap.Hosts {
		if len(host.AvailabilityZone) > 0 {
			if _, present := zoneNameToIdMap[host.AvailabilityZone]; !present {
				zoneSpec := &photon.ZoneCreateSpec{
					Name: host.AvailabilityZone,
				}

				createZoneTask, err := client.Photonclient.Zones.Create(zoneSpec)
				if err != nil {
					return nil, err
				}

				task, err := pollTask(createZoneTask.ID)
				if err != nil {
					return nil, err
				}
				zoneNameToIdMap[host.AvailabilityZone] = task.Entity.ID
			}
		}
	}
	return zoneNameToIdMap, nil
}

func createHostsInBatch(dcMap *manifest.Installation) error {
	hostSpecs, err := createHostSpecs(dcMap)
	if err != nil {
		return err
	}

	createTaskMap := make(map[string]*photon.Task)
	var creationErrors []error
	var pollErrors []error
	for _, spec := range hostSpecs {
		createHostTask, err := client.Photonclient.InfraHosts.Create(&spec)
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
	zoneNameToIdMap, err := createZonesFromDcMap(dcMap)
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
				Username: host.Username,
				Password: host.Password,
				Address:  hostIp,
				Zone:     zoneNameToIdMap[host.AvailabilityZone],
				Tags:     host.Tags,
				Metadata: metaData,
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

func systemInfoJsonHelper(c *cli.Context, id string, client *photon.Client) error {
	if utils.NeedsFormatting(c) {
		deployment, err := client.System.GetSystemInfo()
		if err != nil {
			return err
		}

		utils.FormatObject(deployment, os.Stdout, c)
	}
	return nil
}
