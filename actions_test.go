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
	"io/ioutil"
	"os"
	"testing"
)

func TestAction_IssueCommand(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := NewWithWriter(&b)
	a.IssueCommand(&Command{
		Name:    "foo",
		Message: "bar",
	})

	if got, want := b.String(), "::foo::bar\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_IssueFileCommand(t *testing.T) {
	t.Parallel()

	file, err := ioutil.TempFile(".", ".issue_file_cmd_test_")
	if err != nil {
		t.Fatalf("unable to create a temp env file: %s", err)
	}

	defer os.Remove(file.Name())
	if err = os.Setenv("GITHUB_FOO", file.Name()); err != nil {
		t.Fatalf("unable to set 'GITHUB_FOO' env var: %s", err)
	}

	var b bytes.Buffer
	a := NewWithWriter(&b)

	err = a.IssueFileCommand(&Command{
		Name:    "foo",
		Message: "bar",
	})

	if err != nil {
		t.Errorf("expected nil error, got: %s", err)
	}

	// expect an empty stdout buffer
	if got, want := b.String(), ""; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	// expect the message to be written to the env file
	data, err := ioutil.ReadFile(file.Name())
	if err != nil {
		t.Errorf("unable to read temp env file: %s", err)
	}

	if got, want := string(data), "bar\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

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

	const envGitHubPath = "GITHUB_PATH"
	defer os.Setenv(envGitHubPath, os.Getenv(envGitHubPath))
	os.Unsetenv(envGitHubPath)

	// expect a regular command to be issued when env file is not set.
	var b bytes.Buffer
	a := NewWithWriter(&b)

	a.AddPath("/custom/bin")
	if got, want := b.String(), "::add-path::/custom/bin\n"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	// expect a file command to be issued when env file is set.
	file, err := ioutil.TempFile(".", ".add_path_test_")
	if err != nil {
		t.Fatalf("unable to create a temp env file: %s", err)
	}

	defer os.Remove(file.Name())
	if err = os.Setenv(envGitHubPath, file.Name()); err != nil {
		t.Fatalf("unable to set %q env var: %s", envGitHubPath, err)
	}

	b.Reset()
	a.AddPath("/custom/bin")

	if got, want := b.String(), ""; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	// expect an empty stdout buffer
	if got, want := b.String(), ""; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	// expect the message to be written to the file.
	data, err := ioutil.ReadAll(file)
	if err != nil {
		t.Errorf("unable to read temp env file: %s", err)
	}

	if got, want := string(data), "/custom/bin\n"; got != want {
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

	const envGitHubEnv = "GITHUB_ENV"
	defer os.Setenv(envGitHubEnv, os.Getenv(envGitHubEnv))
	os.Unsetenv(envGitHubEnv)

	// expectations for regular set-env commands
	checks := []struct {
		key, value, want string
	}{
		{"key", "value", "::set-env name=key::value\n"},
		{"key", "this is 100% a special\n\r value!", "::set-env name=key::this is 100%25 a special%0A%0D value!\n"},
	}

	for _, check := range checks {
		var b bytes.Buffer
		a := NewWithWriter(&b)
		a.SetEnv(check.key, check.value)
		if got, want := b.String(), check.want; got != want {
			t.Errorf("SetEnv(%q, %q): expected %q; got %q", check.key, check.value, want, got)
		}
	}

	// expectations for env file env commands
	var b bytes.Buffer
	a := NewWithWriter(&b)
	file, err := ioutil.TempFile(".", ".set_env_test_")
	if err != nil {
		t.Fatalf("unable to create a temp env file: %s", err)
	}

	defer os.Remove(file.Name())
	if err = os.Setenv(envGitHubEnv, file.Name()); err != nil {
		t.Fatalf("unable to set %q env var: %s", envGitHubEnv, err)
	}

	a.SetEnv("key", "value")

	// expect an empty stdout buffer
	if got, want := b.String(), ""; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	// expect the command to be written to the file.
	data, err := ioutil.ReadAll(file)
	if err != nil {
		t.Errorf("unable to read temp env file: %s", err)
	}

	want := "key<<_GitHubActionsFileCommandDelimeter_\nvalue\n_GitHubActionsFileCommandDelimeter_\n"
	if got := string(data); got != want {
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
