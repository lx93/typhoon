import SwiftUI

struct ContentView: View {
    @EnvironmentObject private var vpnController: VPNController
    @State private var brokerURLText = AppConfig.defaultBrokerURL.absoluteString

    var body: some View {
        NavigationStack {
            Form {
                Section("Connection") {
                    LabeledContent("Status", value: vpnController.statusText)

                    if let relay = vpnController.selectedRelayLabel {
                        LabeledContent("Relay", value: relay)
                    }

                    Button(vpnController.isConnected ? "Disconnect" : "Connect") {
                        Task {
                            if vpnController.isConnected {
                                await vpnController.disconnect()
                            } else {
                                await vpnController.connect(brokerURLText: brokerURLText)
                            }
                        }
                    }
                    .disabled(vpnController.isWorking)
                }

                Section("Broker") {
                    TextField("Broker URL", text: $brokerURLText)
                        .keyboardType(.URL)
                        .textInputAutocapitalization(.never)
                        .autocorrectionDisabled()

                    Button("Check Broker") {
                        Task {
                            await vpnController.checkBroker(brokerURLText: brokerURLText)
                        }
                    }
                    .disabled(vpnController.isWorking)
                }

                if let brokerSummary = vpnController.brokerSummary {
                    Section("Broker Result") {
                        Text(brokerSummary)

                        ForEach(vpnController.brokerRelays.prefix(5)) { relay in
                            VStack(alignment: .leading, spacing: 4) {
                                Text("\(relay.publicHost):\(relay.publicPort)")
                                    .font(.headline)
                                Text("\(relay.relayProtocol), \(relay.exitMode), expires \(relay.expiresAt.formatted())")
                                    .font(.footnote)
                                    .foregroundStyle(.secondary)
                            }
                        }
                    }
                }

                if let lastError = vpnController.lastError {
                    Section("Last Error") {
                        Text(lastError)
                            .font(.footnote)
                            .foregroundStyle(.secondary)
                    }
                }
            }
            .navigationTitle("Typhoon")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("Refresh") {
                        Task {
                            await vpnController.load()
                        }
                    }
                }
            }
        }
    }
}

#Preview {
    ContentView()
        .environmentObject(VPNController())
}
