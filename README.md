# sing-box-proxyctl

`sing-box-proxyctl` is a Linux console tool for managing sing-box subscription configs, nodes, latency checks, node switching, and a proxyctl-managed sing-box background process.

中文说明见下方：[中文](#中文)

## English

### What It Does

- Pull a proxy subscription and generate a sing-box config. (Currently supports Vmess and Shadowsocks protocols.)
- List nodes in a readable terminal table.
- Ping nodes concurrently and cache latency/region results.
- Switch the active node and restart sing-box automatically.
- Start, stop, restart, and inspect the managed sing-box process.

This tool does not download or install sing-box core. Install sing-box yourself, or place a local core at:

```bash
./bin/sing-box/sing-box
```

### Install

Download a Linux release asset from GitHub Releases.
Example:

```bash
tar -xzf proxyctl-linux-amd64.tar.gz
chmod +x proxyctl-linux-amd64
sudo mv proxyctl-linux-amd64 /usr/local/bin/proxyctl
```

Check the binary:

```bash
proxyctl version
proxyctl core check
```

### Quick Start

Generate a sing-box config from your subscription:

```bash
proxyctl sub update '<subscription-url>'
```

Use a mixed inbound if you want both SOCKS and HTTP-style proxy environment variables to work through one port:

```bash
proxyctl sub update '<subscription-url>' --inbound mixed --mixed-port 2080
```

Start sing-box:

```bash
proxyctl core start
proxyctl core ps
```

List nodes:

```bash
proxyctl node list
```

Ping nodes:

```bash
proxyctl node ping
proxyctl node ping 0
```

Switch to a node:

```bash
proxyctl use 0
```

`use <id>` updates `route.final` and restarts the managed sing-box process by default.

### Shell Proxy Environment

For `--inbound mixed --mixed-port 2080`:

```bash
export ALL_PROXY=socks5h://127.0.0.1:2080
export HTTP_PROXY=http://127.0.0.1:2080
export HTTPS_PROXY=http://127.0.0.1:2080
```

For `--inbound both --socks-port 1080 --http-port 8080`:

```bash
export ALL_PROXY=socks5h://127.0.0.1:1080
export HTTP_PROXY=http://127.0.0.1:8080
export HTTPS_PROXY=http://127.0.0.1:8080
```

### Commands

Subscription:

```bash
proxyctl sub update '<subscription-url>'
proxyctl sub update '<subscription-url>' -o ./config.json
proxyctl sub update '<subscription-url>' --inbound socks
proxyctl sub update '<subscription-url>' --inbound http
proxyctl sub update '<subscription-url>' --inbound mixed
proxyctl sub update '<subscription-url>' --inbound both
```

Nodes:

```bash
proxyctl node list
proxyctl node ping
proxyctl node ping 0
proxyctl use 0
proxyctl use 0 --no-restart
```

Core process:

```bash
proxyctl core check
proxyctl core status
proxyctl core ps
proxyctl core start
proxyctl core stop
proxyctl core restart
```

Useful flags:

```bash
proxyctl core start -c ./config.json
proxyctl core start --core /usr/local/bin/sing-box
proxyctl core start --pid-file ./proxyctl.pid --log-file ./proxyctl.log
```

### Generated Files

By default, files are created next to `config.json`:

- `config.json`: generated sing-box config
- `.proxyctl-settings.json`: persisted inbound settings
- `.proxyctl-ping.json`: cached latency and region results
- `.proxyctl-sing-box.pid`: managed sing-box pid file
- `.proxyctl-sing-box.log`: managed sing-box log file

### Supported Protocols

- VMess
- Shadowsocks

### Error Handling

Commands print terminal errors and return a non-zero exit code when required operations fail, including subscription fetch errors, invalid configs, missing nodes, invalid node ids, missing sing-box core, and sing-box startup failures such as port binding errors.

## 中文

### 这个工具做什么

`sing-box-proxyctl` 是一个面向 Linux 终端的 sing-box 订阅和节点管理工具。

它可以：

- 拉取订阅并生成 sing-box 配置文件。（目前支持Vmess和Shadowsocks协议）
- 用表格列出节点。
- 并发测速，并缓存延迟和地区结果。
- 切换当前使用的节点，默认重启 sing-box。
- 启动、停止、重启和查看 proxyctl 管理的 sing-box 后台进程。

这个工具不会自动下载或安装 sing-box core。你需要自己安装 `sing-box`，或者把本地 core 放到：

```bash
./bin/sing-box/sing-box
```

### 安装

从 GitHub Releases 下载对应 Linux 架构的压缩包：

- `proxyctl-linux-amd64.tar.gz`
- `proxyctl-linux-arm64.tar.gz`

示例：

```bash
tar -xzf proxyctl-linux-amd64.tar.gz
chmod +x proxyctl-linux-amd64
sudo mv proxyctl-linux-amd64 /usr/local/bin/proxyctl
```

检查安装结果：

```bash
proxyctl version
proxyctl core check
```

### 快速开始

用订阅生成 sing-box 配置：

```bash
proxyctl sub update '<订阅地址>'
```

如果希望命令行工具和脚本更容易走代理，推荐使用 mixed inbound：

```bash
proxyctl sub update '<订阅地址>' --inbound mixed --mixed-port 2080
```

启动 sing-box：

```bash
proxyctl core start
proxyctl core ps
```

查看节点：

```bash
proxyctl node list
```

节点测速：

```bash
proxyctl node ping
proxyctl node ping 0
```

切换节点：

```bash
proxyctl use 0
```

`use <id>` 会更新 `route.final`，并默认重启由 proxyctl 管理的 sing-box 进程。

### 终端代理环境变量

如果使用 `--inbound mixed --mixed-port 2080`：

```bash
export ALL_PROXY=socks5h://127.0.0.1:2080
export HTTP_PROXY=http://127.0.0.1:2080
export HTTPS_PROXY=http://127.0.0.1:2080
```

如果使用 `--inbound both --socks-port 1080 --http-port 8080`：

```bash
export ALL_PROXY=socks5h://127.0.0.1:1080
export HTTP_PROXY=http://127.0.0.1:8080
export HTTPS_PROXY=http://127.0.0.1:8080
```

### 常用命令

订阅：

```bash
proxyctl sub update '<订阅地址>'
proxyctl sub update '<订阅地址>' -o ./config.json
proxyctl sub update '<订阅地址>' --inbound socks
proxyctl sub update '<订阅地址>' --inbound http
proxyctl sub update '<订阅地址>' --inbound mixed
proxyctl sub update '<订阅地址>' --inbound both
```

节点：

```bash
proxyctl node list
proxyctl node ping
proxyctl node ping 0
proxyctl use 0
proxyctl use 0 --no-restart
```

sing-box 后台进程：

```bash
proxyctl core check
proxyctl core status
proxyctl core ps
proxyctl core start
proxyctl core stop
proxyctl core restart
```

常用参数：

```bash
proxyctl core start -c ./config.json
proxyctl core start --core /usr/local/bin/sing-box
proxyctl core start --pid-file ./proxyctl.pid --log-file ./proxyctl.log
```

### 生成的文件

默认情况下，以下文件会生成在 `config.json` 同级目录：

- `config.json`：生成的 sing-box 配置
- `.proxyctl-settings.json`：持久化的 inbound 设置
- `.proxyctl-ping.json`：节点延迟和地区缓存
- `.proxyctl-sing-box.pid`：proxyctl 管理的 sing-box pid 文件
- `.proxyctl-sing-box.log`：proxyctl 管理的 sing-box 日志文件

### 支持的协议

- VMess
- Shadowsocks

### 错误处理

当订阅拉取失败、配置无效、没有节点、节点 id 错误、sing-box core 不存在、sing-box 启动失败或端口占用时，命令会在终端输出错误并返回非 0 退出码。
