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
	"sort"
)

type VM_NetworkIPs struct {
	vm  photon.VM
	ips string
}

type ipsSorter []VM_NetworkIPs

func (ip ipsSorter) Len() int           { return len(ip) }
func (ip ipsSorter) Swap(i, j int)      { ip[i], ip[j] = ip[j], ip[i] }
func (ip ipsSorter) Less(i, j int) bool { return ip[i].ips < ip[j].ips }

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
					"   Requires system administrator access for viewing all system information",
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
			{
				Name:      "list-vms",
				Usage:     "Lists all VMs",
				ArgsUsage: " ",
				Description: "List all VMs associated with all tenants and projects.\n" +
					"   Requires system administrator access.",
				Action: func(c *cli.Context) {
					err := listSystemVms(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "configure-nsx",
				Usage:     "Configure NSX for the system",
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
					cli.StringFlag{
						Name:  "dns-server-addresses",
						Usage: "Comma-separated list of DNS server addresses",
					},
				},
				Action: func(c *cli.Context) {
					err := configureNSX(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "enable-service-type",
				Aliases:   []string{"enable-cluster-type"},
				Usage:     "Enable service type for the system",
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
					err := enableSystemServiceType(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "disable-service-type",
				Aliases:   []string{"disable-cluster-type"},
				Usage:     "Disable service type for the system",
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
					err := disableSystemServiceType(c)
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

	systemInfo, err := client.Photonclient.System.GetSystemInfo()

	if err != nil {
		return err
	}

	if utils.NeedsFormatting(c) {
		utils.FormatObject(systemInfo, w, c)
		return nil
	}

	if c.GlobalIsSet("non-interactive") {
		fmt.Printf("%s\t%s\t%s\t%s\n", systemInfo.BaseVersion,
			systemInfo.FullVersion, systemInfo.GitCommitHash, systemInfo.NetworkType)
	} else {
		fmt.Printf("\n  Base Version:                %s\n", systemInfo.BaseVersion)
		fmt.Printf("  Full Version:                %s\n", systemInfo.FullVersion)
		fmt.Printf("  Git Commit Hash:             %s\n", systemInfo.GitCommitHash)
		fmt.Printf("  Network Type:                %s\n", systemInfo.NetworkType)
		fmt.Printf("\n")
	}
	if systemInfo.State != "" {
		vms, err := client.Photonclient.System.GetSystemVms()
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
		if c.GlobalIsSet("non-interactive") {
			imageDataStores := getCommaSeparatedStringFromStringArray(systemInfo.ImageDatastores)
			securityGroups := getCommaSeparatedStringFromStringArray(systemInfo.Auth.SecurityGroups)

			fmt.Printf("%s\t%s\t%t\t%s\t%s\t%t\t%s\n", systemInfo.State,
				imageDataStores, systemInfo.UseImageDatastoreForVms, systemInfo.SyslogEndpoint,
				systemInfo.NTPEndpoint, systemInfo.LoadBalancerEnabled,
				systemInfo.LoadBalancerAddress)

			fmt.Printf("%s\t%s\t%d\t%s\n", systemInfo.Auth.Endpoint,
				systemInfo.Auth.Domain, systemInfo.Auth.Port, securityGroups)

		} else {
			syslogEndpoint := systemInfo.SyslogEndpoint
			if len(systemInfo.SyslogEndpoint) == 0 {
				syslogEndpoint = "-"
			}
			ntpEndpoint := systemInfo.NTPEndpoint
			if len(systemInfo.NTPEndpoint) == 0 {
				ntpEndpoint = "-"
			}

			fmt.Printf("  State:                       %s\n", systemInfo.State)
			fmt.Printf("\n  Image Datastores:            %s\n", systemInfo.ImageDatastores)
			fmt.Printf("  Use image datastore for vms: %t\n", systemInfo.UseImageDatastoreForVms)
			fmt.Printf("\n  Syslog Endpoint:             %s\n", syslogEndpoint)
			fmt.Printf("  Ntp Endpoint:                %s\n", ntpEndpoint)
			fmt.Printf("\n  LoadBalancer:\n")
			fmt.Printf("    Enabled:                   %t\n", systemInfo.LoadBalancerEnabled)
			if systemInfo.LoadBalancerEnabled {
				fmt.Printf("    Address:                   %s\n", systemInfo.LoadBalancerAddress)
			}

			fmt.Printf("\n  Auth:\n")
			fmt.Printf("    Endpoint:                  %s\n", systemInfo.Auth.Endpoint)
			fmt.Printf("    Domain:                    %s\n", systemInfo.Auth.Domain)
			fmt.Printf("    Port:                      %d\n", systemInfo.Auth.Port)
			fmt.Printf("    SecurityGroups:            %v\n", systemInfo.Auth.SecurityGroups)
		}

		if systemInfo.Stats != nil {
			stats := systemInfo.Stats
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

		if systemInfo.ServiceConfigurations != nil && len(systemInfo.ServiceConfigurations) != 0 {
			if c.GlobalIsSet("non-interactive") {
				serviceConfigurations := []string{}
				for _, c := range systemInfo.ServiceConfigurations {
					serviceConfigurations = append(serviceConfigurations, fmt.Sprintf("%s\t%s", c.Type, c.ImageID))
				}
				scriptServiceConfigurations := strings.Join(serviceConfigurations, ",")
				fmt.Printf("%s\n", scriptServiceConfigurations)
			} else if !utils.NeedsFormatting(c) {
				fmt.Println("\n  Service Configurations:")
				for i, c := range systemInfo.ServiceConfigurations {
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
			err = displayInfoSummary(data, c.GlobalIsSet("non-interactive"))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Sends a pause system task to client
func PauseSystem(c *cli.Context) error {
	var err error
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

	err = systemInfoJsonHelper(c, client.Photonclient)
	if err != nil {
		return err
	}

	return nil
}

// Sends a pause background task to client
func PauseBackgroundTasks(c *cli.Context) error {
	var err error
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

	err = systemInfoJsonHelper(c, client.Photonclient)
	if err != nil {
		return err
	}

	return nil
}

// Sends a resume system task to client
func ResumeSystem(c *cli.Context) error {
	var err error
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

	err = systemInfoJsonHelper(c, client.Photonclient)
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

	err = systemInfoJsonHelper(c, client.Photonclient)
	if err != nil {
		return err
	}

	return nil
}

// Lists all the hosts associated with the deployment
func listSystemVms(c *cli.Context, w io.Writer) error {
	var err error
	client.Photonclient, err = client.GetClient(c)

	vms, err := client.Photonclient.System.GetSystemVms()
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

//Configure NSX for the specified deployment id
func configureNSX(c *cli.Context) error {
	var err error
	client.Photonclient, err = client.GetClient(c)

	if err != nil {
		return err
	}

	nsxAddress := c.String("nsx-address")
	nsxUsername := c.String("nsx-username")
	nsxPassword := c.String("nsx-password")
	floatingIpRootRangeStart := c.String("floating-ip-root-range-start")
	floatingIpRootRangeEnd := c.String("floating-ip-root-range-end")
	t0RouterId := c.String("t0-router-id")
	edgeClusterId := c.String("edge-cluster-id")
	overlayTransportZoneId := c.String("overlay-transport-zone-id")
	tunnelIpPoolId := c.String("tunnel-ip-pool-id")
	hostUplinkPnic := c.String("host-uplink-pnic")
	hostUplinkVlanId := c.Int("host-uplink-vlan-id")
	dnsServerAddresses := c.String("dns-server-addresses")

	if len(nsxAddress) == 0 {
		return fmt.Errorf("Please provide IP address of NSX")
	}
	if len(nsxUsername) == 0 {
		return fmt.Errorf("Please provide NSX username")
	}
	if len(nsxPassword) == 0 {
		return fmt.Errorf("Please provide NSX password")
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
	if len(dnsServerAddresses) == 0 {
		return fmt.Errorf("Please provide list of the DNS server addresses")
	}

	dnsServerAddressList := []string{}
	if dnsServerAddresses != "" {
		dnsServerAddressList = regexp.MustCompile(`\s*,\s*`).Split(dnsServerAddresses, -1)
	}

	if confirmed(c) {
		nsxConfigSpec := &photon.NsxConfigurationSpec{
			NsxAddress:             nsxAddress,
			NsxUsername:            nsxUsername,
			NsxPassword:            nsxPassword,
			FloatingIpRootRange:    photon.IpRange{Start: floatingIpRootRangeStart, End: floatingIpRootRangeEnd},
			T0RouterId:             t0RouterId,
			EdgeClusterId:          edgeClusterId,
			OverlayTransportZoneId: overlayTransportZoneId,
			TunnelIpPoolId:         tunnelIpPoolId,
			HostUplinkPnic:         hostUplinkPnic,
			HostUplinkVlanId:       hostUplinkVlanId,
			DnsServerAddresses:     dnsServerAddressList,
		}

		task, err := client.Photonclient.System.ConfigureNsx(nsxConfigSpec)
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
func enableSystemServiceType(c *cli.Context) error {
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

	if len(serviceType) == 0 {
		return fmt.Errorf("Please provide service type using --type flag")
	}

	if len(imageID) == 0 {
		return fmt.Errorf("Please provide image ID using --image-id flag")
	}

	if confirmed(c) {
		var err error
		client.Photonclient, err = client.GetClient(c)
		if err != nil {
			return err
		}
		serviceConfigSpec := &photon.ServiceConfigurationSpec{
			Type:    serviceType,
			ImageID: imageID,
		}

		task, err := client.Photonclient.System.EnableServiceType(serviceConfigSpec)
		if err != nil {
			return err
		}

		_, err = waitOnTaskOperation(task.ID, c)
		if err != nil {
			return err
		}

		err = systemInfoJsonHelper(c, client.Photonclient)
		if err != nil {
			return err
		}

	} else {
		fmt.Println("Cancelled")
	}
	return nil
}

//Disable service type for the specified deployment id
func disableSystemServiceType(c *cli.Context) error {
	var err error
	serviceType := c.String("type")

	if !c.GlobalIsSet("non-interactive") {
		var err error
		serviceType, err = askForInput("Service Type: ", serviceType)
		if err != nil {
			return err
		}
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

		task, err := client.Photonclient.System.DisableServiceType(serviceConfigSpec)
		if err != nil {
			return err
		}

		_, err = waitOnTaskOperation(task.ID, c)
		if err != nil {
			return err
		}

		err = systemInfoJsonHelper(c, client.Photonclient)
		if err != nil {
			return err
		}

	} else {
		fmt.Println("Cancelled")
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

func systemInfoJsonHelper(c *cli.Context, client *photon.Client) error {
	if utils.NeedsFormatting(c) {
		deployment, err := client.System.GetSystemInfo()
		if err != nil {
			return err
		}

		utils.FormatObject(deployment, os.Stdout, c)
	}
	return nil
}

func displayInfoSummary(data []VM_NetworkIPs, isScripting bool) error {
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
