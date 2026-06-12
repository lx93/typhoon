# Desktop CLI Client

The desktop client MVP is a command-line end-user client. It fetches volunteer
relay candidates from the broker, selects the first usable direct-exit VLESS
Reality Vision relay, generates a sing-box TUN config, and runs sing-box to
route device traffic through that relay.

The first operational target is macOS, but the implementation is in Go so the
broker client, relay selection, sing-box config generation, and process runner
can be reused by Linux and Windows clients later.

## Requirements

- A running Typhoon broker.
- At least one registered volunteer relay.
- A local `sing-box` binary.
- macOS privileges for TUN routing. In practice, run `connect` with `sudo`.

Install sing-box with Homebrew if needed:

```sh
brew install sing-box
```

## Check Relay Selection

```sh
go run ./cmd/client check -broker http://localhost:8080
```

This fetches candidates from:

```http
GET /api/v1/relays?limit=5
```

Then it prints the selected usable relay.

## Generate Config Only

```sh
go run ./cmd/client config \
  -broker http://localhost:8080 \
  -out typhoon-sing-box.json
```

The generated config uses:

- `tun` inbound.
- `auto_route: true`.
- `strict_route: true`.
- DNS servers detoured through the proxy.
- VLESS Reality Vision outbound from the selected relay descriptor.
- Route final set to the proxy outbound.

## Connect on macOS

```sh
sudo go run ./cmd/client connect \
  -broker http://localhost:8080 \
  -sing-box /opt/homebrew/bin/sing-box
```

The client writes a temporary sing-box config, prints the chosen relay, and then
runs:

```sh
sing-box run -c <generated-config>
```

Press `Ctrl-C` to stop. The client forwards the interrupt to sing-box and removes
the temporary config file.

## DNS Troubleshooting

If a raw IP address works through the tunnel but hostnames do not resolve, keep
the generated config on disk and force public DNS resolvers through the proxy:

```sh
sudo go run ./cmd/client connect \
  -broker http://localhost:8080 \
  -sing-box /opt/homebrew/bin/sing-box \
  -dns 1.1.1.1,8.8.8.8 \
  -config-out typhoon-sing-box-debug.json
```

Then confirm the saved config contains:

- DNS servers with `type: tcp`.
- DNS servers with `detour: proxy`.
- A route rule with `protocol: dns` and `action: hijack-dns`.
- `default_domain_resolver: dns-0`.

After connecting, test name resolution with:

```sh
nslookup facebook.com
curl -I https://facebook.com
```

## Reuse Notes

Linux should be able to reuse most of the Go code directly. Windows should reuse
the broker, relay selection, config generation, and command contract, but may
need additional install checks around the Windows tunnel driver used by sing-box.
