package com.sharehodl.di

import android.content.Context
import com.sharehodl.service.BlockchainService
import com.sharehodl.service.CryptoService
import com.sharehodl.service.KeystoreService
import com.sharehodl.service.SettingsPreferences
import com.sharehodl.service.api.MultiChainApiService
import com.sharehodl.service.api.PriceService
import com.sharehodl.service.tx.CosmosTransactionBuilder
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

/**
 * Hilt module for providing app-wide dependencies
 */
@Module
@InstallIn(SingletonComponent::class)
object AppModule {

    @Provides
    @Singleton
    fun provideCryptoService(): CryptoService {
        return CryptoService()
    }

    @Provides
    @Singleton
    fun provideKeystoreService(
        @ApplicationContext context: Context
    ): KeystoreService {
        return KeystoreService(context)
    }

    @Provides
    @Singleton
    fun provideBlockchainService(): BlockchainService {
        return BlockchainService()
    }

    @Provides
    @Singleton
    fun provideMultiChainApiService(): MultiChainApiService {
        return MultiChainApiService()
    }

    @Provides
    @Singleton
    fun providePriceService(): PriceService {
        return PriceService()
    }

    @Provides
    @Singleton
    fun provideCosmosTransactionBuilder(
        cryptoService: CryptoService
    ): CosmosTransactionBuilder {
        return CosmosTransactionBuilder(cryptoService)
    }

    @Provides
    @Singleton
    fun provideSettingsPreferences(
        @ApplicationContext context: Context
    ): SettingsPreferences {
        return SettingsPreferences(context)
    }
}
