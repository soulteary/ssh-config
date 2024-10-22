# SSH Config Tool

[![codecov](https://codecov.io/gh/soulteary/ssh-config/branch/main/graph/badge.svg?token=W816DX12V8)](https://codecov.io/gh/soulteary/ssh-config) [![Go Report Card](https://goreportcard.com/badge/github.com/soulteary/ssh-config)](https://goreportcard.com/report/github.com/soulteary/ssh-config) [![CodeQL Advanced](https://github.com/soulteary/ssh-config/actions/workflows/codeql.yml/badge.svg)](https://github.com/soulteary/ssh-config/actions/workflows/codeql.yml) [![Release](https://github.com/soulteary/ssh-config/actions/workflows/build.yml/badge.svg)](https://github.com/soulteary/ssh-config/actions/workflows/build.yml)

**[中文文档](./README_CN.md)**

<img src=".github/github-repo-card.png" >

SSH Config Tool is a command-line utility for managing SSH configuration files. It allows you to manage your SSH config files using more expressive YAML/JSON formats.

## Features

- Supports conversion from YAML/JSON format to standard SSH config format
- Supports conversion from standard SSH config format to YAML/JSON format
- Supports reading configuration from files or standard input (stdin)
- Supports output to files or standard output (stdout)
- Automatically detects input format (YAML/JSON/SSH Config)

## Installation

Use Docker or download the binary file suitable for your system and CPU architecture from the [GitHub release page](https://github.com/soulteary/ssh-config/releases).

## Usage

### Basic Usage

```bash
ssh-config [options] <input_file> <output_file>
```

Or, use Linux pipes to manipulate files:

```bash
cat input_file | ssh-config -to-yaml > output_file
```

### Docker

Download docker image:

```bash
docker pull soulteary/ssh-config:v1.1.1
# or
docker pull ghcr.io/soulteary/ssh-config:v1.1.1
```

Convert file (test.yaml) in the current directory to YAML (abc.yaml):

```bash
docker run --rm -it -v `pwd`:/ssh soulteary/ssh-config:v1.1.1 ssh-config -to-yaml -src /ssh/test.yaml -dest /ssh/abc.yaml
```

Just want to see the conversion results:

```bash
docker run --rm -it -v `pwd`:/ssh soulteary/ssh-config:v1.1.1 ssh-config -to-yaml -src /ssh/test.yaml
```

If you want to use Linux pipelines, you can first enter the Docker interactive command line:

```bash
docker run --rm -it -v `pwd`:/ssh soulteary/ssh-config:v1.1.1 bash
cat /ssh/test.yaml | ssh-config -to-yaml
```

### Options

- `-to-yaml, -to-json, -to-ssh`: Specify output format (yaml/json/config), only one output format can be specified at a time.
- `-src`: Specify the original configuration file or directory to read from
- `-dest`: Specify the path to save the configuration file
- `-help`: View program command-line help

### Examples

1. Convert YAML format to SSH config format:

```bash
ssh-config -to-ssh -src input.yaml -dest output.conf
```

2. Convert SSH config format to JSON format:

```bash
ssh-config -to-json -src ~/.ssh/config -dest output.json
```

3. Read from standard input, output to standard output, and save in YAML format:

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
