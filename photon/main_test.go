// Copyright (c) 2016 VMware, Inc. All Rights Reserved.
//
// This product is licensed to you under the Apache License, Version 2.0 (the "License").
// You may not use this product except in compliance with the License.
//
// This product may include a number of subcomponents with separate copyright notices and
// license terms. Your use of these subcomponents is subject to the terms and conditions
// of the subcomponent's license, as noted in the LICENSE file.

package main

import (
	"testing"

	"github.com/urfave/cli"
)

/*
 * Test struct containing a boolean to determine if the app recognizes global flag
 */
type TestFlagState struct {
	interactive bool
}

/*
 * Tests if global config is interactive if global flag is not set
 * Should return true
 */
func TestNoGlobalFlag(t *testing.T) {
	state := &TestFlagState{interactive: true}
	app := CreateGlobalFlagApp(state)

	args := []string{"program", "test"}
	err := app.Run(args)
	if err != nil {
		t.Error(err)
	}
	if !state.interactive {
		t.Error("Should not see a global flag n or non-interactive, the cli should be interactive")
	}
}

/*
 * Tests if global config is interactive if global flag --non-interactive is set
 * Should return false
 */
func TestGlobalFlag(t *testing.T) {
	state := &TestFlagState{interactive: true}
	app := CreateGlobalFlagApp(state)

	args := []string{"program", "--non-interactive", "test"}
	err := app.Run(args)
	if err != nil {
		t.Error(err)
	}
	if state.interactive {
		t.Error("Should see a global flag non-interactive, the cli should not be interactive")
	}
}

/*
 * Tests if global config is interactive if global flag alias -n is set
 * Should return false
 */
func TestGlobalFlagAlias(t *testing.T) {
	state := &TestFlagState{interactive: true}
	app := CreateGlobalFlagApp(state)

	args := []string{"program", "-n", "test"}
	err := app.Run(args)
	if err != nil {
		t.Error(err)
	}
	if state.interactive {
		t.Error("Should see a global flag n, the cli should not be interactive")
	}
}

/*
 * Tests if global config is interactive when using a subcommand
 */
func TestGlobalFlagSubcommand(t *testing.T) {
	state := &TestFlagState{interactive: true}
	app := CreateGlobalFlagApp(state)

	args := []string{"program", "-n", "subcommand", "nested"}
	err := app.Run(args)
	if err != nil {
		t.Error(err)
	}
	if state.interactive {
		t.Error("Should see a global flag n, the cli should not be interactive")
	}
}

/*
 * Creates an application that sets global flag for interaction or scripting
 * t is the test state to determine if the flag is properly set
 * returns the application
 */
func CreateGlobalFlagApp(t *TestFlagState) *cli.App {
	app := BuildApp()
	app.Commands = []cli.Command{
		{
			Name:  "test",
			Usage: "sets flag corresponding to if a flag is set or not",
			Action: func(c *cli.Context) {
				if c.GlobalIsSet("non-interactive") {
					t.interactive = false
				}
			},
		},
		{
			Name:  "subcommand",
			Usage: "subcommand global test",
			Subcommands: []cli.Command{
				{
					Name:  "nested",
					Usage: "subcommand to ensure global flag is global",
					Action: func(c *cli.Context) {
						if c.GlobalIsSet("non-interactive") {
							t.interactive = false
						}
					},
				},
			},
		},
	}
	return app
}
