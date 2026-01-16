package main

import (
        "context"
        "encoding/base64"
        "fmt"
        "net/http"
        "regexp"
        "strings"
        "sync"
        "time"

        "github.com/gorilla/handlers"
        "golang.org/x/crypto/bcrypt"
)

// contextKey type for context keys
type contextKey string

const (
        userContextKey contextKey = "user"
)

// CORS middleware with custom origin checking
func corsMiddleware(config *Config) func(http.Handler) http.Handler {
        // Compile regex patterns for allowed origins (supporting wildcards)
        var allowedPatterns []*regexp.Regexp
        for _, origin := range config.CORSAllowedOrigins {
                // Handle wildcard patterns
                if strings.Contains(origin, "*") {
                        // Convert wildcard to regex
                        pattern := strings.ReplaceAll(origin, "*", ".*")
                        if regex, err := regexp.Compile("^" + pattern + "$"); err == nil {
                                allowedPatterns = append(allowedPatterns, regex)
                        }
                } else {
                        // Exact match
                        pattern := "^" + regexp.QuoteMeta(origin) + "$"
                        if regex, err := regexp.Compile(pattern); err == nil {
                                allowedPatterns = append(allowedPatterns, regex)
                        }
                }
        }

        // Custom origin checker that supports wildcards
        originChecker := func(origin string) bool {
                for _, pattern := range allowedPatterns {
                        if pattern.MatchString(origin) {
                                return true
                        }
                }
                return false
        }

        return handlers.CORS(
                handlers.AllowCredentials(), // Allow cookies
                handlers.AllowedOriginValidator(originChecker), // Use validator for wildcards
                handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
                handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
        )
}

// JWT Auth middleware - checks for valid JWT access token
func jwtAuthMiddleware(db Database, config *Config, logger *Logger) func(http.Handler) http.Handler {
        return func(next http.Handler) http.Handler {
                return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                        // Get JWT token from Authorization header
                        authHeader := r.Header.Get("Authorization")
                        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
                                logger.LogWarning("[JWT AUTH] No JWT token found in Authorization header")
                                http.Error(w, `{"success": false, "error": "No access token"}`, http.StatusUnauthorized)
                                return
                        }

                        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

                        // Validate JWT token
                        claims, err := validateAccessToken(tokenString, config)
                        if err != nil {
                                logger.LogError("[JWT AUTH] Invalid JWT token: %s", err.Error())
                                http.Error(w, `{"success": false, "error": "Invalid access token"}`, http.StatusUnauthorized)
                                return
                        }

                        // Get user data
                        user, err := db.GetUserByID(claims.UserID)
                        if err != nil {
                                logger.LogError("[JWT AUTH] Failed to get user data for user %s: %s", claims.UserID, err.Error())
                                http.Error(w, `{"success": false, "error": "User not found"}`, http.StatusInternalServerError)
                                return
                        }

                        logger.LogInfo("[JWT AUTH] JWT valid for user: %s", user.Nickname)

                        // Add user to request context
                        ctx := context.WithValue(r.Context(), userContextKey, user)
                        next.ServeHTTP(w, r.WithContext(ctx))
                })
        }
}

// Get user from context (helper function)
func getUserFromContext(ctx context.Context) (*User, bool) {
        user, ok := ctx.Value(userContextKey).(*User)
        return user, ok
}

// Admin context key
const (
        adminContextKey contextKey = "admin"
)

