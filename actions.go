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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sethvargo/go-envconfig"
)

var (
	// osExit allows `os.Exit()` to be stubbed during testing.
	osExit = os.Exit
)

const (
	addMaskCmd = "add-mask"

	envCmd    = "env"
	outputCmd = "output"
	pathCmd   = "path"
	stateCmd  = "state"

	// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#multiline-strings
	multiLineFileDelim = "_GitHubActionsFileCommandDelimeter_"
	multilineFileCmd   = "%s<<" + multiLineFileDelim + EOF + "%s" + EOF + multiLineFileDelim // ${name}<<${delimiter}${os.EOL}${convertedVal}${os.EOL}${delimiter}

	addMatcherCmd    = "add-matcher"
	removeMatcherCmd = "remove-matcher"

	groupCmd    = "group"
	endGroupCmd = "endgroup"

	stepSummaryCmd = "step-summary"

	debugCmd   = "debug"
	noticeCmd  = "notice"
	warningCmd = "warning"
	errorCmd   = "error"

	errFileCmdFmt = "unable to write command to the environment file: %s"
)

// New creates a new wrapper with helpers for outputting information in GitHub
// actions format.
func New(opts ...Option) *Action {
	a := &Action{
		w:      os.Stdout,
		getenv: os.Getenv,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		a = opt(a)
	}

	return a
}

// Action is an internal wrapper around GitHub Actions' output and magic
// strings.
type Action struct {
	w          io.Writer
	fields     CommandProperties
	getenv     GetenvFunc
	httpClient *http.Client
}

// IssueCommand issues a new GitHub actions Command. It panics if it cannot
// write to the output stream.
func (c *Action) IssueCommand(cmd *Command) {
	if _, err := fmt.Fprint(c.w, cmd.String()+EOF); err != nil {
		panic(fmt.Errorf("failed to issue command: %w", err))
	}
}

// IssueFileCommand issues a new GitHub actions Command using environment files.
// It panics if writing to the file fails.
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
func (c *Action) IssueFileCommand(cmd *Command) {
	if err := c.issueFileCommand(cmd); err != nil {
		panic(err)
	}
}

// issueFileCommand is an internal-only helper that issues the command and
// returns an error to make testing easier.
func (c *Action) issueFileCommand(cmd *Command) (retErr error) {
	e := strings.ReplaceAll(cmd.Name, "-", "_")
	e = strings.ToUpper(e)
	e = "GITHUB_" + e

	filepath := c.getenv(e)
	msg := []byte(cmd.Message + EOF)
	f, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		retErr = fmt.Errorf(errFileCmdFmt, err)
		return
	}

	defer func() {
		if err := f.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	if _, err := f.Write(msg); err != nil {
		retErr = fmt.Errorf(errFileCmdFmt, err)
		return
	}
	return
}

// AddMask adds a new field mask for the given string "p". After called, future
// attempts to log "p" will be replaced with "***" in log output. It panics if
// it cannot write to the output stream.
func (c *Action) AddMask(p string) {
	// ::add-mask::<p>
	c.IssueCommand(&Command{
		Name:    addMaskCmd,
		Message: p,
	})
}

// AddMatcher adds a new matcher with the given file path. It panics if it
// cannot write to the output stream.
func (c *Action) AddMatcher(p string) {
	// ::add-matcher::<p>
	c.IssueCommand(&Command{
		Name:    addMatcherCmd,
		Message: p,
	})
}

// RemoveMatcher removes a matcher with the given owner name. It panics if it
// cannot write to the output stream.
func (c *Action) RemoveMatcher(o string) {
	// ::remove-matcher owner=<o>::
	c.IssueCommand(&Command{
		Name: removeMatcherCmd,
		Properties: CommandProperties{
			"owner": o,
		},
	})
}

// AddPath adds the string "p" to the path for the invocation. It panics if it
// cannot write to the output file.
//
// https://docs.github.com/en/free-pro-team@latest/actions/reference/workflow-commands-for-github-actions#adding-a-system-path
// https://github.blog/changelog/2020-10-01-github-actions-deprecating-set-env-and-add-path-commands/
func (c *Action) AddPath(p string) {
	c.IssueFileCommand(&Command{
		Name:    pathCmd,
		Message: p,
	})
}

