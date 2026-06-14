package com.typhoon.client.vpn

import android.net.VpnService
import android.os.ParcelFileDescriptor
import com.typhoon.client.model.RelayDescriptor

interface ProxyEngine {
    suspend fun start(
        relay: RelayDescriptor,
        configJson: String,
        tunFd: ParcelFileDescriptor,
        vpnService: VpnService,
    )

    fun stop()
}

object ProxyEngineFactory {
    fun create(): ProxyEngine = LibboxProxyEngine()
}

class LibboxProxyEngine : ProxyEngine {
    override suspend fun start(
        relay: RelayDescriptor,
        configJson: String,
        tunFd: ParcelFileDescriptor,
        vpnService: VpnService,
    ) {
        val libboxAvailable = runCatching {
            Class.forName("io.nekohasekai.libbox.Libbox")
        }.isSuccess

        if (!libboxAvailable) {
            throw IllegalStateException(
                "Android libbox is not linked yet. Build sing-box libbox for Android and copy the generated AAR into android/app/libs/libbox.aar.",
            )
        }

        throw IllegalStateException(
            "Android libbox is linked, but the Typhoon Android adapter still needs to be wired to the generated libbox API.",
        )
    }

    override fun stop() = Unit
}
