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
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"

	"golang.org/x/crypto/ssh/terminal"
)

// Creates a cli.Command for clusters
// Subcommands: create;              Usage: cluster create [<options>]
//              show;                Usage: cluster show <id>
//              list;                Usage: cluster list [<options>]
//              list_vms;            Usage: cluster list_vms <id>
//              resize;              Usage: cluster resize <id> <new worker count> [<options>]
//              delete;              Usage: cluster delete <id>
//              trigger-maintenance; Usage: cluster trigger-maintenance <id>
//              cert-to-file;        Usage: cluster cert-to-file <id> <file_path>
func GetClusterCommand() cli.Command {
	command := cli.Command{
		Name:  "cluster",
		Usage: "Options for clusters",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new cluster",
				ArgsUsage: " ",
				Description: "Create a new Kubernetes cluster or Harbor Docker registry. \n\n" +
					"   Non-interactive mode Example: \n" +
					"   photon cluster create -n k8-cluster -k KUBERNETES --dns 10.0.0.1 \\ \n" +
					"     --gateway 192.0.2.1 --netmask 255.255.255.0 --master-ip 192.0.2.20 \\ \n" +
					"     --container-network 10.2.0.0/16 --etcd1 192.0.2.21 \\ \n" +
					"     -c 1 -v cluster-vm -d small-disk --ssh-key ~/.ssh/id_dsa.pub",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name",
					},
					cli.StringFlag{
						Name:  "project, p",
						Usage: "Project name",
					},
					cli.StringFlag{
						Name:  "name, n",
						Usage: "Cluster name",
					},
					cli.StringFlag{
						Name:  "type, k",
						Usage: "Cluster type (KUBERNETES or HARBOR)",
					},
					cli.StringFlag{
						Name:  "vm_flavor, v",
						Usage: "VM flavor name",
					},
					cli.StringFlag{
						Name:  "disk_flavor, d",
						Usage: "Disk flavor name",
					},
					cli.StringFlag{
						Name:  "network_id, w",
						Usage: "VM network ID",
					},
					cli.IntFlag{
						Name:  "worker_count, c",
						Usage: "Worker count",
					},
					cli.StringFlag{
						Name:  "dns",
						Usage: "VM network DNS server IP address",
					},
					cli.StringFlag{
						Name:  "gateway",
						Usage: "VM network gateway IP address",
					},
					cli.StringFlag{
						Name:  "netmask",
						Usage: "VM network netmask",
					},
					cli.StringFlag{
						Name:  "master-ip",
						Usage: "Kubernetes master IP address (required for Kubernetes clusters)",
					},
					cli.StringFlag{
						Name:  "container-network",
						Usage: "CIDR representation of the container network, e.g. '10.2.0.0/16' (required for Kubernetes clusters)",
					},
					cli.StringFlag{
						Name:  "etcd1",
						Usage: "Static IP address with which to create etcd node 1 (required for Kubernetes)",
					},
					cli.StringFlag{
						Name:  "etcd2",
						Usage: "Static IP address with which to create etcd node 2",
					},
					cli.StringFlag{
						Name:  "etcd3",
						Usage: "Static IP address with which to create etcd node 3",
					},
					cli.StringFlag{
						Name:  "ssh-key",
						Usage: "The file path of the SSH key",
					},
					cli.StringFlag{
						Name:  "registry-ca-cert",
						Usage: "The file path of the file containing the CA certificate for a docker registry (optional)",
					},
					cli.StringFlag{
						Name:  "admin-password",
						Usage: "The Harbor registry admin password (optional)",
					},
					cli.IntFlag{
						Name:  "batchSize",
						Usage: "Batch size for expanding worker nodes",
					},
					cli.BoolFlag{
						Name:  "wait-for-ready",
						Usage: "Wait synchronously for the cluster to become ready and expanded fully",
					},
				},
				Action: func(c *cli.Context) {
					err := createCluster(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show information about a cluster. ",
				ArgsUsage: "cluster-id",
				Description: "List the cluster's name, state, type, workerCount \n" +
					"   and all the extended properties. Also, list the master and \n" +
					"   etcd VM information about this cluster. For each VM, list the \n" +
					"   vm's ID, name and IP. \n\n" +
					"   Example: photon cluster show 9b159e92-9495-49a4-af58-53ad4764f616",
				Action: func(c *cli.Context) {
					err := showCluster(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "list",
				Usage:     "List clusters",
				ArgsUsage: " ",
				Description: "List all clusters in the current project. Attributes include \n" +
					"   ID, Name, Type, State and Worker Count \n\n" +
					"   Example: photon cluster list",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "tenant, t",
						Usage: "Tenant name",
					},
					cli.StringFlag{
						Name:  "project, p",
						Usage: "Project name",
					},
					cli.BoolFlag{
						Name:  "summary, s",
						Usage: "Summary view",
					},
				},
				Action: func(c *cli.Context) {
					err := listClusters(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:        "list_vms",
				Usage:       "List the VMs associated with a cluster",
				ArgsUsage:   "cluster-id",
				Description: "Example: photon cluster list_vms 9b159e92-9495-49a4-af58-53ad4764f616",
				Action: func(c *cli.Context) {
					err := listVms(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "resize",
				Usage:     "Resize a cluster",
				ArgsUsage: "cluster-id new-worker-count",
				Description: "Resize the cluster worker size to be the desired size. \n" +
					"   Note: 1.  The cluster's worker size can only be scaled up. \n" +
					"   2. The cluster is resized by batches with the batchSize parameter \n" +
					"   that was specified when the cluster was created. \n\n" +
					"   Example: photon cluster resize 9b159e92-9495-49a4-af58-53ad4764f616 5 ",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "wait-for-ready",
						Usage: "Wait synchronously for the cluster to become ready and expanded fully",
					},
				},
				Action: func(c *cli.Context) {
					err := resizeCluster(c, os.Stdout)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "delete",
				Usage:     "Delete a cluster",
				ArgsUsage: "cluster-id",
				Description: "This call deletes the specified cluster if there are still \n" +
					"   remaining VMs belong to the specified cluster. The remaining VMs \n" +
					"   will be stopped and deleted. \n\n" +
					"   Example: photon cluster delete 9b159e92-9495-49a4-af58-53ad4764f616",
				Action: func(c *cli.Context) {
					err := deleteCluster(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:        "trigger-maintenance",
				Usage:       "Start a background process to recreate failed VMs in a cluster",
				ArgsUsage:   "cluster-id",
				Description: "Example: photon cluster trigger-maintenance 9b159e92-9495-49a4-af58-53ad4764f616",
				Action: func(c *cli.Context) {
					err := triggerMaintenance(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "cert-to-file",
				Usage:     "Save the CA Certificate to a file with the specified path if the certificate exists",
				ArgsUsage: "cluster-id file_path",
				Description: "If a cluster has a CA certificate, this extracts it and saves \n" +
					"   it to a file. If the specified file path doesn't exist, it will create \n" +
					"   a new file with the specified pathThis is useful when using using Harbor, \n" +
					"   which uses a self-signed CA certificate. You can extract the CA certificate \n" +
					"   with this command, and use it as input when creating a Kubernetes cluster. \n\n" +
					"   Example: photon cluster cert-to-_file 9b159e92-9495-49a4-af58-53ad4764f616 ./user/cert",
				Action: func(c *cli.Context) {
					err := certToFile(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
		},
	}
	return command
}

// Sends a "create cluster" request to the API client based on the cli.Context
// Returns an error if one occurred
func createCluster(c *cli.Context, w io.Writer) error {
	err := checkArgNum(c.Args(), 0, "cluster create [<options>]")
	if err != nil {
		return err
	}

	tenantName := c.String("tenant")
	projectName := c.String("project")
	name := c.String("name")
	cluster_type := c.String("type")
	vm_flavor := c.String("vm_flavor")
	disk_flavor := c.String("disk_flavor")
	network_id := c.String("network_id")
	worker_count := c.Int("worker_count")
	dns := c.String("dns")
	gateway := c.String("gateway")
	netmask := c.String("netmask")
	master_ip := c.String("master-ip")
	container_network := c.String("container-network")
	etcd1 := c.String("etcd1")
	etcd2 := c.String("etcd2")
	etcd3 := c.String("etcd3")
	batch_size := c.Int("batchSize")
	ssh_key := c.String("ssh-key")
	ca_cert := c.String("registry-ca-cert")
	admin_password := c.String("admin-password")

	wait_for_ready := c.IsSet("wait-for-ready")

	const DEFAULT_WORKER_COUNT = 1

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	tenant, err := verifyTenant(tenantName)
	if err != nil {
		return err
	}

	project, err := verifyProject(tenant.ID, projectName)
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") {
		name, err = askForInput("Cluster name: ", name)
		if err != nil {
			return err
		}
		cluster_type, err = askForInput("Cluster type: ", cluster_type)
		if err != nil {
			return err
		}
		if worker_count == 0 && cluster_type != "HARBOR" {
			worker_count_string, err := askForInput("Worker count: ", "")
			if err != nil {
				return err
			}
			worker_count, err = strconv.Atoi(worker_count_string)
			if err != nil {
				return fmt.Errorf("Please supply a valid worker count")
			}
		}
	}

	if len(name) == 0 || len(cluster_type) == 0 {
		return fmt.Errorf("Provide a valid cluster name and type")
	}

	if worker_count == 0 && cluster_type != "HARBOR" {
		worker_count = DEFAULT_WORKER_COUNT
	}

	if !c.GlobalIsSet("non-interactive") {
		dns, err = askForInput("Cluster DNS server: ", dns)
		if err != nil {
			return err
		}
		gateway, err = askForInput("Cluster network gateway: ", gateway)
		if err != nil {
			return err
		}
		netmask, err = askForInput("Cluster network netmask: ", netmask)
		if err != nil {
			return err
		}
		ssh_key, err = askForInput("Cluster ssh key file path (leave blank for none): ", ssh_key)
		if err != nil {
			return err
		}
	}

	if len(dns) == 0 || len(gateway) == 0 || len(netmask) == 0 {
		return fmt.Errorf("Provide a valid DNS, gateway, and netmask")
	}

	extended_properties := make(map[string]string)
	extended_properties[photon.ExtendedPropertyDNS] = dns
	extended_properties[photon.ExtendedPropertyGateway] = gateway
	extended_properties[photon.ExtendedPropertyNetMask] = netmask
	if len(ssh_key) != 0 {
		ssh_key_content, err := readSSHKey(ssh_key)
		if err == nil {
			extended_properties[photon.ExtendedPropertySSHKey] = ssh_key_content
		} else {
			return err
		}
	}

	if len(ca_cert) != 0 {
		ca_cert_content, err := readCACert(ca_cert)
		if err == nil {
			extended_properties[photon.ExtendedPropertyRegistryCACert] = ca_cert_content
		} else {
			return err
		}
	}

	cluster_type = strings.ToUpper(cluster_type)
	switch cluster_type {
	case "KUBERNETES":
		if !c.GlobalIsSet("non-interactive") {
			master_ip, err = askForInput("Kubernetes master static IP address: ", master_ip)
			if err != nil {
				return err
			}
			container_network, err = askForInput("Kubernetes worker network ID: ", container_network)
			if err != nil {
				return err
			}
			etcd1, err = askForInput("etcd server 1 static IP address: ", etcd1)
			if err != nil {
				return err
			}
			etcd2, err = askForInput("etcd server 2 static IP address (leave blank for none): ", etcd2)
			if err != nil {
				return err
			}
			if len(etcd2) != 0 {
				etcd3, err = askForInput("etcd server 3 static IP address (leave blank for none): ", etcd3)
				if err != nil {
					return err
				}
			}
		}

		extended_properties[photon.ExtendedPropertyMasterIP] = master_ip
		extended_properties[photon.ExtendedPropertyContainerNetwork] = container_network
		extended_properties[photon.ExtendedPropertyETCDIP1] = etcd1
		if len(etcd2) != 0 {
			extended_properties[photon.ExtendedPropertyETCDIP2] = etcd2
			if len(etcd3) != 0 {
				extended_properties[photon.ExtendedPropertyETCDIP3] = etcd3
			}
		}
	case "HARBOR":
		if !c.GlobalIsSet("non-interactive") {
			master_ip, err = askForInput("Harbor master static IP address: ", master_ip)
			if err != nil {
				return err
			}
			if len(admin_password) == 0 {
				fmt.Printf("Harbor registry admin password: ")
				bytePassword, err := terminal.ReadPassword(0)
				if err != nil {
					return err
				}
				admin_password = string(bytePassword)
				fmt.Printf("\n")
			}
		}
		extended_properties[photon.ExtendedPropertyMasterIP] = master_ip
		extended_properties[photon.ExtendedPropertyAdminPassword] = admin_password
	default:
		return fmt.Errorf("Unsupported cluster type: %s", cluster_type)
	}

	clusterSpec := photon.ClusterCreateSpec{}
	clusterSpec.Name = name
	clusterSpec.Type = cluster_type
	clusterSpec.VMFlavor = vm_flavor
	clusterSpec.DiskFlavor = disk_flavor
	clusterSpec.NetworkID = network_id
	clusterSpec.WorkerCount = worker_count
	clusterSpec.BatchSizeWorker = batch_size
	clusterSpec.ExtendedProperties = extended_properties

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("\n")
		fmt.Printf("Creating cluster: %s (%s)\n", clusterSpec.Name, clusterSpec.Type)
		if len(clusterSpec.VMFlavor) != 0 {
			fmt.Printf("  VM flavor: %s\n", clusterSpec.VMFlavor)
		}
		if len(clusterSpec.DiskFlavor) != 0 {
			fmt.Printf("  Disk flavor: %s\n", clusterSpec.DiskFlavor)
		}
		if clusterSpec.Type != "HARBOR" {
			fmt.Printf("  Worker count: %d\n", clusterSpec.WorkerCount)
		}
		if clusterSpec.BatchSizeWorker != 0 {
			fmt.Printf("  Batch size: %d\n", clusterSpec.BatchSizeWorker)
		}
		fmt.Printf("\n")
	}

	if confirmed(c.GlobalIsSet("non-interactive")) {
		createTask, err := client.Esxclient.Projects.CreateCluster(project.ID, &clusterSpec)
		if err != nil {
			return err
		}

		_, err = waitOnTaskOperation(createTask.ID, c)
		if err != nil {
			return err
		}

		if wait_for_ready {
			if !utils.NeedsFormatting(c) {
				fmt.Printf("Waiting for cluster %s to become ready\n", createTask.Entity.ID)
			}
			cluster, err := waitForCluster(createTask.Entity.ID)
			if err != nil {
				return err
			}

			if utils.NeedsFormatting(c) {
				utils.FormatObject(cluster, w, c)
			} else {
				fmt.Printf("Cluster %s is ready\n", cluster.ID)
			}

		} else {
			fmt.Println("Note: the cluster has been created with minimal resources. You can use the cluster now.")
			fmt.Println("A background task is running to gradually expand the cluster to its target capacity.")
			fmt.Printf("You can run 'cluster show %s' to see the state of the cluster.\n", createTask.Entity.ID)
		}
	} else {
		fmt.Println("Cancelled")
	}

	return nil
}

// Sends a "show cluster" request to the API client based on the cli.Context
// Returns an error if one occurred
func showCluster(c *cli.Context, w io.Writer) error {
	err := checkArgNum(c.Args(), 1, "cluster show <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	cluster, err := client.Esxclient.Clusters.Get(id)
	if err != nil {
		return err
	}

	vms, err := client.Esxclient.Clusters.GetVMs(id)
	if err != nil {
		return err
	}

	var master_vms []photon.VM
	for _, vm := range vms.Items {
		for _, tag := range vm.Tags {
			if strings.Count(tag, ":") == 2 && !strings.Contains(strings.ToLower(tag), "worker") {
				master_vms = append(master_vms, vm)
				break
			}
		}
	}

	if c.GlobalIsSet("non-interactive") {
		extendedProperties := strings.Trim(strings.TrimLeft(fmt.Sprint(cluster.ExtendedProperties), "map"), "[]")
		if cluster.ErrorReason != "" {
			fmt.Printf("%s\t%s\t%s\t%s\t%d\t%s\t%s\n", cluster.ID, cluster.Name, cluster.State, cluster.Type,
				cluster.WorkerCount, cluster.ErrorReason, extendedProperties)
		} else {
			fmt.Printf("%s\t%s\t%s\t%s\t%d\t%s\n", cluster.ID, cluster.Name, cluster.State, cluster.Type,
				cluster.WorkerCount, extendedProperties)
		}
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(cluster, w, c)
	} else {
		fmt.Println("Cluster ID:            ", cluster.ID)
		fmt.Println("  Name:                ", cluster.Name)
		fmt.Println("  State:               ", cluster.State)
		fmt.Println("  Type:                ", cluster.Type)
		fmt.Println("  Worker count:        ", cluster.WorkerCount)
		if cluster.ErrorReason != "" {
			fmt.Println("  Error Reason:        ", cluster.ErrorReason)
		}
		fmt.Println("  Extended Properties: ", cluster.ExtendedProperties)
		fmt.Println()
	}

	err = printClusterVMs(master_vms, w, c)
	if err != nil {
		return err
	}

	return nil
}

// Sends a "list clusters" request to the API client based on the cli.Context
// Returns an error if one occurred
func listClusters(c *cli.Context, w io.Writer) error {
	err := checkArgNum(c.Args(), 0, "cluster list [<options>]")
	if err != nil {
		return err
	}

	tenantName := c.String("tenant")
	projectName := c.String("project")
	summaryView := c.IsSet("summary")

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	tenant, err := verifyTenant(tenantName)
	if err != nil {
		return err
	}

	project, err := verifyProject(tenant.ID, projectName)
	if err != nil {
		return err
	}

	clusterList, err := client.Esxclient.Projects.GetClusters(project.ID)
	if err != nil {
		return err
	}

	err = printClusterList(clusterList.Items, w, c, summaryView)
	if err != nil {
		return err
	}

	return nil
}

// Sends a "list VMs for cluster" request to the API client based on the cli.Context
// Returns an error if one occurred
func listVms(c *cli.Context, w io.Writer) error {
	err := checkArgNum(c.Args(), 1, "cluster list_vms <id>")
	if err != nil {
		return err
	}
	cluster_id := c.Args().First()

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	vms, err := client.Esxclient.Clusters.GetVMs(cluster_id)
	if err != nil {
		return err
	}

	err = printVMList(vms.Items, w, c, false)
	if err != nil {
		return err
	}

	return nil
}

// Sends a "resize cluster" request to the API client based on the cli.Context
// Returns an error if one occurred
func resizeCluster(c *cli.Context, w io.Writer) error {
	err := checkArgNum(c.Args(), 2, "cluster resize <id> <new worker count> [<options>]")
	if err != nil {
		return err
	}

	cluster_id := c.Args()[0]
	worker_count_string := c.Args()[1]
	worker_count, err := strconv.Atoi(worker_count_string)
	wait_for_ready := c.IsSet("wait-for-ready")

	if len(cluster_id) == 0 || err != nil || worker_count <= 0 {
		return fmt.Errorf("Provide a valid cluster ID and worker count")
	}

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("\nResizing cluster %s to worker count %d\n", cluster_id, worker_count)
	}

	if confirmed(c.GlobalIsSet("non-interactive")) {
		resizeSpec := photon.ClusterResizeOperation{}
		resizeSpec.NewWorkerCount = worker_count
		resizeTask, err := client.Esxclient.Clusters.Resize(cluster_id, &resizeSpec)
		if err != nil {
			return err
		}

		_, err = waitOnTaskOperation(resizeTask.ID, c)
		if err != nil {
			return err
		}

		if wait_for_ready {
			cluster, err := waitForCluster(cluster_id)
			if err != nil {
				return err
			}
			if utils.NeedsFormatting(c) {
				utils.FormatObject(cluster, w, c)
			} else {
				fmt.Printf("Cluster %s is ready\n", cluster.ID)
			}
		} else {
			fmt.Println("Note: A background task is running to gradually resize the cluster to its target capacity.")
			fmt.Printf("You may continue to use the cluster. You can run 'cluster show %s'\n", resizeTask.Entity.ID)
			fmt.Println("to see the state of the cluster. If the resize operation is still in progress, the cluster state")
			fmt.Println("will show as RESIZING. Once the cluster is resized, the cluster state will show as READY.")
		}
	} else {
		fmt.Println("Cancelled")
	}

	return nil
}

// Sends a "delete cluster" request to the API client based on the cli.Context
// Returns an error if one occurred
func deleteCluster(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "cluster delete <id>")
	if err != nil {
		return nil
	}

	cluster_id := c.Args().First()

	if len(cluster_id) == 0 {
		return fmt.Errorf("Please provide a valid cluster ID")
	}

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("\nDeleting cluster %s\n", cluster_id)
	}

	if confirmed(c.GlobalIsSet("non-interactive")) {
		deleteTask, err := client.Esxclient.Clusters.Delete(cluster_id)
		if err != nil {
			return err
		}

		_, err = waitOnTaskOperation(deleteTask.ID, c)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("Cancelled")
	}

	return nil
}

// Sends a cluster trigger_maintenance request to the API client based on the cli.Context.
// Returns an error if one occurred.
func triggerMaintenance(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "cluster trigger-maintenance <id>")
	if err != nil {
		return nil
	}

	clusterId := c.Args().First()

	if len(clusterId) == 0 {
		return fmt.Errorf("Please provide a valid cluster ID")
	}

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("Maintenance triggered for cluster %s\n", clusterId)
	}

	task, err := client.Esxclient.Clusters.TriggerMaintenance(clusterId)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Helper routine which waits for a cluster to enter the READY state.
func waitForCluster(id string) (cluster *photon.Cluster, err error) {
	start := time.Now()
	numErr := 0

	taskPollTimeout := 60 * time.Minute
	taskPollDelay := 2 * time.Second
	taskRetryCount := 3

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		displayTaskProgress(start)
	}()

	for time.Since(start) < taskPollTimeout {
		cluster, err = client.Esxclient.Clusters.Get(id)
		if err != nil {
			numErr++
			if numErr > taskRetryCount {
				endAnimation = true
				wg.Wait()
				return
			}
		}
		switch strings.ToUpper(cluster.State) {
		case "ERROR":
			endAnimation = true
			wg.Wait()
			err = fmt.Errorf("Cluster %s entered ERROR state", id)
			return
		case "READY":
			endAnimation = true
			wg.Wait()
			return
		}

		time.Sleep(taskPollDelay)
	}

	endAnimation = true
	wg.Wait()
	err = fmt.Errorf("Timed out while waiting for cluster to enter READY state")
	return
}

// This is a helper function for reading the ssh key from a file.
func readSSHKey(filename string) (result string, err error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer func() {
		e := file.Close()
		if e != nil {
			err = e
		}
	}()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	// Read just the first line because the key file should only by one line long.
	scanner.Scan()
	keystring := scanner.Text()
	keystring = strings.TrimSpace(keystring)
	if err := scanner.Err(); err != nil {
		return "", err
	}
	err = validateSSHKey(keystring)
	if err != nil {
		return "", err
	}
	return keystring, nil
}

// This is a helper function to validate that a key is a valid ssh key
func validateSSHKey(key string) error {
	if len(key) == 0 {
		return fmt.Errorf("The ssh-key file provided has no content")
	}
	// Other validation test can go here if desired in the future
	return nil
}

// This is a helper function for reading the CA Cert from a file.
func readCACert(filename string) (result string, err error) {
	certData, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	cert := string(certData)
	cert = strings.TrimSpace(cert)
	err = validateCert(cert)
	if err != nil {
		return "", err
	}
	return cert, nil
}

// This is helper function to validate that a string is a valid certificate
func validateCert(cert string) error {
	beginCert := "-----BEGIN CERTIFICATE-----"
	endCert := "-----END CERTIFICATE-----"
	if !strings.Contains(cert, beginCert) || !strings.Contains(cert, endCert) {
		return fmt.Errorf("The certificate provided does not have a valid format.")
	}
	return nil
}

func certToFile(c *cli.Context) error {
	err := checkArgNum(c.Args(), 2, "cluster cert-to-file <id> <file_path>")
	if err != nil {
		return err
	}
	id := c.Args().First()
	filePath := c.Args()[1]

	client.Esxclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	cluster, err := client.Esxclient.Clusters.Get(id)
	if err != nil {
		return err
	}

	cert := ""

	// Case: Kubernetes
	if cluster.ExtendedProperties["registry_ca_cert"] != "" {
		cert = cluster.ExtendedProperties["registry_ca_cert"]
		err := ioutil.WriteFile(filePath, []byte(cert), 0644)
		if err != nil {
			return err
		}
		return nil
	}

	// Case: Harbor
	if cluster.ExtendedProperties["ca_cert"] != "" {
		cert = cluster.ExtendedProperties["ca_cert"]
		err := ioutil.WriteFile(filePath, []byte(cert), 0644)
		if err != nil {
			return err
		}
		return nil
	}

	// Extended Property doesn't contain either registry_ca_cert or ca_cert
	return fmt.Errorf("There is no certificate associated with this cluster")
}
