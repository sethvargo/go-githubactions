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
)

// Option is a modifier for an Action
type Option func(*Action)

func OptWriter(w io.Writer) Option {
	return func(a *Action) {
		a.w = w
	}
}

func OptFields(fields CommandProperties) Option {
	return func(a *Action) {
		a.fields = fields
	}
}
