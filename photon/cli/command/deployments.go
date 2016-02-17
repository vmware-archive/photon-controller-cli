package command

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/photon/cli/client"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
)

// Creates a cli.Command for deployments
// Subcommands: create;     Usage: deployment create [<options>]
//              delete;     Usage: deployment delete <id>
//              list;       Usage: deployment list
//              show;       Usage: deployment show <id>
//              list-hosts; Usage: deployment list-hosts <id>
//              list-vms;   Usage: deployment list-vms <id>
//              prepare-deployment-migration;   Usage: deployment prepare migration <sourceDeploymentAddress> <id>
//              finalize-deployment-migration;  Usage: deployment finalize migration <sourceDeploymentAddress> <id>
//              pause_system;                   Usage: deployment pause_system <id>
//              pause_background_tasks;         Usage: deployment pause_background_tasks <id>
//              resume_system;                  Usage: deployment resume_system <id>
func GetDeploymentsCommand() cli.Command {
	command := cli.Command{
		Name:  "deployment",
		Usage: "options for deployment",
		Subcommands: []cli.Command{
			{
				Name:  "create",
				Usage: "Create a new deployment",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "image_datastores, i",
						Usage: "Comma-separated image Datastore Names",
					},
					cli.StringFlag{
						Name:  "oauth_endpoint, o",
						Usage: "Oauth Endpoint",
					},
					cli.StringFlag{
						Name:  "oauth_tenant, t",
						Usage: "Oauth Tenant name",
					},
					cli.StringFlag{
						Name:  "oauth_username, u",
						Usage: "Oauth Username",
					},
					cli.StringFlag{
						Name:  "oauth_password, p",
						Usage: "Oauth Password",
					},
					cli.StringFlag{
						Name:  "syslog_endpoint, s",
						Usage: "Syslog Endpoint",
					},
					cli.StringFlag{
						Name:  "ntp_endpoint, n",
						Usage: "Ntp Endpoint",
					},
					cli.BoolFlag{
						Name:  "use_image_datastore_for_vms, v",
						Usage: "Use image Datastore for VMs",
					},
					cli.BoolFlag{
						Name:  "enable_auth, a",
						Usage: "Enable authentication/authorization for deployment",
					},
				},
				Action: func(c *cli.Context) {
					err := createDeployment(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "delete",
				Usage: "Delete a deployment by id",
				Action: func(c *cli.Context) {
					err := deleteDeployment(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "list",
				Usage: "Lists all the deployments",
				Action: func(c *cli.Context) {
					err := listDeployments(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "show",
				Usage: "Show deployment info",
				Action: func(c *cli.Context) {
					err := showDeployment(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "list-hosts",
				Usage: "Lists all the hosts associated with the deployment",
				Action: func(c *cli.Context) {
					err := listDeploymentHosts(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "list-vms",
				Usage: "Lists all the vms associated with the deployment",
				Action: func(c *cli.Context) {
					err := listDeploymentVms(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "prepare-deployment-migration",
				Usage: "Prepares migration of deployment from source to destination",
				Action: func(c *cli.Context) {
					err := initializeMigrateDeployment(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "finalize-deployment-migration",
				Usage: "Finalizes migration of deployment from source to destination",
				Action: func(c *cli.Context) {
					err := finalizeMigrateDeployment(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "update-image-datastores",
				Usage: "Updates the list of image datastores",
				Action: func(c *cli.Context) {
					err := updateImageDatastores(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "pause_system",
				Usage: "Pause system under the deployment",
				Action: func(c *cli.Context) {
					err := pauseSystem(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "pause_background_tasks",
				Usage: "Pause system's background tasks under the deployment",
				Action: func(c *cli.Context) {
					err := pauseBackgroundTasks(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
			{
				Name:  "resume_system",
				Usage: "Resume system under the deployment",
				Action: func(c *cli.Context) {
					err := resumeSystem(c)
					if err != nil {
						log.Fatal("Error: ", err)
					}
				},
			},
		},
	}
	return command
}

// Sends a create deployment task to client
func createDeployment(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "deployment create [<options>]")
	if err != nil {
		return err
	}

	imageDatastoreNames := c.String("image_datastores")
	oauthEndpoint := c.String("oauth_endpoint")
	oauthTenant := c.String("oauth_tenant")
	oauthUsername := c.String("oauth_username")
	oauthPassword := c.String("oauth_password")
	syslogEndpoint := c.String("syslog_endpoint")
	ntpEndpoint := c.String("ntp_endpoint")
	useDatastoreVMs := c.Bool("use_image_datastore_for_vms")
	enableAuth := c.Bool("enable_auth")

	if !c.GlobalIsSet("non-interactive") {
		var err error
		imageDatastoreNames, err =
			askForInput("Comma-separated image datastore names: ", imageDatastoreNames)
		if err != nil {
			return err
		}
		oauthEndpoint, err = askForInput("OAuth Endpoint: ", oauthEndpoint)
		if err != nil {
			return err
		}
		oauthTenant, err = askForInput("OAuth Tenant: ", oauthTenant)
		if err != nil {
			return err
		}
		oauthUsername, err = askForInput("OAuth Username: ", oauthUsername)
		if err != nil {
			return err
		}
		oauthPassword, err = askForInput("OAuth Password: ", oauthPassword)
		if err != nil {
			return err
		}
		syslogEndpoint, err = askForInput("Syslog Endpoint: ", syslogEndpoint)
		if err != nil {
			return err
		}
		ntpEndpoint, err = askForInput("Ntp Endpoint: ", ntpEndpoint)
		if err != nil {
			return err
		}
	}

	if len(imageDatastoreNames) == 0 {
		return fmt.Errorf("Image datastore names cannot be nil.")
	}
	imageDatastoreList := []string{}
	imageDatastoreList = regexp.MustCompile(`\s*,\s*`).Split(imageDatastoreNames, -1)

	authInfo := &photon.AuthInfo{
		Enabled:  enableAuth,
		Tenant:   oauthTenant,
		Endpoint: oauthEndpoint,
		Username: oauthUsername,
		Password: oauthPassword,
	}

	deploymentSpec := &photon.DeploymentCreateSpec{
		Auth:                    authInfo,
		ImageDatastores:         imageDatastoreList,
		NTPEndpoint:             ntpEndpoint,
		SyslogEndpoint:          syslogEndpoint,
		UseImageDatastoreForVms: useDatastoreVMs,
	}

	if len(ntpEndpoint) == 0 {
		deploymentSpec.NTPEndpoint = nil
	}

	if len(syslogEndpoint) == 0 {
		deploymentSpec.SyslogEndpoint = nil
	}

	if confirmed(c.GlobalIsSet("non-interactive")) {
		client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
		if err != nil {
			return err
		}

		createTask, err := client.Esxclient.Deployments.Create(deploymentSpec)
		if err != nil {
			return err
		}
		if c.GlobalIsSet("non-interactive") {
			task, err := client.Esxclient.Tasks.Wait(createTask.ID)
			if err != nil {
				return nil
			}
			fmt.Printf("%s\n", task.Entity.ID)
		} else {
			task, err := pollTask(createTask.ID)
			if err != nil {
				return err
			}
			fmt.Printf("Created deployment %s\n", task.Entity.ID)
		}

		return nil
	} else {
		fmt.Println("OK, canceled")
		return nil
	}
}

// Sends a delete deployment task to client
func deleteDeployment(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "deployment delete <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	deleteTask, err := client.Esxclient.Deployments.Delete(id)
	if err != nil {
		return err
	}

	err = waitOnTaskOperation(deleteTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Retrieves a list of deployments
func listDeployments(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "deployment list")
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

	if c.GlobalIsSet("non-interactive") {
		fmt.Println(len(deployments.Items))
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
func showDeployment(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "deployment show <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	deployment, err := client.Esxclient.Deployments.Get(id)
	if err != nil {
		return err
	}

	authEndpoint := deployment.Auth.Endpoint
	if len(deployment.Auth.Endpoint) == 0 {
		authEndpoint = "-"
	}
	authTenant := deployment.Auth.Tenant
	if len(deployment.Auth.Tenant) == 0 {
		authTenant = "-"
	}
	syslogEndpoint := deployment.SyslogEndpoint
	if len(deployment.SyslogEndpoint) == 0 {
		syslogEndpoint = "-"
	}
	ntpEndpoint := deployment.NTPEndpoint
	if len(deployment.NTPEndpoint) == 0 {
		ntpEndpoint = "-"
	}

	if c.GlobalIsSet("non-interactive") {
		fmt.Printf("%s\t%s\t%s\t%t\t%t\t%s\t%s\t%s\t%s\n", deployment.ID, deployment.State, deployment.ImageDatastores,
			deployment.UseImageDatastoreForVms, deployment.Auth.Enabled, authEndpoint, authTenant, syslogEndpoint, ntpEndpoint)

	} else {
		fmt.Printf("Deployment ID: %s\n", deployment.ID)
		fmt.Printf("  State:                       %s\n", deployment.State)
		fmt.Printf("  Image Datastores:            %s\n", deployment.ImageDatastores)
		fmt.Printf("  Use image datastore for vms: %t\n\n", deployment.UseImageDatastoreForVms)
		fmt.Printf("  Auth Enabled:                %t\n", deployment.Auth.Enabled)
		fmt.Printf("  Auth Endpoint:               %s\n", authEndpoint)
		fmt.Printf("  Auth Tenant:                 %s\n\n", authTenant)
		fmt.Printf("  Syslog Endpoint:             %s\n", syslogEndpoint)
		fmt.Printf("  Ntp Endpoint:                %s\n", ntpEndpoint)
	}

	return nil
}

// Lists all the hosts associated with the deployment
func listDeploymentHosts(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "deployment list-hosts <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	hosts, err := client.Esxclient.Deployments.GetHosts(id)
	if err != nil {
		return err
	}

	err = printHostList(hosts.Items, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Lists all the hosts associated with the deployment
func listDeploymentVms(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "deployment list-vms <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	vms, err := client.Esxclient.Deployments.GetVms(id)
	if err != nil {
		return err
	}

	err = printVMList(vms.Items, c.GlobalIsSet("non-interactive"), false)
	if err != nil {
		return err
	}

	return nil
}

// Sends a initialize migrate deployment task to client
func initializeMigrateDeployment(c *cli.Context) error {
	err := checkArgNum(c.Args(), 2, "deployment prepare migration <sourceDeploymentAddress> <id>")
	if err != nil {
		return err
	}
	sourceDeploymentAddress := c.Args()[0]
	id := c.Args()[1]

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	initializeMigrateTask, err := client.Esxclient.Deployments.InitializeDeploymentMigration(sourceDeploymentAddress, id)
	if err != nil {
		return err
	}

	err = waitOnTaskOperation(initializeMigrateTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Sends a finalize migrate deployment task to client
func finalizeMigrateDeployment(c *cli.Context) error {
	err := checkArgNum(c.Args(), 2, "deployment finalize migration <sourceDeploymentAddress> <id>")
	if err != nil {
		return err
	}
	sourceDeploymentAddress := c.Args()[0]
	id := c.Args()[1]

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	finalizeMigrateTask, err := client.Esxclient.Deployments.FinalizeDeploymentMigration(sourceDeploymentAddress, id)
	if err != nil {
		return err
	}

	err = waitOnTaskOperation(finalizeMigrateTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Update the image datastores using the information carried in cli.Context.
func updateImageDatastores(c *cli.Context) error {
	err := checkArgNum(c.Args(), 2, "deployment update-image-datastores <id> <comma separated image datastores>")
	if err != nil {
		return err
	}

	id := c.Args().First()
	imageDataStores := &photon.ImageDatastores{
		Items: regexp.MustCompile(`\s*,\s*`).Split(c.Args()[1], -1),
	}

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	task, err := client.Esxclient.Deployments.UpdateImageDatastores(id, imageDataStores)
	if err != nil {
		return err
	}

	fmt.Printf("Image datastores of deployment %s is finished\n", task.Entity.ID)
	return nil
}

// Sends a pause system task to client
func pauseSystem(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "deployment pause_system <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	pauseSystemTask, err := client.Esxclient.Deployments.PauseSystem(id)
	if err != nil {
		return err
	}

	err = waitOnTaskOperation(pauseSystemTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Sends a pause background task to client
func pauseBackgroundTasks(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "deployment pause_background_tasks <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	pauseBackgroundTask, err := client.Esxclient.Deployments.PauseBackgroundTasks(id)
	if err != nil {
		return err
	}

	err = waitOnTaskOperation(pauseBackgroundTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Sends a resume system task to client
func resumeSystem(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "deployment resume_system <id>")
	if err != nil {
		return err
	}
	id := c.Args().First()

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	resumeSystemTask, err := client.Esxclient.Deployments.ResumeSystem(id)
	if err != nil {
		return err
	}

	err = waitOnTaskOperation(resumeSystemTask.ID, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}
