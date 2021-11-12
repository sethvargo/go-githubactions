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
	"fmt"
	"sort"
	"strings"
)

const (
	cmdSeparator        = "::"
	cmdPropertiesPrefix = " "
)

// CommandProperties is a named "map[string]string" type to hold key-value pairs
// passed to an actions command.
type CommandProperties map[string]string

// String encodes the CommandProperties to a string as comma separated
// 'key=value' pairs. The pairs are joined in a chronological order.
func (props *CommandProperties) String() string {
	l := make([]string, 0, len(*props))
	for k, v := range *props {
		l = append(l, fmt.Sprintf("%s=%s", k, escapeProperty(v)))
	}

	sort.Strings(l)
	return strings.Join(l, ",")
}

// Command can be issued by a GitHub action by writing to `stdout` with
// following format.
//
// ::name key=value,key=value::message
//
//  Examples:
//    ::warning::This is the message
//    ::set-env name=MY_VAR::some value
type Command struct {
	Name       string
	Message    string
	Properties CommandProperties
}

// String encodes the Command to a string in the following format:
//
// ::name key=value,key=value::message
func (cmd *Command) String() string {
	// https://github.com/actions/toolkit/blob/9ad01e4fd30025e8858650d38e95cfe9193a3222/packages/core/src/command.ts#L43-L45
	if cmd.Name == "" {
		cmd.Name = "missing.command"
	}

	var builder strings.Builder
	builder.WriteString(cmdSeparator)
	builder.WriteString(cmd.Name)
	if len(cmd.Properties) > 0 {
		builder.WriteString(cmdPropertiesPrefix)
		builder.WriteString(cmd.Properties.String())
	}

	builder.WriteString(cmdSeparator)
	builder.WriteString(escapeData(cmd.Message))
	return builder.String()
}

// escapeData escapes string values for presentation in the output of a command.
// This is a not-so-well-documented requirement of commands that define a
// message:
//
// https://github.com/actions/toolkit/blob/9ad01e4fd30025e8858650d38e95cfe9193a3222/packages/core/src/command.ts#L74
//
// The equivalent toolkit function can be found here:
//
// https://github.com/actions/toolkit/blob/9ad01e4fd30025e8858650d38e95cfe9193a3222/packages/core/src/command.ts#L92
//
func escapeData(v string) string {
	v = strings.ReplaceAll(v, "%", "%25")
	v = strings.ReplaceAll(v, "\r", "%0D")
	v = strings.ReplaceAll(v, "\n", "%0A")
	return v
}

// escapeData escapes command property values for presentation in the output of
// a command.
//
// https://github.com/actions/toolkit/blob/9ad01e4fd30025e8858650d38e95cfe9193a3222/packages/core/src/command.ts#L68
//
// The equivalent toolkit function can be found here:
//
// https://github.com/actions/toolkit/blob/1cc56db0ff126f4d65aeb83798852e02a2c180c3/packages/core/src/command.ts#L99-L106
func escapeProperty(v string) string {
	v = strings.ReplaceAll(v, "%", "%25")
	v = strings.ReplaceAll(v, "\r", "%0D")
	v = strings.ReplaceAll(v, "\n", "%0A")
	v = strings.ReplaceAll(v, ":", "%3A")
	v = strings.ReplaceAll(v, ",", "%2C")
	return v
}
