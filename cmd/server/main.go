/*-------------------------------------------------------------------------
 *
 * Imagineer - TTRPG Campaign Intelligence Platform
 *
 * Copyright (c) 2025 - 2026
 * This software is released under The MIT License
 *
 *-------------------------------------------------------------------------
 */

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/antonypegg/imagineer/internal/api"
	"github.com/antonypegg/imagineer/internal/auth"
	"github.com/antonypegg/imagineer/internal/database"
)

func main() {
	fmt.Println("Imagineer - TTRPG Campaign Intelligence Platform")
	fmt.Println("Version: 0.1.0-dev")

	// Get configuration from environment or defaults
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	configPath := os.Getenv("DB_CONFIG")
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	// Load database configuration
	dbConfig, err := database.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load database config: %v", err)
	}

	// Connect to database
	ctx := context.Background()
	db, err := database.Connect(ctx, dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database")

	// Initialize OAuth handler and JWT secret if configuration is available
	var authHandler *auth.AuthHandler
	jwtSecret := os.Getenv("JWT_SECRET")
	oauthConfig, err := auth.NewOAuthConfigFromEnv()
	if err != nil {
		log.Printf("OAuth not configured: %v (authentication endpoints will be disabled)", err)
	} else {
		if jwtSecret == "" {
			log.Printf("JWT_SECRET not set (authentication endpoints will be disabled)")
		} else {
			jwtExpiryHrs := 24 // Default to 24 hours
			if expiryStr := os.Getenv("JWT_EXPIRY_HOURS"); expiryStr != "" {
				if parsed, err := strconv.Atoi(expiryStr); err == nil && parsed > 0 {
					jwtExpiryHrs = parsed
				}
			}
			authHandler = auth.NewAuthHandler(db, oauthConfig, []byte(jwtSecret), jwtExpiryHrs)
			log.Println("OAuth authentication enabled")
		}
	}

	// Log authentication middleware status
	if jwtSecret != "" {
		log.Println("JWT authentication middleware enabled for protected routes")
	} else {
		log.Println("JWT_SECRET not set - protected routes will be unauthenticated")
	}

	// Create router
	router := api.NewRouter(db, authHandler, jwtSecret)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
