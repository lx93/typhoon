import Foundation
import NetworkExtension
import TyphoonKit

@MainActor
final class VPNController: ObservableObject {
    @Published private(set) var status: NEVPNStatus = .invalid
    @Published private(set) var isWorking = false
    @Published private(set) var lastError: String?
    @Published private(set) var selectedRelayLabel: String?
    @Published private(set) var brokerSummary: String?
    @Published private(set) var brokerRelays: [RelayDescriptor] = []

    private var manager: NETunnelProviderManager?
    private let relaySelector = RelaySelector()

    var isConnected: Bool {
        status == .connected || status == .connecting || status == .reasserting
    }

    var statusText: String {
        switch status {
        case .invalid:
            return "Not configured"
        case .disconnected:
            return "Disconnected"
        case .connecting:
            return "Connecting"
        case .connected:
            return "Connected"
        case .reasserting:
            return "Reconnecting"
        case .disconnecting:
            return "Disconnecting"
        @unknown default:
            return "Unknown"
        }
    }

    init() {
        NotificationCenter.default.addObserver(
            forName: .NEVPNStatusDidChange,
            object: nil,
            queue: .main
        ) { [weak self] _ in
            Task { @MainActor in
                self?.refreshStatus()
            }
        }
    }

    func load() async {
        isWorking = true
        defer { isWorking = false }

        do {
            guard Self.canUseNetworkExtensionPreferences else {
                status = .invalid
                return
            }

            manager = try await loadOrCreateManager()
            refreshStatus()
        } catch {
            lastError = VPNControllerError.message(for: error)
        }
    }

    func connect(brokerURLText: String) async {
        isWorking = true
        defer { isWorking = false }

        do {
            guard let brokerURL = URL(string: brokerURLText), brokerURL.scheme != nil else {
                throw VPNControllerError.invalidBrokerURL
            }
            guard Self.canUseNetworkExtensionPreferences else {
                throw VPNControllerError.networkExtensionUnavailableInSimulator
            }

            let manager = try await loadOrCreateManager()
            try await configure(manager: manager, brokerURL: brokerURL)
            try manager.connection.startVPNTunnel()
            self.manager = manager
            selectedRelayLabel = nil
            refreshStatus()
        } catch {
            lastError = VPNControllerError.message(for: error)
        }
    }

    func checkBroker(brokerURLText: String) async {
        isWorking = true
        defer { isWorking = false }

        do {
            guard let brokerURL = URL(string: brokerURLText), brokerURL.scheme != nil else {
                throw VPNControllerError.invalidBrokerURL
            }

            let response = try await BrokerClient(baseURL: brokerURL).listRelays(limit: 10)
            let usableRelays = relaySelector.orderedCandidates(from: response.relays, now: response.serverTime)

            brokerRelays = response.relays
            brokerSummary = "\(response.count) relays returned, \(usableRelays.count) usable for this MVP."
            lastError = nil
        } catch {
            brokerRelays = []
            brokerSummary = nil
            lastError = VPNControllerError.message(for: error)
        }
    }

    func disconnect() async {
        manager?.connection.stopVPNTunnel()
        refreshStatus()
    }

    private func loadOrCreateManager() async throws -> NETunnelProviderManager {
        let managers = try await NETunnelProviderManager.loadAllFromPreferences()
        if let existing = managers.first(where: { $0.localizedDescription == AppConfig.vpnProfileName }) {
            return existing
        }

        let manager = NETunnelProviderManager()
        manager.localizedDescription = AppConfig.vpnProfileName
        return manager
    }

    private func configure(manager: NETunnelProviderManager, brokerURL: URL) async throws {
        let tunnelProtocol = NETunnelProviderProtocol()
        tunnelProtocol.providerBundleIdentifier = AppConfig.packetTunnelBundleIdentifier
        tunnelProtocol.serverAddress = brokerURL.host ?? brokerURL.absoluteString
        tunnelProtocol.providerConfiguration = [
            AppConfig.providerBrokerURLKey: brokerURL.absoluteString
        ]

        manager.protocolConfiguration = tunnelProtocol
        manager.isEnabled = true
        try await manager.saveToPreferences()
        try await manager.loadFromPreferences()
    }

    private func refreshStatus() {
        status = manager?.connection.status ?? .invalid
    }

    private static var canUseNetworkExtensionPreferences: Bool {
        #if targetEnvironment(simulator)
        false
        #else
        true
        #endif
    }
}

enum VPNControllerError: LocalizedError {
    case invalidBrokerURL
    case networkExtensionUnavailableInSimulator

    var errorDescription: String? {
        switch self {
        case .invalidBrokerURL:
            return "Enter a valid broker URL."
        case .networkExtensionUnavailableInSimulator:
            return "The iOS simulator can check the broker, but it cannot install or start this Packet Tunnel VPN profile. Run the app on a signed physical iPhone with the Network Extension packet-tunnel entitlement to test Connect."
        }
    }

    static func message(for error: Error) -> String {
        let nsError = error as NSError
        if nsError.domain == "NEConfigurationErrorDomain", nsError.code == 11 {
            return "Network Extension preferences are unavailable. On the simulator this is expected; on a real iPhone, confirm the app and Packet Tunnel extension are signed with the packet-tunnel Network Extension entitlement and matching App Group."
        }

        return error.localizedDescription
    }
}
