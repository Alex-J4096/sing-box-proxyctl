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
```

Implemented protocol parsing:

- VMess: `vmess://...`
- Shadowsocks: `ss://...`

Generated config includes:

- SOCKS inbound on `127.0.0.1:1080`
- sing-box outbounds for parsed nodes
- `route.final` pointing to the default node

### Node List

Read nodes from the sing-box config and render them in a terminal table.

```bash
go run . node list
go run . node list -c ./config.json
```

Columns:

- `id`: node index, used by future `use <id>` and current `node ping <id>`
- `use`: `*` marks the current `route.final` node
- `type`: outbound type, such as `vmess` or `shadowsocks`
- `tag`: sing-box outbound tag
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

## Files

- `config.json`: generated sing-box config
- `.proxyctl-ping.json`: cached node latency and region results

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

- Implement `use <id>` to switch `route.final` to a selected node.
- Add `node ping --sort` or list sorting by latency.
- Add a persistent app config directory, for example `~/.config/proxyctl`.
- Store subscription URL and support `sub update` without passing the URL every time.
- Add `sub info` to show subscription metadata if available.
- Add better validation for generated sing-box config.
- Support more protocols, such as Trojan, VLESS, and Hysteria2.
- Support multiple inbounds and custom listen ports.
- Add tests for link parsing, config generation, ping cache loading, and command behavior.
