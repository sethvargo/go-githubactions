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

package githubactions

import (
	"bytes"
	"testing"
)

func TestWithWriter(t *testing.T) {
	t.Parallel()

	a := &Action{}
	var b bytes.Buffer
	opt := WithWriter(&b)

	opt(a)
	a.IssueCommand(&Command{
		Name:    "foo",
		Message: "bar",
	})

	if got, want := b.String(), "::foo::bar"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestWithFields(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := &Action{w: &b}
	opt := WithFields(map[string]string{"baz": "quux"})

	opt(a)
	a.IssueCommand(&Command{
		Name:       "foo",
		Message:    "bar",
		Properties: a.fields,
	})

	if got, want := b.String(), "::foo baz=quux::bar"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestWithGetenv(t *testing.T) {
	t.Parallel()

	a := &Action{}
	opt := WithGetenv(func(k string) string {
		return "sentinel"
	})

	opt(a)
	if got, want := a.Getenv("any"), "sentinel"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}
