package command

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"

	"github.com/vmware/photon-controller-cli/photon/cli/client"
)

// Creates a cli.Command for tasks
// Subcommands: list; Usage: task list [<options>]
//              show; Usage: task show <id>
//              monitor; Usage: task monitor <id>
func GetTasksCommand() cli.Command {
	command := cli.Command{
		Name:  "task",
		Usage: "options for task",
		Subcommands: []cli.Command{
			{
				Name:  "list",
				Usage: "list all tasks",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "entityId, e",
						Usage: "specify entity ID for filtering",
					},
					cli.StringFlag{
						Name:  "entityKind, k",
						Usage: "specify entity kind for filtering(tenant, project, vm etc)",
					},
					cli.StringFlag{
						Name:  "state, s",
						Usage: "specify task state for filtering",
					},
				},
				Action: func(c *cli.Context) {
					err := listTasks(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:  "show",
				Usage: "Show task info with specified ID",
				Action: func(c *cli.Context) {
					err := showTask(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:  "monitor",
				Usage: "Monitor task progress with specified ID",
				Action: func(c *cli.Context) {
					err := monitorTask(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
		},
	}
	return command
}

// Retrieves a list of tasks, returns an error if one occurred
func listTasks(c *cli.Context) error {
	err := checkArgNum(c.Args(), 0, "task list <options>")
	if err != nil {
		return err
	}
	entityId := c.String("entityId")
	entityKind := c.String("entityKind")
	state := c.String("state")

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	options := &photon.TaskGetOptions{
		State:      state,
		EntityID:   entityId,
		EntityKind: entityKind,
	}
	taskList, err := client.Esxclient.Tasks.GetAll(options)
	if err != nil {
		return err
	}

	err = printTaskList(taskList.Items, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}
	return nil
}

// Show the task current state, returns an error if one occurred
func showTask(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "task show <task id>")
	if err != nil {
		return err
	}
	id := c.Args()[0]

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	var apiErrorList []photon.ApiError
	task, err := client.Esxclient.Tasks.Get(id)
	if err != nil {
		if task == nil {
			return err
		} else {
			apiErrorList = getTaskAPIErrorList(task)
		}
	}

	if c.GlobalIsSet("non-interactive") {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\n", task.ID, task.State, task.Entity.ID, task.Entity.Kind, apiErrorList)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "Task:\t%s\n", task.ID)
		fmt.Fprintf(w, "Entity:\t%s %s\n", task.Entity.Kind, task.Entity.ID)
		fmt.Fprintf(w, "State:\t%s\n\n", task.State)
		err := w.Flush()
		if err != nil {
			return err
		}
		if len(apiErrorList) != 0 {
			fmt.Printf("The following error was encountered while running the task:\n\n")
			fmt.Printf("%s\n\n", apiErrorList)
		}
	}

	return nil
}

// Track the progress of the task, returns an error if one occurred
func monitorTask(c *cli.Context) error {
	err := checkArgNum(c.Args(), 1, "task monitor <task id>")
	if err != nil {
		return err
	}
	id := c.Args()[0]

	client.Esxclient, err = client.GetClient(c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		task, err := client.Esxclient.Tasks.Wait(id)
		if err != nil {
			return err
		}
		fmt.Printf("%s\t%s\t%s\t%s\n", task.ID, task.State, task.Entity.ID, task.Entity.Kind)
	} else {
		task, err := pollTask(id)
		if err != nil {
			return err
		}
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "Task:\t%s\n", task.ID)
		fmt.Fprintf(w, "Entity:\t%s %s\n", task.Entity.Kind, task.Entity.ID)
		fmt.Fprintf(w, "State:\t%s\n", task.State)
		err = w.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}
