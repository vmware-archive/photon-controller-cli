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
	"fmt"
	"log"
	"net"
	"regexp"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
	"github.com/vmware/photon-controller-cli/photon/cli/client"

	"errors"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/gopkg.in/yaml.v2"
)

type SecurityGroups struct {
	Items []string `yaml:"items"`
}

type Deployment struct {
	NTPEndpoint             interface{} `yaml:"ntp_endpoint"`
	UseImageDatastoreForVms bool        `yaml:"use_image_datastore_for_vms"`
	SyslogEndpoint          interface{} `yaml:"syslog_endpoint"`
	StatsStoreEndpoint      string      `yaml:"stats_store_endpoint"`
	StatsPort               int         `yaml:"stats_port"`
	StatsEnabled            bool        `yaml:"stats_enabled"`
	ImageDatastores         string      `yaml:"image_datastores"`
	AuthEnabled             bool        `yaml:"auth_enabled"`
	AuthEndpoint            string      `yaml:"oauth_endpoint"`
	AuthUsername            string      `yaml:"oauth_username"`
	AuthPassword            string      `yaml:"oauth_password"`
	AuthTenant              string      `yaml:"oauth_tenant"`
	AuthPort                int         `yaml:"oauth_port"`
	AuthSecurityGroups      []string    `yaml:"oauth_security_groups"`
}

type Host struct {
	Username         string            `yaml:"username"`
	Password         string            `yaml:"password"`
	IpRanges         string            `yaml:"address_ranges"`
	AvailabilityZone string            `yaml:"availability_zone"`
	Tags             []string          `yaml:"usage_tags"`
	Metadata         map[string]string `yaml:"metadata"`
}

type DcMap struct {
	Deployment Deployment `yaml:"deployment"`
	Hosts      []Host     `yaml:"hosts"`
}

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
				Usage: "Deploy Phonton using DC Map",
				Action: func(c *cli.Context) {
					err := deploy(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "destroy",
				Usage: "destroy Phonton deployment",
				Action: func(c *cli.Context) {
					err := destroy(c)
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
	dcMap, err := getDcMap(file)
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
	err = doDeploy(deploymentID)
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
		err = doDetroy(deployment.ID)
		if err != nil {
			return err
		}
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

	return nil
}

// Starts the recurring copy state of source system into destination
func deploymentMigrationPrepare(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "system migration prepare <sourceAddress>")
	if err != nil {
		return err
	}
	sourceAddress := c.Args().First()

	deployments, err := client.Esxclient.Deployments.GetAll()

	// Initialize deployment migration
	for _, deployment := range deployments.Items {
		initializeMigrate, err := client.Esxclient.Deployments.InitializeDeploymentMigration(sourceAddress, deployment.ID)
		if err != nil {
			return err
		}

		_, err = pollTask(initializeMigrate.ID)
		if err != nil {
			return err
		}
		fmt.Printf("Deployment '%s' migration started '%s'.\n", deployment.ID, sourceAddress)

		return nil
	}

	return nil
}

// Finishes the copy state of source system into destination and makes this system the active one
func deploymentMigrationFinalize(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "system migration finalize <sourceAddress>")
	if err != nil {
		return err
	}
	sourceAddress := c.Args().First()

	deployments, err := client.Esxclient.Deployments.GetAll()

	// Initialize deployment migration
	for _, deployment := range deployments.Items {
		finalizeMigrate, err := client.Esxclient.Deployments.FinalizeDeploymentMigration(sourceAddress, deployment.ID)
		if err != nil {
			return err
		}

		_, err = pollTask(finalizeMigrate.ID)
		if err != nil {
			return err
		}
		fmt.Printf("Deployment '%s' migration started '%s' to finalize.\n", deployment.ID, sourceAddress)

		return nil
	}

	return nil
}

func getDcMap(file string) (res *DcMap, err error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	res = &DcMap{}
	err = yaml.Unmarshal(buf, res)
	return
}

func createDeploymentFromDcMap(dcMap *DcMap) (deploymentID string, err error) {
	authInfo := &photon.AuthInfo{
		Enabled:        dcMap.Deployment.AuthEnabled,
		Endpoint:       dcMap.Deployment.AuthEndpoint,
		Port:           dcMap.Deployment.AuthPort,
		Tenant:         dcMap.Deployment.AuthTenant,
		Username:       dcMap.Deployment.AuthUsername,
		Password:       dcMap.Deployment.AuthPassword,
		SecurityGroups: dcMap.Deployment.AuthSecurityGroups,
	}

	statsInfo := &photon.StatsInfo{
		Enabled:       dcMap.Deployment.StatsEnabled,
		StoreEndpoint: dcMap.Deployment.StatsStoreEndpoint,
		StorePort:     dcMap.Deployment.StatsPort,
	}

	deploymentSpec := &photon.DeploymentCreateSpec{
		Auth:            authInfo,
		ImageDatastores: regexp.MustCompile(`\s*,\s*`).Split(dcMap.Deployment.ImageDatastores, -1),
		NTPEndpoint:     dcMap.Deployment.NTPEndpoint,
		SyslogEndpoint:  dcMap.Deployment.SyslogEndpoint,
		Stats:           statsInfo,
		UseImageDatastoreForVms: dcMap.Deployment.UseImageDatastoreForVms,
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

func createAvailabilityZonesFromDcMap(dcMap *DcMap) (map[string]string, error) {
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

func createHostsFromDcMap(dcMap *DcMap, deploymentID string) error {
	availabilityZoneNameToIdMap, err := createAvailabilityZonesFromDcMap(dcMap)
	if err != nil {
		return err
	}

	var hostSpecs []photon.HostCreateSpec
	var managementNetworkIps []string
	for _, host := range dcMap.Hosts {
		hostIps, err := parseIpRanges(host.IpRanges)
		if err != nil {
			return err
		}
		if hostIps == nil || len(hostIps) == 0 {
			return errors.New("Host IP Address missing in DC Map")
		}

		if host.Metadata != nil {
			if managementVmIps, exists := host.Metadata["MANAGEMENT_VM_IPS"]; exists {
				managementNetworkIps, err = parseIpRanges(managementVmIps)
				if err != nil {
					return err
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

func doDeploy(deploymentID string) error {
	deployTask, err := client.Esxclient.Deployments.Deploy(deploymentID)
	if err != nil {
		return err
	}

	_, err = pollTask(deployTask.ID)
	if err != nil {
		return err
	}
	fmt.Printf("Deployment '%s' is deployed.\n", deploymentID)

	return nil
}

func doDetroy(deploymentID string) error {
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
