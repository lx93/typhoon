package com.typhoon.client.vpn

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.net.VpnService
import android.os.Build
import android.os.ParcelFileDescriptor
import androidx.core.app.NotificationCompat
import com.typhoon.client.MainActivity
import com.typhoon.client.R
import com.typhoon.client.config.AppConfig
import com.typhoon.client.model.RelayDescriptor
import com.typhoon.client.model.RelaySelector
import com.typhoon.client.net.BrokerClient
import com.typhoon.client.net.SingBoxConfiguration
import com.typhoon.client.state.ConnectionStatus
import com.typhoon.client.state.TyphoonStatusStore
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.Job
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext

class TyphoonVpnService : VpnService() {
    private val serviceScope = CoroutineScope(SupervisorJob() + Dispatchers.Main.immediate)
    private val relaySelector = RelaySelector()
    private var connectJob: Job? = null
    private var engine: ProxyEngine? = null
    private var tunFd: ParcelFileDescriptor? = null

    override fun onCreate() {
        super.onCreate()
        TyphoonStatusStore.initialize(applicationContext)
        createNotificationChannel()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_CONNECT -> {
                val brokerUrl = intent.getStringExtra(EXTRA_BROKER_URL).orEmpty()
                connectJob?.cancel()
                connectJob = serviceScope.launch {
                    connect(brokerUrl.ifBlank { AppConfig.DEFAULT_BROKER_URL })
                }
            }
            ACTION_DISCONNECT -> disconnect()
        }
        return START_STICKY
    }

    override fun onRevoke() {
        disconnect()
        super.onRevoke()
    }

    override fun onDestroy() {
        disconnect()
        connectJob?.cancel()
        super.onDestroy()
    }

    private suspend fun connect(brokerUrl: String) {
        TyphoonStatusStore.setBrokerUrl(brokerUrl)
        TyphoonStatusStore.clearError()
        TyphoonStatusStore.setStatus(ConnectionStatus.PREPARING, relayLabel = null, lastError = null)
        startForeground(NOTIFICATION_ID, notification("Preparing Typhoon VPN"))

        try {
            TyphoonStatusStore.setStatus(ConnectionStatus.CONNECTING)
            TyphoonStatusStore.appendLog("fetching relays from $brokerUrl")
            val relayResponse = BrokerClient(brokerUrl).listRelays(AppConfig.RELAY_LIMIT)
            val candidates = relaySelector.orderedCandidates(relayResponse.relays, relayResponse.serverInstant)
            TyphoonStatusStore.appendLog(
                "broker returned ${relayResponse.relays.size} relays; ${candidates.size} usable",
            )
            check(candidates.isNotEmpty()) { "No usable VLESS Reality Vision direct-exit relay is available." }

            val relay = connectFirstAvailable(candidates)
            TyphoonStatusStore.setStatus(
                ConnectionStatus.CONNECTED,
                relayLabel = "${relay.publicHost}:${relay.publicPort}",
                lastError = null,
            )
            updateNotification("Connected through ${relay.publicHost}:${relay.publicPort}")
        } catch (error: CancellationException) {
            throw error
        } catch (error: Throwable) {
            cleanupActiveTunnel()
            TyphoonStatusStore.fail(error.message ?: "VPN connection failed")
            stopForeground(STOP_FOREGROUND_REMOVE)
            stopSelf()
        }
    }

    private suspend fun connectFirstAvailable(candidates: List<RelayDescriptor>): RelayDescriptor {
        var lastError: Throwable? = null
        for (relay in candidates) {
            try {
                TyphoonStatusStore.appendLog("trying relay ${relay.id} at ${relay.publicHost}:${relay.publicPort}")
                val config = SingBoxConfiguration(relay = relay).encodedJsonString()
                val fd = establishTunnel()
                tunFd = fd
                val proxyEngine = ProxyEngineFactory.create()
                proxyEngine.start(
                    relay = relay,
                    configJson = config,
                    tunFd = fd,
                    vpnService = this,
                )
                engine = proxyEngine
                tunFd = fd
                return relay
            } catch (error: Throwable) {
                lastError = error
                TyphoonStatusStore.appendLog("relay ${relay.id} failed: ${error.message ?: error::class.java.simpleName}")
                cleanupActiveTunnel()
            }
        }

        throw IllegalStateException("All relay connection attempts failed. Last error: ${lastError?.message ?: "unknown"}")
    }

    private suspend fun establishTunnel(): ParcelFileDescriptor = withContext(Dispatchers.IO) {
        Builder()
            .setSession(AppConfig.VPN_SESSION_NAME)
            .setMtu(1500)
            .addAddress("172.19.0.1", 30)
            .addAddress("fdfe:dcba:9876::1", 126)
            .addDnsServer("1.1.1.1")
            .addDnsServer("8.8.8.8")
            .addRoute("0.0.0.0", 0)
            .addRoute("::", 0)
            .allowFamily(android.system.OsConstants.AF_INET)
            .allowFamily(android.system.OsConstants.AF_INET6)
            .establish()
            ?: throw IllegalStateException("Android did not return a VPN tunnel file descriptor.")
    }

    private fun disconnect() {
        TyphoonStatusStore.setStatus(ConnectionStatus.DISCONNECTING)
        connectJob?.cancel()
        cleanupActiveTunnel()
        stopForeground(STOP_FOREGROUND_REMOVE)
        TyphoonStatusStore.setStatus(ConnectionStatus.DISCONNECTED, relayLabel = null, lastError = null)
        stopSelf()
    }

    private fun cleanupActiveTunnel() {
        engine?.stop()
        engine = null
        tunFd.closeQuietly()
        tunFd = null
    }

    private fun ParcelFileDescriptor?.closeQuietly() {
        if (this == null) return
        runCatching { close() }
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) return
        val manager = getSystemService(NotificationManager::class.java)
        val channel = NotificationChannel(
            NOTIFICATION_CHANNEL_ID,
            getString(R.string.vpn_notification_channel),
            NotificationManager.IMPORTANCE_LOW,
        )
        manager.createNotificationChannel(channel)
    }

    private fun updateNotification(message: String) {
        val manager = getSystemService(NotificationManager::class.java)
        manager.notify(NOTIFICATION_ID, notification(message))
    }

    private fun notification(message: String): Notification {
        val launchIntent = Intent(this, MainActivity::class.java)
        val pendingIntent = PendingIntent.getActivity(
            this,
            0,
            launchIntent,
            PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT,
        )
        return NotificationCompat.Builder(this, NOTIFICATION_CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_vpn)
            .setContentTitle(getString(R.string.vpn_notification_title))
            .setContentText(message)
            .setContentIntent(pendingIntent)
            .setOngoing(true)
            .setOnlyAlertOnce(true)
            .build()
    }

    companion object {
        private const val ACTION_CONNECT = "com.typhoon.client.action.CONNECT"
        private const val ACTION_DISCONNECT = "com.typhoon.client.action.DISCONNECT"
        private const val EXTRA_BROKER_URL = "broker_url"
        private const val NOTIFICATION_CHANNEL_ID = "typhoon_vpn"
        private const val NOTIFICATION_ID = 2001

        fun connectIntent(context: Context, brokerUrl: String): Intent =
            Intent(context, TyphoonVpnService::class.java).apply {
                action = ACTION_CONNECT
                putExtra(EXTRA_BROKER_URL, brokerUrl)
            }

        fun disconnectIntent(context: Context): Intent =
            Intent(context, TyphoonVpnService::class.java).apply {
                action = ACTION_DISCONNECT
            }
    }
}
