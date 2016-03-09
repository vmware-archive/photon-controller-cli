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
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// test Load when config file does not exist
	err := RemoveConfigFile()
	if err != nil {
		t.Error("Not expecting error removing config file")
	}

	config, err := LoadConfig()
	if err != nil {
		t.Error("Not expecting error reading non existing config file.")
	}

	configExpected := &Configuration{}
	if *config != *configExpected {
		t.Error("Expected to get empty config object when reading non existing config file.")
	}

	// test Load when config file is non json
	nonJson := "<target>http://localhost:9080</target>\n"
	err = ChangeConfigFileContents(nonJson)
	if err != nil {
		t.Error("Not expecting error writing string to config file")
	}

	config, err = LoadConfig()
	if err == nil {
		t.Error("Expected to receive error trying to load non json file.")
	}

	configExpected = &Configuration{}
	if *config != *configExpected {
		t.Error("Expected to get empty config object when reading non json config file.")
	}

	// test Load when target in config file is valid
	endpoint := "http://localhost:9080"
	configExpected = &Configuration{
		CloudTarget: endpoint,
	}
	err = SaveConfig(configExpected)
	if err != nil {
		t.Error("Not expecting error saving config file")
	}

	config, err = LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading endpoint when endpoint is valid.")
	}

	if *config != *configExpected {
		t.Error("Config read from file not match what is written to file.")
	}

}

func TestSaveConfig(t *testing.T) {
	// test Save to a new config file
	err := RemoveConfigFile()
	if err != nil {
		t.Error("Not expecting error removing config file")
	}

	configExpected := &Configuration{
		CloudTarget: "test-save-1",
	}
	err = SaveConfig(configExpected)
	if err != nil {
		t.Error("Not expecting error when saving to a new config file.")
	}

	config, err := LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file.")
	}

	if *config != *configExpected {
		t.Error("Configuration read from file not match what is written to file.")
	}

	// test Save to an existing config file
	configExpected = &Configuration{
		CloudTarget: "test-write-to-file-2",
	}
	err = SaveConfig(configExpected)
	if err != nil {
		t.Error("Not expecting error when saving to an existing config file.")
	}

	config, err = LoadConfig()
	if err != nil {
		t.Error("Not expecting error loading config file.")
	}

	if *config != *configExpected {
		t.Error("Configuration read from file not match what is written to file.")
	}
}
