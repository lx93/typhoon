import Foundation

public struct RelaySelector: Sendable {
    public init() {}

    public func orderedCandidates(from relays: [RelayDescriptor], now: Date = Date()) -> [RelayDescriptor] {
        let usable = relays.filter { $0.isUsable(now: now) }
        return usable.filter { Self.isIPv6Literal($0.publicHost) } +
            usable.filter { !Self.isIPv6Literal($0.publicHost) }
    }

    public func selectFirstUsable(from relays: [RelayDescriptor], now: Date = Date()) -> RelayDescriptor? {
        orderedCandidates(from: relays, now: now).first
    }

    private static func isIPv6Literal(_ host: String) -> Bool {
        let trimmed = host.trimmingCharacters(in: CharacterSet(charactersIn: "[]"))
        return trimmed.contains(":")
    }
}
