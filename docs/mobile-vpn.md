# Mobile VPN Client Notes

The mobile client should proxy all device traffic through VPN mode.

## Android

Use `VpnService` to create a TUN interface. A userspace network bridge then forwards traffic from the TUN interface into the VLESS Reality client engine.

Likely implementation choices:

- Android `VpnService` for VPN lifecycle.
- A tun2socks-style bridge for TCP/UDP packets.
- An Xray-compatible client core for VLESS Reality Vision.

## iOS

Use `NetworkExtension` with a packet tunnel provider.

Likely implementation choices:

- `NEPacketTunnelProvider` for VPN lifecycle.
- A userspace bridge from packet tunnel flow to the VLESS Reality client engine.
- An Xray-compatible client core packaged in a way App Store distribution can support.

## Relay Selection

The client should request several relay candidates and attempt them in order.

The MVP can select by:

- Not expired.
- Protocol match.
- Direct exit mode.
- Broker order.

Later selection should consider:

- Latency.
- Recent success rate.
- Capacity.
- Geographic diversity.
- Volunteer reputation.

## Client Safety

The client should treat volunteers like untrusted network providers:

- Prefer HTTPS destinations.
- Avoid leaking broker tokens to relays.
- Rotate relay credentials.
- Fail closed when no relay is available.
- Surface clear status when traffic is routed through a volunteer.
