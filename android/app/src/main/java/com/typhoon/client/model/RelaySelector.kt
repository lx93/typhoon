package com.typhoon.client.model

import java.time.Instant

class RelaySelector {
    fun orderedCandidates(relays: List<RelayDescriptor>, now: Instant): List<RelayDescriptor> {
        val usable = relays.filter { it.isUsable(now) }
        return usable.filter { it.publicHost.isIPv6Literal() } +
            usable.filterNot { it.publicHost.isIPv6Literal() }
    }

    fun selectFirstUsable(relays: List<RelayDescriptor>, now: Instant): RelayDescriptor? =
        orderedCandidates(relays, now).firstOrNull()

    private fun String.isIPv6Literal(): Boolean =
        trim('[', ']').contains(":")
}
