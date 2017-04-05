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
	"errors"

	"github.com/vmware/photon-controller-cli/photon/client"

	"github.com/urfave/cli"
)

const (
	PHYSICAL         = "PHYSICAL"
	SOFTWARE_DEFINED = "SOFTWARE_DEFINED"
	NOT_AVAILABLE    = "NOT_AVAILABLE"
)

func isSoftwareDefinedNetwork(c *cli.Context) (sdnEnabled bool, err error) {
	client.Photonclient, err = client.GetClient(c)
	if err != nil {
		return
	}

	info, err := client.Photonclient.Info.Get()
	if err != nil {
		return
	}

	if info.NetworkType == NOT_AVAILABLE {
		err = errors.New("Network type is missing")
	} else {
		sdnEnabled = (info.NetworkType == SOFTWARE_DEFINED)
	}
	return
}
