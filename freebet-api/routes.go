package main

import (
        "net/http"

        "github.com/gorilla/mux"
)

// SetupRoutes configures all routes and middleware
func SetupRoutes(db Database, config *Config, logger *Logger) *mux.Router {
        // Create router
        router := mux.NewRouter()

        // Create handler instance
        handler := NewHandler(db, config, logger)

        // Apply global middleware (excluding logging which is handled in main.go)
        router.Use(mux.MiddlewareFunc(contentTypeMiddleware)) // JSON content type
        router.Use(mux.MiddlewareFunc(securityHeadersMiddleware(config))) // Security headers
        router.Use(mux.MiddlewareFunc(corsMiddleware(config))) // CORS
        router.Use(mux.MiddlewareFunc(recoveryMiddleware(logger))) // Panic recovery
        router.Use(mux.MiddlewareFunc(rateLimitMiddleware(config, logger))) // Rate limiting

        // Root endpoint (no auth required)
        router.HandleFunc("/", handler.rootHandler).Methods("GET")

        // API routes
        api := router.PathPrefix("/api").Subrouter()
        api.HandleFunc("/health", handler.healthHandler).Methods("GET")
        // api.HandleFunc("/analytics", handler.analyticsHandler).Methods("GET") // Temporarily disabled

        // Auth routes (no auth required - handle JWT validation internally)
        auth := api.PathPrefix("/auth").Subrouter()
        auth.HandleFunc("/register", handler.registerHandler).Methods("POST")
        auth.HandleFunc("/login", handler.loginHandler).Methods("POST")
        auth.HandleFunc("/user", handler.userHandler).Methods("GET")          // Validates JWT access token
        auth.HandleFunc("/logout", handler.logoutHandler).Methods("POST")     // Clears refresh token cookie
        auth.HandleFunc("/refresh", handler.refreshTokenHandler).Methods("POST") // Refreshes access token
        auth.HandleFunc("/topup", handler.topupHandler).Methods("POST")       // Validates JWT access token
        auth.HandleFunc("/change-password", handler.changePasswordHandler).Methods("POST") // Validates JWT access token

        // Google OAuth routes
        auth.HandleFunc("/google", handler.googleLoginHandler).Methods("GET")      // Initiates OAuth flow
        auth.HandleFunc("/google/callback", handler.googleCallbackHandler).Methods("GET") // OAuth callback

        // Bets routes (handle session check internally like Node.js)
        api.HandleFunc("/bets", handler.getBetsHandler).Methods("GET")
        api.HandleFunc("/bets", handler.placeBetHandler).Methods("POST")

        // Matches routes (no auth required)
        api.HandleFunc("/matches", handler.getMatchesHandler).Methods("GET")

        // Players routes (no auth required)
        api.HandleFunc("/players", handler.getPlayersHandler).Methods("GET")

        // Admin sync routes (require admin auth)
        adminSync := api.PathPrefix("").Subrouter()
        adminSync.Use(mux.MiddlewareFunc(adminAuthMiddleware(db, logger)))
        adminSync.HandleFunc("/odds/sync", handler.oddsSyncHandler).Methods("POST")
        adminSync.HandleFunc("/scores/sync", handler.scoresSyncHandler).Methods("POST")
        adminSync.HandleFunc("/calc", handler.calcHandler).Methods("POST")

        // Add OPTIONS handler for CORS preflight requests
        router.Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
        })

        return router
}
