package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func sampleConfig() *Config {
	return &Config{
		DefaultAgent:  "claude",
		DefaultLayout: "quad",
		Agents: map[string]string{
			"claude": "claude",
			"codex":  "codex --flag",
		},
		Layouts: []Layout{
			{Name: "quad", Arrange: "even-horizontal", Panes: []string{"agent:claude", "agent:codex", "", "echo hi"}},
			{Name: "solo", Arrange: "even-vertical", Panes: []string{"agent:default"}},
		},
	}
}

func TestResolveBlank(t *testing.T) {
	// "none" is always blank; "" is blank only when no default is configured.
	c := sampleConfig()
	c.DefaultLayout = ""
	for _, name := range []string{"", "none"} {
		l, err := c.Resolve(name)
		if err != nil {
			t.Fatalf("Resolve(%q) error: %v", name, err)
		}
		if l.Name != BlankLayoutName || !reflect.DeepEqual(l.Panes, []string{""}) {
			t.Errorf("Resolve(%q) = %+v, want blank single pane", name, l)
		}
	}

	// "none" stays blank even when a default preset exists.
	if l, _ := sampleConfig().Resolve("none"); l.Name != BlankLayoutName {
		t.Errorf("Resolve(none) with default set = %q, want blank", l.Name)
	}
}

func TestResolveEmptyUsesDefault(t *testing.T) {
	c := sampleConfig()
	// default_layout is "quad", so "" resolves to quad, not blank.
	// (Overriding DefaultLayout to a real preset to exercise the branch.)
	c.DefaultLayout = "solo"
	l, err := c.Resolve("")
	if err != nil {
		t.Fatal(err)
	}
	if l.Name != "solo" {
		t.Errorf("Resolve(\"\") = %q, want solo (the default)", l.Name)
	}
}

func TestResolveExpandsAgents(t *testing.T) {
	c := sampleConfig()
	l, err := c.Resolve("quad")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"claude", "codex --flag", "", "echo hi"}
	if !reflect.DeepEqual(l.Panes, want) {
		t.Errorf("Resolve(quad) panes = %v, want %v", l.Panes, want)
	}
}

func TestResolveUnknown(t *testing.T) {
	if _, err := sampleConfig().Resolve("nope"); err == nil {
		t.Fatal("expected error for unknown layout")
	}
}

func TestResolveDefaultAgent(t *testing.T) {
	c := sampleConfig() // DefaultAgent = "claude"; solo = ["agent:default"]
	l, err := c.Resolve("solo")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(l.Panes, []string{"claude"}) {
		t.Errorf("agent:default resolved to %v, want [claude]", l.Panes)
	}

	// Switching default_agent redirects agent:default without touching layouts.
	c.DefaultAgent = "codex"
	l, _ = c.Resolve("solo")
	if !reflect.DeepEqual(l.Panes, []string{"codex --flag"}) {
		t.Errorf("agent:default after switch = %v, want [codex --flag]", l.Panes)
	}
}

func TestResolveDefaultAgentUnsetErrors(t *testing.T) {
	c := sampleConfig()
	c.DefaultAgent = "" // no config default and no runtime override applied
	if _, err := c.Resolve("solo"); err == nil {
		t.Fatal("expected error resolving agent:default with no agent set")
	}
}

func TestStaleDefaultAgentRescuedByOverride(t *testing.T) {
	// A config default_agent that names a missing agent must not break Load;
	// it only matters at Resolve, and a runtime override must be able to win.
	c := sampleConfig()
	c.DefaultAgent = "ghost" // not in [agents]

	if err := c.validate(); err != nil {
		t.Fatalf("stale default_agent should pass validate (deferred): %v", err)
	}
	if _, err := c.Resolve("solo"); err == nil {
		t.Error("expected Resolve error while default_agent is stale and unresolved")
	}

	// The higher-precedence override rescues it.
	if err := c.OverrideDefaultAgent("codex"); err != nil {
		t.Fatal(err)
	}
	l, err := c.Resolve("solo")
	if err != nil || !reflect.DeepEqual(l.Panes, []string{"codex --flag"}) {
		t.Errorf("after override Resolve(solo) = %v (err=%v), want [codex --flag]", l.Panes, err)
	}
}

func TestResolveDefaultAgentEmptyCommand(t *testing.T) {
	// default_agent names a real entry, but its command is empty: must error,
	// not silently launch a blank pane.
	c := sampleConfig()
	c.Agents["claude"] = ""
	if _, err := c.Resolve("solo"); err == nil {
		t.Fatal("expected error when the default agent's command is empty")
	}
}

