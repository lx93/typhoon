# Typhoon iOS MVP

This directory contains the starter iPhone client for routing device traffic through a Typhoon volunteer relay.

## What is included

- `TyphoonKit`: shared Swift package for broker API models, relay filtering, and proxy-engine configuration.
- `TyphoonClient/App`: SwiftUI app shell for VPN permission, broker URL entry, and connect/disconnect.
- `TyphoonClient/PacketTunnel`: `NEPacketTunnelProvider` that fetches relays and starts the embedded sing-box/libbox proxy engine.
- `ThirdParty/README.md`: local instructions for rebuilding the ignored `Libbox.xcframework` artifact.
- `project.yml`: XcodeGen project spec for creating an iOS app target plus Packet Tunnel extension target.

## Current MVP boundary

The app, packet tunnel, relay selection, sing-box configuration, and Libbox-backed `EmbeddedProxyEngine` adapter are in place. The checked-in source expects a local generated artifact at:

```text
ios/ThirdParty/Libbox.xcframework
```

That artifact is intentionally ignored by git because it is large. See `ios/ThirdParty/README.md` to rebuild it from sing-box.

The intended data path is:

```text
iPhone apps
  -> NEPacketTunnelProvider
  -> PacketTunnelProxyEngine
  -> VLESS Reality Vision connection
  -> volunteer public host:port
  -> destination internet
```

The code now compiles for simulator and generic iPhone targets. Real traffic routing still needs validation on a signed physical iPhone with the Network Extension packet-tunnel entitlement.

## Simulator behavior

The iOS simulator can run the SwiftUI shell and use **Check Broker** to fetch volunteer relay descriptors. It should not be used for the **Connect** path. Installing or starting a Packet Tunnel VPN profile depends on Apple's Network Extension preferences service and normally fails in the simulator with:

```text
NEConfigurationErrorDomain Code=11 "IPC failed"
```

Use a physical iPhone signed with a developer team that has the Network Extension packet-tunnel entitlement for end-to-end VPN testing.

## Create the Xcode project

Install Xcode and select it as the active developer directory:

```sh
sudo xcodebuild -license
sudo xcode-select -s /Applications/Xcode.app/Contents/Developer
```

If using XcodeGen:

```sh
brew install xcodegen
cd ios
xcodegen generate
open Typhoon.xcodeproj
```

Before building the Packet Tunnel target with the embedded engine, follow `ThirdParty/README.md` to create the local Libbox framework.

If creating the project manually in Xcode:

1. Create an iOS app target named `TyphoonClient`.
2. Add a Packet Tunnel extension target named `PacketTunnel`.
3. Add the local package at `ios/TyphoonKit` to both targets.
4. Add `TyphoonClient/App` and `TyphoonClient/Shared` to the app target.
5. Add `TyphoonClient/PacketTunnel` and `TyphoonClient/Shared` to the extension target.
6. Use the included `Info.plist` and entitlements files for each target.
7. Link `ThirdParty/Libbox.xcframework`, `UIKit.framework`, and `libresolv.tbd` to the Packet Tunnel extension target.

## Required Apple capabilities

The Packet Tunnel target needs:

- Network Extensions capability with `packet-tunnel-provider`.
- App Group matching `group.com.typhoon.client`.

The app target needs:

- App Group matching `group.com.typhoon.client`.

You will need an Apple developer account/team that can sign NetworkExtension packet tunnel apps on device.

If a physical iPhone shows `NEConfigurationErrorDomain Code=11`, check that both the app and Packet Tunnel extension are signed with:

- Network Extensions capability with `packet-tunnel-provider`.
- App Group matching `group.com.typhoon.client`.
- Bundle identifiers matching `com.typhoon.client` and `com.typhoon.client.PacketTunnel`, or update `AppConfig.packetTunnelBundleIdentifier` if you change them.

## Broker and relay flow

The Packet Tunnel extension reads the broker URL from `NETunnelProviderProtocol.providerConfiguration`, calls:

```http
GET /api/v1/relays?limit=5
```

Then it selects the first relay that is:

- `protocol == "vless-reality-vision"`
- `flow == "xtls-rprx-vision"`
- `exit_mode == "direct"`
- not expired according to broker `server_time`
- complete enough to configure a Reality client

## Embedded engine

`EmbeddedProxyEngine` uses sing-box's generated `Libbox.xcframework` and starts a `tun` inbound with a VLESS Reality Vision outbound generated from the broker relay descriptor.

The adapter:

- Generates VLESS Reality Vision client JSON from `SingBoxConfiguration`.
- Lets libbox install and own the packet tunnel network settings.
- Returns the packet tunnel file descriptor to libbox.
- Stops the service cleanly when the VPN disconnects.
- Returns startup errors so the provider can try the next relay or fail closed.
