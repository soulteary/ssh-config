# SSH Config Tool

[![codecov](https://codecov.io/gh/soulteary/ssh-config/branch/main/graph/badge.svg?token=W816DX12V8)](https://codecov.io/gh/soulteary/ssh-config) [![Go Report Card](https://goreportcard.com/badge/github.com/soulteary/ssh-config)](https://goreportcard.com/report/github.com/soulteary/ssh-config) [![CodeQL Advanced](https://github.com/soulteary/ssh-config/actions/workflows/codeql.yml/badge.svg)](https://github.com/soulteary/ssh-config/actions/workflows/codeql.yml) [![Release](https://github.com/soulteary/ssh-config/actions/workflows/build.yml/badge.svg)](https://github.com/soulteary/ssh-config/actions/workflows/build.yml)

**[English Docs](./README.md)**

<img src=".github/github-repo-card.png" >

SSH Config Tool 是一个用于管理 SSH 配置文件的命令行工具。它允许你使用更具表现力的 YAML/JSON 格式来管理你的 SSH 配置文件。

## 特性

- 支持从 YAML/JSON 格式转换为标准 SSH 配置格式
- 默认使用无损 v3 格式转换，保留注释、顺序、重复指令、引号、换行符和未知指令
- 通过 `-legacy` 保留原有 map 格式转换和目录扫描能力
- 支持从文件输入或标准输入(stdin)读取配置
- 支持输出到文件或标准输出(stdout)
- 自动检测输入格式(YAML/JSON/SSH Config)

## 安装

使用 Docker 或者从 [GitHub 发布页面](https://github.com/soulteary/ssh-config/releases)下载合适你的系统、CPU 架构的二进制文件即可。

也可以选择使用 Homebrew 进行安装。

```bash
brew tap soulteary/tap
brew install soulteary/tap/ssh-config
```

Go 用户可以直接安装 v3 命令：

```bash
go install github.com/soulteary/ssh-config/v3@latest
```

## 使用方法

### 基本用法

```bash
ssh-config [options]
```

不带参数运行时，程序读取 `~/.ssh/config`，并将无损 v3 YAML 输出到标准输出：

```bash
ssh-config
```

需要指定输入和输出文件时，请使用 `-src` 和 `-dest`：

```bash
ssh-config -to-yaml -src input_file -dest output_file
```

或，使用 Linux 管道来操作文件：

```bash
cat input_file | ssh-config -to-yaml > output_file
```

### Docker

下载镜像

```bash
docker pull soulteary/ssh-config:latest
# or
docker pull ghcr.io/soulteary/ssh-config:latest
```

将当前目录的配置文件转换并保存为新的文件：

```bash
docker run --rm -it -v `pwd`:/ssh soulteary/ssh-config:latest ssh-config -to-yaml -src /ssh/test.yaml -dest /ssh/abc.yaml
```

如果你只想看看转换结果：

```bash
docker run --rm -it -v `pwd`:/ssh soulteary/ssh-config:latest ssh-config -to-yaml -src /ssh/test.yaml
```

如果你想使用 Linux 管道来操作文件，可以先进入 Docker 交互式命令行：

```bash
docker run --rm -it -v `pwd`:/ssh soulteary/ssh-config:latest bash
cat /ssh/test.yaml | ssh-config -to-yaml
```

### 选项

- `-to-yaml, -to-json, -to-ssh`: 指定输出格式 (yaml/json/config)，同一时间，输出格式只能指定为一种。
- `-src`: 指定输入文件；省略时无损模式读取 `~/.ssh/config`，旧模式扫描 `~/.ssh`
- `-dest`: 指定要保存的配置文件路径；父目录必须已存在，省略时将转换结果写入标准输出
- `-legacy`: 使用原有的有损 map/array 格式，并启用目录扫描。
- `-help`: 查看程序命令行帮助
- `-version`: 输出发布版本、提交、构建时间和工作树状态

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

4. 通过 v3 YAML 格式无损编辑配置：

```bash
ssh-config -to-yaml -src ~/.ssh/config -dest config.v3.yaml
# 修改 config.v3.yaml 中的 directive 字段；未修改的行会保留原始字节。
ssh-config -to-ssh -src config.v3.yaml -dest ~/.ssh/config
```

程序默认读取原有 YAML/JSON 格式并迁移为 v3 Schema。只有下游仍依赖旧 map/array 输出时才需要使用 `-legacy`；旧文档中已经丢失的重复值和指令顺序无法恢复。

字段结构、字节保持规则、编辑行为、Go API 示例和旧格式迁移边界详见 [v3 无损格式规范](./docs/lossless-schema-v3.md)。
升级脚本和 Go 导入路径前，请阅读 [v2 到 v3 迁移指南](./docs/migration-v3.md)。

## 开发

### 依赖

- Go 1.27+

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

安全漏洞请按照 [SECURITY.md](./SECURITY.md) 中的私密流程报告，不要直接提交公开 issue。

## 许可证

本项目采用 Apache 许可证。详见 [LICENSE](./LICENSE) 文件。

# 使用教程

- [使用结构化数据管理 SSH 配置：SSH Config Tool](https://soulteary.com/2024/10/15/manage-ssh-configuration-using-structure-data-ssh-config-tool.html)

# 感谢

- 好用的 OpenSSH 软件包
  - https://man.openbsd.org/ssh_config
- 颇受启发的配置文件
  - https://github.com/bencromwell/sshush
