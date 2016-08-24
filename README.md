ESXCloud Go Cli
===============

The repo for the ESXCloud go cli.

## Setup

The project requires go version 1.5+. You can download and install go from: https://golang.org/dl/

Decide a folder as the GOPATH, e.g. ~/go.

    1. mkdir -p ~/go/src/github.com/vmware/
    2. cd ~/go/src/github.com/vmware/
    3. git clone (this repo from gerrit or github)
    4. export GOPATH=~/go
    5. export PATH=$PATH:~/go/bin
    6. make tools
    7. godep restore


## Build and Test

To run the test:

      make test

To build the executables:

      make build

The executables are generated under photon-controller-cli/bin folder.

To run and verify the CLI:

       ./bin/photon -v


## Pick up Changes from SDK

When there are changes in SDK, wait for them promoted to **_master_** branch on **_github.com_**.

Follow the steps below:

    1. go get -u github.com/vmware/photon-controller-go-sdk/photon
    2. govendor update github.com/vmware/photon-controller-go-sdk/...

Before comitting the change, carefully inspect the changes to Godeps, for example with git diff or SourceTree.

Then you can commit and submit the change.

## Example Usage

Notes:

* Commands below are shown as you type-them. The "%" indicates a command-line prompt.
* These are illustrative examples. Not all commands or parameters are described.


### Getting help

The photon CLI includes extensive usage description that you can use to
discover various operations.

For instance, you can see the top-level commands with:

    % photon help
    NAME:
       photon - Command line interface for Photon Controller

    USAGE:
       photon [global options] command [command options] [arguments...]

    VERSION:
       Git commit hash: 4607b29

    COMMANDS:
       auth			options for auth
       system		options for system operations
       target		options for target
       tenant		options for tenant
       host			options for host
       deployment		options for deployment
       resource-ticket	options for resource-ticket
       image		options for image
       task			options for task
       flavor		options for flavor
       project		options for project
       disk			options for disk
       vm			options for vm
       network		options for network
       cluster		Options for clusters
       availability-zone	options for availability-zone
       help, h		Shows a list of commands or help for one command

    GLOBAL OPTIONS:
       --non-interactive, -n	trigger for non-interactive mode (scripting)
       --help, -h			show help
       --version, -v		print the version

You can see help for an individual command too:

    % photon tenant --help
    NAME:
       photon tenant - options for tenant

    USAGE:
       photon tenant command [command options] [arguments...]

    COMMANDS:
       create		Create a new tenant
       delete		Delete a tenant
       list			List tenants
       set			Select tenant to work with
       show			Show current tenant
       tasks		Show tenant tasks
       set_security_groups	Set security groups for a tenant
       help, h		Shows a list of commands or help for one command

    OPTIONS:
       --help, -h	show help

### Interactive vs. Non-interactive mode
All commands work in two mode: interactive and non-interactive. Interactive
mode will prompt you for parameters you do not provide on the command-line
and will print human-readable output. Non-interactive mode will not prompt
you and will print machine-readable output.

### IDs
Objects in Photon Controller are given unique IDs, and most commands
refer to them using those IDs.

### Setting a target
Before you can use the photon CLI, you need to tell it which Photon Controller
to use.

Usage: `photon target set <PHOTON-CONTROLLER-URL>`

Example:

    % photon target set https://10.118.96.41
    API target set to 'https://10.118.96.41'

If you are not using HTTPS, specify the port:

    % photon target set http://10.118.96.41:9000
    API target set to 'http://10.118.96.41:9000'

### Tenants

Creating a tenant will tell you the ID of the tenant:

Usage: `photon tenant create <TENANT-NAME>`

Example:

    % photon -n tenant create cloud-dev
    cloud-dev	502f9a79-96b6-451d-bfb9-6292ca5b6cfd

You can list all tenants:

    % photon -n tenant list
    502f9a79-96b6-451d-bfb9-6292ca5b6cfd	cloud-dev

### Set tenant for other commands
Many commands take a --tenant parameter because the object in
question is owned by a tenant. As a convenience, you can avoid
passing that parameter, you can set the tenant for future commands:

Usage: `photon tenant set <TENANT-NAME>`

Example:

    % photon tenant set cloud-dev
    Tenant set to 'cloud-dev'

The tenant will be stored in a configuration file in your home directory,
within a subdirectory named _.photon-config_.

You can see what the current tenant is:

    % photon tenant get
    Current tenant is 'cloud-dev' 502f9a79-96b6-451d-bfb9-6292ca5b6cfd

### Resource tickets
You create resource tickets to control the allocations granted to projects,
which are owned by tenants.

A resource ticket must specify the number of VMs that can be created as well
as the total amount of RAM consumed by those VMs. It's possible to have user-defined
resources as well. These are specified as comma-separated limits, and each
limit is a set of three things:

* Name (e.g. vm.memory)
* Value (e.g. 2000)
* Units (GB, MB, KB, COUNT)

Creating a ticket in the current tenant (see above, or use the --tenant flag):

