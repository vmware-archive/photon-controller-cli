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
	"syscall"
	"time"
	"unicode"

	"github.com/vmware/photon-controller-cli/photon/client"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"

	"golang.org/x/crypto/ssh/terminal"
)

// Creates a cli.Command for services
// Subcommands: create;              Usage: service create [<options>]
//              show;                Usage: service show <id>
//              list;                Usage: service list [<options>]
//              list_vms;            Usage: service list_vms <id>
//              resize;              Usage: service resize <id> <new worker count> [<options>]
//              delete;              Usage: service delete <id>
//              trigger-maintenance; Usage: service trigger-maintenance <id>
//              cert-to-file;        Usage: service cert-to-file <id> <file_path>

func GetServiceCommand() cli.Command {
	command := cli.Command{
		Name:    "service",
		Aliases: []string{"cluster"},
		Usage:   "Options for services",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Usage:     "Create a new service",
				ArgsUsage: " ",
				Description: "Create a new Kubernetes service or Harbor Docker registry. \n\n" +
					"   Example: \n" +
					"   photon service create -n k8-service -k KUBERNETES --dns 10.0.0.1 \\ \n" +
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
						Usage: "Service name",
					},
					cli.StringFlag{
						Name:  "type, k",
						Usage: "Service type (KUBERNETES or HARBOR)",
					},
					cli.StringFlag{
						Name:  "vm_flavor, v",
						Usage: "VM flavor name for master and worker",
					},
					cli.StringFlag{
						Name:  "master-vm-flavor, m",
						Usage: "Override master VM flavor",
					},
					cli.StringFlag{
						Name:  "worker-vm-flavor, w",
						Usage: "Override worker VM flavor",
					},
					cli.StringFlag{
						Name:  "disk_flavor, d",
						Usage: "Disk flavor name",
					},
					cli.StringFlag{
						Name:  "network_id, w",
						Usage: "VM network ID",
					},
					cli.StringFlag{
						Name:  "image-id, i",
						Usage: "Image ID",
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
						Usage: "Kubernetes master IP address (required for Kubernetes services)",
					},
					cli.StringFlag{
						Name:  "load-balancer-ip",
						Usage: "Kubernetes load balancer IP address (required for Kubernetes services)",
					},
					cli.StringFlag{
						Name:  "container-network",
						Usage: "CIDR representation of the container network, e.g. '10.2.0.0/16' (required for Kubernetes services)",
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
						Name: "admin-password",
						Usage: "The Harbor registry admin password (optional). The password " +
							"needs to have at least 7 characters with 1 lowercase " +
							"letter, 1 capital letter and 1 numeric character. If not " +
							"specified, the default user name is admin and the password is " +
							"Harbor12345 ",
					},
					cli.IntFlag{
						Name:  "batchSize",
						Usage: "Batch size for expanding worker nodes",
					},
					cli.BoolFlag{
						Name:  "wait-for-ready",
						Usage: "Wait synchronously for the service to become ready and expanded fully",
					},
				},
				Action: func(c *cli.Context) {
					err := createService(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "show",
				Usage:     "Show information about a service. ",
				ArgsUsage: "service-id",
				Description: "List the service's name, state, type, workerCount \n" +
					"   and all the extended properties. Also, list the master and \n" +
					"   etcd VM information about this service. For each VM, list the \n" +
					"   vm's ID, name and IP. \n\n" +
					"   Example: photon service show 9b159e92-9495-49a4-af58-53ad4764f616",
				Action: func(c *cli.Context) {
					err := showService(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "list",
				Usage:     "List services",
				ArgsUsage: " ",
				Description: "List all services in the current project. Attributes include \n" +
					"   ID, Name, Type, State and Worker Count \n\n" +
					"   Example: photon service list",
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
					err := listServices(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:        "list-vms",
				Usage:       "List the VMs associated with a service",
				ArgsUsage:   "service-id",
				Description: "Example: photon service list_vms 9b159e92-9495-49a4-af58-53ad4764f616",
				Action: func(c *cli.Context) {
					err := listVms(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Hidden:      true,
				Name:        "list_vms",
				Usage:       "List the VMs associated with a service",
				ArgsUsage:   "service-id",
				Description: "Deprecated, use list-vms instead",
				Action: func(c *cli.Context) {
					err := listVms(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "resize",
				Usage:     "Resize a service",
				ArgsUsage: "service-id new-worker-count",
				Description: "Resize the service worker size to be the desired size. \n" +
					"   Note: 1.  The service's worker size can only be scaled up. \n" +
					"   2. The service is resized by batches with the batchSize parameter \n" +
					"   that was specified when the service was created. \n\n" +
					"   Example: photon service resize 9b159e92-9495-49a4-af58-53ad4764f616 5 ",
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "wait-for-ready",
						Usage: "Wait synchronously for the service to become ready and expanded fully",
					},
				},
				Action: func(c *cli.Context) {
					err := resizeService(c, os.Stdout)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "delete",
				Usage:     "Delete a service",
				ArgsUsage: "service-id",
				Description: "This call deletes the specified service if there are still \n" +
					"   remaining VMs belong to the specified service. The remaining VMs \n" +
					"   will be stopped and deleted. \n\n" +
					"   Example: photon service delete 9b159e92-9495-49a4-af58-53ad4764f616",
				Action: func(c *cli.Context) {
					err := deleteService(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:        "trigger-maintenance",
				Usage:       "Start a background process to recreate failed VMs in a service",
				ArgsUsage:   "service-id",
				Description: "Example: photon service trigger-maintenance 9b159e92-9495-49a4-af58-53ad4764f616",
				Action: func(c *cli.Context) {
					err := triggerMaintenance(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:      "cert-to-file",
				Usage:     "Save the CA Certificate to a file with the specified path if the certificate exists",
				ArgsUsage: "service-id file_path",
				Description: "If a service has a CA certificate, this extracts it and saves \n" +
					"   it to a file. If the specified file path doesn't exist, it will create \n" +
					"   a new file with the specified pathThis is useful when using using Harbor, \n" +
					"   which uses a self-signed CA certificate. You can extract the CA certificate \n" +
					"   with this command, and use it as input when creating a Kubernetes service. \n\n" +
					"   Example: photon service cert-to-file 9b159e92-9495-49a4-af58-53ad4764f616 ./user/cert",
				Action: func(c *cli.Context) {
					err := certToFile(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Sends a "create service" request to the API client based on the cli.Context
// Returns an error if one occurred
func createService(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	tenantName := c.String("tenant")
	projectName := c.String("project")
	name := c.String("name")
	service_type := c.String("type")
	vm_flavor := c.String("vm_flavor")
	master_vm_flavor := c.String("master_vm_flavor")
	worker_vm_flavor := c.String("worker_vm_flavor")
	disk_flavor := c.String("disk_flavor")
	network_id := c.String("network_id")
	image_id := c.String("image-id")
	worker_count := c.Int("worker_count")
	dns := c.String("dns")
	gateway := c.String("gateway")
	netmask := c.String("netmask")
	masterIP := c.String("master-ip")
	loadBalancerIP := c.String("load-balancer-ip")
	container_network := c.String("container-network")
	etcd1 := c.String("etcd1")
	etcd2 := c.String("etcd2")
	etcd3 := c.String("etcd3")
	batch_size := c.Int("batchSize")
	ssh_key := c.String("ssh-key")
	ca_cert := c.String("registry-ca-cert")
	admin_password := c.String("admin-password")

	if admin_password != "" {
		result := validateHarborPassword(admin_password)
		if !result {
			return fmt.Errorf("The Harbor password is invalid. It should have at least 7 characters " +
				"with 1 lowercase letter, 1 capital letter and 1 numeric character.")
		}
	}

	wait_for_ready := c.IsSet("wait-for-ready")

	const DEFAULT_WORKER_COUNT = 1

	client.Photonclient, err = client.GetClient(c)
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
		name, err = askForInput("Service name: ", name)
		if err != nil {
			return err
		}
		service_type, err = askForInput("Service type: ", service_type)
		if err != nil {
			return err
		}
		if worker_count == 0 && service_type != "HARBOR" {
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

	if len(name) == 0 || len(service_type) == 0 {
		return fmt.Errorf("Provide a valid service name and type")
	}

	if worker_count == 0 && service_type != "HARBOR" {
		worker_count = DEFAULT_WORKER_COUNT
	}

	if !c.GlobalIsSet("non-interactive") {
		dns, err = askForInput("Service DNS server: ", dns)
		if err != nil {
			return err
		}
		gateway, err = askForInput("Service network gateway: ", gateway)
		if err != nil {
			return err
		}
		netmask, err = askForInput("Service network netmask: ", netmask)
		if err != nil {
			return err
		}
		ssh_key, err = askForInput("Service ssh key file path (leave blank for none): ", ssh_key)
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

	service_type = strings.ToUpper(service_type)
	switch service_type {
	case "KUBERNETES":
		if !c.GlobalIsSet("non-interactive") {
			masterIP, err = askForInput("Kubernetes master static IP address: ", masterIP)
			if err != nil {
				return err
			}
			loadBalancerIP, err = askForInput("Kubernetes load balancer static IP address: ", loadBalancerIP)
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

		extended_properties[photon.ExtendedPropertyMasterIP] = masterIP
		extended_properties[photon.ExtendedPropertyLoadBalancerIP] = loadBalancerIP
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
			masterIP, err = askForInput("Harbor master static IP address: ", masterIP)
			if err != nil {
				return err
			}
			if len(admin_password) == 0 {
				fmt.Printf("Harbor registry admin password: ")
				// Casting syscall.Stdin to int because during
				// Windows cross-compilation syscall.Stdin is incorrectly
				// treated as a String.
				bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return err
				}
				admin_password = string(bytePassword)

				result := validateHarborPassword(admin_password)
				if result != true {
					return fmt.Errorf("The Harbor password is invalid. It should have at least 7 " +
						"characters with 1 lowercase letter, 1 capital letter and 1 numeric " +
						"character.")
				}
				fmt.Printf("\n")
			}
		}
		extended_properties[photon.ExtendedPropertyMasterIP] = masterIP
		extended_properties[photon.ExtendedPropertyAdminPassword] = admin_password
	default:
		return fmt.Errorf("Unsupported service type: %s", service_type)
	}

	serviceSpec := photon.ServiceCreateSpec{}
	serviceSpec.Name = name
	serviceSpec.Type = service_type
	serviceSpec.VMFlavor = ""
	if len(vm_flavor) != 0 {
		serviceSpec.MasterVmFlavor = vm_flavor
		serviceSpec.WorkerVmFlavor = vm_flavor
	}
	if len(master_vm_flavor) != 0 {
		serviceSpec.MasterVmFlavor = master_vm_flavor
	}
	if len(worker_vm_flavor) != 0 {
		serviceSpec.WorkerVmFlavor = worker_vm_flavor
	}
	serviceSpec.DiskFlavor = disk_flavor
	serviceSpec.NetworkID = network_id
	serviceSpec.ImageID = image_id
	serviceSpec.WorkerCount = worker_count
	serviceSpec.BatchSizeWorker = batch_size
	serviceSpec.ExtendedProperties = extended_properties

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("\n")
		fmt.Printf("Creating service: %s (%s)\n", serviceSpec.Name, serviceSpec.Type)
		if len(serviceSpec.VMFlavor) != 0 {
			fmt.Printf("  VM flavor: %s\n", serviceSpec.VMFlavor)
		}
		if len(serviceSpec.DiskFlavor) != 0 {
			fmt.Printf("  Disk flavor: %s\n", serviceSpec.DiskFlavor)
		}
		if serviceSpec.Type != "HARBOR" {
			fmt.Printf("  Worker count: %d\n", serviceSpec.WorkerCount)
		}
		if serviceSpec.BatchSizeWorker != 0 {
			fmt.Printf("  Batch size: %d\n", serviceSpec.BatchSizeWorker)
		}
		fmt.Printf("\n")
	}

	if confirmed(c) {
		createTask, err := client.Photonclient.Projects.CreateService(project.ID, &serviceSpec)
		if err != nil {
			return err
		}

		_, err = waitOnTaskOperation(createTask.ID, c)
		if err != nil {
			return err
		}

		if wait_for_ready {
			if !utils.NeedsFormatting(c) {
				fmt.Printf("Waiting for service %s to become ready\n", createTask.Entity.ID)
			}
			service, err := waitForService(createTask.Entity.ID)
			if err != nil {
				return err
			}

			if utils.NeedsFormatting(c) {
				utils.FormatObject(service, w, c)
			} else {
				fmt.Printf("Service %s is ready\n", service.ID)
			}

		} else {
			fmt.Println("Note: the service has been created with minimal resources. You can use the service now.")
			fmt.Println("A background task is running to gradually expand the service to its target capacity.")
			fmt.Printf("You can run 'service show %s' to see the state of the service.\n", createTask.Entity.ID)
		}
	} else {
		fmt.Println("Cancelled")
	}

	return nil
}

// Sends a "show service" request to the API client based on the cli.Context
// Returns an error if one occurred
func showService(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	service, err := client.Photonclient.Services.Get(id)
	if err != nil {
		return err
	}

	vms, err := client.Photonclient.Services.GetVMs(id)
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
		extendedProperties := strings.Trim(strings.TrimLeft(fmt.Sprint(service.ExtendedProperties), "map"), "[]")
		if service.ErrorReason != "" {
			fmt.Printf("%s\t%s\t%s\t%s\t%d\t%s\t%s\n", service.ID, service.Name, service.State, service.Type,
				service.WorkerCount, service.ErrorReason, extendedProperties)
		} else {
			fmt.Printf("%s\t%s\t%s\t%s\t%d\t%s\n", service.ID, service.Name, service.State, service.Type,
				service.WorkerCount, extendedProperties)
		}
	} else if utils.NeedsFormatting(c) {
		utils.FormatObject(service, w, c)
	} else {
		fmt.Println("Service ID:            ", service.ID)
		fmt.Println("  Name:                ", service.Name)
		fmt.Println("  State:               ", service.State)
		fmt.Println("  Type:                ", service.Type)
		fmt.Println("  Worker count:        ", service.WorkerCount)
		if service.ErrorReason != "" {
			fmt.Println("  Error Reason:        ", service.ErrorReason)
		}
		fmt.Println("  Extended Properties: ", service.ExtendedProperties)
		fmt.Println()
	}

	err = printServiceVMs(master_vms, w, c)
	if err != nil {
		return err
	}

	return nil
}

// Sends a "list services" request to the API client based on the cli.Context
// Returns an error if one occurred
func listServices(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}

	tenantName := c.String("tenant")
	projectName := c.String("project")
	summaryView := c.IsSet("summary")

	client.Photonclient, err = client.GetClient(c)
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

	serviceList, err := client.Photonclient.Projects.GetServices(project.ID)
	if err != nil {
		return err
	}

	err = printServiceList(serviceList.Items, w, c, summaryView)
	if err != nil {
		return err
	}

	return nil
}

// Sends a "list VMs for service" request to the API client based on the cli.Context
// Returns an error if one occurred
func listVms(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	service_id := c.Args().First()

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	vms, err := client.Photonclient.Services.GetVMs(service_id)
	if err != nil {
		return err
	}

	err = printVMList(vms.Items, w, c, false)
	if err != nil {
		return err
	}

	return nil
}

// Sends a "resize service" request to the API client based on the cli.Context
// Returns an error if one occurred
func resizeService(c *cli.Context, w io.Writer) error {
	err := checkArgCount(c, 2)
	if err != nil {
		return err
	}

	service_id := c.Args()[0]
	worker_count_string := c.Args()[1]
	worker_count, err := strconv.Atoi(worker_count_string)
	wait_for_ready := c.IsSet("wait-for-ready")

	if len(service_id) == 0 || err != nil || worker_count <= 0 {
		return fmt.Errorf("Provide a valid service ID and worker count")
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("\nResizing service %s to worker count %d\n", service_id, worker_count)
	}

	if confirmed(c) {
		resizeSpec := photon.ServiceResizeOperation{}
		resizeSpec.NewWorkerCount = worker_count
		resizeTask, err := client.Photonclient.Services.Resize(service_id, &resizeSpec)
		if err != nil {
			return err
		}

		_, err = waitOnTaskOperation(resizeTask.ID, c)
		if err != nil {
			return err
		}

		if wait_for_ready {
			service, err := waitForService(service_id)
			if err != nil {
				return err
			}
			if utils.NeedsFormatting(c) {
				utils.FormatObject(service, w, c)
			} else {
				fmt.Printf("Service %s is ready\n", service.ID)
			}
		} else {
			fmt.Println("Note: A background task is running to gradually resize the service to its target capacity.")
			fmt.Printf("You may continue to use the service. You can run 'service show %s'\n", resizeTask.Entity.ID)
			fmt.Println("to see the state of the service. If the resize operation is still in progress, the service state")
			fmt.Println("will show as RESIZING. Once the service is resized, the service state will show as READY.")
		}
	} else {
		fmt.Println("Cancelled")
	}

	return nil
}

// Sends a "delete service" request to the API client based on the cli.Context
// Returns an error if one occurred
func deleteService(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return nil
	}

	service_id := c.Args().First()

	if len(service_id) == 0 {
		return fmt.Errorf("Please provide a valid service ID")
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("\nDeleting service %s\n", service_id)
	}

	if confirmed(c) {
		deleteTask, err := client.Photonclient.Services.Delete(service_id)
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

// Sends a service trigger_maintenance request to the API client based on the cli.Context.
// Returns an error if one occurred.
func triggerMaintenance(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return nil
	}

	serviceId := c.Args().First()

	if len(serviceId) == 0 {
		return fmt.Errorf("Please provide a valid service ID")
	}

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	if !c.GlobalIsSet("non-interactive") {
		fmt.Printf("Maintenance triggered for service %s\n", serviceId)
	}

	task, err := client.Photonclient.Services.TriggerMaintenance(serviceId)
	if err != nil {
		return err
	}

	_, err = waitOnTaskOperation(task.ID, c)
	if err != nil {
		return err
	}

	return nil
}

// Helper routine which waits for a service to enter the READY state.
func waitForService(id string) (service *photon.Service, err error) {
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
		service, err = client.Photonclient.Services.Get(id)
		if err != nil {
			numErr++
			if numErr > taskRetryCount {
				endAnimation = true
				wg.Wait()
				return
			}
		}
		switch strings.ToUpper(service.State) {
		case "ERROR":
			endAnimation = true
			wg.Wait()
			err = fmt.Errorf("Service %s entered ERROR state", id)
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
	err = fmt.Errorf("Timed out while waiting for service to enter READY state")
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
	err := checkArgCount(c, 2)
	if err != nil {
		return err
	}
	id := c.Args().First()
	filePath := c.Args()[1]

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	service, err := client.Photonclient.Services.Get(id)
	if err != nil {
		return err
	}

	cert := ""

	// Case: Kubernetes
	if service.ExtendedProperties["registry_ca_cert"] != "" {
		cert = service.ExtendedProperties["registry_ca_cert"]
		err := ioutil.WriteFile(filePath, []byte(cert), 0644)
		if err != nil {
			return err
		}
		return nil
	}

	// Case: Harbor
	if service.ExtendedProperties["ca_cert"] != "" {
		cert = service.ExtendedProperties["ca_cert"]
		err := ioutil.WriteFile(filePath, []byte(cert), 0644)
		if err != nil {
			return err
		}
		return nil
	}

	// Extended Property doesn't contain either registry_ca_cert or ca_cert
	return fmt.Errorf("There is no certificate associated with this service")
}

func validateHarborPassword(password string) bool {
	correct := true
	number := false
	upper := false
	lower := false
	count := 0
	for _, letter := range password {
		switch {
		case unicode.IsNumber(letter):
			number = true
			count++
		case unicode.IsUpper(letter):
			upper = true
			count++
		case unicode.IsLower(letter):
			lower = true
			count++
		case letter == ' ':
			correct = false
		default:
			count++
		}
	}
	return correct && number && upper && lower && (count >= 7)
}
