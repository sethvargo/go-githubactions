// Copyright 2020 The Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package githubactions provides an SDK for authoring GitHub Actions in Go. It
// has no external dependencies and provides a Go-like interface for interacting
// with GitHub Actions' build system.
package githubactions

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	addMaskCmd   = "add-mask"
	addPathCmd   = "add-path"
	setEnvCmd    = "set-env"
	setOutputCmd = "set-output"
	saveStateCmd = "save-state"

	addMatcherCmd    = "add-matcher"
	removeMatcherCmd = "remove-matcher"

	groupCmd    = "group"
	endGroupCmd = "endgroup"

	debugCmd   = "debug"
	errorCmd   = "error"
	warningCmd = "warning"
)

// New creates a new wrapper with helpers for outputting information in GitHub
// actions format.
func New() *Action {
	return &Action{w: os.Stdout}
}

// NewWithWriter creates a wrapper using the given writer. This is useful for
// tests. The given writer cannot add any prefixes to the string, since GitHub
// requires these special strings to match a very particular format.
func NewWithWriter(w io.Writer) *Action {
	return &Action{w: w}
}

// Action is an internal wrapper around GitHub Actions' output and magic
// strings.
type Action struct {
	w      io.Writer
	fields CommandProperties
}

// IssueCommand issues a new GitHub actions Command.
func (c *Action) IssueCommand(cmd *Command) {
	fmt.Fprintln(c.w, cmd.String())
}

// AddMask adds a new field mask for the given string "p". After called, future
// attempts to log "p" will be replaced with "***" in log output.
func (c *Action) AddMask(p string) {
	// ::add-mask::<p>
	c.IssueCommand(&Command{
		Name:    addMaskCmd,
		Message: p,
	})
}

// AddMatcher adds a new matcher with the given file path.
func (c *Action) AddMatcher(p string) {
	// ::add-matcher::<p>
	c.IssueCommand(&Command{
		Name:    addMatcherCmd,
		Message: p,
	})
}

// RemoveMatcher removes a matcher with the given owner name.
func (c *Action) RemoveMatcher(o string) {
	// ::remove-matcher owner=<o>::
	c.IssueCommand(&Command{
		Name: removeMatcherCmd,
		Properties: CommandProperties{
			"owner": o,
		},
	})
}

// AddPath adds the string "p" to the path for the invocation.
func (c *Action) AddPath(p string) {
	// ::add-path::<p>
	c.IssueCommand(&Command{
		Name:    addPathCmd,
		Message: p,
	})
}

// SaveState saves state to be used in the "finally" post job entry point.
func (c *Action) SaveState(k, v string) {
	// ::save-state name=<k>::<v>
	c.IssueCommand(&Command{
		Name:    saveStateCmd,
		Message: v,
		Properties: CommandProperties{
			"name": k,
		},
	})
}

// GetInput gets the input by the given name.
func (c *Action) GetInput(i string) string {
	e := strings.ReplaceAll(i, " ", "_")
	e = strings.ToUpper(e)
	e = "INPUT_" + e
	return strings.TrimSpace(os.Getenv(e))
}

// Group starts a new collapsable region up to the next ungroup invocation.
func (c *Action) Group(t string) {
	// ::group::<t>
	c.IssueCommand(&Command{
		Name:    groupCmd,
		Message: t,
	})
}

// EndGroup ends the current group.
func (c *Action) EndGroup() {
	// ::endgroup::
	c.IssueCommand(&Command{
		Name: endGroupCmd,
	})
}

// SetEnv sets an environment variable.
func (c *Action) SetEnv(k, v string) {
	// ::set-env name=<k>::<v>
	c.IssueCommand(&Command{
		Name:    setEnvCmd,
		Message: v,
		Properties: CommandProperties{
			"name": k,
		},
	})
}

// SetOutput sets an output parameter.
func (c *Action) SetOutput(k, v string) {
	// ::set-output name=<k>::<v>
	c.IssueCommand(&Command{
		Name:    setOutputCmd,
		Message: v,
		Properties: CommandProperties{
			"name": k,
		},
	})
}

// Debugf prints a debug-level message. The arguments follow the standard Printf
// arguments.
func (c *Action) Debugf(msg string, args ...interface{}) {
	// ::debug <c.fields>::<msg, args>
	c.IssueCommand(&Command{
		Name:       debugCmd,
		Message:    fmt.Sprintf(msg, args...),
		Properties: c.fields,
	})
}

// Errorf prints a error-level message. The arguments follow the standard Printf
// arguments.
func (c *Action) Errorf(msg string, args ...interface{}) {
	// ::error <c.fields>::<msg, args>
	c.IssueCommand(&Command{
		Name:       errorCmd,
		Message:    fmt.Sprintf(msg, args...),
		Properties: c.fields,
	})
}

// Fatalf prints a error-level message and exits. This is equivalent to Errorf
// followed by os.Exit(1).
func (c *Action) Fatalf(msg string, args ...interface{}) {
	c.Errorf(msg, args...)
	os.Exit(1)
}

// Warningf prints a warning-level message. The arguments follow the standard
// Printf arguments.
func (c *Action) Warningf(msg string, args ...interface{}) {
	// ::warning <c.fields>::<msg, args>
	c.IssueCommand(&Command{
		Name:       warningCmd,
		Message:    fmt.Sprintf(msg, args...),
		Properties: c.fields,
	})
}

// WithFieldsSlice includes the provided fields in log output. "f" must be a
// slice of k=v pairs. The given slice will be sorted.
func (c *Action) WithFieldsSlice(f []string) *Action {
	m := make(CommandProperties, 0)
	for _, s := range f {
		pair := strings.SplitN(s, "=", 2)
		if len(pair) < 2 {
			panic(fmt.Sprintf("%q is not a proper k=v pair!", s))
		}

		m[pair[0]] = pair[1]
	}

	return &Action{
		w:      c.w,
		fields: m,
	}
}

// WithFieldsMap includes the provided fields in log output. The fields in "m"
// are automatically converted to k=v pairs and sorted.
func (c *Action) WithFieldsMap(m map[string]string) *Action {
	// Not changing the function signature to 'map[string]interface{}' or
	// 'CommandProperties' to keep the API backwards-compatible. Perform a
	// manual type coversion instead.
	fields := make(CommandProperties, 0)
	for k, v := range m {
		fields[k] = v
	}

	return &Action{
		w:      c.w,
		fields: fields,
	}
}
