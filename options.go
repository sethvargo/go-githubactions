// Copyright 2021 The Authors
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
	"io"
	"net/http"
)

// Option is a modifier for an Action
type Option func(*Action) *Action

// WithWriter sets the writer function on an Action. By default, this will
// be `os.Stdout` from the standard library.
func WithWriter(w io.Writer) Option {
	return func(a *Action) *Action {
		a.w = w
		return a
	}
}

// WithFields sets the extra command field on an Action.
func WithFields(fields CommandProperties) Option {
	return func(a *Action) *Action {
		a.fields = fields
		return a
	}
}

// WithGetenv sets the `Getenv` function on an Action. By default, this will
// be `os.Getenv` from the standard library.
func WithGetenv(getenv GetenvFunc) Option {
	return func(a *Action) *Action {
		a.getenv = getenv
		return a
	}
}

// WithHTTPClient sets a custom HTTP client on the action. This is only used
// when the action makes output HTTP requests (such as generating an OIDC
// token).
func WithHTTPClient(c *http.Client) Option {
	return func(a *Action) *Action {
		a.httpClient = c
		return a
	}
}
