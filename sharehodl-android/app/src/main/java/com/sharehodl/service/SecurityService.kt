package com.sharehodl.service

import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import java.io.BufferedReader
import java.io.File
import java.io.InputStreamReader
import javax.inject.Inject
import javax.inject.Singleton

/**
 * Security service for detecting root and other security threats
 * IMPORTANT: A determined attacker can bypass these checks, but they provide
 * a reasonable defense against casual attacks and automated malware
 */
@Singleton
class SecurityService @Inject constructor(
    private val context: Context
) {
    // MARK: - Root Detection

    /**
     * Check if the device appears to be rooted
     * Returns true if root indicators are detected
     */
    fun isRooted(): Boolean {
        return checkRootBinaries() ||
                checkSuperuserApps() ||
                checkRootCloakingApps() ||
                checkDangerousProps() ||
                checkRWPaths() ||
                checkSuExists() ||
                checkTestKeys()
    }

    /**
     * Get detailed security status
     */
    fun getSecurityStatus(): SecurityStatus {
        return SecurityStatus(
            isRooted = isRooted(),
            isEmulator = isEmulator(),
            isDebugBuild = isDebugBuild(),
            isDebuggable = isDebuggable()
        )
    }

    // MARK: - Root Checks

    /**
     * Check for common root binaries
     */
    private fun checkRootBinaries(): Boolean {
        val binaryPaths = arrayOf(
            "/system/app/Superuser.apk",
            "/sbin/su",
            "/system/bin/su",
            "/system/xbin/su",
            "/data/local/xbin/su",
            "/data/local/bin/su",
            "/system/sd/xbin/su",
            "/system/bin/failsafe/su",
            "/data/local/su",
            "/su/bin/su",
            "/su/bin",
            "/system/xbin/daemonsu",
            "/system/bin/.ext/.su",
            "/system/usr/we-need-root/su-backup",
            "/system/xbin/mu",
            "/magisk/.core/bin/su"
        )

        for (path in binaryPaths) {
            if (File(path).exists()) {
                return true
            }
        }

        return false
    }

    /**
     * Check for superuser apps
     */
    private fun checkSuperuserApps(): Boolean {
        val rootApps = arrayOf(
            "com.koushikdutta.superuser",
            "com.thirdparty.superuser",
            "eu.chainfire.supersu",
            "com.noshufou.android.su",
            "com.noshufou.android.su.elite",
            "com.yellowes.su",
            "com.topjohnwu.magisk",
            "com.kingroot.kinguser",
            "com.kingo.root",
            "com.smedialink.oneclickroot",
            "com.zhiqupk.root.global",
            "com.alephzain.framaroot"
        )

        val pm = context.packageManager
        for (packageName in rootApps) {
            try {
                pm.getPackageInfo(packageName, 0)
                return true
            } catch (e: PackageManager.NameNotFoundException) {
                // App not installed, continue
            }
        }

        return false
    }

    /**
     * Check for root cloaking apps
     */
    private fun checkRootCloakingApps(): Boolean {
        val cloakingApps = arrayOf(
            "com.devadvance.rootcloak",
            "com.devadvance.rootcloakplus",
            "de.robv.android.xposed.installer",
            "com.saurik.substrate",
            "com.zachspong.temprootremovejb",
            "com.amphoras.hidemyroot",
            "com.amphoras.hidemyrootadfree",
            "com.formyhm.hideroot",
            "com.formyhm.hiderootPremium",
            "io.github.vvb2060.magisk"
        )

        val pm = context.packageManager
        for (packageName in cloakingApps) {
            try {
                pm.getPackageInfo(packageName, 0)
                return true
            } catch (e: PackageManager.NameNotFoundException) {
                // App not installed, continue
            }
        }

        return false
    }

    /**
     * Check for dangerous build properties
     */
    private fun checkDangerousProps(): Boolean {
        val dangerousProps = mapOf(
            "ro.debuggable" to "1",
            "ro.secure" to "0"
        )

        for ((prop, dangerousValue) in dangerousProps) {
            val value = getSystemProperty(prop)
            if (value == dangerousValue) {
                return true
            }
        }

        return false
    }

    /**
     * Check for read-write paths that should be read-only
     */
    private fun checkRWPaths(): Boolean {
        val paths = arrayOf(
            "/system",
            "/system/bin",
            "/system/sbin",
            "/system/xbin",
            "/vendor/bin",
            "/sbin",
            "/etc"
        )

        for (path in paths) {
            val file = File(path)
            if (file.exists() && file.canWrite()) {
                return true
            }
        }

        return false
    }

    /**
     * Check if su command exists and is executable
     */
    private fun checkSuExists(): Boolean {
        return try {
            val process = Runtime.getRuntime().exec(arrayOf("/system/xbin/which", "su"))
            val reader = BufferedReader(InputStreamReader(process.inputStream))
            val result = reader.readLine()
            reader.close()
            result != null
        } catch (e: Exception) {
            false
        }
    }

    /**
     * Check for test-keys in build
     */
    private fun checkTestKeys(): Boolean {
        val buildTags = Build.TAGS
        return buildTags != null && buildTags.contains("test-keys")
    }

    // MARK: - Emulator Detection

    /**
     * Check if running in an emulator
     */
    fun isEmulator(): Boolean {
        return (Build.BRAND.startsWith("generic") && Build.DEVICE.startsWith("generic")) ||
                Build.FINGERPRINT.startsWith("generic") ||
                Build.FINGERPRINT.startsWith("unknown") ||
                Build.HARDWARE.contains("goldfish") ||
                Build.HARDWARE.contains("ranchu") ||
                Build.MODEL.contains("google_sdk") ||
                Build.MODEL.contains("Emulator") ||
                Build.MODEL.contains("Android SDK built for x86") ||
                Build.MANUFACTURER.contains("Genymotion") ||
                Build.PRODUCT.contains("sdk_google") ||
                Build.PRODUCT.contains("google_sdk") ||
                Build.PRODUCT.contains("sdk") ||
                Build.PRODUCT.contains("sdk_x86") ||
                Build.PRODUCT.contains("sdk_gphone64_arm64") ||
                Build.PRODUCT.contains("vbox86p") ||
                Build.PRODUCT.contains("emulator") ||
                Build.PRODUCT.contains("simulator")
    }

    // MARK: - Debug Detection

    /**
     * Check if this is a debug build
     */
    fun isDebugBuild(): Boolean {
        return try {
            val appInfo = context.packageManager.getApplicationInfo(context.packageName, 0)
            (appInfo.flags and android.content.pm.ApplicationInfo.FLAG_DEBUGGABLE) != 0
        } catch (e: Exception) {
            false
        }
    }

    /**
     * Check if the app is debuggable
     */
    fun isDebuggable(): Boolean {
        return android.os.Debug.isDebuggerConnected()
    }

    // MARK: - Helpers

    private fun getSystemProperty(propName: String): String? {
        return try {
            val process = Runtime.getRuntime().exec(arrayOf("getprop", propName))
            val reader = BufferedReader(InputStreamReader(process.inputStream))
            val result = reader.readLine()
            reader.close()
            result
        } catch (e: Exception) {
            null
        }
    }
}

/**
 * Security status data class
 */
data class SecurityStatus(
    val isRooted: Boolean,
    val isEmulator: Boolean,
    val isDebugBuild: Boolean,
    val isDebuggable: Boolean
) {
    val isSecure: Boolean
        get() = !isRooted && !isDebuggable

    val securityIssues: List<String>
        get() {
            val issues = mutableListOf<String>()
            if (isRooted) {
                issues.add("Device appears to be rooted")
            }
            if (isEmulator) {
                issues.add("Running in emulator")
            }
            if (isDebugBuild) {
                issues.add("Debug build detected")
            }
            if (isDebuggable) {
                issues.add("Debugger attached")
            }
            return issues
        }
}
