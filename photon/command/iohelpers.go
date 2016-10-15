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
	"os"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/vmware/photon-controller-cli/photon/client"
	cf "github.com/vmware/photon-controller-cli/photon/configuration"
	"github.com/vmware/photon-controller-cli/photon/utils"

	"github.com/codegangsta/cli"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

type ClusterVM struct {
	ClusterVM photon.VM `json:"vm"`
	IPAddress string    `json:"ipAddress"`
}

// Prompt for input if name is empty
// Read each line as string and remove all spaces
func askForInput(msg string, name string) (string, error) {
	if len(name) != 0 {
		return name, nil
	}

	fmt.Printf(msg)
	consoleReader := bufio.NewReader(os.Stdin)

	line, err := consoleReader.ReadString('\n')
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(line), nil
}

func printHostList(hostList []photon.Host, w io.Writer, c *cli.Context) error {
	if c.GlobalIsSet("non-interactive") {
		for _, host := range hostList {
			tag := strings.Trim(fmt.Sprint(host.Tags), "[]")
			scriptTag := strings.Replace(tag, " ", ",", -1)
			fmt.Printf("%s\t%s\t%s\t%s\n", host.ID, host.State, host.Address, scriptTag)
		}
	} else if c.GlobalString("output") != "" {
		utils.FormatObjects(hostList, w, c)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "ID\tState\tIP\tTags\n")
		for _, host := range hostList {
			fmt.Fprintf(w, "%s\t%v\t%s\t%s\n", host.ID, host.State, host.Address, strings.Trim(fmt.Sprint(host.Tags), "[]"))
		}
		err := w.Flush()
		if err != nil {
			return err
		}
		fmt.Printf("\nTotal: %d\n", len(hostList))
	}

	return nil
}

// Prompt for input if limitsList is empty
func askForLimitList(limitsList []photon.QuotaLineItem) ([]photon.QuotaLineItem, error) {
	if len(limitsList) == 0 {
		for i := 1; ; i++ {
			fmt.Printf("\nLimit %d (ENTER to finish)\n", i)

			key, err := askForInput("Key: ", "")
			if err != nil {
				return limitsList, err
			}
			if len(key) == 0 {
				break
			}

			valueStr, err := askForInput("Value: ", "")
			if err != nil {
				return limitsList, err
			}
			value, err := strconv.ParseFloat(valueStr, 64)
			if err != nil {
				return limitsList, fmt.Errorf("Error: %s. Please provide float as value", err.Error())
			}

			unit, err := askForInput("Unit: ", "")
			if err != nil {
				return limitsList, err
			}
			limitsList = append(limitsList, photon.QuotaLineItem{Key: key, Value: value, Unit: unit})
		}
	}

	if len(limitsList) == 0 {
		return limitsList, fmt.Errorf("Please provide at least 1 limit")
	}

	return limitsList, nil
}

// Prompts the user to confirm action, will repeat until a response is yes, y, no, or n.
func confirmed(isScripting bool) bool {
	if !isScripting {
		response := ""
		for {
			response, _ = askForInput("Are you sure [y/n]? ", response)
			response = strings.ToLower(response)
			if response == "y" || response == "yes" {
				return true
			}
			if response == "n" || response == "no" {
				return false
			}
			fmt.Printf("Please enter \"yes\" or \"no\".\n")
			response = ""
		}
	}
	return true
}

// Prints out the output of tasks
func printTaskList(taskList []photon.Task, c *cli.Context) error {
	if c.GlobalIsSet("non-interactive") {
		for _, task := range taskList {
			fmt.Printf("%s\t%s\t%s\t%d\t%d\n", task.ID, task.State, task.Operation, task.StartedTime, task.EndTime-task.StartedTime)
		}
	} else if utils.NeedsFormatting(c) {
		utils.FormatObjects(taskList, os.Stdout, c)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "\nTask\tStart Time\tDuration\n")

		for _, task := range taskList {
			var duration int64
			startTime := timestampToString(task.StartedTime)
			if task.EndTime-task.StartedTime > 0 {
				duration = (task.EndTime - task.StartedTime) / 1000
			} else {
				duration = 0
			}
			fmt.Fprintf(w, "%s\t%s\t%.2d:%.2d:%.2d\n", task.ID, startTime, duration/3600, (duration/60)%60, duration%60)
			err := w.Flush()
			if err != nil {
				return err
			}
			fmt.Printf("%s, %s\n", task.Operation, task.State)
		}
		if len(taskList) > 0 {
			fmt.Printf("\nYou can run 'photon task show <id>' for more information\n")
		}
		fmt.Printf("Total: %d\n", len(taskList))
	}

	return nil
}