// SaveState saves state to be used in the "finally" post job entry point. It
// panics if it cannot write to the output stream.
//
// On 2022-10-11, GitHub deprecated "::save-state name=<k>::<v>" in favor of
// [environment files].
//
// [environment files]: https://github.blog/changelog/2022-10-11-github-actions-deprecating-save-state-and-set-output-commands/
func (c *Action) SaveState(k, v string) {
	c.IssueFileCommand(&Command{
		Name:    stateCmd,
		Message: fmt.Sprintf(multilineFileCmd, k, v),
	})
}

// GetInput gets the input by the given name. It returns the empty string if the
// input is not defined.
func (c *Action) GetInput(i string) string {
	e := strings.ReplaceAll(i, " ", "_")
	e = strings.ToUpper(e)
	e = "INPUT_" + e
	return strings.TrimSpace(c.getenv(e))
}

// Group starts a new collapsable region up to the next ungroup invocation. It
// panics if it cannot write to the output stream.
func (c *Action) Group(t string) {
	// ::group::<t>
	c.IssueCommand(&Command{
		Name:    groupCmd,
		Message: t,
	})
}

// EndGroup ends the current group. It panics if it cannot write to the output
// stream.
func (c *Action) EndGroup() {
	// ::endgroup::
	c.IssueCommand(&Command{
		Name: endGroupCmd,
	})
}

// AddStepSummary writes the given markdown to the job summary. If a job summary
// already exists, this value is appended.
//
// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#adding-a-job-summary
// https://github.blog/2022-05-09-supercharging-github-actions-with-job-summaries/
func (c *Action) AddStepSummary(markdown string) {
	c.IssueFileCommand(&Command{
		Name:    stepSummaryCmd,
		Message: markdown,
	})
}

// AddStepSummaryTemplate adds a summary template by parsing the given Go
// template using html/template with the given input data. See AddStepSummary
// for caveats.
//
// This primarily exists as a convenience function that renders a template.
func (c *Action) AddStepSummaryTemplate(tmpl string, data any) error {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	c.AddStepSummary(b.String())
	return nil
}

// SetEnv sets an environment variable. It panics if it cannot write to the
// output file.
//
// https://docs.github.com/en/free-pro-team@latest/actions/reference/workflow-commands-for-github-actions#setting-an-environment-variable
// https://github.blog/changelog/2020-10-01-github-actions-deprecating-set-env-and-add-path-commands/
func (c *Action) SetEnv(k, v string) {
	c.IssueFileCommand(&Command{
		Name:    envCmd,
		Message: fmt.Sprintf(multilineFileCmd, k, v),
	})
}

// SetOutput sets an output parameter. It panics if it cannot write to the
// output stream.
//
// On 2022-10-11, GitHub deprecated "::set-output name=<k>::<v>" in favor of
// [environment files].
//
// [environment files]: https://github.blog/changelog/2022-10-11-github-actions-deprecating-save-state-and-set-output-commands/
func (c *Action) SetOutput(k, v string) {
	c.IssueFileCommand(&Command{
		Name:    outputCmd,
		Message: fmt.Sprintf(multilineFileCmd, k, v),
	})
}

// Debugf prints a debug-level message. It follows the standard fmt.Printf
// arguments, appending an OS-specific line break to the end of the message. It
// panics if it cannot write to the output stream.
func (c *Action) Debugf(msg string, args ...any) {
	// ::debug <c.fields>::<msg, args>
	c.IssueCommand(&Command{
		Name:       debugCmd,
		Message:    fmt.Sprintf(msg, args...),
		Properties: c.fields,
	})
}

// Noticef prints a notice-level message. It follows the standard fmt.Printf
// arguments, appending an OS-specific line break to the end of the message. It
// panics if it cannot write to the output stream.
func (c *Action) Noticef(msg string, args ...any) {
	// ::notice <c.fields>::<msg, args>
	c.IssueCommand(&Command{
		Name:       noticeCmd,
		Message:    fmt.Sprintf(msg, args...),
		Properties: c.fields,
	})
}

