package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/budgets/core/internal/config"
	"github.com/budgets/core/internal/database"
	"github.com/budgets/core/internal/encryption"
	"github.com/budgets/core/internal/secrets"
	"github.com/budgets/core/internal/server"
)

// @title Budget Management System API
// @version 1.0
// @description A complete Budget Management System with multi-user support, SSO authentication, and group-based data isolation.

// @contact.name API Support
// @contact.email support@budgets.example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @securityDefinitions.oauth2.accessCode OAuth2
// @authorizationUrl https://accounts.google.com/o/oauth2/auth
// @tokenUrl https://oauth2.googleapis.com/token
// @scope.email Grants access to user email
// @scope.profile Grants access to user profile

func main() {
	provider := secrets.GetProvider()
	cfg, err := config.Load(provider)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	encryptor, err := encryption.NewEncryptor(cfg.Auth.EncryptionKey.Value())
	if err != nil {
		log.Fatalf("Failed to initialize encryptor: %v", err)
	}

	deps := server.BuildDependencies(db.Pool, encryptor)
	srv := server.New(cfg, db, deps)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      srv.Router(),
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeoutSeconds) * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %d", cfg.Server.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeoutSeconds)*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
