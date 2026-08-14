package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// version is stamped at build time with
//
//	-ldflags "-X github.com/timvw/wtx/cmd.version=<tag>"
//
// It is deliberately empty by default. An unstamped build falls back to the
// binary's own build metadata, which is more informative than any constant a
// release process would have to remember to edit.
var version string

const (
	// develVersion is what the toolchain records in Main.Version for a build
	// that was not resolved through the module proxy -- an ordinary local
	// `go build`, for instance. It identifies nothing, so it is treated as
	// absent and resolution falls through to the VCS revision.
	develVersion = "(devel)"

	// devVersion is reported when nothing else is available.
	devVersion = "dev"

	// revisionLen is how much of a VCS revision to keep: twelve hex digits,
	// git's customary abbreviation for a human-readable commit.
	revisionLen = 12
)

// Version reports the version of this binary, in descending order of
// precedence: a version stamped in at build time, the module version or source
// revision the toolchain embedded, or a development placeholder.
//
// It always returns a non-empty string. Build metadata is unavailable in some
// build modes, and a binary that cannot say what it is is worse than one that
// admits to being a development build.
func Version() string {
	if version != "" {
		return version
	}

	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return devVersion
	}

	if v := bi.Main.Version; v != "" && v != develVersion {
		return v
	}

	if v := versionFromVCS(bi.Settings); v != "" {
		return v
	}

	return devVersion
}

// versionFromVCS derives a version from the VCS stamps the toolchain embeds
// during a `go build` in a checkout. It returns "" when the build carries no
// revision, which happens for source tarballs and for -buildvcs=false.
func versionFromVCS(settings []debug.BuildSetting) string {
	var revision string
	var modified bool

	for _, s := range settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}

	if revision == "" {
		return ""
	}
	if len(revision) > revisionLen {
		revision = revision[:revisionLen]
	}
	if modified {
		revision += "+dirty"
	}

	return revision
}

// newVersionCmd builds the `wtx version` command.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the wtx version",
		Long: `Print the version of this wtx binary.

The version is whatever was stamped in at build time, falling back to the
module version or commit the binary was built from, and finally to "dev".`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), Version())
			return err
		},
	}
}