// Warningf prints a warning-level message. It follows the standard fmt.Printf
// arguments, appending an OS-specific line break to the end of the message. It
// panics if it cannot write to the output stream.
func (c *Action) Warningf(msg string, args ...any) {
	// ::warning <c.fields>::<msg, args>
	c.IssueCommand(&Command{
		Name:       warningCmd,
		Message:    fmt.Sprintf(msg, args...),
		Properties: c.fields,
	})
}

// Errorf prints a error-level message. It follows the standard fmt.Printf
// arguments, appending an OS-specific line break to the end of the message. It
// panics if it cannot write to the output stream.
func (c *Action) Errorf(msg string, args ...any) {
	// ::error <c.fields>::<msg, args>
	c.IssueCommand(&Command{
		Name:       errorCmd,
		Message:    fmt.Sprintf(msg, args...),
		Properties: c.fields,
	})
}

// Fatalf prints a error-level message and exits. This is equivalent to Errorf
// followed by os.Exit(1).
func (c *Action) Fatalf(msg string, args ...any) {
	c.Errorf(msg, args...)
	osExit(1)
}

// Infof prints message to stdout without any level annotations. It follows the
// standard fmt.Printf arguments, appending an OS-specific line break to the end
// of the message. It panics if it cannot write to the output stream.
func (c *Action) Infof(msg string, args ...any) {
	if _, err := fmt.Fprintf(c.w, msg+EOF, args...); err != nil {
		panic(fmt.Errorf("failed to write info command: %w", err))
	}
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
		w:          c.w,
		fields:     m,
		getenv:     c.getenv,
		httpClient: c.httpClient,
	}
}

// idTokenResponse is the response from minting an ID token.
type idTokenResponse struct {
	Value string `json:"value,omitempty"`
}

// GetIDToken returns the GitHub OIDC token from the GitHub Actions runtime.
func (c *Action) GetIDToken(ctx context.Context, audience string) (string, error) {
	requestURL := c.getenv("ACTIONS_ID_TOKEN_REQUEST_URL")
	if requestURL == "" {
		return "", fmt.Errorf("missing ACTIONS_ID_TOKEN_REQUEST_URL in environment")
	}

	requestToken := c.getenv("ACTIONS_ID_TOKEN_REQUEST_TOKEN")
	if requestToken == "" {
		return "", fmt.Errorf("missing ACTIONS_ID_TOKEN_REQUEST_TOKEN in environment")
	}

	u, err := url.Parse(requestURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse request URL: %w", err)
	}
	if audience != "" {
		q := u.Query()
		q.Set("audience", audience)
		u.RawQuery = q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+requestToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// This has moved to the io package in Go 1.16, but we still support up to Go
	// 1.13 for now.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1000))
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	body = bytes.TrimSpace(body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("non-successful response from minting OIDC token: %s", body)
	}

	var tokenResp idTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to process response as JSON: %w", err)
	}
	return tokenResp.Value, nil
}

// Getenv retrieves the value of the environment variable named by the key.
// It uses an internal function that can be set with `WithGetenv`.
func (c *Action) Getenv(key string) string {
	return c.getenv(key)
}

// GetenvFunc is an abstraction to make tests feasible for commands that
// interact with environment variables.
type GetenvFunc func(key string) string

