package main

import (
        "context"
        "fmt"
        "net/http"
        "os"
        "os/signal"
        "syscall"
        "time"
)

func main() {
        // Load configuration
        config, err := loadConfig()
        if err != nil {
                fmt.Printf("[ERROR] Failed to load configuration: %v\n", err)
                os.Exit(1)
        }

        // Initialize logger
        logger := NewLogger(config.LogLevel)

        // Log startup information
        logger.LogStartup("FREEBET.GURU Go API", fmt.Sprintf("%d", config.Port))
        logger.LogInfo("Environment: %s", config.Env)

        // Initialize database
        db, err := NewPostgresDB(config.DatabaseURL, config, logger)
        if err != nil {
                logger.LogError("Failed to connect to database: %s", err.Error())
                os.Exit(1)
        }
        defer db.Close()

        // Test database connection
        if err := db.Ping(); err != nil {
                logger.LogError("Database ping failed: %s", err.Error())
                os.Exit(1)
        }
        logger.LogSuccess("Database connection established")

        // Log database statistics on startup
        stats, err := db.GetDatabaseStats()
        if err == nil {
                logger.LogSystem("DATABASE", "Initial stats - Users: %d, Sessions: %d, Bets: %d, Matches: %d",
                        stats["users"], stats["sessions"], stats["bets"], stats["matches"])
        } else {
                logger.LogWarning("Failed to get initial database stats: %s", err.Error())
        }

        // Setup routes with logging middleware
        router := SetupRoutes(db, config, logger)
        
        // Wrap with logging middleware
        handler := logger.Middleware(router)

        // Create HTTP server
        server := &http.Server{
                Addr:         fmt.Sprintf(":%d", config.Port),
                Handler:      handler,
                ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
                WriteTimeout: time.Duration(config.WriteTimeout) * time.Second,
                IdleTimeout:  time.Duration(config.IdleTimeout) * time.Second,
        }

        // Start server in a goroutine
        go func() {
                logger.LogInfo("Server starting on port %d", config.Port)
                
                if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                        logger.LogError("Server failed to start: %s", err.Error())
                        os.Exit(1)
                }
        }()

        // Setup graceful shutdown
        quit := make(chan os.Signal, 1)
        signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

        // Wait for interrupt signal
        <-quit
        logger.LogWarning("Shutdown signal received, shutting down gracefully...")

        // Give outstanding requests 30 seconds to complete
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        // Attempt graceful shutdown
        if err := server.Shutdown(ctx); err != nil {
                logger.LogError("Server forced to shutdown: %s", err.Error())
                os.Exit(1)
        }

        // Log final metrics and shutdown info
        logger.LogMetrics()
        logger.LogShutdown()
        logger.LogSuccess("Server shutdown complete")
}
