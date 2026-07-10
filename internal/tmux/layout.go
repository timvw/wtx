package tmux

import "strings"

// tmux treats "." and ":" as window/pane and session separators in targets, so
// a name containing them cannot be addressed. We percent-encode them (and "%"
// itself) rather than flatten, keeping the mapping injective so distinct
// branches like "a.b" and "a-b" never collide onto one session.
var componentEscaper = strings.NewReplacer("%", "%25", ".", "%2E", ":", "%3A")

// SessionName derives a stable, collision-free tmux session name from a repo
// and branch, e.g. "wtx/feature/x". Slashes are kept (tmux allows them, and
// repo names contain none, so the first slash always separates the two parts).
func SessionName(repo, branch string) string {
	return componentEscaper.Replace(repo) + "/" + componentEscaper.Replace(branch)
}

// BuildConstructionSequence returns the ordered, index-free tmux invocations
// that create and arrange numPanes panes for a session rooted at dir. The
// new-session and split-window commands print each new pane's stable ID via
// -P -F '#{pane_id}'; select-layout (emitted only for >1 pane) prints nothing.
// It performs no I/O.
func BuildConstructionSequence(session, dir, arrange string, numPanes int) [][]string {
	if numPanes < 1 {
		numPanes = 1
	}
	seq := [][]string{
		{"new-session", "-d", "-P", "-F", "#{pane_id}", "-s", session, "-c", dir},
	}
	for i := 1; i < numPanes; i++ {
		seq = append(seq,
			[]string{"split-window", "-P", "-F", "#{pane_id}", "-t", session, "-c", dir})
	}
	if numPanes > 1 && arrange != "" {
		seq = append(seq, []string{"select-layout", "-t", session, arrange})
	}
	return seq
}

// BuildPaneCommandSequence returns the tmux invocations that run each pane's
// command and set focus, given the captured pane IDs. paneIDs is in pane
// creation order with paneIDs[i] the ID for panes[i]. An empty command leaves
// that pane a plain login shell. It performs no I/O.
func BuildPaneCommandSequence(paneIDs, panes []string) [][]string {
	seq := make([][]string, 0, len(panes)*2+1)
	for i, cmd := range panes {
		if cmd == "" {
			continue
		}
		seq = append(seq,
			[]string{"send-keys", "-t", paneIDs[i], "-l", "--", cmd},
			[]string{"send-keys", "-t", paneIDs[i], "Enter"},
		)
	}
	if len(paneIDs) > 0 {
		seq = append(seq, []string{"select-pane", "-t", paneIDs[0]})
	}
	return seq
}
