package com.typhoon.client.net

import com.typhoon.client.model.RelayDescriptor
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.net.InetSocketAddress
import java.net.Socket

object RelayReachability {
    suspend fun checkTcp(relay: RelayDescriptor, timeoutMillis: Int = 5_000) {
        withContext(Dispatchers.IO) {
            val host = relay.publicHost.trim().removePrefix("[").removeSuffix("]")
            Socket().use { socket ->
                socket.connect(InetSocketAddress(host, relay.publicPort), timeoutMillis)
            }
        }
    }
}
