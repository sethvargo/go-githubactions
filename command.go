package githubactions

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// CommandProperties is a named "map[string]interface{}" type to hold key-value
// pairs passed to an actions command.
type CommandProperties map[string]interface{}

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
	Message    interface{}
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

	const cmdSeparator = "::"
	var builder strings.Builder
	builder.WriteString(cmdSeparator)
	builder.WriteString(cmd.Name)
	if len(cmd.Properties) > 0 {
		builder.WriteString(" ")
		builder.WriteString(cmd.Properties.String())
	}

	builder.WriteString(cmdSeparator)
	builder.WriteString(escapeData(cmd.Message))
	return builder.String()
}

// toCommandValue sanitizes an input into a string so it can be passed with a
// Command safely.
//
// The equivalent toolkit function can be found here:
//
// https://github.com/actions/toolkit/blob/9ad01e4fd30025e8858650d38e95cfe9193a3222/packages/core/src/command.ts#L83-L90
func toCommandValue(i interface{}) string {
	switch v := i.(type) {
	case nil:
		return ""
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	default:
		data, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}

		return string(data)
	}
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
func escapeData(i interface{}) string {
	v := toCommandValue(i)
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
func escapeProperty(i interface{}) string {
	v := toCommandValue(i)
	v = strings.ReplaceAll(v, "%", "%25")
	v = strings.ReplaceAll(v, "\r", "%0D")
	v = strings.ReplaceAll(v, "\n", "%0A")
	v = strings.ReplaceAll(v, ":", "%3A")
	v = strings.ReplaceAll(v, ",", "%2C")
	return v
}
