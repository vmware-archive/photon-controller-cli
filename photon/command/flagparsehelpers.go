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
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/vmware/photon-controller-go-sdk/photon"
)

// Get limitsList from -limits/-l string flag
func parseLimitsListFromFlag(limits string) ([]photon.QuotaLineItem, error) {
	var limitsList []photon.QuotaLineItem
	if len(limits) != 0 {
		limitsListOri := regexp.MustCompile(`\s*,\s*`).Split(limits, -1)
		for i := 0; i < len(limitsListOri); i++ {
			limit := strings.Fields(limitsListOri[i])
			if len(limit) != 3 {
				return limitsList, fmt.Errorf("Error parsing limits, should be: <key> <value> <unit>, <key> <value> <unit>...")
			}

			key := limit[0]
			value, err := strconv.ParseFloat(limit[1], 64)
			if err != nil {
				return limitsList, fmt.Errorf("Error: %s. Please provide float as value", err.Error())
			}
			unit := limit[2]

			limitsList = append(limitsList, photon.QuotaLineItem{Key: key, Value: value, Unit: unit})
		}
	}
	return limitsList, nil
}

// Get affinitiesList from -affinities/-a string flag
func parseAffinitiesListFromFlag(affinities string) ([]photon.LocalitySpec, error) {
	var affinitiesList []photon.LocalitySpec
	if len(affinities) != 0 {
		affinitiesListOri := regexp.MustCompile(`\s*,\s*`).Split(affinities, -1)
		for i := 0; i < len(affinitiesListOri); i++ {
			affinity := regexp.MustCompile(`\s*:\s*`).Split(affinitiesListOri[i], -1)
			if len(affinity) != 2 {
				return affinitiesList, fmt.Errorf("Error parsing affinities, should be: <kind> <id>, <kind> <id>...")
			}

			kind := affinity[0]
			id := affinity[1]

			affinitiesList = append(affinitiesList, photon.LocalitySpec{Kind: kind, ID: id})
		}
	}
	return affinitiesList, nil
}

// Get disksList from -disks/-d string flag
func parseDisksListFromFlag(disks string) ([]photon.AttachedDisk, error) {
	var disksList []photon.AttachedDisk
	if len(disks) != 0 {
		disksListOri := regexp.MustCompile(`\s*,\s*`).Split(disks, -1)
		for i := 0; i < len(disksListOri); i++ {
			disk := strings.Fields(disksListOri[i])
			if len(disk) != 3 {
				return disksList, fmt.Errorf("Error parsing disks, should be: <name> <flavor> <boot=true/capacity>...")
			}

			name := disk[0]
			flavor := disk[1]
			if disk[2] == "boot=true" {
				disksList = append(disksList, photon.AttachedDisk{Name: name, Flavor: flavor, Kind: "ephemeral-disk", BootDisk: true})
			} else {
				capacity, err := strconv.Atoi(disk[2])
				if err != nil {
					return disksList, err
				}
				disksList = append(disksList, photon.AttachedDisk{Name: name, Flavor: flavor, Kind: "ephemeral-disk", BootDisk: false, CapacityGB: capacity})
			}
		}
	}
	return disksList, nil
}

// Get environment Map from -environment/-e string flag
func parseMapFromFlag(cmdFlag string) (map[string]string, error) {
	newMap := make(map[string]string)
	if len(cmdFlag) != 0 {
		entries := regexp.MustCompile(`\s*,\s*`).Split(cmdFlag, -1)
		for i := 0; i < len(entries); i++ {
			entry := regexp.MustCompile(`\s*:\s*`).Split(entries[i], -1)
			if len(entry) != 2 {
				return newMap, fmt.Errorf("Error parsing the command flag, should be: <key>:<value>, <key>:<value>...")
			}

			key := entry[0]
			value := entry[1]

			newMap[key] = value
		}
	}
	return newMap, nil
}