func printQuotaList(w *tabwriter.Writer, qliList []photon.QuotaLineItem, colEntry ...string) {
	for i := 0; i < len(qliList); i++ {
		if i == 0 {
			for j := 0; j < len(colEntry); j++ {
				fmt.Fprintf(w, "%s\t", colEntry[j])
			}
			fmt.Fprintf(w, "%s %g %s\n", qliList[i].Key, qliList[i].Value, qliList[i].Unit)
		} else {
			for j := 0; j < len(colEntry); j++ {
				fmt.Fprintf(w, "\t")
			}
			fmt.Fprintf(w, "%s %g %s\n", qliList[i].Key, qliList[i].Value, qliList[i].Unit)
		}
	}
}

func quotaLineItemListToString(qliList []photon.QuotaLineItem) string {
	scriptUsage := []string{}
	for _, u := range qliList {
		scriptUsage = append(scriptUsage, fmt.Sprintf("%s:%g:%s", u.Key, u.Value, u.Unit))
	}
	return strings.Join(scriptUsage, ",")
}

// Converts milliseconds since the epoch (as used in tasks) to a string
// If the time is zero or negative, returns the string "-"
// We do this because a timestamp of zero in a task means "undefined"
func timestampToString(timestamp int64) string {
	if timestamp <= 0 {
		return "-"
	} else {
		return time.Unix(timestamp/1000, 0).Format("2006-01-02 03:04:05.00")
	}
}

func checkArgNum(args []string, num int, usage string) error {
	if len(args) < num {
		return fmt.Errorf("Please provide argument. Usage: %s", usage)
	}
	if len(args) > num {
		return fmt.Errorf("Unknown arguments: %v. Usage: %s", args[num:], usage)
	}
	return nil
}

// Prompt for input if disksList is empty
func askForVMDiskList(disksList []photon.AttachedDisk) ([]photon.AttachedDisk, error) {
	if len(disksList) == 0 {
		for i := 1; ; i++ {
			fmt.Printf("\nDisk %d (ENTER to finish)\n", i)

			name, err := askForInput("Name: ", "")
			if err != nil {
				return disksList, err
			}
			if len(name) == 0 {
				break
			}

			flavor, err := askForInput("Flavor: ", "")
			if err != nil {
				return disksList, err
			}

			boot := false
			response := ""
			for {
				response, _ = askForInput("Boot disk? [y/n]: ", response)
				response = strings.ToLower(response)
				if response == "y" || response == "yes" {
					boot = true
					break
				}
				if response == "n" || response == "no" {
					break
				}
				fmt.Printf("Please enter \"yes\" or \"no\".\n")
				response = ""
			}

			var disk photon.AttachedDisk
			if boot {
				disk = photon.AttachedDisk{
					Name:     name,
					Flavor:   flavor,
					Kind:     "ephemeral-disk",
					BootDisk: true,
				}
			} else {
				capacityGB, err := askForInput("Capacity in GB: ", "")
				if err != nil {
					return disksList, err
				}
				capacity, err := strconv.Atoi(capacityGB)
				if err != nil {
					return disksList, fmt.Errorf("Error: %s. Please provide int as value", err.Error())
				}
				disk = photon.AttachedDisk{
					Name:       name,
					Flavor:     flavor,
					Kind:       "ephemeral-disk",
					BootDisk:   false,
					CapacityGB: capacity,
				}
			}

			disksList = append(disksList, disk)
		}
	}

	if len(disksList) == 0 {
		return disksList, fmt.Errorf("Please provide at least 1 disk")
	}

	return disksList, nil
}

func printVMList(vmList []photon.VM, w io.Writer, c *cli.Context, summaryView bool) error {
	stateCount := make(map[string]int)
	for _, vm := range vmList {
		stateCount[vm.State]++
	}

	if c.GlobalIsSet("non-interactive") {
		if !summaryView {
			for _, vm := range vmList {
				fmt.Printf("%s\t%s\t%s\n", vm.ID, vm.Name, vm.State)
			}
		}
	} else if c.GlobalString("output") != "" {
		utils.FormatObjects(vmList, w, c)
	} else {
		if !summaryView {
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 4, 4, 2, ' ', 0)
			fmt.Fprintf(w, "ID\tName\tState\n")
			for _, vm := range vmList {
				fmt.Fprintf(w, "%s\t%s\t%s\n", vm.ID, vm.Name, vm.State)
			}
			err := w.Flush()
			if err != nil {
				return err
			}
		}
		fmt.Printf("\nTotal: %d\n", len(vmList))
		for key, value := range stateCount {
			fmt.Printf("%s: %d\n", key, value)
		}
	}
	return nil
}

