import Foundation

enum AppConfig {
    static let vpnProfileName = "Typhoon Volunteer VPN"
    static let appGroupIdentifier = "group.com.typhoon.client"
    static let packetTunnelBundleIdentifier = "com.typhoon.client.PacketTunnel"
    static let providerBrokerURLKey = "broker_url"
    static let defaultBrokerURL = URL(string: "http://54.238.185.205:8080/")!
}
