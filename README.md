# proxyctl

`proxyctl` is a Linux terminal CLI tool for managing proxy subscriptions and sing-box nodes.

Current stack:

- Go
- Cobra
- pterm
- sing-box config output

## Features

### Subscription Update

Pull a proxy subscription, parse supported nodes, and generate a sing-box config file.

```bash
go run . sub update <subscription-url>
go run . sub update <subscription-url> -o ./config.json
go run . sub update <subscription-url> --inbound mixed --mixed-port 2080
go run . sub update <subscription-url> --inbound both --socks-port 1080 --http-port 8080
```

Implemented protocol parsing:

- VMess: `vmess://...`
- Shadowsocks: `ss://...`

Generated config includes:

- Configurable inbound on `127.0.0.1`
- sing-box outbounds for parsed nodes
- `route.final` pointing to the default node

Inbound modes:

- `socks`: SOCKS inbound only, default `127.0.0.1:1080`
- `http`: HTTP inbound only, default `127.0.0.1:8080`
- `mixed`: sing-box mixed inbound, default `127.0.0.1:2080`
- `both`: SOCKS and HTTP inbounds

For terminal package managers and shell scripts, `mixed` or `both` is usually more convenient than SOCKS only.

Inbound options are saved to `.proxyctl-settings.json` next to the generated config. Future `sub update` runs reuse the last saved inbound settings unless flags are passed again.

Example shell proxy environment:

```bash
# --inbound both
export ALL_PROXY=socks5h://127.0.0.1:1080
export HTTP_PROXY=http://127.0.0.1:8080
export HTTPS_PROXY=http://127.0.0.1:8080

# --inbound mixed
export ALL_PROXY=socks5h://127.0.0.1:2080
export HTTP_PROXY=http://127.0.0.1:2080
export HTTPS_PROXY=http://127.0.0.1:2080
```

### Node List

Read nodes from the sing-box config and render them in a terminal table.

```bash
go run . node list
go run . node list -c ./config.json
```

Columns:

- `id`: node index, used by `use <id>` and `node ping <id>`
- `use`: `*` marks the current `route.final` node
- `type`: outbound type, such as `vmess` or `shadowsocks`
- `name`: friendly node name derived from the subscription item
- `region`: last known server country from ping cache
- `latency`: last TCP ping latency
- `status`: last ping status
- `transport`: `tcp`, `ws`, etc.
- `tls`: `on` or `off`

Optional ping cache path:

```bash
go run . node list --ping-cache ./.proxyctl-ping.json
```

### Node Ping

Run concurrent TCP latency checks for nodes and save results for `node list`.

```bash
go run . node ping
go run . node ping 0
go run . node ping -j 16 -t 2s
```

Flags:

- `-c, --config`: sing-box config path, default `./config.json`
- `--ping-cache`: ping result cache path
- `-j, --concurrency`: max concurrent workers, default `8`
- `-t, --timeout`: TCP ping and region lookup timeout, default `3s`

Ping results are saved to `.proxyctl-ping.json` by default, next to the config file.

Region lookup stores only the country name. If lookup fails, the region is `unknown`.

### Node Use

Switch the active node by updating `route.final` in the sing-box config, then restart the managed sing-box process.

```bash
go run . use <id>
go run . use 0
go run . use 0 -c ./config.json
go run . use 0 --no-restart
```

The `id` is the 0-based node index shown by `node list`.

Use `--no-restart` if you only want to update the config file.

Restart behavior:

- Uses `sing-box` from `$PATH`, or `./bin/sing-box/sing-box` if no PATH binary exists
- Stops only the process recorded in `.proxyctl-sing-box.pid`
- Starts sing-box with `run -c <config>`
- Writes logs to `.proxyctl-sing-box.log`

Optional restart flags:

- `--core`: custom sing-box core path
- `--pid-file`: custom managed pid file
- `--log-file`: custom log file

### Version

Print build and platform information.

```bash
go run . version
go run . version --verbose
```

Release builds can inject version metadata with ldflags:

```bash
go build -ldflags "-X github.com/Alex-J4096/proxyctl/cmd.Version=1.0.0 -X github.com/Alex-J4096/proxyctl/cmd.GitCommit=$(git rev-parse --short HEAD) -X github.com/Alex-J4096/proxyctl/cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

### Core Management

Check sing-box core availability and manage the proxyctl-managed sing-box background process.

```bash
go run . core check
go run . core status
go run . core ps
go run . core start
go run . core stop
go run . core restart
go run . core restart -c ./config.json
```

Detection checks:

- `sing-box` from `$PATH`
- Local binary at `./bin/sing-box/sing-box`

This command only reports the current core status. It does not download, install, or modify sing-box.

Process management commands:

- `core start`: start sing-box in the background with `run -c <config>`
- `core stop`: stop the process recorded in `.proxyctl-sing-box.pid`
- `core restart`: stop then start the managed process
- `core ps`: show the managed background process status
- `core status`: show both core availability and managed process status

Shared flags:

- `-c, --config`: sing-box config path, default `./config.json`
- `--core`: custom sing-box core path
- `--pid-file`: custom managed pid file
- `--log-file`: custom log file

`use <id>` uses the same managed process strategy after switching nodes.

## Files

- `config.json`: generated sing-box config
- `.proxyctl-ping.json`: cached node latency and region results
- `.proxyctl-settings.json`: persisted proxyctl settings, including inbound preferences
- `.proxyctl-sing-box.pid`: pid file for the proxyctl-managed sing-box process
- `.proxyctl-sing-box.log`: log file for the proxyctl-managed sing-box process

## Error Handling

Commands print terminal errors and exit with a non-zero status when required operations fail.

Handled failure cases include:

- Subscription fetch failure or non-2xx HTTP response
- Empty subscription response
- Invalid supported node links
- No supported nodes found
- Missing or invalid config file
- Invalid node id
- Missing or non-executable sing-box core
- sing-box startup failure, including port binding errors reported by the core

## Development

Run tests:

```bash
go test ./...
```

If the Go build cache is not writable in the current environment:

```bash
GOCACHE=/private/tmp/proxyctl-gocache go test ./...
```

## TODO

- Add `node ping --sort` or list sorting by latency.
- Add a persistent app config directory, for example `~/.config/proxyctl`.
- Store subscription URL and support `sub update` without passing the URL every time.
- Add `sub info` to show subscription metadata if available.
- Add better validation for generated sing-box config.
- Support more protocols, such as Trojan, VLESS, and Hysteria2.
- Support multiple inbounds and custom listen ports.
- Add tests for link parsing, config generation, ping cache loading, and command behavior.
