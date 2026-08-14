## Purpose

Defines how a version of `wtx` is published: what pushing a version tag produces, which platforms get downloadable artifacts, how those artifacts are named, verified, and licensed, and what a user can rely on when they download one.

## ADDED Requirements

### Requirement: Pushing a version tag publishes a release

Pushing a tag of the form `vMAJOR.MINOR.PATCH`, where each component is one or more decimal digits, SHALL trigger an automated release pipeline that builds the executable for every supported platform and publishes a GitHub Release for that tag.

No other event SHALL trigger the release pipeline. In particular, a push to any branch, a pull request, and a tag that does not match the version format SHALL NOT publish anything.

Publishing is externally visible and effectively irreversible — downstream tooling consumes published tags and assets — so the trigger is deliberately narrow.

#### Scenario: Version tag publishes a release

- **WHEN** a tag matching `vMAJOR.MINOR.PATCH` is pushed to the repository
- **THEN** the release pipeline runs
- **AND** on success a GitHub Release exists for that tag with the built artifacts attached

#### Scenario: Non-version tag publishes nothing

- **WHEN** a tag that does not match `vMAJOR.MINOR.PATCH` is pushed
- **THEN** the release pipeline does not run
- **AND** no GitHub Release is created

#### Scenario: Branch push publishes nothing

- **WHEN** a commit is pushed to any branch, including the default branch
- **THEN** the release pipeline does not run

### Requirement: A release provides archives for every supported platform

A release SHALL include one downloadable archive per supported operating-system and architecture pair. The supported pairs SHALL be:

- Linux on `amd64` and on `arm64`
- macOS on `amd64` and on `arm64`
- Windows on `amd64`

Windows on `arm64` is deliberately excluded. Each archive SHALL contain the `wtx` executable for exactly that platform, built without requiring a C toolchain or any dynamically linked library beyond what the target operating system provides, so that extracting the archive and running the executable is sufficient to use it.

#### Scenario: Every supported platform has an archive

- **WHEN** a release is published
- **THEN** it carries one archive for each of the five supported operating-system and architecture pairs

#### Scenario: An extracted binary runs without further installation

- **WHEN** a user downloads the archive for their platform and extracts it
- **THEN** the extracted `wtx` executable runs on that platform
- **AND** no runtime, interpreter, or additional shared library has to be installed first

#### Scenario: No archive is published for an unsupported platform

- **WHEN** a release is published
- **THEN** it carries no archive for Windows on `arm64`

### Requirement: Release archives are named predictably and use the platform's conventional format

Archive file names SHALL encode the project name, the released version, the operating system, and the architecture, in a form that is stable across releases so that scripted downloads can construct a URL from a version string alone.

Archives for Linux and macOS SHALL use `tar.gz`; archives for Windows SHALL use `zip`.

#### Scenario: Name encodes project, version, and platform

- **WHEN** a release asset is listed
- **THEN** its file name contains the project name, the released version, the operating system, and the architecture

#### Scenario: Names are constructible from a version

- **WHEN** a script knows the released version and its target platform
- **THEN** it can construct the asset's download URL without listing the release's contents first

#### Scenario: Archive format matches the platform convention

- **WHEN** the Linux and macOS assets are inspected
- **THEN** they are `tar.gz` archives
- **AND** the Windows asset is a `zip` archive

### Requirement: A release publishes checksums for its artifacts

A release SHALL include a checksum file covering every published archive, so that a downloader can verify an archive's integrity without trusting the transport alone.

#### Scenario: Checksum file is published

- **WHEN** a release is published
- **THEN** it includes a checksum file listing a cryptographic digest for each published archive

#### Scenario: Published digests match the published archives

- **WHEN** a user downloads an archive and the checksum file from the same release
- **AND** computes the archive's digest
- **THEN** the computed digest equals the one recorded in the checksum file

### Requirement: A released binary reports the released version

The executable inside a release archive SHALL report the version of the tag it was built from, with the leading `v` of the tag omitted from the reported version being permitted but the version otherwise identifying the tag unambiguously.

The version SHALL be stamped in at build time rather than read from a file that a human has to remember to edit, so that the tag is the single source of truth for what a release is called.

#### Scenario: Version subcommand reports the released version

- **WHEN** the executable from the archive published for tag `vX.Y.Z` is run as `wtx version`
- **THEN** it prints a version that identifies `X.Y.Z`

#### Scenario: No hand-edited version constant is required

- **WHEN** a maintainer publishes a release
- **THEN** pushing the version tag is sufficient
- **AND** no source file has to be edited to record the version number

### Requirement: The project declares its license and ships it with every release

The repository SHALL contain an MIT license file at its root naming Tim Van Wassenhove as the copyright holder, and SHALL contain a README describing what `wtx` is and how to install a released binary.

Every release archive SHALL contain both files alongside the executable, so that a downloaded archive states the terms it is distributed under without the recipient having to visit the repository.

#### Scenario: License is present in the repository

- **WHEN** the repository is checked out
- **THEN** an MIT license file naming Tim Van Wassenhove is present at the root

#### Scenario: Archives carry the license and README

- **WHEN** any release archive is extracted
- **THEN** it contains the license file and the README in addition to the `wtx` executable

#### Scenario: README documents installation from a release

- **WHEN** a user reads the README
- **THEN** it states what `wtx` is
- **AND** it describes how to obtain and install a released binary

### Requirement: A release records what changed

A release SHALL carry release notes derived from the commits between it and the previous release, so that a user can tell what changed without reading the commit log themselves. Purely internal commits — documentation, test, and chore commits, and merge commits — SHALL be omitted from those notes.

#### Scenario: Release notes list the changes

- **WHEN** a release is published
- **THEN** its notes list the user-relevant commits made since the previous release

#### Scenario: Housekeeping commits are omitted

- **WHEN** the commits since the previous release include documentation, test, chore, or merge commits
- **THEN** those commits do not appear in the release notes

### Requirement: A failed release pipeline publishes nothing partial that cannot be corrected

If any stage of the release pipeline fails, the pipeline SHALL fail visibly rather than publishing a release that is missing artifacts without indication. Re-running the pipeline for the same tag after a transient failure SHALL be safe: it SHALL converge on a complete release rather than failing because an artifact from the earlier attempt already exists.

#### Scenario: Build failure fails the release

- **WHEN** the build for any supported platform fails during a release run
- **THEN** the release pipeline reports failure

#### Scenario: Re-running a release for the same tag is safe

- **WHEN** a release run fails partway through after publishing some assets
- **AND** the pipeline is re-run for the same tag
- **THEN** it completes and the release carries the full, correct set of assets
- **AND** the re-run does not fail merely because an asset was already present
