package com.sharehodl.ui.theme

import android.app.Activity
import android.os.Build
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.SideEffect
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalView
import androidx.core.view.WindowCompat

// ShareHODL Professional Financial Theme Colors
// Primary - Trust Blue (Professional, Secure)
val PrimaryBlue = Color(0xFF1976D2)
val PrimaryBlueLight = Color(0xFF2196F3)
val PrimaryBlueDark = Color(0xFF1565C0)

// Accent Colors
val GainGreen = Color(0xFF00C853)      // Positive/Profit
val LossRed = Color(0xFFFF1744)        // Negative/Loss
val WarningOrange = Color(0xFFF59E0B)  // Warnings/Alerts
val NeutralGray = Color(0xFF9E9E9E)    // Neutral

// Dark Theme - Modern Financial App
val DarkBackground = Color(0xFF0D1117)     // GitHub-style dark
val DarkSurface = Color(0xFF161B22)        // Elevated surface
val DarkSurfaceVariant = Color(0xFF21262D) // Cards, containers
val DarkOnSurface = Color(0xFFFFFFFF)
val DarkOnSurfaceVariant = Color(0xFFB0B0B0)

// Light Theme - Clean Professional
val LightBackground = Color(0xFFF8FAFC)
val LightSurface = Color(0xFFFFFFFF)
val LightSurfaceVariant = Color(0xFFF1F5F9)
val LightOnSurface = Color(0xFF1A1A1A)
val LightOnSurfaceVariant = Color(0xFF64748B)

private val DarkColorScheme = darkColorScheme(
    primary = PrimaryBlue,
    onPrimary = Color.White,
    primaryContainer = PrimaryBlueDark,
    onPrimaryContainer = Color.White,
    secondary = GainGreen,
    onSecondary = Color.White,
    secondaryContainer = GainGreen.copy(alpha = 0.2f),
    onSecondaryContainer = GainGreen,
    tertiary = WarningOrange,
    onTertiary = Color.Black,
    tertiaryContainer = WarningOrange.copy(alpha = 0.2f),
    onTertiaryContainer = WarningOrange,
    background = DarkBackground,
    onBackground = DarkOnSurface,
    surface = DarkSurface,
    onSurface = DarkOnSurface,
    surfaceVariant = DarkSurfaceVariant,
    onSurfaceVariant = DarkOnSurfaceVariant,
    surfaceTint = PrimaryBlue,
    error = LossRed,
    onError = Color.White,
    errorContainer = LossRed.copy(alpha = 0.2f),
    onErrorContainer = LossRed,
    outline = Color(0xFF30363D),
    outlineVariant = Color(0xFF21262D)
)

private val LightColorScheme = lightColorScheme(
    primary = PrimaryBlue,
    onPrimary = Color.White,
    primaryContainer = Color(0xFFDBEAFE),
    onPrimaryContainer = PrimaryBlueDark,
    secondary = GainGreen,
    onSecondary = Color.White,
    secondaryContainer = Color(0xFFD1FAE5),
    onSecondaryContainer = Color(0xFF065F46),
    tertiary = WarningOrange,
    onTertiary = Color.White,
    tertiaryContainer = Color(0xFFFEF3C7),
    onTertiaryContainer = Color(0xFF78350F),
    background = LightBackground,
    onBackground = LightOnSurface,
    surface = LightSurface,
    onSurface = LightOnSurface,
    surfaceVariant = LightSurfaceVariant,
    onSurfaceVariant = LightOnSurfaceVariant,
    surfaceTint = PrimaryBlue,
    error = LossRed,
    onError = Color.White,
    errorContainer = Color(0xFFFEE2E2),
    onErrorContainer = Color(0xFF991B1B),
    outline = Color(0xFFE2E8F0),
    outlineVariant = Color(0xFFF1F5F9)
)

@Composable
fun ShareHODLTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    dynamicColor: Boolean = false,
    content: @Composable () -> Unit
) {
    val colorScheme = when {
        dynamicColor && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> {
            val context = LocalContext.current
            if (darkTheme) dynamicDarkColorScheme(context) else dynamicLightColorScheme(context)
        }
        darkTheme -> DarkColorScheme
        else -> LightColorScheme
    }

    val view = LocalView.current
    if (!view.isInEditMode) {
        SideEffect {
            val window = (view.context as Activity).window
            // Use dark status bar color for modern look
            window.statusBarColor = if (darkTheme) DarkBackground.toArgb() else LightBackground.toArgb()
            WindowCompat.getInsetsController(window, view).isAppearanceLightStatusBars = !darkTheme
        }
    }

    MaterialTheme(
        colorScheme = colorScheme,
        typography = Typography,
        content = content
    )
}

// Utility extension colors for easy access
object ShareHODLColors {
    val gainGreen = GainGreen
    val lossRed = LossRed
    val warningOrange = WarningOrange
    val neutralGray = NeutralGray
    val primaryBlue = PrimaryBlue

    // Sector colors
    val technologyColor = Color(0xFF3B82F6)
    val healthcareColor = Color(0xFF10B981)
    val financeColor = Color(0xFF8B5CF6)
    val consumerColor = Color(0xFFF59E0B)
    val energyColor = Color(0xFFEF4444)
    val industrialColor = Color(0xFF6B7280)
    val realEstateColor = Color(0xFF14B8A6)
    val utilitiesColor = Color(0xFF06B6D4)
    val materialsColor = Color(0xFFEC4899)
    val communicationsColor = Color(0xFF6366F1)
}
