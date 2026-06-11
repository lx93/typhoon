import Foundation
import Network
import NetworkExtension
import OSLog
import TyphoonKit

protocol PacketTunnelProxyEngine: AnyObject {
    func start(relay: RelayDescriptor, tunnelProvider: NEPacketTunnelProvider) async throws
    func stop()
}

#if canImport(Libbox)
import Libbox

final class EmbeddedProxyEngine: PacketTunnelProxyEngine {
    private let logger = Logger(subsystem: "com.typhoon.client.PacketTunnel", category: "EmbeddedProxyEngine")
    private var commandServer: LibboxCommandServer?
    private var platformInterface: LibboxPacketTunnelPlatformInterface?
    private var activeRelay: RelayDescriptor?

    func start(relay: RelayDescriptor, tunnelProvider: NEPacketTunnelProvider) async throws {
        activeRelay = relay

        let configuration = try SingBoxConfiguration(relay: relay).encodedJSONString()
        let directories = try EngineDirectories.make()

        let setupOptions = LibboxSetupOptions()
        setupOptions.basePath = directories.base.path
        setupOptions.workingPath = directories.working.path
        setupOptions.tempPath = directories.temporary.path
        setupOptions.logMaxLines = 3000
        setupOptions.debug = true
        setupOptions.crashReportSource = "TyphoonPacketTunnel"
        setupOptions.oomKillerEnabled = true

        var setupError: NSError?
        LibboxSetup(setupOptions, &setupError)
        if let setupError {
            throw PacketTunnelProxyEngineError.engineStartFailed(setupError.localizedDescription)
        }
        LibboxPromoteOOMDraft()

        let platformInterface = LibboxPacketTunnelPlatformInterface(tunnelProvider: tunnelProvider)
        var commandServerError: NSError?
        guard let commandServer = LibboxNewCommandServer(platformInterface, platformInterface, &commandServerError) else {
            throw PacketTunnelProxyEngineError.engineStartFailed(commandServerError?.localizedDescription ?? "Unable to create libbox command server.")
        }

        do {
            try commandServer.start()
            try commandServer.startOrReloadService(configuration, options: LibboxOverrideOptions())
        } catch {
            commandServer.close()
            throw PacketTunnelProxyEngineError.engineStartFailed(error.localizedDescription)
        }

        self.platformInterface = platformInterface
        self.commandServer = commandServer
        logger.info("libbox started for relay \(relay.id, privacy: .public)")
    }

    func stop() {
        try? commandServer?.closeService()
        commandServer?.close()
        platformInterface?.reset()
        commandServer = nil
        platformInterface = nil
        activeRelay = nil
    }
}

#else

final class EmbeddedProxyEngine: PacketTunnelProxyEngine {
    private var activeRelay: RelayDescriptor?

    func start(relay: RelayDescriptor, tunnelProvider _: NEPacketTunnelProvider) async throws {
        activeRelay = relay
        _ = try SingBoxConfiguration(relay: relay).encodedJSON()

        throw PacketTunnelProxyEngineError.engineNotLinked
    }

    func stop() {
        activeRelay = nil
    }
}

#endif

enum PacketTunnelProxyEngineError: LocalizedError {
    case engineNotLinked
    case engineStartFailed(String)

    var errorDescription: String? {
        switch self {
        case .engineNotLinked:
            return "Libbox.xcframework is not linked yet. Build sing-box lib_apple and add Libbox to the PacketTunnel target."
        case .engineStartFailed(let message):
            return "The embedded VLESS Reality Vision engine failed to start: \(message)"
        }
    }
}
