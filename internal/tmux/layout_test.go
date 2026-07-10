package tmux

import (
	"reflect"
	"testing"
)

func TestSessionName(t *testing.T) {
	tests := []struct {
		repo, branch, want string
	}{
		{"wtx", "feature/x", "wtx/feature/x"},   // slashes kept
		{"my.repo", "v1.2", "my%2Erepo/v1%2E2"}, // dots encoded
		{"repo", "fix:bug", "repo/fix%3Abug"},   // colons encoded
		{"a.b", "c/d.e:f", "a%2Eb/c/d%2Ee%3Af"}, // mixed
	}
	for _, tc := range tests {
		if got := SessionName(tc.repo, tc.branch); got != tc.want {
			t.Errorf("SessionName(%q,%q)=%q want %q", tc.repo, tc.branch, got, tc.want)
		}
	}
}

func TestSessionNameInjective(t *testing.T) {
	// The pre-fix flattening collapsed these onto one name; encoding must not.
	if SessionName("r", "a.b") == SessionName("r", "a-b") {
		t.Error("SessionName must not collide a.b with a-b")
	}
}

func TestBuildConstructionSequenceSinglePane(t *testing.T) {
	seq := BuildConstructionSequence("s", "/dir", "even-horizontal", 1)
	want := [][]string{
		{"new-session", "-d", "-P", "-F", "#{pane_id}", "-s", "s", "-c", "/dir"},
	}
	if !reflect.DeepEqual(seq, want) {
		t.Fatalf("single pane sequence = %v, want %v (no split, no select-layout)", seq, want)
	}
}

func TestBuildConstructionSequenceMultiPane(t *testing.T) {
	seq := BuildConstructionSequence("s", "/dir", "tiled", 3)
	if len(seq) != 4 { // 1 new-session + 2 splits + 1 select-layout
		t.Fatalf("expected 4 construction commands, got %d: %v", len(seq), seq)
	}
	if seq[0][0] != "new-session" {
		t.Errorf("first command = %q, want new-session", seq[0][0])
	}
	if seq[1][0] != "split-window" || seq[2][0] != "split-window" {
		t.Errorf("commands 1,2 = %q,%q, want two split-window", seq[1][0], seq[2][0])
	}
	last := seq[3]
	if !reflect.DeepEqual(last, []string{"select-layout", "-t", "s", "tiled"}) {
		t.Errorf("last command = %v, want select-layout ... tiled", last)
	}
}

func TestBuildConstructionSequenceEmptyArrangeSkipsSelectLayout(t *testing.T) {
	seq := BuildConstructionSequence("s", "/dir", "", 2)
	for _, cmd := range seq {
		if cmd[0] == "select-layout" {
			t.Fatalf("empty arrange should not emit select-layout: %v", seq)
		}
	}
	if len(seq) != 2 {
		t.Fatalf("expected 2 commands (new-session + split), got %d", len(seq))
	}
}

func TestBuildPaneCommandSequence(t *testing.T) {
	ids := []string{"%1", "%2", "%3"}
	panes := []string{"claude", "", "codex"}
	seq := BuildPaneCommandSequence(ids, panes)

	// %1 send+enter, %3 send+enter (%2 blank skipped), then select-pane %1.
	want := [][]string{
		{"send-keys", "-t", "%1", "-l", "--", "claude"},
		{"send-keys", "-t", "%1", "Enter"},
		{"send-keys", "-t", "%3", "-l", "--", "codex"},
		{"send-keys", "-t", "%3", "Enter"},
		{"select-pane", "-t", "%1"},
	}
	if !reflect.DeepEqual(seq, want) {
		t.Fatalf("pane command sequence =\n%v\nwant\n%v", seq, want)
	}
}

func TestBuildPaneCommandSequenceAllBlank(t *testing.T) {
	seq := BuildPaneCommandSequence([]string{"%1", "%2"}, []string{"", ""})
	// No send-keys, but focus is still set on the first pane.
	want := [][]string{{"select-pane", "-t", "%1"}}
	if !reflect.DeepEqual(seq, want) {
		t.Fatalf("all-blank sequence = %v, want just select-pane", seq)
	}
}
