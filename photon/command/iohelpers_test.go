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
	"regexp"
	"testing"
)

func TestTimestampToString(t *testing.T) {

	timeString := timestampToString(-1)
	if timeString != "-" {
		t.Error("timestampToString didn't return '-'")
	}
	timeString = timestampToString(0)
	if timeString != "-" {
		t.Error("timestampToString didn't return '-'")
	}
	timeString = timestampToString(1)
	matched, err := regexp.MatchString(`^[\d]+-[\d]+-[\d]+ [\d]+:[\d]+:[\d]+\.[\d]+$`, timeString)
	if !matched || err != nil {
		t.Error("timestampToString didn't return a timestamp")
		//("2006-01-02 03:04:05.00")
	}
}
