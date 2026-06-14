package com.typhoon.client.net

import org.junit.Assert.assertEquals
import org.junit.Assert.assertThrows
import org.junit.Test

class BrokerClientTest {
    @Test
    fun buildsRelayListUrlForPlainBroker() {
        assertEquals(
            "http://localhost:8080/api/v1/relays?limit=5",
            BrokerClient.relayListUrl("http://localhost:8080", 5),
        )
    }

    @Test
    fun preservesBrokerBasePath() {
        assertEquals(
            "https://example.com/typhoon/api/v1/relays?limit=10",
            BrokerClient.relayListUrl("https://example.com/typhoon/", 10),
        )
    }

    @Test
    fun defaultsInvalidLimitToFive() {
        assertEquals(
            "https://example.com/api/v1/relays?limit=5",
            BrokerClient.relayListUrl("https://example.com", 0),
        )
    }

    @Test
    fun replacesExistingLimitQuery() {
        assertEquals(
            "https://example.com/api/v1/relays?foo=bar&limit=8",
            BrokerClient.relayListUrl("https://example.com?foo=bar&limit=1", 8),
        )
    }

    @Test
    fun rejectsBrokerWithoutHost() {
        assertThrows(IllegalArgumentException::class.java) {
            BrokerClient.relayListUrl("localhost:8080", 5)
        }
    }
}
