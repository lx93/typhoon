package com.typhoon.client.state

import android.content.Context
import com.typhoon.client.config.AppConfig
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import java.time.LocalTime
import java.time.format.DateTimeFormatter

object TyphoonStatusStore {
    private const val KEY_STATUS = "status"
    private const val KEY_BROKER_URL = "broker_url"
    private const val KEY_RELAY_LABEL = "relay_label"
    private const val KEY_LAST_ERROR = "last_error"
    private const val KEY_LOG_LINES = "log_lines"
    private const val MAX_LOG_LINES = 80

    private val timeFormatter = DateTimeFormatter.ofPattern("HH:mm:ss")
    private val state = MutableStateFlow(TyphoonUiState(brokerUrl = AppConfig.DEFAULT_BROKER_URL))
    private var appContext: Context? = null

    val uiState: StateFlow<TyphoonUiState> = state.asStateFlow()

    fun initialize(context: Context) {
        if (appContext != null) return
        appContext = context.applicationContext
        val prefs = context.getSharedPreferences(AppConfig.STATUS_PREFS, Context.MODE_PRIVATE)
        val restoredStatus = runCatching {
            ConnectionStatus.valueOf(prefs.getString(KEY_STATUS, ConnectionStatus.DISCONNECTED.name)!!)
        }.getOrDefault(ConnectionStatus.DISCONNECTED)
        state.value = TyphoonUiState(
            status = if (restoredStatus == ConnectionStatus.CONNECTED) ConnectionStatus.DISCONNECTED else restoredStatus,
            brokerUrl = prefs.getString(KEY_BROKER_URL, AppConfig.DEFAULT_BROKER_URL) ?: AppConfig.DEFAULT_BROKER_URL,
            relayLabel = prefs.getString(KEY_RELAY_LABEL, null),
            lastError = prefs.getString(KEY_LAST_ERROR, null),
            logLines = prefs.getString(KEY_LOG_LINES, null)?.lines()?.filter { it.isNotBlank() }.orEmpty(),
        )
    }

    fun setBrokerUrl(brokerUrl: String) {
        state.update { it.copy(brokerUrl = brokerUrl) }
        persist()
    }

    fun setStatus(
        status: ConnectionStatus,
        relayLabel: String? = state.value.relayLabel,
        lastError: String? = state.value.lastError,
    ) {
        state.update {
            it.copy(
                status = status,
                relayLabel = relayLabel,
                lastError = lastError,
            )
        }
        appendLog(status.label)
    }

    fun appendLog(message: String) {
        val timestamp = LocalTime.now().format(timeFormatter)
        state.update {
            it.copy(logLines = (it.logLines + "[$timestamp] $message").takeLast(MAX_LOG_LINES))
        }
        persist()
    }

    fun fail(message: String) {
        state.update {
            it.copy(
                status = ConnectionStatus.FAILED,
                lastError = message,
                relayLabel = null,
                logLines = (it.logLines + "[${LocalTime.now().format(timeFormatter)}] error: $message")
                    .takeLast(MAX_LOG_LINES),
            )
        }
        persist()
    }

    fun clearError() {
        state.update { it.copy(lastError = null) }
        persist()
    }

    private fun persist() {
        val context = appContext ?: return
        val current = state.value
        context.getSharedPreferences(AppConfig.STATUS_PREFS, Context.MODE_PRIVATE)
            .edit()
            .putString(KEY_STATUS, current.status.name)
            .putString(KEY_BROKER_URL, current.brokerUrl)
            .putString(KEY_RELAY_LABEL, current.relayLabel)
            .putString(KEY_LAST_ERROR, current.lastError)
            .putString(KEY_LOG_LINES, current.logLines.joinToString("\n"))
            .apply()
    }
}
