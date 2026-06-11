import Foundation

public struct SingBoxConfiguration: Equatable, Sendable {
    public let relay: RelayDescriptor
    public let tunnelIPv4Address: String
    public let tunnelIPv6Address: String
    public let dnsServers: [String]
    public let mtu: Int

    public init(
        relay: RelayDescriptor,
        tunnelIPv4Address: String = "172.19.0.1/30",
        tunnelIPv6Address: String = "fdfe:dcba:9876::1/126",
        dnsServers: [String] = ["1.1.1.1", "8.8.8.8"],
        mtu: Int = 1500
    ) {
        self.relay = relay
        self.tunnelIPv4Address = tunnelIPv4Address
        self.tunnelIPv6Address = tunnelIPv6Address
        self.dnsServers = dnsServers
        self.mtu = mtu
    }

    public func encodedJSON() throws -> Data {
        try JSONSerialization.data(
            withJSONObject: makeJSONObject(),
            options: [.prettyPrinted, .sortedKeys]
        )
    }

    public func encodedJSONString() throws -> String {
        String(decoding: try encodedJSON(), as: UTF8.self)
    }

    public func makeJSONObject() -> [String: Any] {
        [
            "log": [
                "level": "info",
                "timestamp": true
            ],
            "dns": [
                "servers": dnsServers.enumerated().map { index, server in
                    [
                        "tag": "dns-\(index)",
                        "type": "udp",
                        "server": server,
                        "detour": "proxy"
                    ]
                },
                "final": "dns-0"
            ],
            "inbounds": [
                [
                    "type": "tun",
                    "tag": "tun-in",
                    "address": [
                        tunnelIPv4Address,
                        tunnelIPv6Address
                    ],
                    "mtu": mtu,
                    "auto_route": true,
                    "strict_route": true,
                    "stack": "system",
                    "dns_mode": "hijack",
                    "endpoint_independent_nat": true
                ] as [String: Any]
            ],
            "outbounds": [
                [
                    "type": "vless",
                    "tag": "proxy",
                    "server": relay.publicHost,
                    "server_port": relay.publicPort,
                    "uuid": relay.clientID,
                    "flow": relay.flow,
                    "network": "tcp",
                    "packet_encoding": "xudp",
                    "tls": [
                        "enabled": true,
                        "server_name": relay.serverName,
                        "utls": [
                            "enabled": true,
                            "fingerprint": "chrome"
                        ],
                        "reality": [
                            "enabled": true,
                            "public_key": relay.realityPublicKey,
                            "short_id": relay.shortID
                        ]
                    ] as [String: Any]
                ] as [String: Any],
                [
                    "type": "direct",
                    "tag": "direct"
                ],
                [
                    "type": "block",
                    "tag": "block"
                ]
            ],
            "route": [
                "auto_detect_interface": true,
                "final": "proxy"
            ]
        ]
    }
}
