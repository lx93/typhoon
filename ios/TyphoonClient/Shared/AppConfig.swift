import Foundation

enum AppConfig {
    static let vpnProfileName = "Typhoon Volunteer VPN"
    static let appGroupIdentifier = "group.com.typhoon.client"
    static let packetTunnelBundleIdentifier = "com.typhoon.client.PacketTunnel"
    static let providerBrokerURLKey = "broker_url"
    static let defaultBrokerURL = URL(string: "http://127.0.0.1:8080")!
}