func TestOverrideDefaultAgent(t *testing.T) {
	c := sampleConfig()

	// Empty keeps the config value.
	if err := c.OverrideDefaultAgent(""); err != nil || c.DefaultAgent != "claude" {
		t.Errorf("empty override changed agent to %q (err=%v)", c.DefaultAgent, err)
	}
	// Valid name switches it.
	if err := c.OverrideDefaultAgent("codex"); err != nil || c.DefaultAgent != "codex" {
		t.Errorf("override to codex failed: agent=%q err=%v", c.DefaultAgent, err)
	}
	// Unknown name errors and leaves the value unchanged.
	if err := c.OverrideDefaultAgent("ghost"); err == nil {
		t.Error("expected error overriding to unknown agent")
	}
	if c.DefaultAgent != "codex" {
		t.Errorf("failed override mutated agent to %q", c.DefaultAgent)
	}
}

func TestExpandPanes(t *testing.T) {
	vars := TemplateVars{Branch: "feature/x", Path: "/wt/x", Repo: "wtx", Session: "wtx/feature/x"}
	in := []string{
		"claude",                         // no template, passthrough
		"claude \"work on {{.Branch}}\"", // branch context
		"cat {{.Path}}/TASK.md",          // path context
	}
	got, err := ExpandPanes(in, vars)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"claude",
		"claude \"work on feature/x\"",
		"cat /wt/x/TASK.md",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ExpandPanes = %v, want %v", got, want)
	}
}

func TestExpandPanesShellQuote(t *testing.T) {
	// A hostile branch name must not break out of the shell command.
	vars := TemplateVars{Branch: `x"; rm -rf ~; echo "`}
	got, err := ExpandPanes([]string{"claude {{shq .Branch}}"}, vars)
	if err != nil {
		t.Fatal(err)
	}
	want := `claude 'x"; rm -rf ~; echo "'`
	if got[0] != want {
		t.Errorf("shq expansion = %q, want %q", got[0], want)
	}
}

func TestShellQuote(t *testing.T) {
	if got := ShellQuote("it's"); got != `'it'\''s'` {
		t.Errorf("ShellQuote(\"it's\") = %q", got)
	}
}

func TestExpandPanesMalformed(t *testing.T) {
	if _, err := ExpandPanes([]string{"{{.Branch"}, TemplateVars{}); err == nil {
		t.Fatal("expected error for malformed template")
	}
}

func TestExpandPanesUnknownKey(t *testing.T) {
	if _, err := ExpandPanes([]string{"{{.Nope}}"}, TemplateVars{}); err == nil {
		t.Fatal("expected error for unknown template field")
	}
}

func TestAgentRef(t *testing.T) {
	if name, ok := AgentRef("agent:claude"); !ok || name != "claude" {
		t.Errorf("AgentRef(agent:claude) = %q,%v", name, ok)
	}
	if _, ok := AgentRef("plain command"); ok {
		t.Error("AgentRef should reject a plain command")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{"ok", func(*Config) {}, false},
		{"reserved name", func(c *Config) { c.Layouts[0].Name = BlankLayoutName }, true},
		{"bad arrange", func(c *Config) { c.Layouts[0].Arrange = "diagonal" }, true},
		{"no panes", func(c *Config) { c.Layouts[0].Panes = nil }, true},
		{"unknown agent", func(c *Config) { c.Layouts[0].Panes = []string{"agent:ghost"} }, true},
		{"default not defined", func(c *Config) { c.DefaultLayout = "ghost" }, true},
		{"default none ok", func(c *Config) { c.DefaultLayout = BlankLayoutName }, false},
		{"empty name", func(c *Config) { c.Layouts[0].Name = "  " }, true},
		{"duplicate name", func(c *Config) { c.Layouts[1].Name = c.Layouts[0].Name }, true},
		// agent:default and default_agent are always deferred to Resolve (they
		// may be overridden at runtime), so neither an unset nor a stale
		// default_agent fails validation.
		{"agent:default unset deferred", func(c *Config) { c.DefaultAgent = "" }, false},
		{"stale default_agent deferred", func(c *Config) { c.DefaultAgent = "ghost" }, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := sampleConfig()
			tc.mutate(c)
			err := c.validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestLoadWritesStarter(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("WTX_HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "config.toml")); err != nil {
		t.Errorf("starter config not written: %v", err)
	}
	// Starter must be internally consistent (its default resolves to a preset).
	if _, err := cfg.Resolve(""); err != nil {
		t.Errorf("starter default_layout does not resolve: %v", err)
	}
}

func TestPathHonoursWTXHome(t *testing.T) {
	t.Setenv("WTX_HOME", "/custom/place")
	got, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if got != "/custom/place/config.toml" {
		t.Errorf("Path() = %q, want /custom/place/config.toml", got)
	}
}