func printClusterList(clusterList []photon.Cluster, w io.Writer, c *cli.Context, summaryView bool) error {
	stateCount := make(map[string]int)
	for _, cluster := range clusterList {
		stateCount[cluster.State]++
	}

	if c.GlobalIsSet("non-interactive") {
		if !summaryView {
			for _, cluster := range clusterList {
				fmt.Printf("%s\t%s\t%s\t%s\t%d\n", cluster.ID, cluster.Name, cluster.Type, cluster.State, cluster.WorkerCount)
			}
		}
	} else if c.GlobalString("output") != "" {
		utils.FormatObjects(clusterList, w, c)
	} else {
		if !summaryView {
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 4, 4, 2, ' ', 0)
			fmt.Fprintf(w, "ID\tName\tType\tState\tWorker Count\n")
			for _, cluster := range clusterList {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\n", cluster.ID, cluster.Name, cluster.Type, cluster.State, cluster.WorkerCount)
			}
			err := w.Flush()
			if err != nil {
				return err
			}
		}
		fmt.Printf("\nTotal: %d\n", len(clusterList))
		for key, value := range stateCount {
			fmt.Printf("%s: %d\n", key, value)
		}
	}

	return nil
}

func printClusterVMs(vms []photon.VM, w io.Writer, c *cli.Context) (err error) {
	clusterVMs := []ClusterVM{}
	for _, vm := range vms {
		ipAddr := "-"
		networks, err := getVMNetworks(vm.ID, c)
		if err != nil {
			continue
		}
		for _, nt := range networks {
			network := nt.(map[string]interface{})
			if val, ok := network["network"]; !ok || val == nil {
				continue
			}
			if val, ok := network["ipAddress"]; ok && val != nil {
				ipAddr = val.(string)
				break
			}
		}
		clusterVM := ClusterVM{
			vm,
			ipAddr,
		}
		clusterVMs = append(clusterVMs, clusterVM)

	}

	if c.GlobalIsSet("non-interactive") {
		fmt.Printf("%d\n", len(vms))
		for _, clusterVM := range clusterVMs {
			fmt.Printf("%s\t%s\t%s\n", clusterVM.ClusterVM.ID,
				clusterVM.ClusterVM.Name, clusterVM.IPAddress)
		}
	} else if utils.NeedsFormatting(c) {
		utils.FormatObjects(clusterVMs, w, c)
	} else {
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "VM ID\tVM Name\tVM IP\n")
		for _, clusterVM := range clusterVMs {
			fmt.Fprintf(w, "%s\t%s\t%s\n", clusterVM.ClusterVM.ID,
				clusterVM.ClusterVM.Name, clusterVM.IPAddress)
		}
		err := w.Flush()
		if err != nil {
			return err
		}
	}
	return nil
}

func getVMNetworks(id string, c *cli.Context) (networks []interface{}, err error) {
	task, err := client.Esxclient.VMs.GetNetworks(id)
	if err != nil {
		return nil, err
	}

	if utils.IsNonInteractive(c) {
		task, err = client.Esxclient.Tasks.Wait(task.ID)
		if err != nil {
			return nil, err
		}
	} else {
		task, err = pollTask(task.ID)
		if err != nil {
			return nil, err
		}
	}
	networkConnections := task.ResourceProperties.(map[string]interface{})
	networks = networkConnections["networkConnections"].([]interface{})
	return networks, nil
}

func printVMNetworks(networks []interface{}, isScripting bool) error {
	w := new(tabwriter.Writer)
	if !isScripting {
		w.Init(os.Stdout, 4, 4, 2, ' ', 0)
		fmt.Fprintf(w, "Network\tMAC Address\tIP Address\tNetmask\tIsConnected\n")
	}
	for _, nt := range networks {
		networkName := "-"
		macAddr := "-"
		ipAddr := "-"
		netMask := "-"
		isConnected := "-"

		network := nt.(map[string]interface{})
		if val, ok := network["network"]; ok && val != nil {
			networkName = val.(string)
		}
		if val, ok := network["macAddress"]; ok && val != nil {
			macAddr = val.(string)
		}
		if val, ok := network["ipAddress"]; ok && val != nil {
			ipAddr = val.(string)
		}
		if val, ok := network["netmask"]; ok && val != nil {
			netMask = val.(string)
		}
		if val, ok := network["isConnected"]; ok && val != nil {
			isConnected = val.(string)
		}
		if isScripting {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", networkName, macAddr, ipAddr, netMask, isConnected)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", networkName, macAddr, ipAddr, netMask, isConnected)
		}
	}
	if !isScripting {
		err := w.Flush()
		if err != nil {
			return err
		}
		fmt.Printf("\nTotal: %d\n", len(networks))
	}
	return nil
}

