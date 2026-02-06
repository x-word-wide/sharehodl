package com.sharehodl

import android.app.Application
import dagger.hilt.android.HiltAndroidApp

/**
 * Main Application class for ShareHODL Wallet
 */
@HiltAndroidApp
class ShareHODLApplication : Application() {

    override fun onCreate() {
        super.onCreate()
        // Initialize any app-wide services here
    }
}