Usage `photon resource-ticket create --name <RESOURCE-TICKET-NAME> --limits "<LIMITS>"`

Example:

    % photon -n resource-ticket create --name cloud-dev-resources --limits "vm.memory 2000 GB, vm 1000 COUNT"
    32ad527e-d21a-4b2a-a235-b0883bd64354

Creating a ticket with user-defined resources:

    % photon -n resource-ticket create --name cloud-dev-resources --limits "vm.memory 2000 GB, vm 1000 COUNT vm.potrzebie 250 COUNT"
    32ad527e-d21a-4b2a-a235-b0883bd64354


Viewing tickets:

    % photon -n resource-ticket list
    1
    32ad527e-d21a-4b2a-a235-b0883bd64354	cloud-dev-resources vm.memory:2000:GB,vm:1000:COUNT

    % photon -n resource-ticket show cloud-dev-resources
    cloud-dev-resources	32ad527e-d21a-4b2a-a235-b0883bd64354 vm.memory:2000:GB,vm:1000:COUNT vm.memory:0:GB,vm:0:COUNT

    % photon resource-ticket show cloud-dev-resources
    ID                                    Name                 Limit              Usage
    32ad527e-d21a-4b2a-a235-b0883bd64354  cloud-dev-resources  vm.memory 2000 GB  vm.memory 1000 GB
                                                               vm 1000 COUNT      vm 500 COUNT

### Projects
A project is owned by a tenant and all VMs are created within a project. Each
project is associated with a resource ticket that controls the total resources
that can be used. See above for more information about resource tickets.

A project has a set of limits. These are specified just like the resource ticket above,
but they must not exceed the limits in the associated resource ticket.

Creating a project:

Usage: `photon project create --resource-ticket <RESOURCE-TICKET-NAME> --name <PROJECT-NAME> --limits <LIMITS>`

    % photon -n project create --resource-ticket cloud-dev-resources --name cloud-dev-staging --limits "vm.memory 1000 GB, vm 500 COUNT"
    fabb9236-d0a4-4d30-8935-ee65d6729f78

Viewing projects:

    % photon -n project list
    fabb9236-d0a4-4d30-8935-ee65d6729f78 cloud-dev-staging vm.memory:1000:GB,vm:500:COUNT vm.memory:0:GB,vm:0:COUNT

Setting the project (applies to commands that require a project, like creating a VM).
If you prefer, you can pass the --project flag:

    % photon -n project set cloud-dev-staging

### Flavors
When a VM is made, it is described using two kinds of flavors: VM and disk. The flavors
describes how many resources are consumed by the VM from the resource ticket.

The cost argument specifies a set of costs, each separated by commas. Each cost consists of three value:

* Name
* Value, which will be subtracted from the resource ticket when the VM is created
* Units: GB, MB, KB, B, or COUNT

Note that VM flavors must specify at least the vm.cpu and vm.memory costs. Other
user-defined costs may be included as well, if desired. They should match the
resources in the resource ticket.

Creating a VM flavor with with 1 VM, 1 CPU and 2 GB RAM:

Usage: `photon flavor create --name <FLAVOR-NAME> --kind <KIND> --cost <COST>`

Example:

    % photon -n flavor create --name "cloud-vm-small" --kind "vm" --cost "vm 1.0 COUNT, vm.cpu 1.0 COUNT, vm.memory 2.0 GB"
    ddfb5be0-3355-46d3-9f2f-e28750eb201b

Creating a VM flavor with user-defined attributes:

    % photon -n flavor create --name "cloud-vm-small" --kind "vm" --cost "vm 1.0 COUNT, vm.cpu 1.0 COUNT, vm.memory 2.0 GB vm.potrzebie 10"
    ddfb5be0-3355-46d3-9f2f-e28750eb201b

Creating a disk flavor:

    % photon -n flavor create --name "cloud-disk" --kind "ephemeral-disk" --cost "ephemeral-disk 1.0 COUNT"
    78efc53a-88ce-4f09-9b5d-49662d21e56c

Viewing flavors:

    % photon -n flavor list
    78efc53a-88ce-4f09-9b5d-49662d21e56c	cloud-disk	ephemeral-disk	ephemeral-disk:1:COUNT
    ddfb5be0-3355-46d3-9f2f-e28750eb201b	cloud-vm-small	vm	vm:1:COUNT,vm.cpu:1:COUNT,vm.memory:2:GB

    % photon flavor show ddfb5be0-3355-46d3-9f2f-e28750eb201b
    Flavor ID: ddfb5be0-3355-46d3-9f2f-e28750eb201b
      Name:  cloud-vm-small
      Kind:  vm
      Cost:  [vm 1 COUNT vm.cpu 1 COUNT vm.memory 2 GB]
      State: READY

### Images

Uploading an image (OVA, OVF, or VMDK). The replication type is either EAGER or ON_DEMAND

Usage: `photon image create <IMAGE-FILENAME> -n <IMAGE-NAME> -i <TYPE>`

