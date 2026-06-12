import Foundation
import NetworkExtension
import OSLog
import TyphoonKit

final class PacketTunnelProvider: NEPacketTunnelProvider {
    private let logger = Logger(subsystem: "com.typhoon.client.PacketTunnel", category: "PacketTunnel")
    private var engine: PacketTunnelProxyEngine?
    private let selector = RelaySelector()

    override func startTunnel(
        options: [String: NSObject]?,
        completionHandler: @escaping (Error?) -> Void
    ) {
        Task {
            do {
                TunnelDiagnostics.clear()
                TunnelDiagnostics.recordEvent("Starting packet tunnel")
                logger.info("Starting packet tunnel")
                let brokerURL = try resolveBrokerURL()
                TunnelDiagnostics.recordEvent("Resolved broker URL: \(brokerURL.absoluteString)")
                logger.info("Resolved broker URL: \(brokerURL.absoluteString, privacy: .public)")

                let broker = BrokerClient(baseURL: brokerURL)
                let response = try await broker.listRelays(limit: 5)
                let candidates = selector.orderedCandidates(from: response.relays, now: response.serverTime)
                TunnelDiagnostics.recordEvent("Broker returned \(response.relays.count) relays; \(candidates.count) usable candidates")
                logger.info("Broker returned \(response.relays.count, privacy: .public) relays; \(candidates.count, privacy: .public) usable candidates")

                guard candidates.isEmpty == false else {
                    throw PacketTunnelError.noUsableRelay
                }

                let relay = try await connectFirstAvailableRelay(candidates)
                TunnelDiagnostics.recordEvent("Started tunnel through relay \(relay.id)")
                logger.info("Started tunnel through relay \(relay.id, privacy: .public)")
                completionHandler(nil)
            } catch {
                TunnelDiagnostics.recordError(error.localizedDescription)
                logger.error("Failed to start tunnel: \(error.localizedDescription, privacy: .public)")
                completionHandler(error)
            }
        }
    }

    override func stopTunnel(
        with reason: NEProviderStopReason,
        completionHandler: @escaping () -> Void
    ) {
        engine?.stop()
        engine = nil
        completionHandler()
    }

    private func resolveBrokerURL() throws -> URL {
        guard
            let tunnelProtocol = protocolConfiguration as? NETunnelProviderProtocol,
            let providerConfiguration = tunnelProtocol.providerConfiguration,
            let urlString = providerConfiguration[AppConfig.providerBrokerURLKey] as? String,
            let url = URL(string: urlString)
        else {
            return AppConfig.defaultBrokerURL
        }

        return url
    }

    private func connectFirstAvailableRelay(_ candidates: [RelayDescriptor]) async throws -> RelayDescriptor {
        var lastError: Error?

        for relay in candidates {
            do {
                TunnelDiagnostics.recordEvent("Trying relay \(relay.id) at \(relay.publicHost):\(relay.publicPort)")
                logger.info("Trying relay \(relay.id, privacy: .public) at \(relay.publicHost, privacy: .public):\(relay.publicPort, privacy: .public)")
                let engine = EmbeddedProxyEngine()
                try await engine.start(relay: relay, tunnelProvider: self)
                self.engine = engine

                return relay
            } catch {
                lastError = error
                TunnelDiagnostics.recordError("Relay \(relay.id) failed: \(error.localizedDescription)")
                self.engine?.stop()
                self.engine = nil
                logger.warning("Relay \(relay.id, privacy: .public) failed: \(error.localizedDescription, privacy: .public)")
            }
        }

        throw PacketTunnelError.allRelaysFailed(lastError?.localizedDescription)
    }
}

enum PacketTunnelError: LocalizedError {
    case noUsableRelay
    case allRelaysFailed(String?)

    var errorDescription: String? {
        switch self {
        case .noUsableRelay:
            return "No usable VLESS Reality Vision direct-exit relay is available."
        case .allRelaysFailed(let message):
            if let message {
                return "All relay connection attempts failed. Last error: \(message)"
            }
            return "All relay connection attempts failed."
        }
    }
}
