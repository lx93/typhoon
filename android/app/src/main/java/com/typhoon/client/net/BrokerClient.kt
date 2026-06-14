package com.typhoon.client.net

import com.typhoon.client.model.ErrorResponse
import com.typhoon.client.model.RelayListResponse
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.Json
import java.io.IOException
import java.net.HttpURLConnection
import java.net.URI
import java.net.URL
import java.net.URLEncoder

class BrokerClient(
    private val baseUrl: String,
    private val json: Json = Json { ignoreUnknownKeys = true },
) {
    suspend fun listRelays(limit: Int = 5): RelayListResponse = withContext(Dispatchers.IO) {
        val url = URL(relayListUrl(baseUrl, limit))
        val connection = (url.openConnection() as HttpURLConnection).apply {
            requestMethod = "GET"
            connectTimeout = 10_000
            readTimeout = 15_000
        }

        try {
            val status = connection.responseCode
            val stream = if (status in 200..299) {
                connection.inputStream
            } else {
                connection.errorStream ?: connection.inputStream
            }
            val body = stream.bufferedReader().use { it.readText() }
            if (status !in 200..299) {
                val apiError = runCatching { json.decodeFromString<ErrorResponse>(body).error }.getOrNull()
                throw IOException("broker list relays: ${apiError?.ifBlank { null } ?: body.ifBlank { connection.responseMessage }}")
            }
            json.decodeFromString<RelayListResponse>(body)
        } finally {
            connection.disconnect()
        }
    }

    companion object {
        fun relayListUrl(baseUrl: String, limit: Int): String {
            val trimmed = baseUrl.trim()
            require(trimmed.isNotBlank()) { "broker URL is required" }

            val uri = URI(trimmed)
            require(!uri.scheme.isNullOrBlank() && !uri.host.isNullOrBlank()) {
                "broker URL must include scheme and host"
            }

            val basePath = uri.rawPath.orEmpty().trim('/')
            val relayPath = listOf(basePath, "api/v1/relays")
                .filter { it.isNotBlank() }
                .joinToString(separator = "/", prefix = "/")
            val safeLimit = if (limit < 1) 5 else limit
            val query = appendLimit(uri.rawQuery, safeLimit)
            return URI(uri.scheme, uri.userInfo, uri.host, uri.port, relayPath, query, null).toString()
        }

        private fun appendLimit(rawQuery: String?, limit: Int): String {
            val encodedLimit = URLEncoder.encode(limit.toString(), Charsets.UTF_8.name())
            val existing = rawQuery
                ?.split("&")
                ?.filter { it.isNotBlank() }
                ?.filterNot { it.substringBefore("=") == "limit" }
                .orEmpty()
            return (existing + "limit=$encodedLimit")
                .joinToString("&")
        }
    }
}
