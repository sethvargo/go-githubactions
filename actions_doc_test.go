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

package githubactions_test

import (
	"context"

	"github.com/sethvargo/go-githubactions"
)

var (
	a = githubactions.New()
)

func ExampleNew() {
	a = githubactions.New()
}

func ExampleAction_AddMask() {
	a := githubactions.New()
	a.AddMask("my-password")
}

func ExampleAction_AddPath() {
	a := githubactions.New()
	a.AddPath("/tmp/myapp")
}

func ExampleAction_GetInput() {
	a := githubactions.New()
	a.GetInput("foo")
}

func ExampleAction_Group() {
	a := githubactions.New()
	a.Group("My group")
}

func ExampleAction_EndGroup() {
	a := githubactions.New()
	a.Group("My group")

	// work

	a.EndGroup()
}

func ExampleAction_AddStepSummary() {
	a := githubactions.New()
	a.AddStepSummary(`
## Heading

- :rocket:
- :moon:
`)
}

func ExampleAction_AddStepSummaryTemplate() {
	a := githubactions.New()
	if err := a.AddStepSummaryTemplate(`
## Heading

- {{.Input}}
- :moon:
`, map[string]string{
		"Input": ":rocket:",
	}); err != nil {
		// handle error
	}
}

func ExampleAction_Debugf() {
	a := githubactions.New()
	a.Debugf("a debug message")
}

func ExampleAction_Debugf_fieldsMap() {
	a := githubactions.New()
	m := map[string]string{
		"file": "app.go",
		"line": "100",
	}
	a.WithFieldsMap(m).Debugf("a debug message")
}

func ExampleAction_Debugf_fieldsSlice() {
	a := githubactions.New()
	s := []string{"file=app.go", "line=100"}
	a.WithFieldsSlice(s).Debugf("a debug message")
}

func ExampleAction_Warningf() {
	a := githubactions.New()
	a.Warningf("a warning message")
}

func ExampleAction_Warningf_fieldsMap() {
	a := githubactions.New()
	m := map[string]string{
		"file": "app.go",
		"line": "100",
	}
	a.WithFieldsMap(m).Warningf("a warning message")
}

func ExampleAction_Warningf_fieldsSlice() {
	a := githubactions.New()
	s := []string{"file=app.go", "line=100"}
	a.WithFieldsSlice(s).Warningf("a warning message")
}

func ExampleAction_Errorf() {
	a := githubactions.New()
	a.Errorf("an error message")
}

func ExampleAction_Errorf_fieldsMap() {
	a := githubactions.New()
	m := map[string]string{
		"file": "app.go",
		"line": "100",
	}
	a.WithFieldsMap(m).Errorf("an error message")
}

func ExampleAction_Errorf_fieldsSlice() {
	a := githubactions.New()
	s := []string{"file=app.go", "line=100"}
	a.WithFieldsSlice(s).Errorf("an error message")
}

func ExampleAction_SetEnv() {
	a := githubactions.New()
	a.SetEnv("MY_THING", "my value")
}

func ExampleAction_SetOutput() {
	a := githubactions.New()
	a.SetOutput("filepath", "/tmp/file-xyz1234")
}

func ExampleAction_GetIDToken() {
	ctx := context.Background()

	a := githubactions.New()
	token, err := a.GetIDToken(ctx, "my-aud")
	if err != nil {
		// handle error
	}
	_ = token
}
