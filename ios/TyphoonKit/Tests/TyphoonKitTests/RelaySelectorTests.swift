import XCTest
@testable import TyphoonKit

final class RelaySelectorTests: XCTestCase {
    func testSelectsFirstUsableDirectVLESSRealityRelay() {
        let now = Date(timeIntervalSince1970: 1_800_000_000)
        let expired = relay(
            id: "expired",
            expiresAt: now.addingTimeInterval(-1)
        )
        let usable = relay(
            id: "usable",
            expiresAt: now.addingTimeInterval(60)
        )

        let selected = RelaySelector().selectFirstUsable(from: [expired, usable], now: now)

        XCTAssertEqual(selected?.id, "usable")
    }

    func testRejectsWrongProtocol() {
        let now = Date(timeIntervalSince1970: 1_800_000_000)
        let wrongProtocol = relay(
            id: "wrong",
            relayProtocol: "socks5",
            expiresAt: now.addingTimeInterval(60)
        )

        XCTAssertNil(RelaySelector().selectFirstUsable(from: [wrongProtocol], now: now))
    }

    func testBuildsProxyEngineConfigFromRelayDescriptor() throws {
        let descriptor = relay(id: "relay-1")
        let config = ProxyEngineConfiguration(relay: descriptor)
        let data = try config.encodedJSON()
        let json = String(decoding: data, as: UTF8.self)

        XCTAssertTrue(json.contains("\"address\" : \"volunteer.example.com\""))
        XCTAssertTrue(json.contains("\"flow\" : \"xtls-rprx-vision\""))
        XCTAssertTrue(json.contains("\"security\" : \"reality\""))
    }

    func testBuildsSingBoxVLESSRealityVisionConfig() throws {
        let descriptor = relay(id: "relay-1")
        let json = try SingBoxConfiguration(relay: descriptor).encodedJSONString()

        XCTAssertTrue(json.contains("\"type\" : \"tun\""))
        XCTAssertTrue(json.contains("\"type\" : \"vless\""))
        XCTAssertTrue(json.contains("\"server\" : \"volunteer.example.com\""))
        XCTAssertTrue(json.contains("\"server_port\" : 443"))
        XCTAssertTrue(json.contains("\"flow\" : \"xtls-rprx-vision\""))
        XCTAssertTrue(json.contains("\"public_key\" : \"reality-public-key\""))
        XCTAssertTrue(json.contains("\"short_id\" : \"5f7a8d9c01ab23cd\""))
        XCTAssertTrue(json.contains("\"fingerprint\" : \"chrome\""))
        XCTAssertTrue(json.contains("\"auto_route\" : true"))
        XCTAssertTrue(json.contains("\"strict_route\" : true"))
        XCTAssertTrue(json.contains("\"default_domain_resolver\" : \"dns-0\""))
        XCTAssertTrue(json.contains("\"protocol\" : \"dns\""))
        XCTAssertTrue(json.contains("\"action\" : \"hijack-dns\""))
        XCTAssertTrue(json.contains("\"type\" : \"tcp\""))
        XCTAssertTrue(json.contains("\"detour\" : \"proxy\""))
    }

    private func relay(
        id: String,
        relayProtocol: String = RelayConstants.protocolVLESSRealityVision,
        flow: String = RelayConstants.flowVision,
        exitMode: String = RelayConstants.exitModeDirect,
        expiresAt: Date = Date(timeIntervalSince1970: 1_800_000_060)
    ) -> RelayDescriptor {
        RelayDescriptor(
            id: id,
            publicHost: "volunteer.example.com",
            publicPort: 443,
            relayProtocol: relayProtocol,
            clientID: "2c08df10-4ef4-4ab9-95c6-cb1e94cdb2ff",
            realityPublicKey: "reality-public-key",
            shortID: "5f7a8d9c01ab23cd",
            serverName: "www.microsoft.com",
            flow: flow,
            exitMode: exitMode,
            maxSessions: 8,
            maxMbps: 20,
            volunteerVersion: "dev",
            registeredAt: Date(timeIntervalSince1970: 1_800_000_000),
            lastHeartbeatAt: Date(timeIntervalSince1970: 1_800_000_000),
            expiresAt: expiresAt
        )
    }
}
