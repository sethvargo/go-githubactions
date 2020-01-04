// Copyright 2019 The Authors
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

func TestAction_AddMask(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a.AddMask("foobar")

	if got, want := b.String(), "::add-mask::foobar\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_AddMatcher(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a.AddMatcher("foobar.json")

	if got, want := b.String(), "::add-matcher::foobar.json\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_RemoveMatcher(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a.RemoveMatcher("foobar")

	if got, want := b.String(), "::remove-matcher owner=foobar::\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_AddPath(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a.AddPath("/custom/bin")

	if got, want := b.String(), "::add-path::/custom/bin\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_SaveState(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a.SaveState("key", "value")

	if got, want := b.String(), "::save-state name=key::value\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_Group(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a.Group("mygroup")

	if got, want := b.String(), "::group::mygroup\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_EndGroup(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a.EndGroup()

	if got, want := b.String(), "::endgroup::\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_SetEnv(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a.SetEnv("key", "value")

	if got, want := b.String(), "::set-env name=key::value\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_SetOutput(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a.SetOutput("key", "value")

	if got, want := b.String(), "::set-output name=key::value\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_Debugf(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a.Debugf("fail: %s", "thing")

	if got, want := b.String(), "::debug::fail: thing\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_Errorf(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a.Errorf("fail: %s", "thing")

	if got, want := b.String(), "::error::fail: thing\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_Warningf(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a.Warningf("fail: %s", "thing")

	if got, want := b.String(), "::warning::fail: thing\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_WithFieldsSlice(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a = a.WithFieldsSlice([]string{"line=100", "file=app.js"})
	a.Debugf("fail: %s", "thing")

	if got, want := b.String(), "::debug file=app.js,line=100::fail: thing\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_WithFieldsMap(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a = a.WithFieldsMap(map[string]string{"line": "100", "file": "app.js"})
	a.Debugf("fail: %s", "thing")

	if got, want := b.String(), "::debug file=app.js,line=100::fail: thing\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}
