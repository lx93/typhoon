package com.typhoon.client

import android.Manifest
import android.content.Context
import android.content.Intent
import android.net.VpnService
import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.compose.setContent
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextFieldDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.typhoon.client.state.ConnectionStatus
import com.typhoon.client.state.TyphoonStatusStore
import com.typhoon.client.state.TyphoonUiState
import com.typhoon.client.vpn.TyphoonVpnService

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        TyphoonStatusStore.initialize(applicationContext)

        setContent {
            TyphoonApp()
        }
    }
}

@Composable
private fun TyphoonApp() {
    val context = LocalContext.current
    val state by TyphoonStatusStore.uiState.collectAsStateWithLifecycle()
    var brokerUrl by remember(state.brokerUrl) { mutableStateOf(state.brokerUrl) }
    val vpnPermissionLauncher = rememberLauncherForActivityResult(ActivityResultContracts.StartActivityForResult()) {
        startVpn(context, brokerUrl)
    }
    val notificationPermissionLauncher = rememberLauncherForActivityResult(ActivityResultContracts.RequestPermission()) {}

    MaterialTheme {
        TyphoonScreen(
            state = state,
            brokerUrl = brokerUrl,
            onBrokerUrlChange = {
                brokerUrl = it
                TyphoonStatusStore.setBrokerUrl(it)
            },
            onToggle = {
                if (state.isConnected || state.isWorking) {
                    stopVpn(context)
                } else {
                    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
                        notificationPermissionLauncher.launch(Manifest.permission.POST_NOTIFICATIONS)
                    }
                    val prepareIntent = VpnService.prepare(context)
                    if (prepareIntent != null) {
                        vpnPermissionLauncher.launch(prepareIntent)
                    } else {
                        startVpn(context, brokerUrl)
                    }
                }
            },
        )
    }
}

@Composable
private fun TyphoonScreen(
    state: TyphoonUiState,
    brokerUrl: String,
    onBrokerUrlChange: (String) -> Unit,
    onToggle: () -> Unit,
) {
    val terminalGreen = Color(0xFF65F58A)
    val dimGreen = Color(0xFF294F35)
    val panelBlack = Color(0xFF07110B)
    val buttonColor = if (state.isConnected || state.isWorking) Color(0xFFB6F579) else terminalGreen

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xFF030604))
            .padding(20.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        Text(
            text = "typhoon://mobile-client",
            color = terminalGreen,
            fontFamily = FontFamily.Monospace,
            fontWeight = FontWeight.Bold,
            fontSize = 22.sp,
        )
        Text(
            text = "status = ${state.status.label}",
            color = Color(0xFFD8FFE0),
            fontFamily = FontFamily.Monospace,
        )
        state.relayLabel?.let {
            Text(
                text = "relay = $it",
                color = Color(0xFFA5F2B5),
                fontFamily = FontFamily.Monospace,
                fontSize = 13.sp,
            )
        }

        OutlinedTextField(
            value = brokerUrl,
            onValueChange = onBrokerUrlChange,
            label = { Text("broker url", fontFamily = FontFamily.Monospace) },
            textStyle = TextStyle(fontFamily = FontFamily.Monospace, color = Color.White),
            singleLine = true,
            modifier = Modifier.fillMaxWidth(),
            colors = TextFieldDefaults.colors(
                focusedContainerColor = panelBlack,
                unfocusedContainerColor = panelBlack,
                focusedIndicatorColor = terminalGreen,
                unfocusedIndicatorColor = dimGreen,
                focusedLabelColor = terminalGreen,
                unfocusedLabelColor = Color(0xFF7DA989),
                cursorColor = terminalGreen,
            ),
        )

        Button(
            onClick = onToggle,
            modifier = Modifier
                .fillMaxWidth()
                .height(58.dp),
            shape = RoundedCornerShape(8.dp),
            colors = ButtonDefaults.buttonColors(
                containerColor = buttonColor,
                contentColor = Color(0xFF061008),
            ),
        ) {
            Text(
                text = when {
                    state.isWorking -> "DISCONNECT"
                    state.isConnected -> "DISCONNECT"
                    else -> "CONNECT"
                },
                fontFamily = FontFamily.Monospace,
                fontWeight = FontWeight.Black,
                letterSpacing = 1.sp,
            )
        }

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .weight(1f)
                .background(panelBlack, RoundedCornerShape(8.dp))
                .border(1.dp, dimGreen, RoundedCornerShape(8.dp))
                .padding(14.dp),
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .verticalScroll(rememberScrollState()),
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                val lines = state.logLines.ifEmpty {
                    listOf("ready. enter broker URL, then connect.")
                }
                lines.forEach { line ->
                    Text(
                        text = "> $line",
                        color = terminalGreen,
                        fontFamily = FontFamily.Monospace,
                        fontSize = 13.sp,
                        lineHeight = 18.sp,
                    )
                }
                state.lastError?.let {
                    Spacer(Modifier.size(8.dp))
                    Text(
                        text = "! $it",
                        color = Color(0xFFFFA0A0),
                        fontFamily = FontFamily.Monospace,
                        fontSize = 13.sp,
                        lineHeight = 18.sp,
                    )
                }
            }
        }

        Text(
            text = if (state.status == ConnectionStatus.CONNECTED) {
                "traffic route: device -> Typhoon VPN -> volunteer relay"
            } else {
                "vpn is fail-closed: no relay, no connection."
            },
            color = Color(0xFF7DA989),
            fontFamily = FontFamily.Monospace,
            fontSize = 12.sp,
            modifier = Modifier.align(Alignment.CenterHorizontally),
        )
    }
}

private fun startVpn(context: Context, brokerUrl: String) {
    val intent = TyphoonVpnService.connectIntent(context, brokerUrl)
    ContextCompat.startForegroundService(context, intent)
}

private fun stopVpn(context: Context) {
    context.startService(TyphoonVpnService.disconnectIntent(context))
}
