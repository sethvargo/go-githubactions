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
}
