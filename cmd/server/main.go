package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/tusharr/go-ai-huggingface/internal/ai"
	"github.com/tusharr/go-ai-huggingface/internal/config"
	"github.com/tusharr/go-ai-huggingface/internal/handler"
	"github.com/tusharr/go-ai-huggingface/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logLevel := logger.ParseLogLevel(cfg.Logger.Level)
	appLogger := logger.NewLogger(logLevel, cfg.Logger.Structured)

	ctx := context.Background()
	appLogger.Info(ctx, "Starting go-ai-huggingface server", map[string]interface{}{
		"version":   "1.0.0",
		"port":      cfg.Server.Port,
		"log_level": cfg.Logger.Level,
		"model":     cfg.HuggingFace.DefaultModel,
	})

	// Initialize services
	aiService := ai.NewHuggingFaceService(&cfg.HuggingFace, appLogger)
	aiHandler := handler.NewAIHandler(aiService, appLogger)

	// Setup routes
	mux := setupRoutes(aiHandler)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		appLogger.Info(ctx, "Server starting", map[string]interface{}{
			"address": server.Addr,
		})

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Error(ctx, "Server failed to start", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
	}()

	appLogger.Info(ctx, "Server started successfully", map[string]interface{}{
		"address": server.Addr,
	})

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info(ctx, "Server is shutting down...", nil)

	// Create a context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.GracefulShutdownTimeout)
	defer cancel()

	// Gracefully shutdown the server
	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Error(ctx, "Server forced to shutdown", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	appLogger.Info(ctx, "Server exited properly", nil)
}

// setupRoutes configures all HTTP routes and middleware
func setupRoutes(aiHandler *handler.AIHandler) http.Handler {
	mux := http.NewServeMux()

	// Health and monitoring endpoints
	mux.HandleFunc("/health", aiHandler.Health)
	mux.HandleFunc("/metrics", aiHandler.Metrics)

	// AI endpoints
	mux.HandleFunc("/v1/text/generate", aiHandler.GenerateText)
	mux.HandleFunc("/v1/text/complete", aiHandler.GenerateCompletion)
	mux.HandleFunc("/v1/text/sentiment", aiHandler.AnalyzeSentiment)
	mux.HandleFunc("/v1/text/summarize", aiHandler.SummarizeText)
	mux.HandleFunc("/v1/models/validate", aiHandler.ValidateModel)

	// API documentation endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"service": "go-ai-huggingface",
			"version": "1.0.0",
			"endpoints": map[string]interface{}{
				"health":            "GET /health",
				"metrics":           "GET /metrics",
				"generate_text":     "POST /v1/text/generate",
				"complete_text":     "POST /v1/text/complete",
				"analyze_sentiment": "POST /v1/text/sentiment",
				"summarize_text":    "POST /v1/text/summarize",
				"validate_model":    "GET /v1/models/validate?model=<model_name>",
			},
			"documentation": "https://github.com/tusharr/go-ai-huggingface",
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	})

	// Apply middleware (order matters!)
	var handler http.Handler = mux

	// Apply CORS middleware
	handler = aiHandler.EnableCORS(handler)

	// Apply rate limiting from configuration
	handler = aiHandler.RateLimiter(cfg.HuggingFace.RateLimitRPM)(handler)

	// Apply request logging (should be last/outermost)
	handler = aiHandler.RequestLogger(handler)

	return handler
}
