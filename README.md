# Typhoon

Typhoon is a Snowflake-like volunteer relay network for helping users reach blocked public websites and apps through temporary volunteers outside China.

The MVP is deliberately narrow:

- Volunteers run a desktop command-line app.
- Volunteers must expose a publicly reachable port, with IPv6 preferred when available.
- Volunteers act as direct exit nodes.
- The broker is a control plane only and does not proxy user traffic.
- China mobile clients proxy all device traffic through VPN mode.
- Relay transport uses Xray-core's VLESS + Reality + Vision support.

Future versions should add dedicated exit servers so volunteers can choose to act as entry relays instead of direct exits.

## Repository Layout

```text
cmd/broker/          Broker HTTP API.
cmd/volunteer/       Volunteer CLI for Xray-backed relay registration.
docs/                Architecture, API, mobile, and rollout notes.
ios/                 iOS VPN client scaffold and shared Swift package.
internal/broker/     Broker store and HTTP handlers.
internal/relay/      Shared relay descriptor models.
internal/volunteer/  Xray config generation helpers.
```

## Quick Start

Start the broker:

```sh
go run ./cmd/broker -addr :8080
```

Run a volunteer relay:

```sh
go run ./cmd/volunteer \
  -broker http://localhost:8080 \
  -public-port 443 \
  -listen-port 443 \
  -xray /path/to/xray
```

By default, the volunteer listens on IPv6 (`::`) and advertises the first global IPv6 address it can find. Pass `-public-host` and, if needed, `-listen-host` when using a DNS name, IPv4 address, or manually chosen IPv6 address. A global IPv6 address still needs inbound firewall/router rules that allow clients to reach the volunteer port.

The volunteer prints colored client connection events by default: green when a client TCP connection opens, red when it closes, plus the client IP, duration, and byte counts. Pass `-connection-log=false` to let Xray bind the public port directly without the observer.

The volunteer command expects an `xray` binary that supports `xray x25519` and `xray run -config`.

List client relay candidates:

```sh
curl http://localhost:8080/api/v1/relays
```

Run the desktop CLI client relay check:

```sh
go run ./cmd/client check -broker http://localhost:8080
```

For the macOS full-device routing MVP, see `docs/desktop-client.md`.

## MVP Warning

In the MVP, volunteers are direct exits. Their public IP can appear to destination websites and apps. That is simple and useful for early testing, but it creates real legal, abuse, and privacy risk for volunteers. The first public rollout should add abuse controls, rate limits, exit policy controls, and preferably dedicated exit servers.
