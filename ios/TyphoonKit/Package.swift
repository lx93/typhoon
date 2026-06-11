// swift-tools-version: 5.9

import PackageDescription

let package = Package(
    name: "TyphoonKit",
    platforms: [
        .iOS(.v16),
        .macOS(.v13)
    ],
    products: [
        .library(
            name: "TyphoonKit",
            targets: ["TyphoonKit"]
        )
    ],
    targets: [
        .target(
            name: "TyphoonKit"
        ),
        .testTarget(
            name: "TyphoonKitTests",
            dependencies: ["TyphoonKit"]
        )
    ]
)
