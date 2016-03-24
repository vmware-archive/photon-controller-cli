// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package manifest_test

import (
	. "github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/vmware/photon-controller-cli/Godeps/_workspace/src/github.com/onsi/gomega"

	"testing"
)

func TestManifest(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Manifest Suite")
}
