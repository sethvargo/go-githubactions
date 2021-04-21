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
	"bytes"
	"testing"
)

func TestOptWriter(t *testing.T) {
	t.Parallel()

	a := &Action{}
	var b bytes.Buffer
	opt := OptWriter(&b)

	opt(a)
	a.IssueCommand(&Command{
		Name:    "foo",
		Message: "bar",
	})

	if got, want := b.String(), "::foo::bar\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestOptFields(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := &Action{w: &b}
	opt := OptFields(map[string]string{"baz": "quux"})

	opt(a)
	a.IssueCommand(&Command{
		Name:       "foo",
		Message:    "bar",
		Properties: a.fields,
	})

	if got, want := b.String(), "::foo baz=quux::bar\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestOptGetenv(t *testing.T) {
	t.Parallel()

	a := &Action{}
	opt := OptGetenv(func(k string) string {
		return "sentinel"
	})

	opt(a)
	if got, want := a.getenv("any"), "sentinel"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}
