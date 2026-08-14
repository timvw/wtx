## Purpose

Defines the command-line surface of the `wtx` executable: how it is invoked, how it presents help and usage, how it reports its own version, and what it writes to which stream with which exit code. This is the contract every future `wtx` subcommand plugs into.

## ADDED Requirements

### Requirement: The project builds an executable named wtx

The project SHALL build a single self-contained executable whose invocation name is `wtx`. The executable SHALL be runnable on the platforms the Go toolchain targets by default, without requiring a runtime or interpreter to be installed alongside it.

#### Scenario: The executable builds from a clean checkout

- **WHEN** the repository is checked out and the project's build step is run
- **THEN** the build succeeds
- **AND** it produces an executable named `wtx`

#### Scenario: The executable runs

- **WHEN** the built `wtx` executable is invoked
- **THEN** it runs to completion without requiring any additional runtime to be installed

### Requirement: Invoking wtx with no arguments prints help and succeeds

Running `wtx` with no arguments SHALL print the root help text and exit with status `0`. Bare invocation SHALL NOT be treated as an error, and SHALL NOT perform any action beyond printing help.

#### Scenario: Bare invocation

- **WHEN** `wtx` is run with no arguments
- **THEN** the root help text is printed
- **AND** the exit status is `0`

#### Scenario: Bare invocation has no side effects

- **WHEN** `wtx` is run with no arguments
- **THEN** no files are created or modified and no external process is started

### Requirement: Help is reachable as both a subcommand and a flag

The executable SHALL provide its usage information through a `help` subcommand and through the `-h` and `--help` flags. All three SHALL be accepted at the root level, and `-h`/`--help` SHALL additionally be accepted on every subcommand. Requesting help SHALL exit with status `0`.

The root help text SHALL name the program, describe in at least one sentence what `wtx` is for, and list the available subcommands with a one-line description of each.

#### Scenario: Help subcommand

- **WHEN** `wtx help` is run
- **THEN** the root help text is printed
- **AND** the exit status is `0`

#### Scenario: Help flags

- **WHEN** `wtx --help` is run, or `wtx -h` is run
- **THEN** the root help text is printed
- **AND** the exit status is `0`

#### Scenario: Root help lists available subcommands

- **WHEN** the root help text is printed
- **THEN** it names the program `wtx`
- **AND** it contains a description of what `wtx` is for
- **AND** it lists `help` and `version` among the available commands, each with a one-line description

#### Scenario: Help for a specific subcommand

- **WHEN** `wtx help version` is run, or `wtx version --help` is run
- **THEN** usage information for the `version` subcommand is printed
- **AND** the exit status is `0`

#### Scenario: Help for an unknown subcommand

- **WHEN** `wtx help no-such-command` is run
- **THEN** a diagnostic naming the unknown command is written to standard error
- **AND** the exit status is non-zero

### Requirement: wtx reports its own version

The executable SHALL report its version through a `version` subcommand and through a `--version` flag on the root command. Both SHALL print the same version string, and both SHALL exit with status `0`.

The version output SHALL be a single line written to standard output, SHALL be non-empty, and SHALL contain no leading or trailing whitespace beyond the terminating newline. Output SHALL be stable enough for a script to consume: invoking the same binary twice SHALL produce identical output.

#### Scenario: Version subcommand

- **WHEN** `wtx version` is run
- **THEN** a single non-empty line containing the version is written to standard output
- **AND** nothing is written to standard error
- **AND** the exit status is `0`

#### Scenario: Version flag matches the version subcommand

- **WHEN** `wtx --version` is run
- **THEN** the version string it prints is identical to the one printed by `wtx version`

#### Scenario: Version output is stable

- **WHEN** the same `wtx` binary is invoked as `wtx version` twice
- **THEN** both invocations write byte-identical output

### Requirement: The reported version reflects how the binary was built

The version string SHALL be derived from the binary's own build provenance rather than from a value that must be edited by hand for each release. Resolution SHALL follow this precedence:

1. A version explicitly stamped into the binary at build time, when present.
2. Otherwise, the version or source revision recorded in the binary's embedded build metadata.
3. Otherwise, a placeholder indicating an unversioned development build.

The executable SHALL NOT fail, panic, or report an empty version when no build metadata is available.

#### Scenario: Release build reports the stamped version

- **WHEN** the binary is built with an explicit version stamped in at build time
- **AND** `wtx version` is run
- **THEN** the output is the stamped version

#### Scenario: Installed build reports build metadata

- **WHEN** the binary is built or installed by the Go toolchain without an explicitly stamped version
- **AND** the toolchain recorded a module version or source revision in the binary
- **THEN** `wtx version` reports a version derived from that recorded metadata

#### Scenario: Build with no version information available

- **WHEN** no version was stamped in and no usable build metadata is embedded in the binary
- **AND** `wtx version` is run
- **THEN** a non-empty placeholder identifying the build as a development build is printed
- **AND** the exit status is `0`

### Requirement: Exit status and output streams follow shell conventions

The executable SHALL exit with status `0` when the requested command succeeds and with a non-zero status when it fails, so that shell scripts and CI steps can branch on the result. Diagnostics for failures SHALL be written to standard error, keeping standard output free of error text so it remains machine-consumable.

An unrecognised command or an unrecognised flag SHALL be treated as a usage error: a diagnostic naming the offending input SHALL be written to standard error and the exit status SHALL be non-zero.

#### Scenario: Unknown subcommand

- **WHEN** `wtx no-such-command` is run
- **THEN** a diagnostic naming `no-such-command` is written to standard error
- **AND** standard output contains no error text
- **AND** the exit status is non-zero

#### Scenario: Unknown flag

- **WHEN** `wtx --no-such-flag` is run
- **THEN** a diagnostic naming the unrecognised flag is written to standard error
- **AND** the exit status is non-zero

#### Scenario: Successful command

- **WHEN** any `wtx` command completes successfully
- **THEN** the exit status is `0`
