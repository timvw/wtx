// Package config loads wtx layout presets and agent definitions. The model is
// deliberately close to kwt's: an [agents] name->command table plus layout
// presets whose panes are literal commands, "" (blank shell), or agent:<name>.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
)

// BlankLayoutName is reserved: it always means a single plain-shell pane and
// can never name a preset.
const BlankLayoutName = "none"

// ValidArranges is the set of tmux select-layout presets wtx accepts.
var ValidArranges = map[string]bool{
	"even-horizontal": true,
	"even-vertical":   true,
	"tiled":           true,
	"main-vertical":   true,
	"main-horizontal": true,
}

// Layout is a named tmux arrangement. Each pane is a literal shell command, an
// empty string (plain login shell), or an agent:<name> reference.
type Layout struct {
	Name    string   `toml:"name"`
	Arrange string   `toml:"arrange"`
	Panes   []string `toml:"panes"`
}

// DefaultAgentRef is the reserved agent name that indirects through the
// configured default_agent, so a layout can say agent:default and stay
// independent of which assistant is chosen.
const DefaultAgentRef = "default"

// Config is the on-disk wtx configuration.
type Config struct {
	DefaultAgent  string            `toml:"default_agent"`
	DefaultLayout string            `toml:"default_layout"`
	Agents        map[string]string `toml:"agents"`
	Layouts       []Layout          `toml:"layouts"`
}

// OverrideDefaultAgent replaces the default agent when name is non-empty,
// validating that it names a configured agent. An empty name keeps whatever the
// config specified. Used to apply the --agent flag and $WTX_AGENT at runtime.
func (c *Config) OverrideDefaultAgent(name string) error {
	if name == "" {
		return nil
	}
	if cmd, ok := c.Agents[name]; !ok || strings.TrimSpace(cmd) == "" {
		return fmt.Errorf("agent %q is not defined in [agents]", name)
	}
	c.DefaultAgent = name
	return nil
}

// agentCommand resolves an agent reference to its command. The reserved name
// "default" indirects through DefaultAgent. It reports false when the reference
// (or the default it points at) is not a configured agent.
func (c *Config) agentCommand(ref string) (string, bool) {
	name := ref
	if ref == DefaultAgentRef {
		if c.DefaultAgent == "" {
			return "", false
		}
		name = c.DefaultAgent
	}
	cmd, ok := c.Agents[name]
	return cmd, ok
}

// Path returns the config file location, honouring $WTX_HOME then XDG. It
// errors rather than returning a relative path when no base can be resolved,
// so Load never writes a starter config into the caller's cwd by accident.
func Path() (string, error) {
	if home := os.Getenv("WTX_HOME"); home != "" {
		return filepath.Join(home, "config.toml"), nil
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot determine config directory: set WTX_HOME or XDG_CONFIG_HOME: %w", err)
		}
		base = filepath.Join(h, ".config")
	}
	return filepath.Join(base, "wtx", "config.toml"), nil
}

// Load reads the config, writing a starter file the first time so users have
// something concrete to edit instead of hidden binary defaults.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := writeStarter(path); err != nil {
			return nil, err
		}
	}
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	names := map[string]bool{}
	for _, l := range c.Layouts {
		if strings.TrimSpace(l.Name) == "" {
			return fmt.Errorf("layout has an empty name")
		}
		if l.Name == BlankLayoutName {
			return fmt.Errorf("layout name %q is reserved for the blank session", BlankLayoutName)
		}
		if names[l.Name] {
			return fmt.Errorf("duplicate layout name %q", l.Name)
		}
		if !ValidArranges[l.Arrange] {
			return fmt.Errorf("layout %q has invalid arrange %q (valid: %s)", l.Name, l.Arrange, arrangeList())
		}
		if len(l.Panes) == 0 {
			return fmt.Errorf("layout %q has no panes", l.Name)
		}
		for _, pane := range l.Panes {
			if agent, ok := AgentRef(pane); ok {
				if agent == "" {
					return fmt.Errorf("layout %q has an empty agent reference", l.Name)
				}
				if agent == DefaultAgentRef {
					// agent:default (and the default_agent it points at) is
					// resolved after any --agent / $WTX_AGENT override, so its
					// validity is checked in Resolve, not here.
					continue
				}
				cmd, ok := c.agentCommand(agent)
				if !ok {
					return fmt.Errorf("layout %q references unknown agent %q", l.Name, agent)
				}
				if strings.TrimSpace(cmd) == "" {
					return fmt.Errorf("layout %q references agent %q with an empty command", l.Name, agent)
				}
			}
		}
		names[l.Name] = true
	}
	if c.DefaultLayout != "" && c.DefaultLayout != BlankLayoutName && !names[c.DefaultLayout] {
		return fmt.Errorf("default_layout %q is not a defined preset", c.DefaultLayout)
	}
	return nil
}