// Clears the tenant in the config if it has a matching id
func clearConfigTenant(id string) error {
	config, err := cf.LoadConfig()
	if err != nil {
		return err
	}

	if config == nil || config.Tenant == nil {
		return nil
	}

	if len(id) == 0 || config.Tenant.ID == id {
		config.Tenant = nil
		err = cf.SaveConfig(config)
		if err != nil {
			return err
		}
		err = clearConfigProject("")
		if err != nil {
			return err
		}
	}
	return nil
}

// Finds the id of the tenant based on name, returns empty string with an error if it is not found
func findTenantID(name string) (string, error) {
	tenants, err := client.Esxclient.Tenants.GetAll()
	if err != nil {
		return "", err
	}
	var found bool
	var id string
	for _, tenant := range tenants.Items {
		if tenant.Name == name {
			found = true
			id = tenant.ID
			break
		}
	}
	if !found {
		return "", fmt.Errorf("Tenant name '%s' not found", name)
	}
	return id, nil
}

// Verifies and gets tenant name and id for commands specifying tenant
// Returns tenant in config file if name is empty
func verifyTenant(name string) (*cf.TenantConfiguration, error) {
	if len(name) != 0 {
		tenantID, err := findTenantID(name)
		if len(tenantID) == 0 || err != nil {
			return nil, err
		}
		return &cf.TenantConfiguration{Name: name, ID: tenantID}, nil
	}

	config, err := cf.LoadConfig()
	if err != nil {
		return nil, err
	}
	if config.Tenant == nil {
		return nil, fmt.Errorf("Error: Set tenant first using 'tenant set <name>' or '-t <name>' option")
	}

	return config.Tenant, nil
}

// Clears the project in the config if it has a matching id
func clearConfigProject(id string) error {
	config, err := cf.LoadConfig()
	if err != nil {
		return err
	}

	if config == nil || config.Project == nil {
		return nil
	}

	if len(id) == 0 || config.Project.ID == id {
		config.Project = nil
		err = cf.SaveConfig(config)
		if err != nil {
			return err
		}
	}
	return nil
}

// Finds the rt based on tenant id and rt name, returns nil with an error if it is not found
func findResourceTicket(tenantID string, name string) (*photon.ResourceTicket, error) {
	tickets, err := client.Esxclient.Tenants.GetResourceTickets(tenantID, &photon.ResourceTicketGetOptions{Name: name})
	if err != nil {
		return nil, err
	}
	rtList := tickets.Items

	if len(rtList) < 1 {
		return nil, fmt.Errorf("Error: Cannot find resource ticket named '%s'", name)
	}
	if len(rtList) > 1 {
		return nil, fmt.Errorf("Error: Found more than 1 resource ticket named '%s'", name)
	}

	return &rtList[0], nil
}

// Finds the project based on tenant id and project name, returns nil with an error if it is not found
func findProject(tenantID string, name string) (*photon.ProjectCompact, error) {
	tickets, err := client.Esxclient.Tenants.GetProjects(tenantID, &photon.ProjectGetOptions{Name: name})
	if err != nil {
		return nil, err
	}
	pList := tickets.Items

	if len(pList) < 1 {
		return nil, fmt.Errorf("Error: Cannot find project named '%s'", name)
	}
	if len(pList) > 1 {
		return nil, fmt.Errorf("Error: Found more than 1 projects named '%s'", name)
	}

	return &pList[0], nil
}

// Verifies and gets project name and id for commands specifying project
// Returns project in config file if name is empty
func verifyProject(tenantID string, name string) (*cf.ProjectConfiguration, error) {
	if len(name) != 0 {
		project, err := findProject(tenantID, name)
		if err != nil {
			return nil, err
		}
		return &cf.ProjectConfiguration{Name: name, ID: project.ID}, nil
	}

	config, err := cf.LoadConfig()
	if err != nil {
		return nil, err
	}
	if config.Project == nil {
		return nil, fmt.Errorf("Error: Set project first using 'project set <name>' or '-p <name>' option")
	}

	return config.Project, nil
}

// Return APIErrors of a task
func getTaskAPIErrorList(task *photon.Task) []photon.ApiError {
	var apiErrorList []photon.ApiError
	for i := 0; task != nil && i < len(task.Steps); i++ {
		apiErrorList = append(apiErrorList, task.Steps[i].Errors...)
	}
	return apiErrorList
}

