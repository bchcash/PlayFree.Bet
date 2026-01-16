package main

import (
        "fmt"
        "os"
        "strconv"
        "strings"
        "time"

        "github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
        // Server configuration
        Port    int    `json:"port"`
        Env     string `json:"env"`
        LogLevel string `json:"log_level"`

        // Database configuration
        DatabaseURL string `json:"database_url"`

        // Authentication configuration
        BcryptCost           int           `json:"bcrypt_cost"`
        JWTSecret            string        `json:"jwt_secret"`
        JWTAccessTokenTTL    time.Duration `json:"jwt_access_token_ttl"`
        JWTRefreshTokenTTL   time.Duration `json:"jwt_refresh_token_ttl"`
        CookieName           string        `json:"cookie_name"`         // For refresh tokens
        CookieSecure         bool          `json:"cookie_secure"`
        CookieHTTPOnly       bool          `json:"cookie_http_only"`
        CookieSameSite       string        `json:"cookie_same_site"`

        // Game/Business logic constants
        InitialBalance     float64 `json:"initial_balance"`
        TopupAmount        float64 `json:"topup_amount"`
        MaxTopupBalance    float64 `json:"max_topup_balance"`
        MinPasswordLength  int     `json:"min_password_length"`

        // Betting limits
        MinBetAmount      float64 `json:"min_bet_amount"`
        MaxBetAmount      float64 `json:"max_bet_amount"`

        // CORS configuration
        CORSAllowedOrigins []string `json:"cors_allowed_origins"`
        CORSCredentials    bool     `json:"cors_credentials"`

        // Pagination defaults
        DefaultPlayerLimit int `json:"default_player_limit"`
        MaxPlayerLimit     int `json:"max_player_limit"`

        // Server timeouts (seconds)
        ReadTimeout       int `json:"read_timeout"`
        WriteTimeout      int `json:"write_timeout"`
        IdleTimeout       int `json:"idle_timeout"`

        // Rate limiting
        RateLimitRequests int `json:"rate_limit_requests"`
        RateLimitWindow   int `json:"rate_limit_window"`

        // Database connection pool
        DBMaxConns        int `json:"db_max_conns"`
        DBMinConns        int `json:"db_min_conns"`
        DBMaxLifetime     int `json:"db_max_lifetime"`
        DBMaxIdleTime     int `json:"db_max_idle_time"`

        // HSTS configuration
        HSTSMaxAge        int `json:"hsts_max_age"`

        // Odds API configuration
        OddsAPIKey        string `json:"odds_api_key"`

        // Google OAuth configuration
        GoogleClientID     string `json:"google_client_id"`
        GoogleClientSecret string `json:"google_client_secret"`
        GoogleRedirectURL  string `json:"google_redirect_url"`

        // Telegram configuration
        TelegramBotToken  string `json:"telegram_bot_token"`
        TelegramChannelID string `json:"telegram_channel_id"`
}