// Admin auth middleware - checks for valid Basic Auth admin credentials
func adminAuthMiddleware(db Database, logger *Logger) func(http.Handler) http.Handler {
        return func(next http.Handler) http.Handler {
                return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                        // Get Basic Auth header
                        authHeader := r.Header.Get("Authorization")
                        if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
                                logger.LogWarning("[ADMIN AUTH] Missing Basic Auth header")
                                http.Error(w, `{"ok": false, "error": "Unauthorized", "message": "Basic authentication required"}`, http.StatusUnauthorized)
                                return
                        }

                        // Decode Basic Auth
                        encoded := strings.TrimPrefix(authHeader, "Basic ")
                        decoded, err := base64.StdEncoding.DecodeString(encoded)
                        if err != nil {
                                logger.LogWarning("[ADMIN AUTH] Invalid base64 encoding: %s", err.Error())
                                http.Error(w, `{"ok": false, "error": "Unauthorized", "message": "Invalid authentication encoding"}`, http.StatusUnauthorized)
                                return
                        }

                        // Parse username:password
                        parts := strings.SplitN(string(decoded), ":", 2)
                        if len(parts) != 2 {
                                logger.LogWarning("[ADMIN AUTH] Invalid Basic Auth format")
                                http.Error(w, `{"ok": false, "error": "Unauthorized", "message": "Invalid authentication format"}`, http.StatusUnauthorized)
                                return
                        }

                        username := parts[0]
                        password := parts[1]

                        logger.LogAuth("[ADMIN AUTH] Attempting authentication for admin: %s", username)

                        // Get admin from database
                        admin, err := db.GetAdminByUsername(username)
                        if err != nil {
                                logger.LogWarning("[ADMIN AUTH] Admin not found: %s", username)
                                http.Error(w, `{"ok": false, "error": "Unauthorized", "message": "Invalid username or password"}`, http.StatusUnauthorized)
                                return
                        }

                        // Verify password
                        err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password))
                        if err != nil {
                                logger.LogWarning("[ADMIN AUTH] Invalid password for admin: %s", username)
                                http.Error(w, `{"ok": false, "error": "Unauthorized", "message": "Invalid username or password"}`, http.StatusUnauthorized)
                                return
                        }

                        // Update last login
                        if err := db.UpdateAdminLastLogin(admin.ID); err != nil {
                                logger.LogWarning("[ADMIN AUTH] Failed to update last login: %s", err.Error())
                                // Don't fail the request, just log
                        }

                        logger.LogSuccess("[ADMIN AUTH] Admin authenticated: %s", admin.Username)

                        // Add admin to request context
                        ctx := context.WithValue(r.Context(), adminContextKey, admin)
                        next.ServeHTTP(w, r.WithContext(ctx))
                })
        }
}

// Get admin from context (helper function)
func getAdminFromContext(ctx context.Context) (*Admin, bool) {
        admin, ok := ctx.Value(adminContextKey).(*Admin)
        return admin, ok
}

// Content-Type middleware - ensures JSON content type for API responses
func contentTypeMiddleware(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Header().Set("Content-Type", "application/json")
                next.ServeHTTP(w, r)
        })
}

// Recovery middleware - catches panics and returns 500
func recoveryMiddleware(logger *Logger) func(http.Handler) http.Handler {
        return func(next http.Handler) http.Handler {
                return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                        defer func() {
                                if err := recover(); err != nil {
                                        logger.LogError("[RECOVERY] Panic recovered:", err)
                                        http.Error(w, `{"success": false, "error": "Internal server error"}`, http.StatusInternalServerError)
                                }
                        }()
                        next.ServeHTTP(w, r)
                })
        }
}

// Request ID middleware - adds unique request ID to each request
func requestIDMiddleware(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                // Request ID is handled in logger middleware, just pass through
                next.ServeHTTP(w, r)
        })
}

// Security headers middleware
func securityHeadersMiddleware(config *Config) func(http.Handler) http.Handler {
        return func(next http.Handler) http.Handler {
                return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                        // Security headers
                        w.Header().Set("X-Content-Type-Options", "nosniff")
                        w.Header().Set("X-Frame-Options", "DENY")
                        w.Header().Set("X-XSS-Protection", "1; mode=block")

                        // HSTS in production (configurable max-age)
                        if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
                                w.Header().Set("Strict-Transport-Security", fmt.Sprintf("max-age=%d; includeSubDomains", config.HSTSMaxAge))
                        }

                        next.ServeHTTP(w, r)
                })
        }
}

// Rate limiting middleware (basic implementation)
func rateLimitMiddleware(config *Config, logger *Logger) func(http.Handler) http.Handler {
        // Simple in-memory rate limiter (for demo purposes)
        // In production, use Redis or similar
        var mu sync.RWMutex
        requests := make(map[string]int)
        resetTime := make(map[string]int64)

        return func(next http.Handler) http.Handler {
                return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                        // Get client IP
                        clientIP := r.RemoteAddr
                        if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
                                clientIP = forwarded
                        } else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
                                clientIP = realIP
                        }

                        // Rate limiting with configurable window and requests
                        now := time.Now().Unix()
                        windowStart := now - int64(config.RateLimitWindow) // Configurable window

                        mu.Lock()
                        // Reset counter if window expired
                        if resetTime[clientIP] < windowStart {
                                requests[clientIP] = 0
                                resetTime[clientIP] = now
                        }

                        // Check rate limit
                        if requests[clientIP] >= config.RateLimitRequests {
                                mu.Unlock()
                                logger.LogWarning("[RATE LIMIT] Rate limit exceeded for IP:", clientIP)
                                http.Error(w, `{"success": false, "error": "Rate limit exceeded"}`, http.StatusTooManyRequests)
                                return
                        }

                        requests[clientIP]++
                        mu.Unlock()
                        next.ServeHTTP(w, r)
                })
        }
}
