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
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(
		WithWriter(&b),
		nil,
		WithGetenv(func(key string) string {
			return key
		}),
	)

	a.IssueCommand(&Command{
		Name:    "foo",
		Message: "bar",
	})

	if got, want := b.String(), "::foo::bar"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	if got, want := a.GetInput("quux"), "INPUT_QUUX"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_IssueCommand(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a.IssueCommand(&Command{
		Name:    "foo",
		Message: "bar",
	})

	if got, want := b.String(), "::foo::bar"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_IssueFileCommand(t *testing.T) {
	t.Parallel()

	file, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("unable to create a temp env file: %s", err)
	}

	defer os.Remove(file.Name())

	fakeGetenvFunc := newFakeGetenvFunc(t, "GITHUB_FOO", file.Name())
	var b bytes.Buffer
	a := New(WithWriter(&b), WithGetenv(fakeGetenvFunc))

	if err := a.issueFileCommand(&Command{
		Name:    "foo",
		Message: "bar",
	}); err != nil {
		t.Errorf("expected nil error, got: %s", err)
	}

	// expect an empty stdout buffer
	if got, want := b.String(), ""; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	// expect the message to be written to the env file
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Errorf("unable to read temp env file: %s", err)
	}

	if got, want := string(data), "bar"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_AddMask(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a.AddMask("foobar")

	if got, want := b.String(), "::add-mask::foobar"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_AddMatcher(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a.AddMatcher("foobar.json")

	if got, want := b.String(), "::add-matcher::foobar.json"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_RemoveMatcher(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a.RemoveMatcher("foobar")

	if got, want := b.String(), "::remove-matcher owner=foobar::"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_AddPath(t *testing.T) {
	t.Parallel()

	// expect a file command to be issued when env file is set.
	file, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("unable to create a temp env file: %s", err)
	}
	defer os.Remove(file.Name())

	fakeGetenvFunc := newFakeGetenvFunc(t, "GITHUB_PATH", file.Name())
	var b bytes.Buffer
	a := New(WithWriter(&b), WithGetenv(fakeGetenvFunc))

	a.AddPath("/custom/bin")

	if got, want := b.String(), ""; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	// expect an empty stdout buffer
	if got, want := b.String(), ""; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	// expect the message to be written to the file.
	data, err := io.ReadAll(file)
	if err != nil {
		t.Errorf("unable to read temp env file: %s", err)
	}

	if got, want := string(data), "/custom/bin"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_SaveState(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a.SaveState("key", "value")

	if got, want := b.String(), "::save-state name=key::value"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_GetInput(t *testing.T) {
	t.Parallel()

	fakeGetenvFunc := newFakeGetenvFunc(t, "INPUT_FOO", "bar")

	var b bytes.Buffer
	a := New(WithWriter(&b), WithGetenv(fakeGetenvFunc))
	if got, want := a.GetInput("foo"), "bar"; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_Group(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a.Group("mygroup")

	if got, want := b.String(), "::group::mygroup"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_EndGroup(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a.EndGroup()

	if got, want := b.String(), "::endgroup::"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_AddStepSummary(t *testing.T) {
	t.Parallel()

	// expectations for env file env commands
	var b bytes.Buffer
	file, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("unable to create a temp env file: %s", err)
	}

	defer os.Remove(file.Name())
	fakeGetenvFunc := newFakeGetenvFunc(t, "GITHUB_STEP_SUMMARY", file.Name())
	a := New(WithWriter(&b), WithGetenv(fakeGetenvFunc))
	a.AddStepSummary(`
## This is

some markdown
`)
	a.AddStepSummary(`
- content
`)

	// expect an empty stdout buffer
	if got, want := b.String(), ""; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	// expect the command to be written to the file.
	data, err := io.ReadAll(file)
	if err != nil {
		t.Errorf("unable to read temp summary file: %s", err)
	}

	want := "\n## This is\n\nsome markdown\n" + EOF + "\n- content\n" + EOF
	if got := string(data); got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_AddStepSummaryTemplate(t *testing.T) {
	t.Parallel()

	// expectations for env file env commands
	var b bytes.Buffer
	file, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("unable to create a temp env file: %s", err)
	}

	defer os.Remove(file.Name())
	fakeGetenvFunc := newFakeGetenvFunc(t, "GITHUB_STEP_SUMMARY", file.Name())
	a := New(WithWriter(&b), WithGetenv(fakeGetenvFunc))
	a.AddStepSummaryTemplate(`
## This is

{{.Input}}
- content
`, map[string]string{
		"Input": "some markdown",
	})

	// expect an empty stdout buffer
	if got, want := b.String(), ""; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	// expect the command to be written to the file.
	data, err := io.ReadAll(file)
	if err != nil {
		t.Errorf("unable to read temp summary file: %s", err)
	}

	want := "\n## This is\n\nsome markdown\n- content\n" + EOF
	if got := string(data); got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_SetEnv(t *testing.T) {
	t.Parallel()

	// expectations for env file env commands
	var b bytes.Buffer
	file, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatalf("unable to create a temp env file: %s", err)
	}

	defer os.Remove(file.Name())
	fakeGetenvFunc := newFakeGetenvFunc(t, "GITHUB_ENV", file.Name())
	a := New(WithWriter(&b), WithGetenv(fakeGetenvFunc))
	a.SetEnv("key", "value")
	a.SetEnv("key2", "value2")

	// expect an empty stdout buffer
	if got, want := b.String(), ""; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	// expect the command to be written to the file.
	data, err := io.ReadAll(file)
	if err != nil {
		t.Errorf("unable to read temp env file: %s", err)
	}

	want := "key<<_GitHubActionsFileCommandDelimeter_" + EOF + "value" + EOF + "_GitHubActionsFileCommandDelimeter_" + EOF
	want += "key2<<_GitHubActionsFileCommandDelimeter_" + EOF + "value2" + EOF + "_GitHubActionsFileCommandDelimeter_" + EOF
	if got := string(data); got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_SetOutput(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a.SetOutput("key", "value")

	if got, want := b.String(), "::set-output name=key::value"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_Debugf(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a.Debugf("fail: %s", "thing")

	if got, want := b.String(), "::debug::fail: thing"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_Noticef(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a.Noticef("fail: %s", "thing")

	if got, want := b.String(), "::notice::fail: thing"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_Warningf(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a.Warningf("fail: %s", "thing")

	if got, want := b.String(), "::warning::fail: thing"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_Errorf(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a.Errorf("fail: %s", "thing")

	if got, want := b.String(), "::error::fail: thing"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_Fatalf(t *testing.T) {
	// NOTE: This test case cannot be `t.Parallel()` because it patches a
	//       global `osExit`, so could impact other (concurrent) test runs.
	calls := []int{}
	finalizer := osExitMock(&calls)
	defer finalizer()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a.Fatalf("fail: %s", "bring")

	if got, want := b.String(), "::error::fail: bring"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}

	if got, want := calls, []int{1}; !reflect.DeepEqual(got, want) {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_Infof(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a.Infof("info: %s", "thing")

	if got, want := b.String(), "info: thing"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_WithFieldsSlice(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a = a.WithFieldsSlice([]string{"line=100", "file=app.js"})
	a.Debugf("fail: %s", "thing")

	if got, want := b.String(), "::debug file=app.js,line=100::fail: thing"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_WithFieldsSlice_Panic(t *testing.T) {
	t.Parallel()

	defer func() {
		want := `"no-equals" is not a proper k=v pair!`
		if got := recover(); got != want {
			t.Errorf("expected %q to be %q", got, want)
		}
	}()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a = a.WithFieldsSlice([]string{"no-equals"})
	a.Debugf("fail: %s", "thing")
}

func TestAction_WithFieldsMap(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	a := New(WithWriter(&b))
	a = a.WithFieldsMap(map[string]string{"line": "100", "file": "app.js"})
	a.Debugf("fail: %s", "thing")

	if got, want := b.String(), "::debug file=app.js,line=100::fail: thing"+EOF; got != want {
		t.Errorf("expected %q to be %q", got, want)
	}
}

func TestAction_GetIDToken(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if len(token) < 7 {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		token = token[7:]

		if token != "my-valid-token" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		aud := r.URL.Query().Get("audience")
		if aud != "" {
			fmt.Fprintf(w, `{"value":"token.%s"}`, aud)
			return
		}

		fmt.Fprint(w, `{"value":"token"}`)
	}))

	cases := []struct {
		name     string
		url      string
		token    string
		audience string
		expResp  string
		expErr   string
	}{
		{
			name:   "missing_url",
			url:    "",
			token:  "my-valid-token",
			expErr: "missing ACTIONS_ID_TOKEN_REQUEST_URL",
		},
		{
			name:   "missing_token",
			url:    "http://example.com",
			token:  "",
			expErr: "missing ACTIONS_ID_TOKEN_REQUEST_TOKEN",
		},
		{
			name:   "invalid_url",
			url:    "not-valid",
			token:  "my-valid-token",
			expErr: "failed to make HTTP request",
		},
		{
			name:   "invalid_token",
			url:    srv.URL,
			token:  "abcd1234",
			expErr: "non-successful response from minting OIDC token",
		},
		{
			name:     "no_audience",
			url:      srv.URL,
			token:    "my-valid-token",
			audience: "",
			expResp:  "token",
		},
		{
			name:     "audience",
			url:      srv.URL,
			token:    "my-valid-token",
			audience: "my-aud",
			expResp:  "token.my-aud",
		},
		{
			name:     "audience_special_chars",
			url:      srv.URL,
			token:    "my-valid-token",
			audience: "th!$my%a_d",
			expResp:  "token.th!$my%a_d",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			getEnvFunc := func(k string) string {
				switch k {
				case "ACTIONS_ID_TOKEN_REQUEST_URL":
					return tc.url
				case "ACTIONS_ID_TOKEN_REQUEST_TOKEN":
					return tc.token
				default:
					return ""
				}
			}

			a := New(WithGetenv(getEnvFunc))
			result, err := a.GetIDToken(ctx, tc.audience)
			if err != nil {
				if tc.expErr == "" {
					t.Fatal(err)
				}

				if got, want := err.Error(), tc.expErr; !strings.Contains(got, want) {
					t.Errorf("expected %q to be contain %q", got, want)
				}
			} else if tc.expErr != "" {
				t.Errorf("expected error %q, got nothing", tc.expErr)
			}

			if got, want := result, tc.expResp; got != want {
				t.Errorf("expected %q to be %q", got, want)
			}
		})
	}
}

// newFakeGetenvFunc returns a new GetenvFunc that is expected to be called with
// the provided key. It returns the provided value if the call matches the
// provided key. It reports an error on test t otherwise.
func newFakeGetenvFunc(t *testing.T, wantKey, v string) GetenvFunc {
	return func(gotKey string) string {
		if gotKey != wantKey {
			t.Errorf("expected call GetenvFunc(%q) to be GetenvFunc(%q)", gotKey, wantKey)
		}

		return v
	}
}

func osExitMock(calls *[]int) func() {
	osExit = func(code int) {
		*calls = append(*calls, code)
	}

	finalizer := func() {
		osExit = os.Exit
	}
	return finalizer
}
