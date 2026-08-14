package cmd

import (
	"bytes"
	"runtime/debug"
	"strings"
	"testing"
)

// buildSettings converts a key/value map into the slice shape the toolchain
// embeds, so table cases can be written as plain maps.
func buildSettings(kv map[string]string) []debug.BuildSetting {
	if kv == nil {
		return nil
	}

	settings := make([]debug.BuildSetting, 0, len(kv))
	for k, v := range kv {
		settings = append(settings, debug.BuildSetting{Key: k, Value: v})
	}

	return settings
}

// result is what one in-process execution of the root command produced.
type result struct {
	stdout string
	stderr string
	err    error
}

// run executes a fresh root command with the given arguments, capturing both
// output streams. Running in-process rather than exec'ing a built binary keeps
// the suite fast and platform-independent; main.go is a thin enough shim that
// the error it turns into an exit status is the error returned here.
func run(t *testing.T, args ...string) result {
	t.Helper()

	var stdout, stderr bytes.Buffer

	root := newRootCmd()
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	// SetArgs(nil) would make cobra fall back to os.Args, which under `go
	// test` are the test binary's own flags. An empty slice means "no
	// arguments", which is what a bare `wtx` invocation looks like.
	root.SetArgs(args)

	err := root.Execute()

	return result{stdout: stdout.String(), stderr: stderr.String(), err: err}
}

func TestVersionReturnsStampedValue(t *testing.T) {
	const stamped = "v1.2.3"

	t.Cleanup(func() { version = "" })
	version = stamped

	if got := Version(); got != stamped {
		t.Errorf("Version() = %q, want %q", got, stamped)
	}
}

// The unstamped path depends on how the test binary itself was built, so this
// asserts the guarantee the spec actually makes: never empty, never a panic.
func TestVersionIsNeverEmptyWhenUnstamped(t *testing.T) {
	t.Cleanup(func() { version = "" })
	version = ""

	got := Version()
	if got == "" {
		t.Fatal("Version() returned an empty string")
	}
	if strings.TrimSpace(got) != got {
		t.Errorf("Version() = %q, want no surrounding whitespace", got)
	}
}

func TestVersionFromVCS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		settings map[string]string
		want     string
	}{
		{
			name:     "clean checkout shortens the revision",
			settings: map[string]string{"vcs.revision": "0123456789abcdef0123456789abcdef01234567", "vcs.modified": "false"},
			want:     "0123456789ab",
		},
		{
			name:     "dirty checkout is marked",
			settings: map[string]string{"vcs.revision": "0123456789abcdef0123456789abcdef01234567", "vcs.modified": "true"},
			want:     "0123456789ab+dirty",
		},
		{
			name:     "short revision is left alone",
			settings: map[string]string{"vcs.revision": "abc123"},
			want:     "abc123",
		},
		{
			name:     "no revision yields nothing",
			settings: map[string]string{"vcs.modified": "true"},
			want:     "",
		},
		{
			name:     "no settings at all yields nothing",
			settings: nil,
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := versionFromVCS(buildSettings(tt.settings)); got != tt.want {
				t.Errorf("versionFromVCS() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVersionCommand(t *testing.T) {
	got := run(t, "version")

	if got.err != nil {
		t.Fatalf("wtx version returned error: %v", got.err)
	}
	if got.stderr != "" {
		t.Errorf("wtx version wrote %q to stderr, want nothing", got.stderr)
	}

	line, ok := strings.CutSuffix(got.stdout, "\n")
	if !ok {
		t.Fatalf("wtx version output %q, want a single newline-terminated line", got.stdout)
	}
	if line == "" {
		t.Error("wtx version printed an empty version")
	}
	if strings.Contains(line, "\n") {
		t.Errorf("wtx version output %q, want exactly one line", got.stdout)
	}
	if strings.TrimSpace(line) != line {
		t.Errorf("wtx version printed %q, want no surrounding whitespace", line)
	}
}

func TestVersionCommandRejectsArguments(t *testing.T) {
	if got := run(t, "version", "extra"); got.err == nil {
		t.Error("wtx version extra returned no error, want one")
	}
}

func TestVersionFlagMatchesVersionCommand(t *testing.T) {
	sub := run(t, "version")
	flag := run(t, "--version")

	if flag.err != nil {
		t.Fatalf("wtx --version returned error: %v", flag.err)
	}
	if flag.stdout != sub.stdout {
		t.Errorf("wtx --version printed %q, want %q (identical to wtx version)", flag.stdout, sub.stdout)
	}
}

func TestVersionOutputIsStable(t *testing.T) {
	first := run(t, "version")
	second := run(t, "version")

	if first.stdout != second.stdout {
		t.Errorf("repeated wtx version differed: %q then %q", first.stdout, second.stdout)
	}
}

func TestBareInvocationPrintsHelpAndSucceeds(t *testing.T) {
	got := run(t)

	if got.err != nil {
		t.Fatalf("bare wtx returned error: %v", got.err)
	}
	if !strings.Contains(got.stdout, "Usage:") {
		t.Errorf("bare wtx printed %q, want help text", got.stdout)
	}
}

func TestRootHelpListsCommands(t *testing.T) {
	for _, args := range [][]string{{"help"}, {"--help"}, {"-h"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			got := run(t, args...)

			if got.err != nil {
				t.Fatalf("wtx %s returned error: %v", strings.Join(args, " "), got.err)
			}
			for _, want := range []string{"wtx", "help", "version"} {
				if !strings.Contains(got.stdout, want) {
					t.Errorf("help output does not mention %q:\n%s", want, got.stdout)
				}
			}
		})
	}
}

func TestHelpForSubcommand(t *testing.T) {
	for _, args := range [][]string{{"help", "version"}, {"version", "--help"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			got := run(t, args...)

			if got.err != nil {
				t.Fatalf("wtx %s returned error: %v", strings.Join(args, " "), got.err)
			}
			if !strings.Contains(got.stdout, "wtx version") {
				t.Errorf("output does not show usage for the version command:\n%s", got.stdout)
			}
		})
	}
}

func TestUsageErrors(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantText string
	}{
		{"unknown command", []string{"no-such-command"}, "no-such-command"},
		{"unknown help topic", []string{"help", "no-such-command"}, "no-such-command"},
		{"unknown flag", []string{"--no-such-flag"}, "no-such-flag"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := run(t, tt.args...)

			if got.err == nil {
				t.Fatalf("wtx %s returned no error, want one", strings.Join(tt.args, " "))
			}
			if !strings.Contains(got.err.Error(), tt.wantText) {
				t.Errorf("error %q does not name %q", got.err, tt.wantText)
			}
			// Diagnostics belong on stderr, which main writes the returned
			// error to. Stdout stays clean so it remains machine-consumable.
			if strings.Contains(got.stdout, "Error:") {
				t.Errorf("stdout carries error text:\n%s", got.stdout)
			}
		})
	}
}
