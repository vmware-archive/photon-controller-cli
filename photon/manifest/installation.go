// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package manifest

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"regexp"
)

type Installation struct {
	Deployment deployment `yaml:"deployment"`
	Hosts      []host     `yaml:"hosts"`
}

type deployment struct {
	ResumeSystem            bool            `yaml:"resume_system"`
	ImageDatastores         imageDatastores `yaml:"image_datastores"`
	UseImageDatastoreForVms bool            `yaml:"use_image_datastore_for_vms"`

	SyslogEndpoint interface{} `yaml:"syslog_endpoint"`
	NTPEndpoint    interface{} `yaml:"ntp_endpoint"`

	LoadBalancerEnabled string `yaml:"enable_loadbalancer"`

	StatsEnabled       bool   `yaml:"stats_enabled"`
	StatsStoreEndpoint string `yaml:"stats_store_endpoint"`
	StatsPort          int    `yaml:"stats_port"`

	AuthEnabled        bool     `yaml:"auth_enabled"`
	AuthUsername       string   `yaml:"oauth_username"`
	AuthPassword       string   `yaml:"oauth_password"`
	AuthTenant         string   `yaml:"oauth_tenant"`
	AuthSecurityGroups []string `yaml:"oauth_security_groups"`

	SdnEnabled             bool   `yaml:"sdn_enabled"`
	NetworkManagerAddress  string `yaml:"network_manager_address"`
	NetworkManagerUsername string `yaml:"network_manager_username"`
	NetworkManagerPassword string `yaml:"network_manager_password"`
	NetworkZoneId          string `yaml:"network_zone_id"`
	NetworkTopRouterId     string `yaml:"network_top_router_id"`
	NetworkIpRange         string `yaml:"network_ip_range"`
	NetworkFloatingIpRange string `yaml:"network_floating_ip_range"`
}

type host struct {
	IpRanges         string            `yaml:"address_ranges"`
	Username         string            `yaml:"username"`
	Password         string            `yaml:"password"`
	AvailabilityZone string            `yaml:"availability_zone"`
	Tags             []string          `yaml:"usage_tags"`
	Metadata         map[string]string `yaml:"metadata"`
}

func LoadInstallation(file string) (res *Installation, err error) {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	res = &Installation{}
	err = yaml.Unmarshal(buf, res)
	if err != nil {
		return nil, err
	}
	return
}

type imageDatastores []string

func (d *imageDatastores) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	// try un-marshalling as an array
	var imageArray []string
	err = unmarshal(&imageArray)
	if err == nil {
		*d = imageDatastores(imageArray)
		return nil
	}

	// try un-marshalliong as a comma separated string
	var imageList string
	err = unmarshal(&imageList)
	if err != nil {
		return err
	}

	*d = regexp.MustCompile(`\s*,\s*`).Split(imageList, -1)
	return
}
