package main

import (
	"os"

	"fmt"

	"github.com/vmware/photon-controller-cli/photon/cli/command"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
)

var commandName = ""
var githash = ""

func main() {
	app := BuildApp()
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func BuildApp() *cli.App {
	app := cli.NewApp()
	app.Name = commandName
	app.Usage = "Command line interface for Photon Controller"
	app.Version = "Git commit hash: " + githash
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "non-interactive, n",
			Usage: "trigger for non-interactive mode (scripting)",
		},
	}
	app.Commands = []cli.Command{
		command.GetAuthCommand(),
		command.GetSystemCommand(),
		command.GetTargetCommand(),
		command.GetTenantsCommand(),
		command.GetHostsCommand(),
		command.GetDeploymentsCommand(),
		command.GetResourceTicketCommand(),
		command.GetImagesCommand(),
		command.GetTasksCommand(),
		command.GetFlavorsCommand(),
		command.GetProjectsCommand(),
		command.GetDiskCommand(),
		command.GetVMCommand(),
		command.GetNetworksCommand(),
		command.GetClusterCommand(),
		command.GetAvailabilityZonesCommand(),
	}
	return app
}