// loadConfig loads configuration from environment variables with defaults
func loadConfig() (*Config, error) {
        // Load .env file if it exists (ignore error if file doesn't exist)
        godotenv.Load()

        config := &Config{
                // Server defaults
                Port:      getEnvInt("API_PORT", 3001),
                Env:       getEnvString("NODE_ENV", "development"),
                LogLevel:  getEnvString("LOG_LEVEL", "INFO"),

                // Database (required) - prefer EXTERNAL_DATABASE_URL if set
                DatabaseURL: getEnvStringWithFallback("EXTERNAL_DATABASE_URL", "DATABASE_URL", ""),

                // Authentication defaults (from environment)
                BcryptCost:           getEnvInt("BCRYPT_COST", 12), // bcrypt.DefaultCost is 10, we use 12 for better security
                JWTSecret:            getEnvString("JWT_SECRET", "your-super-secret-jwt-key-change-in-production"), // Must be set in production
                JWTAccessTokenTTL:    getEnvDuration("JWT_ACCESS_TOKEN_TTL", 15*time.Minute), // 15 minutes
                JWTRefreshTokenTTL:   getEnvDuration("JWT_REFRESH_TOKEN_TTL", 7*24*time.Hour), // 7 days
                CookieName:           getEnvString("COOKIE_NAME", "refresh_token"), // Changed from session_token
                CookieSecure:         getEnvBool("COOKIE_SECURE", false), // true in production
                CookieHTTPOnly:       getEnvBool("COOKIE_HTTP_ONLY", true), // Always true for security
                CookieSameSite:       getEnvString("COOKIE_SAME_SITE", "strict"), // CSRF protection: "strict", "lax", "none"

                // Game/Business logic constants (from environment)
                InitialBalance:     getEnvFloat64("INITIAL_BALANCE", 10000.0), // $10,000 starting balance
                TopupAmount:        getEnvFloat64("TOPUP_AMOUNT", 10000.0), // $10,000 topup amount
                MaxTopupBalance:   getEnvFloat64("MAX_TOPUP_BALANCE", 500.0), // Can only topup if balance < $500
                MinPasswordLength:  getEnvInt("MIN_PASSWORD_LENGTH", 6), // Minimum password length

                // Betting limits (from environment)
                MinBetAmount:       getEnvFloat64("MIN_BET_AMOUNT", 1.0), // Minimum bet amount
                MaxBetAmount:       getEnvFloat64("MAX_BET_AMOUNT", 100000.0), // Maximum bet amount

                // CORS configuration from environment
                CORSAllowedOrigins: getEnvCORSOrigins("CORS_ALLOWED_ORIGINS",
                        // Default values for development (with wildcard support)
                        []string{
                                "http://localhost:*",                    // Localhost with any port
                                "https://*.freebet.guru",                // All freebet.guru subdomains
                                "https://*.xn--80adjb6a.xn--p1ai",       // All лудик.рф subdomains (IDN)
                                "https://*.repl.co",                     // Replit domains
                                "https://*.replit.dev",                  // Replit dev domains
                                "https://*.replit.app",                  // Replit app domains
                                "https://*.picard.replit.dev",           // Replit picard subdomains
                        }),
                CORSCredentials: getEnvBool("CORS_CREDENTIALS", true), // Allow cookies/credentials

                // Pagination defaults (from environment)
                DefaultPlayerLimit: getEnvInt("PAGINATION_DEFAULT_LIMIT", 50),
                MaxPlayerLimit:     getEnvInt("PAGINATION_MAX_LIMIT", 100),

                // Server timeouts (seconds, from environment)
                ReadTimeout:        getEnvInt("READ_TIMEOUT", 15),
                WriteTimeout:       getEnvInt("WRITE_TIMEOUT", 15),
                IdleTimeout:        getEnvInt("IDLE_TIMEOUT", 60),

                // Rate limiting (from environment)
                RateLimitRequests:  getEnvInt("RATE_LIMIT_REQUESTS", 100), // Requests per window
                RateLimitWindow:    getEnvInt("RATE_LIMIT_WINDOW", 60),    // Window in seconds

                // Database connection pool (from environment)
                DBMaxConns:         getEnvInt("DB_MAX_CONNS", 10),
                DBMinConns:         getEnvInt("DB_MIN_CONNS", 1),
                DBMaxLifetime:      getEnvInt("DB_MAX_LIFETIME", 3600),     // 1 hour in seconds
                DBMaxIdleTime:      getEnvInt("DB_MAX_IDLE_TIME", 1800),    // 30 minutes in seconds

                // HSTS configuration (from environment)
                HSTSMaxAge:         getEnvInt("HSTS_MAX_AGE", 31536000), // 1 year in seconds

                // Odds API configuration (from environment)
                OddsAPIKey:         getEnvString("ODDS_API_KEY", ""),

                // Google OAuth configuration (from environment)
                GoogleClientID:     getEnvString("GOOGLE_CLIENT_ID", ""),
                GoogleClientSecret: getEnvString("GOOGLE_CLIENT_SECRET", ""),
                GoogleRedirectURL:  getEnvString("GOOGLE_REDIRECT_URL", "http://localhost:3001/api/auth/google/callback"),

                // Telegram configuration (from environment)
                TelegramBotToken:   getEnvString("TELEGRAM_BOT_TOKEN", ""),
                TelegramChannelID:  getEnvString("TELEGRAM_CHANNEL_ID", ""),
        }

        // Validate required configuration
        if config.DatabaseURL == "" {
                return nil, fmt.Errorf("DATABASE_URL environment variable is required")
        }

        // Environment-specific overrides
        if config.Env == "production" {
                config.CookieSecure = true // HTTPS only in production
        }

        return config, nil
}

// Helper functions for environment variable parsing
func getEnvString(key, defaultValue string) string {
        if value := os.Getenv(key); value != "" {
                return value
        }
        return defaultValue
}

func getEnvStringWithFallback(primaryKey, fallbackKey, defaultValue string) string {
        if value := os.Getenv(primaryKey); value != "" {
                return value
        }
        if value := os.Getenv(fallbackKey); value != "" {
                return value
        }
        return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
        if value := os.Getenv(key); value != "" {
                if intValue, err := strconv.Atoi(value); err == nil {
                        return intValue
                }
        }
        return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
        if value := os.Getenv(key); value != "" {
                if boolValue, err := strconv.ParseBool(value); err == nil {
                        return boolValue
                }
        }
        return defaultValue
}

func getEnvFloat64(key string, defaultValue float64) float64 {
        if value := os.Getenv(key); value != "" {
                if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
                        return floatValue
                }
        }
        return defaultValue
}

// getEnvDuration parses duration from environment variable
// Format: "7d" (days), "24h" (hours), "60m" (minutes), "30s" (seconds)
// Examples: "7d", "24h", "60m", "30s", "168h" (7 days)
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
        if value := os.Getenv(key); value != "" {
                // Try parsing as Go duration format (e.g., "168h", "7d")
                if duration, err := time.ParseDuration(value); err == nil {
                        return duration
                }
                // Try parsing as days (e.g., "7d" -> 7 days)
                if strings.HasSuffix(value, "d") {
                        if days, err := strconv.Atoi(strings.TrimSuffix(value, "d")); err == nil {
                                return time.Duration(days) * 24 * time.Hour
                        }
                }
        }
        return defaultValue
}

// getEnvCORSOrigins parses CORS_ALLOWED_ORIGINS environment variable
// Format: comma-separated list of origins
// Example: "https://example.com,https://*.example.com,http://localhost:*"
func getEnvCORSOrigins(key string, defaultOrigins []string) []string {
        if value := os.Getenv(key); value != "" {
                // Parse comma-separated values, trim whitespace
                var origins []string
                for _, origin := range strings.Split(value, ",") {
                        origin = strings.TrimSpace(origin)
                        if origin != "" {
                                origins = append(origins, origin)
                        }
                }
                if len(origins) > 0 {
                        return origins
                }
        }
        return defaultOrigins
}
