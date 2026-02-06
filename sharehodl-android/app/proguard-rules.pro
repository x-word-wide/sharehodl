# ShareHODL ProGuard Rules

# Keep BouncyCastle crypto classes
-keep class org.bouncycastle.** { *; }
-dontwarn org.bouncycastle.**

# Web3j Crypto library
-keep class org.web3j.crypto.** { *; }
-dontwarn org.web3j.**

# Keep our crypto service (prevent inlining of security checks)
-keep class com.sharehodl.service.CryptoService { *; }
-keep class com.sharehodl.service.CryptoService$Account { *; }
-keep class com.sharehodl.service.KeystoreService { *; }
-keep class com.sharehodl.service.BlockchainService { *; }

# Keep model classes for JSON serialization
-keep class com.sharehodl.model.** { *; }
-keep class com.sharehodl.service.Balance { *; }
-keep class com.sharehodl.service.Validator { *; }
-keep class com.sharehodl.service.Delegation { *; }
-keep class com.sharehodl.service.Rewards { *; }
-keep class com.sharehodl.service.Proposal { *; }
-keep class com.sharehodl.service.Transaction { *; }

# Keep ExtendedKey for crypto operations
-keep class com.sharehodl.service.ExtendedKey { *; }

# Remove logging in release builds (security)
-assumenosideeffects class android.util.Log {
    public static int v(...);
    public static int d(...);
    public static int i(...);
}

# Keep Hilt
-keep class dagger.hilt.** { *; }
-keep class javax.inject.** { *; }
-keepclasseswithmembers class * {
    @dagger.hilt.* <methods>;
}

# Kotlin Coroutines
-keepnames class kotlinx.coroutines.internal.MainDispatcherFactory {}
-keepnames class kotlinx.coroutines.CoroutineExceptionHandler {}

# Retrofit
-keepattributes Signature
-keepattributes *Annotation*
-keep class retrofit2.** { *; }
-keepclasseswithmembers class * {
    @retrofit2.http.* <methods>;
}

# OkHttp
-dontwarn okhttp3.**
-dontwarn okio.**
-keep class okhttp3.** { *; }
-keep interface okhttp3.** { *; }

# ZXing for QR codes
-keep class com.google.zxing.** { *; }
