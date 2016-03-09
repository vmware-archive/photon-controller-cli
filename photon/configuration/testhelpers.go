// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package configuration

import (
	"os"
)

func RemoveConfigFile() error {
	filepath, err := getConfigurationFilePath()
	if err != nil {
		return err
	}

	if isFileExist(filepath) {
		err = os.Remove(filepath)
		if err != nil {
			return err
		}
	}

	return nil
}

func ChangeConfigFileContents(content string) error {
	filepath, err := getConfigurationFilePath()
	if err != nil {
		return err
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	err = file.Sync()
	if err != nil {
		return err
	}

	return nil
}
