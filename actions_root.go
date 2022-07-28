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

package githubactions

import (
	"context"
)

var (
	defaultAction = New()
)

// IssueCommand issues an arbitrary GitHub actions Command.
func IssueCommand(cmd *Command) {
	defaultAction.IssueCommand(cmd)
}

// IssueFileCommand issues a new GitHub actions Command using environment files.
func IssueFileCommand(cmd *Command) {
	defaultAction.IssueFileCommand(cmd)
}

// AddMask adds a new field mask for the given string "p". After called, future
// attempts to log "p" will be replaced with "***" in log output.
func AddMask(p string) {
	defaultAction.AddMask(p)
}

// AddMatcher adds a new matcher with the given file path.
func AddMatcher(p string) {
	defaultAction.AddMatcher(p)
}

// RemoveMatcher removes a matcher with the given owner name.
func RemoveMatcher(o string) {
	defaultAction.RemoveMatcher(o)
}

// AddPath adds the string "p" to the path for the invocation.
func AddPath(p string) {
	defaultAction.AddPath(p)
}

// SaveState saves state to be used in the "finally" post job entry point.
func SaveState(k, v string) {
	defaultAction.SaveState(k, v)
}

// GetInput gets the input by the given name.
func GetInput(i string) string {
	return defaultAction.GetInput(i)
}

// Group starts a new collapsable region up to the next ungroup invocation.
func Group(t string) {
	defaultAction.Group(t)
}

// EndGroup ends the current group.
func EndGroup() {
	defaultAction.EndGroup()
}

// AddStepSummary writes the given markdown to the job summary. If a job summary
// already exists, this value is appended.
func AddStepSummary(markdown string) {
	defaultAction.AddStepSummary(markdown)
}

// AddStepSummaryTemplate adds a summary template by parsing the given Go
// template using html/template with the given input data. See AddStepSummary
// for caveats.
//
// This primarily exists as a convenience function that renders a template.
func AddStepSummaryTemplate(tmpl string, data any) error {
	return defaultAction.AddStepSummaryTemplate(tmpl, data)
}

// SetEnv sets an environment variable.
func SetEnv(k, v string) {
	defaultAction.SetEnv(k, v)
}

// SetOutput sets an output parameter.
func SetOutput(k, v string) {
	defaultAction.SetOutput(k, v)
}

// Debugf prints a debug-level message. The arguments follow the standard Printf
// arguments.
func Debugf(msg string, args ...any) {
	defaultAction.Debugf(msg, args...)
}

// Noticef prints a notice-level message. The arguments follow the standard
// Printf arguments.
func Noticef(msg string, args ...any) {
	defaultAction.Noticef(msg, args...)
}

// Errorf prints a error-level message. The arguments follow the standard Printf
// arguments.
func Errorf(msg string, args ...any) {
	defaultAction.Errorf(msg, args...)
}

// Fatalf prints a error-level message and exits. This is equivalent to Errorf
// followed by os.Exit(1).
func Fatalf(msg string, args ...any) {
	defaultAction.Fatalf(msg, args...)
}

// Infof prints a info-level message. The arguments follow the standard Printf
// arguments.
func Infof(msg string, args ...any) {
	defaultAction.Infof(msg, args...)
}

// Warningf prints a warning-level message. The arguments follow the standard
// Printf arguments.
func Warningf(msg string, args ...any) {
	defaultAction.Warningf(msg, args...)
}

// WithFieldsSlice includes the provided fields in log output. "f" must be a
// slice of k=v pairs. The given slice will be sorted.
func WithFieldsSlice(f []string) *Action {
	return defaultAction.WithFieldsSlice(f)
}

// WithFieldsMap includes the provided fields in log output. The fields in "m"
// are automatically converted to k=v pairs and sorted.
func WithFieldsMap(m map[string]string) *Action {
	return defaultAction.WithFieldsMap(m)
}

// GetIDToken returns the GitHub OIDC token from the GitHub Actions runtime.
func GetIDToken(ctx context.Context, audience string) (string, error) {
	return defaultAction.GetIDToken(ctx, audience)
}

func Context() (*GitHubContext, error) {
	return defaultAction.Context()
}
