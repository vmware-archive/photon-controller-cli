// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package utils

/**
 * These utilities format output in a variety of ways.
 *
 * The goal is to have multiple methods of output so that it's easy to script the CLI
 * in whatever way a user wants. Currently we just implement JSON output, but we plan
 * to implement a subset of the JSONPath spec so that we can implement:
 * - output just a single value (e.g. ID) from an object
 * - output a list of objects as a table with the columns specified by the user
 *
 * In order to make life easier for callers, they pass us the CLI context and we examine
 * the arguments in here. Note that the arguments are global arguments (they occur before
 * the subcommand) because they apply uniformly to all subcommands.
 */

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/codegangsta/cli"
)

// Called by main to validate the output arguments
// Currently it only validates the --output argument. Eventually when we support
// json path, it will validate that as well.
func ValidateArgs(c *cli.Context) error {
	if c.GlobalBool("non-interactive") == true && c.GlobalString("output") != "" {
		return fmt.Errorf("--non-interactive and --output are mutually exclusive")
	}
	if c.GlobalString("output") != "" && c.GlobalString("output") != "json" {
		return fmt.Errorf("output type must be 'json'")
	}
	if c.GlobalBool("detail") == true && c.GlobalString("output") != "" {
		return fmt.Errorf("--detail and --output are mutually exclusive")
	}
	if c.GlobalBool("detail") == true && c.GlobalBool("non-interactive") == true {
		return fmt.Errorf("--detail and --non-interactive are mutually exclusive")
	}
	return nil
}

// Tells the caller if the user has requested custom formatting
func NeedsFormatting(c *cli.Context) bool {
	return c.GlobalString("output") != ""
}

// Outputs the given object (image, list of images, VM, etc...) as specified by the user
// Currently we only support JSON output, but we will support more formats later.
func FormatObject(o interface{}, w io.Writer, c *cli.Context) {
	outputType := c.GlobalString("output")
	switch outputType {
	case "json":
		formatObjectJson(o, w)
	default:
		fmt.Fprintf(w, "Unknown output type: '%s'", outputType)
	}
}

// Just like FormatObject, but if the incoming object is nil, it prints
// it as an empty list, not "null". This makes for nicer JSON output for
// list commands, which return lists of objects
func FormatObjects(o interface{}, w io.Writer, c *cli.Context) {
	value := reflect.ValueOf(o)
	kind := value.Kind()
	if (kind == reflect.Array || kind == reflect.Slice) && value.Len() == 0 {
		FormatObject(new([0]int), w, c)
	} else {
		FormatObject(o, w, c)
	}
}

// Ouptut an object as JSON
func formatObjectJson(o interface{}, w io.Writer) {
	jsonBytes, err := json.Marshal(o)
	if err != nil {
		fmt.Fprintf(w, "Cannot convert output to JSON: %s", err)
		return
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, jsonBytes, "", "  ")
	if err != nil {
		fmt.Fprintf(w, "Cannot format JSON output: %s", err)
		return
	}
	fmt.Fprintf(w, "%s\n", string(prettyJSON.Bytes()))
}
