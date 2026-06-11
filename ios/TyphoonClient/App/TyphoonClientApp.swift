import SwiftUI

@main
struct TyphoonClientApp: App {
    @StateObject private var vpnController = VPNController()

    var body: some Scene {
        WindowGroup {
            ContentView()
                .environmentObject(vpnController)
                .task {
                    await vpnController.load()
                }
        }
    }
}