Example:

    % photon image create photon.ova -n photon-os -i EAGER
    Created image 'photon-os' ID: 8d0b9383-ff64-4112-85db-e8111e2269fc

Viewing images:

    % photon image list
    ID                                    Name       State  Size(Byte)   Replication_type  ReplicationProgress  SeedingProgress
    8d0b9383-ff64-4112-85db-e8111e2269fc  photon-os  READY  16777216146  EAGER             100.0%               100.0%

    Total: 1

    % photon image show 8d0b9383-ff64-4112-85db-e8111e2269fc
    Image ID: 8d0b9383-ff64-4112-85db-e8111e2269fc
      Name:                   photon-os
      State:                  READY
      Size:                   16777216146 Byte(s)
      Image Replication Type: EAGER
      Settings:
        scsi0.virtualDev : lsilogic
        ethernet0.virtualDev : vmxnet3

Deleting images:

    % photon image delete 8d0b9383-ff64-4112-85db-e8111e2269fc
    Are you sure [y/n]? y
    DELETE_IMAGE completed for 'image' entity 8d0b9383-ff64-4112-85db-e8111e2269fc

Note that if you delete an image that is being used by a VM, it will
go into the PENDING_DELETE state. It will be deleted once all VMs that
are using it have also been deleted.

    % photon image show 8d0b9383-ff64-4112-85db-e8111e2269fc
    Image ID: 8d0b9383-ff64-4112-85db-e8111e2269fc
      Name:                       kube
      State:                      PENDING_DELETE
      Size:                       16777216146 Byte(s)
      Image Replication Type:     EAGER
      Image Replication Progress: 100%
      Image Seeding Progress:     100%
      Settings:

### VMs

When you create a VM, you must specify both the VM and disk flavors. The disks parameter
lists a set of disks, separated by commas. Each disk is described by three values:

* name
* flavor
* Either "boot=true" or a size in GB for the disk

Usage: `photon -n vm create --name <VM-NAME> --image <IMAGE-ID> --flavor <VM-FLAVOR> --disk <DISK-DESCRIPTION>`

    % photon -n vm create --name vm-1 --image 8d0b9383-ff64-4112-85db-e8111e2269fc --flavor cloud-vm-small --disks "disk-1 cloud-disk boot=true"
    86911d88-a037-4576-9649-4df579abb88c

Starting a VM:

    % photon vm start 86911d88-a037-4576-9649-4df579abb88c
    START_VM completed for 'vm' entity 86911d88-a037-4576-9649-4df579abb88c

Viewing VMs. Note that the IP address will only be shown in the VM tools are installed on the VM:

    % photon vm list
    Using target 'http://10.118.96.41:9000'
    ID                                    Name  State
    86911d88-a037-4576-9649-4df579abb88c  vm-1  STARTED

    Total: 1
    STARTED: 1

    % photon vm show 86911d88-a037-4576-9649-4df579abb88c
    Using target 'http://10.118.96.41:9000'
    VM ID:  86911d88-a037-4576-9649-4df579abb88c
      Name:         vm-1
      State:        STARTED
      Flavor:       cloud-vm-small
      Source Image: 8d0b9383-ff64-4112-85db-e8111e2269fc
      Host:         10.160.98.190
      Datastore:    56d62db1-e77c3b0d-7ebe-005056a7d183
      Metadata:     map[]
      Disks:
        Disk 1:
          ID:        2000d3a5-aaba-40c1-b08e-ba8a70be6112
          Name:      disk-1
          Kind:      ephemeral-disk
          Flavor:    78efc53a-88ce-4f09-9b5d-49662d21e56c
          Capacity:  15
          Boot:      true
        Networks: 1
          Name:        VM Network
          IP Address:

Note that when the VM is created, it consumes some of the resources allocated
to the project, based on the definitions in the flavor:

    % photon project list
    Using target 'http://10.118.96.41:9000'
    ID                                    Name               Limit              Usage
    fabb9236-d0a4-4d30-8935-ee65d6729f78  cloud-dev-staging  vm.memory 1000 GB  vm.cpu 1 COUNT
                                                             vm 500 COUNT       vm.memory 2 GB
                                                                                vm 1 COUNT
                                                                                ephemeral-disk.capacity 15 GB
                                                                                ephemeral-disk 1 COUNT

    Total projects: 1

### Hosts

Adding an ESX host:

Usage: `photon host create -u <USER-NAME> -p <PASSWORD> -i <ADDRESS> --tag <CLOUD|MGMT> -d <DEPLOYMENT-ID>``

    % photon -n host create -u root -p MY-PASSWORD -i 10.160.105.139 --tag 'CLOUD' -d prod-deployment
    3a159e73-854f-4598-937f-909d503b1dc6

Viewing hosts:

    % photon deployment list-hosts prod-deployment
    ID                                    State  IP              Tags
    3a159e73-854f-4598-937f-909d503b1dc6  READY  10.160.105.139  CLOUD
    a5411f8c-84b6-4b58-9670-7728db7c4cac  READY  10.160.98.190   CLOUD

    Total: 2
