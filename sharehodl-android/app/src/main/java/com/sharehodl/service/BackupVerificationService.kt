package com.sharehodl.service

import android.content.Context
import android.content.SharedPreferences
import androidx.security.crypto.EncryptedSharedPreferences
import androidx.security.crypto.MasterKey
import java.util.Date

/**
 * Service for verifying user has backed up their recovery phrase
 * SECURITY: Uses EncryptedSharedPreferences for secure storage
 */
class BackupVerificationService(context: Context) {

    private val masterKey = MasterKey.Builder(context.applicationContext)
        .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
        .build()

    private val prefs: SharedPreferences = EncryptedSharedPreferences.create(
        context.applicationContext,
        PREFS_NAME,
        masterKey,
        EncryptedSharedPreferences.PrefKeyEncryptionScheme.AES256_SIV,
        EncryptedSharedPreferences.PrefValueEncryptionScheme.AES256_GCM
    )

    /**
     * Check if backup has been verified
     */
    val isBackupVerified: Boolean
        get() = prefs.getBoolean(KEY_BACKUP_VERIFIED, false)

    /**
     * Date when backup was verified (epoch millis)
     */
    val backupVerifiedDate: Long?
        get() {
            val time = prefs.getLong(KEY_BACKUP_VERIFIED_DATE, -1)
            return if (time == -1L) null else time
        }

    /**
     * Mark backup as verified
     */
    fun markBackupVerified() {
        prefs.edit()
            .putBoolean(KEY_BACKUP_VERIFIED, true)
            .putLong(KEY_BACKUP_VERIFIED_DATE, System.currentTimeMillis())
            .apply()
    }

    /**
     * Reset backup verification (for testing or wallet reset)
     */
    fun resetBackupVerification() {
        prefs.edit()
            .remove(KEY_BACKUP_VERIFIED)
            .remove(KEY_BACKUP_VERIFIED_DATE)
            .apply()
    }

    /**
     * Generate a verification challenge with random word positions
     * @param mnemonic The full recovery phrase
     * @param challengeCount Number of words to verify (default 3)
     * @return List of word challenges
     */
    fun generateChallenge(mnemonic: String, challengeCount: Int = 3): List<WordChallenge> {
        val words = mnemonic.split(" ")
        if (words.size < challengeCount) return emptyList()

        // Generate random unique positions
        val positions = mutableSetOf<Int>()
        while (positions.size < challengeCount) {
            positions.add((0 until words.size).random())
        }

        // Create challenges sorted by position
        return positions.sorted().map { position ->
            WordChallenge(
                wordNumber = position + 1,  // 1-indexed for user display
                correctWord = words[position],
                userAnswer = ""
            )
        }
    }

    /**
     * Verify user answers against the challenge
     * @param challenges The word challenges with user answers
     * @return True if all answers are correct
     */
    fun verifyChallenge(challenges: List<WordChallenge>): Boolean {
        return challenges.all { challenge ->
            challenge.userAnswer.lowercase().trim() == challenge.correctWord.lowercase()
        }
    }

    companion object {
        private const val PREFS_NAME = "backup_verification"
        private const val KEY_BACKUP_VERIFIED = "backup_verified"
        private const val KEY_BACKUP_VERIFIED_DATE = "backup_verified_date"

        @Volatile
        private var instance: BackupVerificationService? = null

        fun getInstance(context: Context): BackupVerificationService {
            return instance ?: synchronized(this) {
                instance ?: BackupVerificationService(context.applicationContext).also {
                    instance = it
                }
            }
        }
    }
}

/**
 * Represents a single word verification challenge
 */
data class WordChallenge(
    val wordNumber: Int,      // 1-indexed position (e.g., "Word #5")
    val correctWord: String,  // The correct answer
    var userAnswer: String    // User's input
) {
    /**
     * Check if user's answer is correct
     */
    val isCorrect: Boolean
        get() = userAnswer.lowercase().trim() == correctWord.lowercase()

    /**
     * Display string for the challenge
     */
    val displayText: String
        get() = "Word #$wordNumber"
}
