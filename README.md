# wtx

Tmux workspaces on top of [`wt`](https://github.com/timvw/wt)-managed git worktrees, with per-pane code-assistant layouts.

`wtx` is early: the command-line skeleton is in place, and the workspace commands are still being built. This README documents what exists today.

## Install

### From a release

Download the archive for your platform from the [latest release](https://github.com/timvw/wtx/releases/latest), extract it, and put the `wtx` binary somewhere on your `PATH`:

```sh
# Linux / macOS — adjust the version, OS and architecture to match the asset
tar -xzf wtx_0.1.0_darwin_arm64.tar.gz
sudo install -m 755 wtx /usr/local/bin/wtx
```

On Windows, extract the `.zip` and place `wtx.exe` in a directory on your `PATH`.

Archives are published for Linux and macOS on `amd64` and `arm64`, and for Windows on `amd64`. Each carries a static binary with no runtime dependencies, alongside this README and the license.

To verify a download, fetch `checksums.txt` from the same release and check it:

```sh
sha256sum -c checksums.txt --ignore-missing
```

### From source

Requires the Go toolchain version declared in [`go.mod`](go.mod).

```sh
git clone https://github.com/timvw/wtx.git
cd wtx
make build      # produces ./wtx with the version stamped in
```

## Usage

```sh
wtx             # print help
wtx version     # print the version of this binary
wtx help <cmd>  # help for a specific command
```

## Development

```sh
make fmt-check  # fail if anything is not gofmt-formatted
make vet        # go vet ./...
make lint       # golangci-lint, using .golangci.yml
make test       # go test ./...
```

CI runs exactly these checks on every pull request, plus the same test suite on Linux, macOS and Windows and a cross-compile of every platform a release publishes.

## License

MIT — see [LICENSE](LICENSE).
