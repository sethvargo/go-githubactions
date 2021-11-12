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

import "testing"

func TestCommandProperties_String(t *testing.T) {
	t.Parallel()

	props := CommandProperties{"hello": "world"}
	if got, want := props.String(), "hello=world"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	props["foo"] = "bar"
	if got, want := props.String(), "foo=bar,hello=world"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestCommand_String(t *testing.T) {
	t.Parallel()

	cmd := Command{Name: "foo"}
	if got, want := cmd.String(), "::foo::"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	cmd = Command{Name: "foo", Message: "bar"}
	if got, want := cmd.String(), "::foo::bar"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	cmd = Command{
		Name:       "foo",
		Message:    "bar",
		Properties: CommandProperties{"bar": "foo"},
	}
	if got, want := cmd.String(), "::foo bar=foo::bar"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	cmd = Command{Message: "quux"}
	if got, want := cmd.String(), "::missing.command::quux"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}
