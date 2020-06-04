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

// githubactions provides an SDK for authoring GitHub Actions in Go. It has no
// external dependencies and provides a Go-like interface for interacting with
// GitHub Actions' build system.
package githubactions

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const (
	addMaskFmt   = "::add-mask::%s\n"
	addPathFmt   = "::add-path::%s\n"
	setEnvFmt    = "::set-env name=%s::%s\n"
	setOutputFmt = "::set-output name=%s::%s\n"
	saveStateFmt = "::save-state name=%s::%s\n"

	addMatcherFmt    = "::add-matcher::%s\n"
	removeMatcherFmt = "::remove-matcher owner=%s::\n"

	groupFmt    = "::group::%s\n"
	endGroupFmt = "::endgroup::\n"

	debugFmt   = "::debug%s::%s\n"
	errorFmt   = "::error%s::%s\n"
	warningFmt = "::warning%s::%s\n"
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
	fields string
}

// AddMask adds a new field mask for the given string "p". After called, future
// attempts to log "p" will be replaced with "***" in log output.
func (c *Action) AddMask(p string) {
	fmt.Fprintf(c.w, addMaskFmt, p)
}

// AddMatcher adds a new matcher with the given file path.
func (c *Action) AddMatcher(p string) {
	fmt.Fprintf(c.w, addMatcherFmt, p)
}

// RemoveMatcher removes a matcher with the given owner name.
func (c *Action) RemoveMatcher(o string) {
	fmt.Fprintf(c.w, removeMatcherFmt, o)
}

// AddPath adds the string "p" to the path for the invocation.
func (c *Action) AddPath(p string) {
	fmt.Fprintf(c.w, addPathFmt, p)
}

// SaveState saves state to be used in the "finally" post job entry point.
func (c *Action) SaveState(k, v string) {
	fmt.Fprintf(c.w, saveStateFmt, k, v)
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
	fmt.Fprintf(c.w, groupFmt, t)
}

// EndGroup ends the current group.
func (c *Action) EndGroup() {
	fmt.Fprint(c.w, endGroupFmt)
}

// SetEnv sets an environment variable.
func (c *Action) SetEnv(k, v string) {
	fmt.Fprintf(c.w, setEnvFmt, k, v)
}

// SetOutput sets an output parameter.
func (c *Action) SetOutput(k, v string) {
	// escape sequences that GitHub actions will unescape when the output is used.
	// The list can update. For future reference, list can be found in JS/TS toolkit's
	// core/src/command.ts#escapeData().
	// https://github.com/actions/toolkit/blob/master/packages/core/src/command.ts
	v = strings.ReplaceAll(v, "%", "%25")
	v = strings.ReplaceAll(v, "\n", "%0A")
	v = strings.ReplaceAll(v, "\r", "%0D")
	fmt.Fprintf(c.w, setOutputFmt, k, v)
}

// Debugf prints a debug-level message. The arguments follow the standard Printf
// arguments.
func (c *Action) Debugf(msg string, args ...interface{}) {
	fmt.Fprintf(c.w, debugFmt, c.fields, fmt.Sprintf(msg, args...))
}

// Errorf prints a error-level message. The arguments follow the standard Printf
// arguments.
func (c *Action) Errorf(msg string, args ...interface{}) {
	fmt.Fprintf(c.w, errorFmt, c.fields, fmt.Sprintf(msg, args...))
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
	fmt.Fprintf(c.w, warningFmt, c.fields, fmt.Sprintf(msg, args...))
}

// WithFieldsSlice includes the provided fields in log output. "f" must be a
// slice of k=v pairs. The given slice will be sorted.
func (c *Action) WithFieldsSlice(f []string) *Action {
	sort.Strings(f)
	return &Action{
		w:      c.w,
		fields: " " + strings.Join(f, ","),
	}
}

// WithFieldsMap includes the provided fields in log output. The fields in "m"
// are automatically converted to k=v pairs and sorted.
func (c *Action) WithFieldsMap(m map[string]string) *Action {
	l := make([]string, 0, len(m))
	for k, v := range m {
		l = append(l, fmt.Sprintf("%s=%s", k, v))
	}
	return c.WithFieldsSlice(l)
}
