package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

type FaucetService struct {
	clientCtx    client.Context
	dailyLimits  map[string]sdk.Coins
	requestCount map[string]time.Time
	config       *Config
}

type Config struct {
	ChainID         string
	NodeURL         string
	FaucetMnemonic  string
	DailyLimit      string
	RequestLimit    string
	Port            string
}

type FaucetRequest struct {
	Address string `json:"address"`
	Denom   string `json:"denom,omitempty"`
}

type FaucetResponse struct {
	Success bool   `json:"success"`
	TxHash  string `json:"tx_hash,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type StatusResponse struct {
	ChainID        string            `json:"chain_id"`
	FaucetAddress  string            `json:"faucet_address"`
	AvailableCoins sdk.Coins         `json:"available_coins"`
	DailyLimit     string            `json:"daily_limit"`
	RequestLimit   string            `json:"request_limit"`
	Uptime         string            `json:"uptime"`
	RequestStats   map[string]int    `json:"request_stats"`
}

var startTime = time.Now()

func loadConfig() *Config {
	return &Config{
		ChainID:        getEnv("CHAIN_ID", "sharehodl-local-1"),
		NodeURL:        getEnv("NODE_URL", "http://localhost:1317"),
		FaucetMnemonic: getEnv("FAUCET_MNEMONIC", "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon art"),
		DailyLimit:     getEnv("DAILY_LIMIT", "1000000000hodl,100000000shodl"),
		RequestLimit:   getEnv("REQUEST_LIMIT", "100000000hodl,10000000shodl"),
		Port:           getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (f *FaucetService) setupClientContext() error {
	// Create codec
	cdc := codec.NewProtoCodec(nil)
	
	// Create keyring
	kr := keyring.NewInMemory(cdc)
	
	// Import faucet account from mnemonic
	_, err := kr.NewAccount("faucet", f.config.FaucetMnemonic, "", sdk.FullFundraiserPath, hd.Secp256k1)
	if err != nil {
		return fmt.Errorf("failed to create faucet account: %w", err)
	}

	// Setup client context
	f.clientCtx = client.Context{}.
		WithCodec(cdc).
		WithChainID(f.config.ChainID).
		WithKeyring(kr).
		WithBroadcastMode("block")

	return nil
}

func (f *FaucetService) isRateLimited(address string) bool {
	if lastRequest, exists := f.requestCount[address]; exists {
		return time.Since(lastRequest) < 24*time.Hour
	}
	return false
}

func (f *FaucetService) parseCoins(coinStr string) (sdk.Coins, error) {
	if coinStr == "" {
		return sdk.Coins{}, nil
	}
	
	var coins sdk.Coins
	coinStrs := strings.Split(coinStr, ",")
	
	for _, coinStr := range coinStrs {
		coin, err := sdk.ParseCoinNormalized(strings.TrimSpace(coinStr))
		if err != nil {
			return nil, err
		}
		coins = coins.Add(coin)
	}
	
	return coins, nil
}

func (f *FaucetService) sendTokens(ctx context.Context, toAddress string, coins sdk.Coins) (string, error) {
	// Get faucet address
	faucetInfo, err := f.clientCtx.Keyring.Key("faucet")
	if err != nil {
		return "", err
	}

	faucetAddr, err := faucetInfo.GetAddress()
	if err != nil {
		return "", err
	}

	// Parse recipient address
	recipientAddr, err := sdk.AccAddressFromBech32(toAddress)
	if err != nil {
		return "", err
	}

	// Create send message
	msg := banktypes.NewMsgSend(faucetAddr, recipientAddr, coins)

	// Create transaction factory
	txf := tx.Factory{}.
		WithChainID(f.config.ChainID).
		WithKeybase(f.clientCtx.Keyring).
		WithTxConfig(f.clientCtx.TxConfig).
		WithAccountRetriever(f.clientCtx.AccountRetriever).
		WithSignMode(f.clientCtx.TxConfig.SignModeHandler().DefaultMode())

	// Build and sign transaction
	txBuilder, err := tx.BuildUnsignedTx(txf, msg)
	if err != nil {
		return "", err
	}

	err = tx.Sign(txf, "faucet", txBuilder, true)
	if err != nil {
		return "", err
	}

	// Broadcast transaction
	txBytes, err := f.clientCtx.TxConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return "", err
	}

	res, err := f.clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return "", err
	}

	if res.Code != 0 {
		return "", fmt.Errorf("transaction failed: %s", res.RawLog)
	}

	return res.TxHash, nil
}

func (f *FaucetService) handleFaucetRequest(w http.ResponseWriter, r *http.Request) {
	var req FaucetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		f.writeErrorResponse(w, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Validate address format
	if _, err := sdk.AccAddressFromBech32(req.Address); err != nil {
		f.writeErrorResponse(w, http.StatusBadRequest, "Invalid address format", err)
		return
	}

	// Check rate limiting
	if f.isRateLimited(req.Address) {
		f.writeErrorResponse(w, http.StatusTooManyRequests, "Daily limit exceeded. Try again in 24 hours", nil)
		return
	}

	// Parse request limit
	requestCoins, err := f.parseCoins(f.config.RequestLimit)
	if err != nil {
		f.writeErrorResponse(w, http.StatusInternalServerError, "Invalid faucet configuration", err)
		return
	}

	// Filter by requested denomination if specified
	var coinsToSend sdk.Coins
	if req.Denom != "" {
		for _, coin := range requestCoins {
			if coin.Denom == req.Denom {
				coinsToSend = sdk.NewCoins(coin)
				break
			}
		}
		if coinsToSend.IsZero() {
			f.writeErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Denomination %s not available", req.Denom), nil)
			return
		}
	} else {
		coinsToSend = requestCoins
	}

	// Send tokens
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	txHash, err := f.sendTokens(ctx, req.Address, coinsToSend)
	if err != nil {
		f.writeErrorResponse(w, http.StatusInternalServerError, "Failed to send tokens", err)
		return
	}

	// Update rate limiting
	f.requestCount[req.Address] = time.Now()

	// Send success response
	response := FaucetResponse{
		Success: true,
		TxHash:  txHash,
		Message: fmt.Sprintf("Successfully sent %s to %s", coinsToSend.String(), req.Address),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	
	log.Printf("Sent %s to %s (tx: %s)", coinsToSend.String(), req.Address, txHash)
}

func (f *FaucetService) handleStatus(w http.ResponseWriter, r *http.Request) {
	// Get faucet address
	faucetInfo, err := f.clientCtx.Keyring.Key("faucet")
	if err != nil {
		f.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get faucet info", err)
		return
	}

	faucetAddr, err := faucetInfo.GetAddress()
	if err != nil {
		f.writeErrorResponse(w, http.StatusInternalServerError, "Failed to get faucet address", err)
		return
	}

	// Calculate uptime
	uptime := time.Since(startTime).String()

	// Count requests in the last 24 hours
	now := time.Now()
	recentRequests := 0
	for _, lastRequest := range f.requestCount {
		if now.Sub(lastRequest) < 24*time.Hour {
			recentRequests++
		}
	}

	response := StatusResponse{
		ChainID:       f.config.ChainID,
		FaucetAddress: faucetAddr.String(),
		DailyLimit:    f.config.DailyLimit,
		RequestLimit:  f.config.RequestLimit,
		Uptime:        uptime,
		RequestStats:  map[string]int{"last_24h": recentRequests},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (f *FaucetService) writeErrorResponse(w http.ResponseWriter, statusCode int, message string, err error) {
	var errorDetail string
	if err != nil {
		errorDetail = err.Error()
		log.Printf("Error: %s - %v", message, err)
	}

	response := FaucetResponse{
		Success: false,
		Message: message,
		Error:   errorDetail,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Load configuration
	config := loadConfig()

	// Initialize faucet service
	faucet := &FaucetService{
		config:       config,
		dailyLimits:  make(map[string]sdk.Coins),
		requestCount: make(map[string]time.Time),
	}

	// Setup client context
	if err := faucet.setupClientContext(); err != nil {
		log.Fatalf("Failed to setup client context: %v", err)
	}

	// Setup routes
	r := mux.NewRouter()
	r.HandleFunc("/faucet", faucet.handleFaucetRequest).Methods("POST")
	r.HandleFunc("/status", faucet.handleStatus).Methods("GET")
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(r)

	// Start server
	log.Printf("Starting faucet service on port %s", config.Port)
	log.Printf("Chain ID: %s", config.ChainID)
	log.Printf("Node URL: %s", config.NodeURL)
	
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}