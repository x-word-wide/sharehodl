// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "ShareHODL",
    platforms: [
        .iOS(.v16),
        .macOS(.v13)
    ],
    products: [
        .library(
            name: "ShareHODL",
            targets: ["ShareHODL"]
        )
    ],
    dependencies: [
        // secp256k1 - Swift wrapper for libsecp256k1
        .package(url: "https://github.com/GigaBitcoin/secp256k1.swift", from: "0.15.0"),
        // CryptoSwift for additional crypto utilities (RIPEMD160, PBKDF2)
        .package(url: "https://github.com/krzyzanowskim/CryptoSwift", from: "1.8.0")
    ],
    targets: [
        .target(
            name: "ShareHODL",
            dependencies: [
                .product(name: "P256K", package: "secp256k1.swift"),
                .product(name: "CryptoSwift", package: "CryptoSwift")
            ],
            path: "ShareHODL",
            resources: [
                .process("Resources/bip39-english.txt")
            ]
        )
    ]
)
