package com.typhoon.client.state

enum class ConnectionStatus(val label: String) {
    DISCONNECTED("Disconnected"),
    PREPARING("Preparing VPN"),
    CONNECTING("Connecting"),
    CONNECTED("Connected"),
    DISCONNECTING("Disconnecting"),
    FAILED("Failed"),
}

data class TyphoonUiState(
    val status: ConnectionStatus = ConnectionStatus.DISCONNECTED,
    val brokerUrl: String = "",
    val relayLabel: String? = null,
    val lastError: String? = null,
    val logLines: List<String> = emptyList(),
) {
    val isWorking: Boolean
        get() = status == ConnectionStatus.PREPARING ||
            status == ConnectionStatus.CONNECTING ||
            status == ConnectionStatus.DISCONNECTING

    val isConnected: Boolean
        get() = status == ConnectionStatus.CONNECTED
}
