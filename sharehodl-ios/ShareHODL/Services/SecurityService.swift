import Foundation
import UIKit

/// Security service for detecting jailbreak and other security threats
/// IMPORTANT: A determined attacker can bypass these checks, but they provide
/// a reasonable defense against casual attacks and automated malware
final class SecurityService {
    static let shared = SecurityService()

    private init() {}

    // MARK: - Jailbreak Detection

    /// Check if the device appears to be jailbroken
    /// Returns true if jailbreak indicators are detected
    func isJailbroken() -> Bool {
        #if targetEnvironment(simulator)
        // Don't flag simulators as jailbroken
        return false
        #else
        return checkJailbreakFiles() ||
               checkJailbreakPaths() ||
               checkSandboxIntegrity() ||
               checkSymbolicLinks() ||
               checkWritePermissions()
        #endif
    }

    /// Get detailed security status
    func getSecurityStatus() -> SecurityStatus {
        return SecurityStatus(
            isJailbroken: isJailbroken(),
            isDebuggerAttached: isDebuggerAttached(),
            isRunningInEmulator: isRunningInEmulator()
        )
    }

    // MARK: - Jailbreak Checks

    /// Check for common jailbreak files
    private func checkJailbreakFiles() -> Bool {
        let jailbreakFiles = [
            "/Applications/Cydia.app",
            "/Applications/Sileo.app",
            "/Applications/Zebra.app",
            "/Applications/Installer.app",
            "/Applications/blackra1n.app",
            "/Applications/FakeCarrier.app",
            "/Applications/Icy.app",
            "/Applications/IntelliScreen.app",
            "/Applications/MxTube.app",
            "/Applications/RockApp.app",
            "/Applications/SBSettings.app",
            "/Applications/WinterBoard.app",
            "/Library/MobileSubstrate/MobileSubstrate.dylib",
            "/Library/MobileSubstrate/DynamicLibraries/LiveClock.plist",
            "/Library/MobileSubstrate/DynamicLibraries/Veency.plist",
            "/System/Library/LaunchDaemons/com.ikey.bbot.plist",
            "/System/Library/LaunchDaemons/com.saurik.Cydia.Startup.plist",
            "/bin/bash",
            "/bin/sh",
            "/usr/sbin/sshd",
            "/usr/bin/sshd",
            "/usr/libexec/sftp-server",
            "/etc/apt",
            "/private/var/lib/apt",
            "/private/var/lib/cydia",
            "/private/var/stash",
            "/private/var/tmp/cydia.log",
            "/var/cache/apt",
            "/var/lib/apt",
            "/var/lib/cydia"
        ]

        for path in jailbreakFiles {
            if FileManager.default.fileExists(atPath: path) {
                return true
            }
        }

        return false
    }

    /// Check if we can access paths that should be inaccessible
    private func checkJailbreakPaths() -> Bool {
        let restrictedPaths = [
            "/private/var/mobile/Library/Caches/com.saurik.Cydia",
            "/var/log/syslog",
            "/private/var/log/syslog"
        ]

        for path in restrictedPaths {
            if FileManager.default.isReadableFile(atPath: path) {
                return true
            }
        }

        return false
    }

    /// Check sandbox integrity
    private func checkSandboxIntegrity() -> Bool {
        // Try to write outside the sandbox
        let testPath = "/private/jailbreak_test_\(UUID().uuidString)"
        do {
            try "test".write(toFile: testPath, atomically: true, encoding: .utf8)
            try FileManager.default.removeItem(atPath: testPath)
            return true // We shouldn't be able to write here
        } catch {
            return false // Expected behavior
        }
    }

    /// Check for symbolic links that indicate jailbreak
    private func checkSymbolicLinks() -> Bool {
        let suspiciousPaths = [
            "/Applications",
            "/Library/Ringtones",
            "/Library/Wallpaper",
            "/usr/arm-apple-darwin9",
            "/usr/include",
            "/usr/libexec",
            "/usr/share"
        ]

        for path in suspiciousPaths {
            var isDirectory: ObjCBool = false
            if FileManager.default.fileExists(atPath: path, isDirectory: &isDirectory) {
                // Check if it's a symbolic link
                do {
                    let attributes = try FileManager.default.attributesOfItem(atPath: path)
                    if attributes[.type] as? FileAttributeType == .typeSymbolicLink {
                        return true
                    }
                } catch {
                    continue
                }
            }
        }

        return false
    }

    /// Check if we have unexpected write permissions
    private func checkWritePermissions() -> Bool {
        let restrictedPaths = [
            "/",
            "/private",
            "/etc"
        ]

        for path in restrictedPaths {
            if FileManager.default.isWritableFile(atPath: path) {
                return true
            }
        }

        return false
    }

    // MARK: - Debugger Detection

    /// Check if a debugger is attached
    func isDebuggerAttached() -> Bool {
        var info = kinfo_proc()
        var size = MemoryLayout<kinfo_proc>.stride
        var mib: [Int32] = [CTL_KERN, KERN_PROC, KERN_PROC_PID, getpid()]

        let result = sysctl(&mib, UInt32(mib.count), &info, &size, nil, 0)
        guard result == 0 else { return false }

        return (info.kp_proc.p_flag & P_TRACED) != 0
    }

    // MARK: - Emulator Detection

    /// Check if running in an emulator (not simulator)
    func isRunningInEmulator() -> Bool {
        #if targetEnvironment(simulator)
        return true
        #else
        // Check for Frida
        if let _ = dlopen("FridaGadget", RTLD_NOW) {
            return true
        }

        // Check for suspicious environment variables
        let suspiciousEnvVars = ["DYLD_INSERT_LIBRARIES", "DYLD_LIBRARY_PATH"]
        for envVar in suspiciousEnvVars {
            if ProcessInfo.processInfo.environment[envVar] != nil {
                return true
            }
        }

        return false
        #endif
    }

    // MARK: - URL Scheme Check

    /// Check if suspicious URL schemes can be opened
    func canOpenSuspiciousURLs() -> Bool {
        let suspiciousSchemes = [
            "cydia://",
            "sileo://",
            "zbra://",
            "filza://"
        ]

        for scheme in suspiciousSchemes {
            if let url = URL(string: scheme),
               UIApplication.shared.canOpenURL(url) {
                return true
            }
        }

        return false
    }
}

// MARK: - Security Status

struct SecurityStatus {
    let isJailbroken: Bool
    let isDebuggerAttached: Bool
    let isRunningInEmulator: Bool

    var isSecure: Bool {
        return !isJailbroken && !isDebuggerAttached
    }

    var securityIssues: [String] {
        var issues: [String] = []
        if isJailbroken {
            issues.append("Device appears to be jailbroken")
        }
        if isDebuggerAttached {
            issues.append("Debugger detected")
        }
        if isRunningInEmulator {
            issues.append("Running in emulator/simulator")
        }
        return issues
    }
}
