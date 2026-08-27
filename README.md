# SSH Config Tool

[![codecov](https://codecov.io/gh/soulteary/ssh-config/branch/main/graph/badge.svg?token=W816DX12V8)](https://codecov.io/gh/soulteary/ssh-config) [![Go Report Card](https://goreportcard.com/badge/github.com/soulteary/ssh-config)](https://goreportcard.com/report/github.com/soulteary/ssh-config) [![CodeQL Advanced](https://github.com/soulteary/ssh-config/actions/workflows/codeql.yml/badge.svg)](https://github.com/soulteary/ssh-config/actions/workflows/codeql.yml) [![Release](https://github.com/soulteary/ssh-config/actions/workflows/build.yml/badge.svg)](https://github.com/soulteary/ssh-config/actions/workflows/build.yml)

**[中文文档](./README_CN.md)**

<img src=".github/github-repo-card.png" >

SSH Config Tool is a command-line utility for managing SSH configuration files. It allows you to manage your SSH config files using more expressive YAML/JSON formats while still being able to round-trip them back to classic `ssh_config` syntax.

## Features

- Converts YAML/JSON representations into standard SSH config files
- Converts classic SSH config files into YAML or JSON for easier editing and review
- Uses the lossless v3 schema by default, preserving comments, ordering, repeated directives, quoting, line endings, and unknown directives
- Keeps the previous map-based conversion and directory scan available through `-legacy`
- Supports reading configuration from files or standard input (stdin)
- Supports output to files or standard output (stdout)
- Automatically detects the input format (YAML/JSON/SSH Config); the default mode preserves the underlying SSH document bytes, while only `-legacy` normalizes trailing blank lines

## Installation

Use Docker or download the binary file suitable for your system and CPU architecture from the [GitHub release page](https://github.com/soulteary/ssh-config/releases).

Alternatively, you can install it via Homebrew.

```bash
brew tap soulteary/tap
brew install soulteary/tap/ssh-config
```

Go users can install the v3 command directly:

```bash
go install github.com/soulteary/ssh-config/v3@latest
```

## Usage

### Basic Usage

```bash
ssh-config [options]
```

Run without arguments to export `~/.ssh/config` as lossless v3 YAML on standard output:

```bash
ssh-config -lossless
```

Or, use Linux pipes to manipulate files:

```bash
cat input_file | ssh-config -lossless -to-yaml > output_file
```

### Docker

Download docker image:

```bash
docker pull soulteary/ssh-config:latest
# or
docker pull ghcr.io/soulteary/ssh-config:latest
```

Convert file (test.yaml) in the current directory to YAML (abc.yaml):

```bash
docker run --rm -it --user "$(id -u):$(id -g)" -v "$(pwd):/ssh" soulteary/ssh-config:latest -lossless -to-yaml -src /ssh/test.yaml -dest /ssh/abc.yaml
```

Passing the host UID and GID keeps files written to the bind mount owned by
the current user. It is not required when the result is written only to
standard output.

Just want to see the conversion results:

```bash
docker run --rm -it -v "$(pwd):/ssh" soulteary/ssh-config:latest -lossless -to-yaml -src /ssh/test.yaml
```

If you want to use Linux pipelines, you can first enter the Docker interactive command line:

```bash
docker run --rm -it --entrypoint bash -v "$(pwd):/ssh" soulteary/ssh-config:latest
cat /ssh/test.yaml | ssh-config -lossless -to-yaml
```

### Options

- `-to-yaml, -to-json, -to-ssh`: Specify output format (yaml/json/config), only one output format can be specified at a time.
- `-src`: Specify the source file. When omitted, lossless mode reads `~/.ssh/config`; legacy mode scans `~/.ssh`.
- `-dest`: Specify the path to save the configuration file. Its parent directory must already exist. When omitted, the converted result is written to standard output.
- `-document-path`: Select a document by its `path` when `-to-ssh` reads a multi-document v3 schema.
- `-legacy`: Use the previous lossy map/array formats. This mode also enables directory scanning.
- `-lossless`: Deprecated compatibility alias; lossless conversion is already the default in v3.
- `-help`: View program command-line help
- `-version`: Print release, commit, build, and tree-state metadata

### Examples

1. Export the primary SSH configuration as lossless v3 YAML (default behaviour):

```bash
ssh-config
```

2. Convert YAML format to SSH config format:

```bash
ssh-config -lossless -to-ssh -src input.yaml -dest output.conf
```

3. Convert SSH config format to JSON format:

```bash
ssh-config -lossless -to-json -src ~/.ssh/config -dest output.json
```

4. Read from standard input, output to standard output, and save in YAML format:

```bash
cat input.conf | ssh-config -lossless -to-yaml > output.yaml
```

5. Losslessly edit a configuration through the v3 YAML representation:

```bash
ssh-config -lossless -to-yaml -src ~/.ssh/config -dest config.v3.yaml
# Edit directive fields in config.v3.yaml. Unchanged lines retain their exact bytes.
ssh-config -lossless -to-ssh -src config.v3.yaml -dest ~/.ssh/config
```

The previous YAML/JSON formats remain readable and are migrated to schema v3 by default. Use `-legacy` only when an existing consumer still requires the old map/array output. Repeated values and directive ordering already absent from a legacy document cannot be reconstructed.

See the [lossless schema v3 specification](./docs/lossless-schema-v3.md) for node shapes, byte-preservation rules, editing behavior, API examples, and migration limits.
See the [v2 to v3 migration guide](./docs/migration-v3.md) before updating scripts or Go imports.

## Development

### Dependencies

- Go 1.27+

### Build

```bash
go build
```

### Test

```bash
go test -v ./... -covermode=atomic -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html
```

## Contributing

Issues and pull requests are welcome.

Please report vulnerabilities through the private process in [SECURITY.md](./SECURITY.md), not through a public issue.

## License

This project is licensed under the Apache License. See the [LICENSE](./LICENSE) file for details.

# Guide

- [SSH Config Tool: Use structured data to manage SSH configuration](https://soulteary.com/2024/10/15/manage-ssh-configuration-using-structure-data-ssh-config-tool.html)

# Credits

- Useful OpenSSH software
  - https://man.openbsd.org/ssh_config
- Inspiration for the definition of configuration files
  - https://github.com/bencromwell/sshush
