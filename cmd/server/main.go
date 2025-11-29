package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"qwen3-compatibility/internal/config"
	"qwen3-compatibility/internal/handlers"
	"qwen3-compatibility/internal/middleware"
	"qwen3-compatibility/internal/services"
	"qwen3-compatibility/pkg/client"
)

var (
	cfg *config.Config

	// Version information
	Version   = "v0.0.1"
	GitCommit = "unknown"
	BuildTime = "unknown"
	GoVersion = runtime.Version()
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "qwen3-compatibility",
	Short: "Qwen3 Compatibility Server",
	Long: `A Go implementation providing OpenAI-compatible APIs for Qwen3 services.

Available Endpoints:
  POST /v1/audio/transcriptions  - Audio transcription using Qwen3 ASR models`,
	RunE: runServer,
}

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the compatibility server",
	Long: `Start the HTTP server providing OpenAI-compatible APIs for Qwen3 services.

Available Endpoints:
  POST /v1/audio/transcriptions  - Audio transcription using Qwen3 ASR models`,
	RunE: runServer,
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version, commit hash, build time, and Go version information.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("qwen3-compatibility %s\n", Version)
		fmt.Printf("Git commit: %s\n", GitCommit)
		fmt.Printf("Built: %s\n", BuildTime)
		fmt.Printf("Go version: %s\n", GoVersion)
	},
}

func init() {
	// Initialize flags on root command only
	config.InitializeFlags(rootCmd)
	// Server command inherits flags from root
	rootCmd.AddCommand(serverCmd)
	// Add version command
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	// Bind flags to viper (must happen before Load)
	config.BindFlags(cmd)

	// Load configuration
	loadedCfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := loadedCfg.Validate(); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	cfg = loadedCfg

	// Set gin mode
	if gin.Mode() == gin.DebugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create services and clients
	dashscopeClient := client.NewDashScopeClient(
		cfg.DashScope.Timeout,
	)

	uploadService := services.NewUploadService(dashscopeClient, &cfg.Upload)
	asrService := services.NewASRService(dashscopeClient)

	// Create handlers
	transcriptionHandler := handlers.NewTranscriptionHandler(uploadService, asrService, cfg)

	// Setup router
	router := setupRouter(transcriptionHandler)

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s", cfg.GetServerAddress())
		log.Printf("Max file size: %d bytes", cfg.Upload.MaxFileSize)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("Server shutdown complete")
	return nil
}

func setupRouter(transcriptionHandler *handlers.TranscriptionHandler) *gin.Engine {
	router := gin.New()

	// Add middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.CORS())

	// Setup routes
	api := router.Group("/v1")
	api.Use(middleware.AuthMiddleware()) // Add auth middleware to API routes
	{
		api.POST("/audio/transcriptions", transcriptionHandler.Transcription)
	}

	return router
}