var taskInProgress *photon.Task
var endAnimation bool

// Display state of taskInProgress while endAnimation is not true
// Print format:
// e.g:  0h: 0m: 0s [  ] CREATE_HOST : QUEUED
//       0h: 0m: 0s [= ] CREATE_HOST : CREATE_HOST | Step 1/1
//       0h: 0m: 1s [==] CREATE_HOST : COMPLETED
func displayTaskProgress(start time.Time) {
	cursor := 0
	displayInterval := 500 * time.Millisecond
	for !endAnimation {
		if taskInProgress != nil {
			var taskStatus string
			startedStep := findStartedStep(taskInProgress)

			if startedStep == nil {
				taskStatus = taskInProgress.State
			} else {
				cursor = startedStep.Sequence + 1
				taskStatus = fmt.Sprintf("%s | Step %d/%d",
					startedStep.Operation, startedStep.Sequence+1, len(taskInProgress.Steps))
			}

			fmt.Printf("\r%s\r", strings.Repeat(" ", 100))

			elapsed := int(time.Since(start).Seconds())
			fmt.Printf("%2dh%2dm%2ds ", elapsed/3600, (elapsed/60)%60, elapsed%60)
			fmt.Printf("[%s] ", getProgressBar(cursor, len(taskInProgress.Steps)+1))
			fmt.Printf("%s : %s", taskInProgress.Operation, taskStatus)
		}
		time.Sleep(displayInterval)
	}
	fmt.Printf("\r%s\r", strings.Repeat(" ", 100))
}

// Wait for task to finish and display task progress
func pollTask(id string) (task *photon.Task, err error) {
	return pollTaskWithTimeout(client.Esxclient, id, 30*time.Minute)
}

func pollTaskWithTimeout(api *photon.Client, id string, pollTimeout time.Duration) (task *photon.Task, err error) {
	start := time.Now()
	numErr := 0

	taskPollDelay := 500 * time.Millisecond
	taskRetryCount := 3

	endAnimation = false
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		displayTaskProgress(start)
	}()

	for time.Since(start) < pollTimeout {
		task, err = api.Tasks.Get(id)

		if err != nil {
			switch err.(type) {
			case photon.ApiError:
				apiErrorList := getTaskAPIErrorList(task)
				if len(apiErrorList) != 0 {
					err = fmt.Errorf("%s\nAPI Errors: %s", err.Error(), apiErrorList)
				}
				endAnimation = true
				wg.Wait()
				return
			default:
				apiErrorList := getTaskAPIErrorList(task)
				if len(apiErrorList) != 0 {
					err = fmt.Errorf("%s\nAPI Errors: %s", err.Error(), apiErrorList)
				}

				if task != nil && task.State == "ERROR" {
					endAnimation = true
					wg.Wait()
					return
				}

				numErr++
				if numErr > taskRetryCount {
					endAnimation = true
					wg.Wait()
					return
				}
			}
		} else {
			numErr = 0
			taskInProgress = task
			if task.State == "COMPLETED" {
				endAnimation = true
				wg.Wait()
				return
			}
		}

		time.Sleep(taskPollDelay)
	}

	endAnimation = true
	wg.Wait()
	err = fmt.Errorf("Timed out while waiting for task to complete")
	return
}

func findStartedStep(task *photon.Task) *photon.Step {
	for i := 0; task != nil && i < len(task.Steps); i++ {
		if task.Steps[i].State == "STARTED" {
			return &task.Steps[i]
		}
	}
	return nil
}

func getProgressBar(cursor int, len int) string {
	return strings.Repeat("=", cursor) + strings.Repeat(" ", len-cursor)
}

func waitOnTaskOperation(taskId string, c *cli.Context) (string, error) {
	var task *photon.Task
	var err error
	needsFormatting := utils.NeedsFormatting(c)
	if c.GlobalIsSet("non-interactive") || needsFormatting {
		task, err = client.Esxclient.Tasks.Wait(taskId)
		if err != nil {
			return "", err
		}
		if !needsFormatting {
			fmt.Println(task.Entity.ID)
		}
	} else {
		task, err = pollTask(taskId)
		if err != nil {
			return "", err
		}
		fmt.Printf("%s completed for '%s' entity %s\n", task.Operation, task.Entity.Kind, task.Entity.ID)
	}
	return task.Entity.ID, err
}

func getCommaSeparatedStringFromStringArray(arr []string) string {
	res := ""
	for _, element := range arr {
		res += element + ","
	}
	if res != "" {
		res = strings.TrimSuffix(res, ",")
	}
	return res
}
