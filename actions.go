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
	"io/ioutil"
	"os"
	"strings"
)

const (
	addMaskCmd   = "add-mask"
	setOutputCmd = "set-output"
	saveStateCmd = "save-state"

	addPathCmd = "add-path" // used when issuing the regular command
	pathCmd    = "path"     // used when issuing the file command

	setEnvCmd       = "set-env"        // used when issuing the regular command
	envCmd          = "env"            // used when issuing the file command
	envCmdMsgFmt    = "%s<<%s\n%s\n%s" // ${name}<<${delimiter}${os.EOL}${convertedVal}${os.EOL}${delimiter}
	envCmdDelimiter = "_GitHubActionsFileCommandDelimeter_"

	addMatcherCmd    = "add-matcher"
	removeMatcherCmd = "remove-matcher"

	groupCmd    = "group"
	endGroupCmd = "endgroup"

	debugCmd   = "debug"
	errorCmd   = "error"
	warningCmd = "warning"

	errFileCmdFmt = "unable to write command to the environment file: %s"
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

// IssueFileCommand issues a new GitHub actions Command using environment files.
//
// https://docs.github.com/en/free-pro-team@latest/actions/reference/workflow-commands-for-github-actions#environment-files
//
// The TypeScript equivalent function:
//
// https://github.com/actions/toolkit/blob/4f7fb6513a355689f69f0849edeb369a4dc81729/packages/core/src/file-command.ts#L10-L23
//
// IssueFileCommand currently ignores the 'CommandProperties' field provided
// with the 'Command' argument as it's scope is unclear in the current
// TypeScript implementation.
func (c *Action) IssueFileCommand(cmd *Command) error {
	return c.issueFileCommand(cmd, os.Getenv)
}

func (c *Action) issueFileCommand(cmd *Command, f getenvFunc) error {
	e := strings.ReplaceAll(cmd.Name, "-", "_")
	e = strings.ToUpper(e)
	e = "GITHUB_" + e

	err := ioutil.WriteFile(f(e), []byte(cmd.Message+"\n"), os.ModeAppend)
	if err != nil {
		return fmt.Errorf(errFileCmdFmt, err)
	}

	return nil
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

// AddPath adds the string "p" to the path for the invocation. It attempts to
// issue a file command at first. If that fails, it falls back to the regular
// (now deprecated) 'add-path' command, which may stop working in the future.
// The deprecated fallback may be useful for users running an older version of
// GitHub runner.
//
// https://docs.github.com/en/free-pro-team@latest/actions/reference/workflow-commands-for-github-actions#adding-a-system-path
// https://github.blog/changelog/2020-10-01-github-actions-deprecating-set-env-and-add-path-commands/
func (c *Action) AddPath(p string) {
	c.addPath(p, os.Getenv)
}

func (c *Action) addPath(p string, f getenvFunc) {
	err := c.issueFileCommand(&Command{
		Name:    pathCmd,
		Message: p,
	}, f)

	if err != nil { // use regular command as fallback
		// ::add-path::<p>
		c.IssueCommand(&Command{
			Name:    addPathCmd,
			Message: p,
		})
	}
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
	return c.getInput(i, os.Getenv)
}

func (c *Action) getInput(i string, f getenvFunc) string {
	e := strings.ReplaceAll(i, " ", "_")
	e = strings.ToUpper(e)
	e = "INPUT_" + e
	return strings.TrimSpace(f(e))
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

// SetEnv sets an environment variable. It attempts to issue a file command at
// first. If that fails, it falls back to the regular (now deprecated) 'set-env'
// command, which may stop working in the future. The deprecated fallback may be
// useful for users running an older version of GitHub runner.
//
// https://docs.github.com/en/free-pro-team@latest/actions/reference/workflow-commands-for-github-actions#setting-an-environment-variable
// https://github.blog/changelog/2020-10-01-github-actions-deprecating-set-env-and-add-path-commands/
func (c *Action) SetEnv(k, v string) {
	c.setEnv(k, v, os.Getenv)
}

func (c *Action) setEnv(k, v string, f getenvFunc) {
	err := c.issueFileCommand(&Command{
		Name:    envCmd,
		Message: fmt.Sprintf(envCmdMsgFmt, k, envCmdDelimiter, v, envCmdDelimiter),
	}, f)

	if err != nil { // use regular command as fallback
		// ::set-env name=<k>::<v>
		c.IssueCommand(&Command{
			Name:    setEnvCmd,
			Message: v,
			Properties: CommandProperties{
				"name": k,
			},
		})
	}
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
// slice of k=v pairs. The given slice will be sorted. It panics if any of the
// string in the given slice does not construct a valid 'key=value' pair.
func (c *Action) WithFieldsSlice(f []string) *Action {
	m := make(CommandProperties)
	for _, s := range f {
		pair := strings.SplitN(s, "=", 2)
		if len(pair) < 2 {
			panic(fmt.Sprintf("%q is not a proper k=v pair!", s))
		}

		m[pair[0]] = pair[1]
	}

	return c.WithFieldsMap(m)
}

// WithFieldsMap includes the provided fields in log output. The fields in "m"
// are automatically converted to k=v pairs and sorted.
func (c *Action) WithFieldsMap(m map[string]string) *Action {
	return &Action{
		w:      c.w,
		fields: m,
	}
}

// getenvFunc is an abstraction to make tests feasible for commands that
// interact with environment variables.
type getenvFunc func(k string) string