// GitHubContext of current workflow.
//
// See: https://docs.github.com/en/actions/learn-github-actions/environment-variables
type GitHubContext struct {
	Action           string `env:"GITHUB_ACTION"`
	ActionPath       string `env:"GITHUB_ACTION_PATH"`
	ActionRepository string `env:"GITHUB_ACTION_REPOSITORY"`
	Actions          bool   `env:"GITHUB_ACTIONS"`
	Actor            string `env:"GITHUB_ACTOR"`
	APIURL           string `env:"GITHUB_API_URL,default=https://api.github.com"`
	BaseRef          string `env:"GITHUB_BASE_REF"`
	Env              string `env:"GITHUB_ENV"`
	EventName        string `env:"GITHUB_EVENT_NAME"`
	EventPath        string `env:"GITHUB_EVENT_PATH"`
	GraphqlURL       string `env:"GITHUB_GRAPHQL_URL,default=https://api.github.com/graphql"`
	HeadRef          string `env:"GITHUB_HEAD_REF"`
	Job              string `env:"GITHUB_JOB"`
	Path             string `env:"GITHUB_PATH"`
	Ref              string `env:"GITHUB_REF"`
	RefName          string `env:"GITHUB_REF_NAME"`
	RefProtected     bool   `env:"GITHUB_REF_PROTECTED"`
	RefType          string `env:"GITHUB_REF_TYPE"`

	// Repository is the owner and repository name. For example, octocat/Hello-World
	// It is not recommended to use this field to acquire the repository name
	// but to use the Repo method instead.
	Repository string `env:"GITHUB_REPOSITORY"`

	// RepositoryOwner is the repository owner. For example, octocat
	// It is not recommended to use this field to acquire the repository owner
	// but to use the Repo method instead.
	RepositoryOwner string `env:"GITHUB_REPOSITORY_OWNER"`

	RetentionDays int64  `env:"GITHUB_RETENTION_DAYS"`
	RunAttempt    int64  `env:"GITHUB_RUN_ATTEMPT"`
	RunID         int64  `env:"GITHUB_RUN_ID"`
	RunNumber     int64  `env:"GITHUB_RUN_NUMBER"`
	ServerURL     string `env:"GITHUB_SERVER_URL,default=https://github.com"`
	SHA           string `env:"GITHUB_SHA"`
	StepSummary   string `env:"GITHUB_STEP_SUMMARY"`
	Workflow      string `env:"GITHUB_WORKFLOW"`
	Workspace     string `env:"GITHUB_WORKSPACE"`

	// Event is populated by parsing the file at EventPath, if it exists.
	Event map[string]any
}

// Repo returns the username of the repository owner and repository name.
func (c *GitHubContext) Repo() (string, string) {
	if c == nil {
		return "", ""
	}

	// Based on https://github.com/actions/toolkit/blob/main/packages/github/src/context.ts
	if c.Repository != "" {
		parts := strings.SplitN(c.Repository, "/", 2)
		if len(parts) == 1 {
			return parts[0], ""
		}
		return parts[0], parts[1]
	}

	// If c.Repository is empty attempt to get the repo from the Event data.
	var repoName string
	// NOTE: differs from context.ts. Fall back to GITHUB_REPOSITORY_OWNER
	ownerName := c.RepositoryOwner
	if c.Event != nil {
		if repo, ok := c.Event["repository"].(map[string]any); ok {
			if name, ok := repo["name"].(string); ok {
				repoName = name
			}
			if owner, ok := repo["owner"].(map[string]any); ok {
				if name, ok := owner["name"].(string); ok {
					ownerName = name
				}
			}
		}
	}
	return ownerName, repoName
}

// Context returns the context of current action with the payload object
// that triggered the workflow
func (c *Action) Context() (*GitHubContext, error) {
	ctx := context.Background()
	lookuper := &wrappedLookuper{f: c.getenv}

	var githubContext GitHubContext
	if err := envconfig.ProcessWith(ctx, &githubContext, lookuper); err != nil {
		return nil, fmt.Errorf("could not process github context variables: %w", err)
	}

	if githubContext.EventPath != "" {
		eventData, err := os.ReadFile(githubContext.EventPath)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("could not read event file: %w", err)
		}
		if eventData != nil {
			if err := json.Unmarshal(eventData, &githubContext.Event); err != nil {
				return nil, fmt.Errorf("failed to unmarshal event payload: %w", err)
			}
		}
	}

	return &githubContext, nil
}

// wrappedLookuper creates a lookuper that wraps a given getenv func.
type wrappedLookuper struct {
	f GetenvFunc
}

// Lookup implements a custom lookuper.
func (w *wrappedLookuper) Lookup(key string) (string, bool) {
	if v := w.f(key); v != "" {
		return v, true
	}
	return "", false
}
