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
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/urfave/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"

	"github.com/vmware/photon-controller-cli/photon/client"
)

type stepSorter []photon.Step

func (step stepSorter) Len() int           { return len(step) }
func (step stepSorter) Swap(i, j int)      { step[i], step[j] = step[j], step[i] }
func (step stepSorter) Less(i, j int) bool { return step[i].Sequence < step[j].Sequence }

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
				Name:      "list",
				Usage:     "list all tasks",
				ArgsUsage: " ",
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
				Name:      "show",
				Usage:     "Show task info with specified ID",
				ArgsUsage: "<task-id>",
				Action: func(c *cli.Context) {
					err := showTask(c)
					if err != nil {
						log.Fatal(err)
					}
				},
			},
			{
				Name:      "monitor",
				Usage:     "Monitor task progress with specified ID",
				ArgsUsage: "<task-id>",
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
	err := checkArgCount(c, 0)
	if err != nil {
		return err
	}
	entityId := c.String("entityId")
	entityKind := c.String("entityKind")
	state := c.String("state")

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	options := &photon.TaskGetOptions{
		State:      state,
		EntityID:   entityId,
		EntityKind: entityKind,
	}
	taskList, err := client.Photonclient.Tasks.GetAll(options)
	if err != nil {
		return err
	}

	err = printTaskList(taskList.Items, c)
	if err != nil {
		return err
	}
	return nil
}

// Show the task current state, returns an error if one occurred
func showTask(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args()[0]

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	task, taskError := client.Photonclient.Tasks.Get(id)
	if taskError != nil && task == nil {
		return taskError
	}
	var resourceProperties string
	if task.ResourceProperties != nil {
		a, err := json.Marshal(task.ResourceProperties)
		if err != nil {
			fmt.Println("Error here ")
		}
		resourceProperties = string(a)
	}
	if c.GlobalIsSet("non-interactive") {
		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%d\t%d\t%v\n", task.ID, task.State, task.Entity.ID, task.Entity.Kind,
			task.Operation, task.StartedTime, task.EndTime, resourceProperties)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "Task:\t%s\n", task.ID)
		fmt.Fprintf(w, "Entity:\t%s %s\n", task.Entity.Kind, task.Entity.ID)
		fmt.Fprintf(w, "State:\t%s\n", task.State)
		fmt.Fprintf(w, "Operation:\t%s\n", task.Operation)
		fmt.Fprintf(w, "StartedTime:\t%s\n", timestampToString(task.StartedTime))
		fmt.Fprintf(w, "EndTime:\t%s\n", timestampToString(task.EndTime))
		if task.ResourceProperties != nil {
			fmt.Fprintf(w, "ResourceProperties:\t%v\n", resourceProperties)
		}
		err := w.Flush()
		if err != nil {
			return err
		}
	}
	err = printTaskSteps(task, c.GlobalIsSet("non-interactive"))
	if err != nil {
		return err
	}

	return nil
}

// Track the progress of the task, returns an error if one occurred
func monitorTask(c *cli.Context) error {
	err := checkArgCount(c, 1)
	if err != nil {
		return err
	}
	id := c.Args()[0]

	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return err
	}

	if c.GlobalIsSet("non-interactive") {
		task, err := client.Photonclient.Tasks.Wait(id)
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

func printTaskSteps(task *photon.Task, isScripting bool) error {
	if isScripting {
		for _, step := range task.Steps {
			fmt.Printf("%d\t%s\t%s\t%d\t%d\t%s\t%s\n", step.Sequence, step.Operation, step.State, step.StartedTime,
				step.EndTime, getApiErrorCode(step.Errors, ","), getApiErrorCode(step.Warnings, ","))
		}
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "Steps:\n")
		fmt.Fprintf(w, "\tOperation\tState\tStartedTime\tEndTime\tErrorCode\tWarningCode\n")
		steps := task.Steps
		sort.Sort(stepSorter(steps))
		for _, step := range steps {
			fmt.Fprintf(w, "\t%s\t%s\t%s\t%s\t%s\t%s\n", step.Operation, step.State,
				timestampToString(task.StartedTime),
				timestampToString(task.EndTime),
				getApiErrorCode(step.Errors, ", "), getApiErrorCode(step.Warnings, ", "))
		}
		err := w.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}

func getApiErrorCode(apiErrors []photon.ApiError, delim string) string {
	errors := []string{}
	for _, error := range apiErrors {
		errors = append(errors, fmt.Sprintf("%s", error.Code))
	}
	return strings.Join(errors, delim)
}
