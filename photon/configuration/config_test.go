// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package configuration_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/vmware/photon-controller-cli/photon/configuration"
)

var _ = Describe("Config", func() {
	BeforeEach(func() {
		var err error
		UserConfigDir, err = ioutil.TempDir("", "config-test-")
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		err := RemoveConfigFile()
		err2 := os.Remove(UserConfigDir)
		Expect(err).To(BeNil())
		Expect(err2).To(BeNil())
	})

	Describe("LoadConfig", func() {
		Context("when config file does not exist", func() {
			BeforeEach(func() {
				err := RemoveConfigFile()
				Expect(err).To(BeNil())
			})

			It("retuns empty config and no error", func() {
				config, err := LoadConfig()

				Expect(err).To(BeNil())
				Expect(config).To(BeEquivalentTo(&Configuration{}))
			})
		})

		Context("when config file is not json", func() {
			BeforeEach(func() {
				nonJson := "<target>http://localhost:9080</target>\n"
				err := ChangeConfigFileContents(nonJson)
				Expect(err).To(BeNil())
			})

			It("returns empty config and error", func() {
				config, err := LoadConfig()

				Expect(err).To(
					MatchError("Error loading configuration: invalid character '<' looking for beginning of value"))
				Expect(config).To(BeEquivalentTo(&Configuration{}))
			})
		})

		Context("when config file is valid", func() {
			var (
				configExpected *Configuration
			)
			BeforeEach(func() {
				configExpected = &Configuration{
					CloudTarget: "http://localhost:9080",
				}

				err := SaveConfig(configExpected)
				Expect(err).To(BeNil())
			})

			It("returns the config", func() {
				config, err := LoadConfig()

				Expect(err).To(BeNil())
				Expect(config).To(BeEquivalentTo(configExpected))
			})
		})
	})

	Describe("SaveConfig", func() {
		Context("when config file does not exist", func() {
			BeforeEach(func() {
				err := RemoveConfigFile()
				Expect(err).To(BeNil())
			})

			It("saves to file", func() {
				configExpected := &Configuration{
					CloudTarget: "test-save-1",
				}

				err := SaveConfig(configExpected)
				Expect(err).To(BeNil())

				config, err := LoadConfig()
				Expect(err).To(BeNil())
				Expect(config).To(BeEquivalentTo(configExpected))
			})
		})

		Context("when config file exists", func() {
			BeforeEach(func() {
				config := "{CloudTarget: \"http://localhost:9080\"}"

				err := ChangeConfigFileContents(config)
				Expect(err).To(BeNil())
			})

			It("saves to updates to file", func() {
				configExpected := &Configuration{
					CloudTarget: "test-write-to-file-2",
				}

				err := SaveConfig(configExpected)
				Expect(err).To(BeNil())

				config, err := LoadConfig()
				Expect(err).To(BeNil())
				Expect(config).To(BeEquivalentTo(configExpected))
			})
		})
	})
})
