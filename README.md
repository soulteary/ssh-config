# SSH Config Tool

[![codecov](https://codecov.io/gh/soulteary/ssh-config/branch/main/graph/badge.svg?token=W816DX12V8)](https://codecov.io/gh/soulteary/ssh-config) [![Go Report Card](https://goreportcard.com/badge/github.com/soulteary/ssh-config)](https://goreportcard.com/report/github.com/soulteary/ssh-config) [![CodeQL Advanced](https://github.com/soulteary/ssh-config/actions/workflows/codeql.yml/badge.svg)](https://github.com/soulteary/ssh-config/actions/workflows/codeql.yml) [![Release](https://github.com/soulteary/ssh-config/actions/workflows/build.yml/badge.svg)](https://github.com/soulteary/ssh-config/actions/workflows/build.yml)

**[中文文档](./README_CN.md)**

<img src=".github/github-repo-card.png" >

SSH Config Tool is a command-line utility for managing SSH configuration files. It allows you to manage your SSH config files using more expressive YAML/JSON formats while still being able to round-trip them back to classic `ssh_config` syntax.

## Features

- Converts YAML/JSON representations into standard SSH config files
- Converts classic SSH config files into YAML or JSON for easier editing and review
- Scans a single file or an entire directory tree (such as `~/.ssh`) while skipping key material and other non-config files
- Supports reading configuration from files or standard input (stdin)
- Supports output to files or standard output (stdout), creating parent folders when needed
- Automatically detects the input format (YAML/JSON/SSH Config) and tidies trailing blank lines

## Installation

Use Docker or download the binary file suitable for your system and CPU architecture from the [GitHub release page](https://github.com/soulteary/ssh-config/releases).

## Usage

### Basic Usage

```bash
ssh-config [options]
```

Run without arguments to export all SSH configuration under `~/.ssh` to YAML on standard output:

```bash
ssh-config
```

Or, use Linux pipes to manipulate files:

```bash
cat input_file | ssh-config -to-yaml > output_file
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
docker run --rm -it -v `pwd`:/ssh soulteary/ssh-config:latest ssh-config -to-yaml -src /ssh/test.yaml -dest /ssh/abc.yaml
```

Just want to see the conversion results:

```bash
docker run --rm -it -v `pwd`:/ssh soulteary/ssh-config:latest ssh-config -to-yaml -src /ssh/test.yaml
```

If you want to use Linux pipelines, you can first enter the Docker interactive command line:

```bash
docker run --rm -it -v `pwd`:/ssh soulteary/ssh-config:latest bash
cat /ssh/test.yaml | ssh-config -to-yaml
```

### Options

- `-to-yaml, -to-json, -to-ssh`: Specify output format (yaml/json/config), only one output format can be specified at a time.
- `-src`: Specify the original configuration file or directory to read from. When omitted, the tool scans `~/.ssh`.
- `-dest`: Specify the path to save the configuration file. When omitted, the converted result is written to standard output.
- `-help`: View program command-line help

### Examples

1. Export the SSH configuration for your current user to YAML (default behaviour):

```bash
ssh-config
```

2. Convert YAML format to SSH config format:

```bash
ssh-config -to-ssh -src input.yaml -dest output.conf
```

3. Convert SSH config format to JSON format:

```bash
ssh-config -to-json -src ~/.ssh/config -dest output.json
```

4. Read from standard input, output to standard output, and save in YAML format:

```bash
cat input.conf | ssh-config -to-yaml > output.yaml
```

## Development

### Dependencies

- Go 1.23+

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

## License

This project is licensed under the Apache License. See the [LICENSE](./LICENSE) file for details.

# Guide

- [SSH Config Tool: Use structured data to manage SSH configuration](https://soulteary.com/2024/10/15/manage-ssh-configuration-using-structure-data-ssh-config-tool.html)

# Credits

- Useful OpenSSH software
  - https://man.openbsd.org/ssh_config
- Inspiration for the definition of configuration files
  - https://github.com/bencromwell/sshush