// Resolve selects a layout by name, applying precedence: an explicit name, then
// the configured default, then the blank layout. "" or "none" yields the blank
// single-pane layout. Returned pane commands have agent:<name> expanded.
func (c *Config) Resolve(name string) (Layout, error) {
	if name == "" {
		name = c.DefaultLayout
	}
	if name == "" || name == BlankLayoutName {
		return Layout{Name: BlankLayoutName, Arrange: "", Panes: []string{""}}, nil
	}
	for _, l := range c.Layouts {
		if l.Name != name {
			continue
		}
		panes := make([]string, len(l.Panes))
		for i, pane := range l.Panes {
			if agent, ok := AgentRef(pane); ok {
				cmd, ok := c.agentCommand(agent)
				if !ok || strings.TrimSpace(cmd) == "" {
					if agent == DefaultAgentRef {
						return Layout{}, fmt.Errorf("layout %q uses agent:default but the default agent %q is unset, not defined in [agents], or empty; set default_agent, $WTX_AGENT, or --agent to a valid agent", l.Name, c.DefaultAgent)
					}
					return Layout{}, fmt.Errorf("layout %q references unknown or empty agent %q", l.Name, agent)
				}
				panes[i] = cmd
			} else {
				panes[i] = pane
			}
		}
		return Layout{Name: l.Name, Arrange: l.Arrange, Panes: panes}, nil
	}
	return Layout{}, fmt.Errorf("unknown layout %q (available: %s)", name, c.layoutNames())
}

// TemplateVars are the values available to {{.Field}} placeholders in resolved
// pane commands, so an assistant can launch with worktree context rather than
// just the right working directory.
type TemplateVars struct {
	Branch  string // branch checked out in the worktree
	Path    string // absolute worktree path
	Repo    string // repository name
	Session string // tmux session name
}

// paneFuncs are template helpers available in pane commands. shq/shellquote
// POSIX-quote a value so it is safe to interpolate into a shell command even
// when a branch name contains shell metacharacters like $, ` or ".
var paneFuncs = template.FuncMap{
	"shq":        ShellQuote,
	"shellquote": ShellQuote,
}

// ShellQuote wraps s in single quotes, escaping any embedded single quotes, so
// the result is a single safe shell word.
func ShellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// ExpandPanes renders each pane command as a text/template against vars. Panes
// with no placeholders are returned unchanged, so plain commands never need
// escaping. A malformed template is reported with its pane index. Values are
// NOT auto-quoted; use {{shq .Branch}} when interpolating into shell text.
func ExpandPanes(panes []string, vars TemplateVars) ([]string, error) {
	out := make([]string, len(panes))
	for i, pane := range panes {
		if !strings.Contains(pane, "{{") {
			out[i] = pane
			continue
		}
		tmpl, err := template.New("pane").Option("missingkey=error").Funcs(paneFuncs).Parse(pane)
		if err != nil {
			return nil, fmt.Errorf("pane %d: %w", i, err)
		}
		var b strings.Builder
		if err := tmpl.Execute(&b, vars); err != nil {
			return nil, fmt.Errorf("pane %d: %w", i, err)
		}
		out[i] = b.String()
	}
	return out, nil
}

// AgentRef reports whether a pane string is an agent:<name> reference.
func AgentRef(pane string) (string, bool) {
	name, ok := strings.CutPrefix(pane, "agent:")
	if !ok {
		return "", false
	}
	return strings.TrimSpace(name), true
}

func (c *Config) layoutNames() string {
	out := make([]string, 0, len(c.Layouts))
	for _, l := range c.Layouts {
		out = append(out, l.Name)
	}
	sort.Strings(out)
	if len(out) == 0 {
		return "no presets defined"
	}
	return strings.Join(out, ", ")
}

func arrangeList() string {
	out := make([]string, 0, len(ValidArranges))
	for a := range ValidArranges {
		out = append(out, a)
	}
	sort.Strings(out)
	return strings.Join(out, ", ")
}

func writeStarter(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(starterConfig), 0o644)
}

const starterConfig = `# wtx configuration

# The code assistant launched by layouts that use agent:default. Point this at
# whatever tool you use; change it here to switch assistants everywhere without
# touching any layout.
default_agent = "claude"

# Layout selected when none is given on the command line ("none" = blank shell).
default_layout = "assistant"

# Named commands referenced by layouts as agent:<name>. Configure flags once.
# Commands may use template vars for worktree context: {{.Branch}}, {{.Path}},
# {{.Repo}}, {{.Session}}. Wrap any var interpolated into shell text with shq so
# unusual branch names stay safe, e.g.:
#   claude = 'claude {{shq (printf "work on %s" .Branch)}}'
[agents]
claude = "claude"
codex = "codex"

# Layout presets. Each pane is a literal command, "" (plain shell), or
# agent:<name>. agent:default resolves to the agent named by default_agent.
[[layouts]]
name = "assistant"
arrange = "even-horizontal"
panes = ["agent:default", ""]

# A two-assistant variant, e.g. wtx add feature/x --layout duo
[[layouts]]
name = "duo"
arrange = "even-horizontal"
panes = ["agent:claude", "agent:codex"]
`
