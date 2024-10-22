# SSH Config Tool

[![codecov](https://codecov.io/gh/soulteary/ssh-config/branch/main/graph/badge.svg?token=W816DX12V8)](https://codecov.io/gh/soulteary/ssh-config) [![Go Report Card](https://goreportcard.com/badge/github.com/soulteary/ssh-config)](https://goreportcard.com/report/github.com/soulteary/ssh-config) [![CodeQL Advanced](https://github.com/soulteary/ssh-config/actions/workflows/codeql.yml/badge.svg)](https://github.com/soulteary/ssh-config/actions/workflows/codeql.yml) [![Release](https://github.com/soulteary/ssh-config/actions/workflows/build.yml/badge.svg)](https://github.com/soulteary/ssh-config/actions/workflows/build.yml)

**[English Docs](./README.md)**

<img src=".github/github-repo-card.png" >

SSH Config Tool 是一个用于管理 SSH 配置文件的命令行工具。它允许你使用更具表现力的 YAML/JSON 格式来管理你的 SSH 配置文件。

## 特性

- 支持从 YAML/JSON 格式转换为标准 SSH 配置格式
- 支持从标准 SSH 配置格式转换为 YAML/JSON 格式
- 支持从文件输入或标准输入(stdin)读取配置
- 支持输出到文件或标准输出(stdout)
- 自动检测输入格式(YAML/JSON/SSH Config)

## 安装

使用 Docker 或者从 [GitHub 发布页面](https://github.com/soulteary/ssh-config/releases)下载合适你的系统、CPU 架构的二进制文件即可。

## 使用方法

### 基本用法

```bash
ssh-config [options] <input_file> <output_file>
```

或，使用 Linux 管道来操作文件：

```bash
cat input_file | ssh-config -to-yaml > output_file
```

### Docker

下载镜像

```bash
docker pull soulteary/ssh-config:v1.1.1
# or
docker pull ghcr.io/soulteary/ssh-config:v1.1.1
```

将当前目录的配置文件转换并保存为新的文件：

```bash
docker run --rm -it -v `pwd`:/ssh soulteary/ssh-config:v1.1.1 ssh-config -to-yaml -src /ssh/test.yaml -dest /ssh/abc.yaml
```

如果你只想看看转换结果：

```bash
docker run --rm -it -v `pwd`:/ssh soulteary/ssh-config:v1.1.1 ssh-config -to-yaml -src /ssh/test.yaml
```

如果你想使用 Linux 管道来操作文件，可以先进入 Docker 交互式命令行：

```bash
docker run --rm -it -v `pwd`:/ssh soulteary/ssh-config:v1.1.1 bash
cat /ssh/test.yaml | ssh-config -to-yaml
```

### 选项

- `-to-yaml, -to-json, -to-ssh`: 指定输出格式 (yaml/json/config)，同一时间，输出格式只能指定为一种。
- `-src`: 指定要读取的原始配置文件，或配置目录
- `-dest`: 指定要保存的配置文件路径
- `-help`: 查看程序命令行帮助

### 示例

1. 将 YAML 格式转换为 SSH 配置格式:

```bash
ssh-config -to-ssh -src input.yaml -dest output.conf
```

2. 将 SSH 配置格式转换为 JSON 格式:

```bash
ssh-config -to-json -src ~/.ssh/config -dest output.json
```

3. 从标准输入读取，输出到标准输出，并以 YAML 格式保存:

```bash
cat input.conf | ssh-config -to-yaml > output.yaml
```

## 开发

### 依赖

- Go 1.23+

### 构建

```bash
go build
```

### 测试

```bash
go test -v ./... -covermode=atomic -coverprofile=coverage.out && go tool cover -html=coverage.out -o coverage.html
```

## 贡献

欢迎提交 issues 和 pull requests。

## 许可证

本项目采用 Apache 许可证。详见 [LICENSE](./LICENSE) 文件。

# 使用教程

- [使用结构化数据管理 SSH 配置：SSH Config Tool](https://soulteary.com/2024/10/15/manage-ssh-configuration-using-structure-data-ssh-config-tool.html)

# 感谢

- 好用的 OpenSSH 软件包
  - https://man.openbsd.org/ssh_config
- 颇受启发的配置文件
  - https://github.com/bencromwell/sshush
