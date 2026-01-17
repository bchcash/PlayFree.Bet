package main

import (
        "context"
        "encoding/json"
        "fmt"
        "net"
        "net/http"
        "net/url"
        "regexp"
        "strconv"
        "strings"
        "time"

        "golang.org/x/crypto/bcrypt"
        "golang.org/x/oauth2"
)

// Handler struct contains dependencies
type Handler struct {
        db     Database
        config *Config
        logger *Logger
}

// NewHandler creates a new handler instance
func NewHandler(db Database, config *Config, logger *Logger) *Handler {
        return &Handler{
                db:     db,
                config: config,
                logger: logger,
        }
}

// validateEmail validates email format using regex
func validateEmail(email string) bool {
        emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
        return emailRegex.MatchString(email)
}

// Health check handler
func (h *Handler) healthHandler(w http.ResponseWriter, r *http.Request) {
        // Get database statistics
        stats, err := h.db.GetDatabaseStats()
        databaseStatus := "ok"
        if err != nil {
                h.logger.LogError("Failed to get database stats: %s", err.Error())
                databaseStatus = "error"
                stats = map[string]int{
                        "users":    0,
                        "sessions": 0,
                        "bets":     0,
                        "matches":  0,
                }
        }

        // Log database statistics
        h.logger.LogSystem("DATABASE", "Database stats - Users: %d, Sessions: %d, Bets: %d, Matches: %d",
                stats["users"], stats["sessions"], stats["bets"], stats["matches"])

        // Get real client IP (not local server IP)
        clientIP := h.getClientIP(r)

        // Calculate uptime in seconds
        uptimeSeconds := int64(time.Since(h.logger.startTime).Seconds())

        // Build response for mobile app
        response := HealthResponse{
                // Main fields
                Ok:            true,
                Status:        "ok",
                UptimeSeconds: uptimeSeconds,
                ClientIP:      clientIP,
                Time:          time.Now().Format(time.RFC3339),
                Version:       "1.0.0",

                // Statistics
                UsersCount:    stats["users"],
                BetsCount:     stats["bets"],
                MatchesCount:  stats["matches"],
                DatabaseStatus: databaseStatus,
                Port:          h.config.Port,
        }

        h.writeJSON(w, http.StatusOK, response)
}

// Root endpoint handler
func (h *Handler) rootHandler(w http.ResponseWriter, r *http.Request) {
        response := RootResponse{
                Message:   "FREEBET.GURU Go API Server",
                Endpoints: map[string]string{
                        "health":  "/health",
                        "auth":    "/api/auth/*",
                        "bets":    "/api/bets",
                        "matches": "/api/matches",
                        "players": "/api/players",
                },
        }

        h.writeJSON(w, http.StatusOK, response)
}

// AUTH HANDLERS

// Register handler
func (h *Handler) registerHandler(w http.ResponseWriter, r *http.Request) {
        h.logger.LogAuth("Processing registration request")

        var req RegisterRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                h.writeError(w, http.StatusBadRequest, "Invalid JSON")
                return
        }

        // Validate input
        if req.Email == "" || req.Password == "" || req.Nickname == "" {
                h.writeError(w, http.StatusBadRequest, "Email, password and nickname are required")
                return
        }

        // Validate email format
        if !validateEmail(req.Email) {
                h.writeError(w, http.StatusBadRequest, "Invalid email format")
                return
        }

        // Validate nickname length
        if len(req.Nickname) < 3 || len(req.Nickname) > 10 {
                h.writeError(w, http.StatusBadRequest, "Nickname must be between 3 and 10 characters")
                return
        }

        // Validate age confirmation
        if !req.AgeConfirmed {
                h.writeError(w, http.StatusBadRequest, "You must confirm that you are 18 years or older")
                return
        }

        if len(req.Password) < h.config.MinPasswordLength {
                h.writeError(w, http.StatusBadRequest, fmt.Sprintf("Password must be at least %d characters long", h.config.MinPasswordLength))
                return
        }

        // Check if user exists
        existingUser, _ := h.db.GetUserByEmail(req.Email)
        existingNickname, _ := h.db.GetUserByNickname(req.Nickname)
        if existingUser != nil || existingNickname != nil {
                var errorMsg string
                if existingUser != nil {
                        errorMsg = "User with this email already exists"
                } else {
                        errorMsg = "Nickname is already taken"
                }
                h.writeError(w, http.StatusBadRequest, errorMsg)
                return
        }

        // Hash password
        h.logger.LogAuth("Hashing password for new user: %s", req.Email)
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), h.config.BcryptCost)
        if err != nil {
                h.logger.LogError("Password hashing failed: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Registration failed")
                return
        }

        // Create user
        h.logger.LogAuth("Creating user record: %s", req.Email)
        user, err := h.db.CreateUser(req.Email, string(hashedPassword), req.Nickname, h.config.InitialBalance)
        if err != nil {
                h.logger.LogError("User creation failed: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Registration failed")
                return
        }

        // Generate JWT tokens
        h.logger.LogAuth("Generating JWT tokens for user: %s", user.ID)

        accessToken, err := generateAccessToken(user, h.config)
        if err != nil {
                h.logger.LogError("Access token generation failed: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Registration failed")
                return
        }

        refreshTokenString, err := generateRefreshToken(user.ID, h.config)
        if err != nil {
                h.logger.LogError("Refresh token generation failed: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Registration failed")
                return
        }

        // Store refresh token in database
        expiresAt := time.Now().Add(h.config.JWTRefreshTokenTTL)
        _, err = h.db.CreateRefreshToken(user.ID, refreshTokenString, expiresAt)
        if err != nil {
                h.logger.LogError("Refresh token storage failed: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Registration failed")
                return
        }

        // Set refresh token cookie
        h.setRefreshTokenCookie(w, refreshTokenString)

        h.logger.LogSuccess("Registration successful for user: %s", user.Email)

        response := RegisterResponse{
                Success:   true,
                Message:   "Registration successful! You are now logged in.",
                AccessToken:  accessToken,
                RefreshToken: refreshTokenString,
                User: UserResponse{
                        ID:           user.ID,
                        Email:        user.Email,
                        Nickname:     user.Nickname,
                        Money:        user.Money,
                        Topup:        user.Topup,
                        LastTopupAt:  user.LastTopupAt,
                        AuthProvider: user.AuthProvider,
                },
        }

        h.writeJSON(w, http.StatusOK, response)
}

// Login handler
func (h *Handler) loginHandler(w http.ResponseWriter, r *http.Request) {
        h.logger.LogAuth("Processing login request")

        var req LoginRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                h.writeError(w, http.StatusBadRequest, "Invalid JSON")
                return
        }

        if req.Identifier == "" || req.Password == "" {
                h.writeError(w, http.StatusBadRequest, "Identifier and password are required")
                return
        }

        // Find user by email or nickname
        h.logger.LogAuth("Looking up user: %s", req.Identifier)
        user, err := h.db.GetUserByEmail(req.Identifier)
        if err != nil {
                user, err = h.db.GetUserByNickname(req.Identifier)
        }
        if err != nil {
                h.logger.LogAuth("User not found: %s", req.Identifier)
                h.writeError(w, http.StatusUnauthorized, "Invalid email/nickname or password")
                return
        }

        // Verify password
        h.logger.LogAuth("Verifying password for user: %s", user.ID)
        if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(req.Password)); err != nil {
                h.logger.LogAuth("Invalid password for user: %s", user.ID)
                h.writeError(w, http.StatusUnauthorized, "Invalid email/nickname or password")
                return
        }

        // Generate JWT tokens
        h.logger.LogAuth("Generating JWT tokens for user: %s", user.ID)

        accessToken, err := generateAccessToken(user, h.config)
        if err != nil {
                h.logger.LogError("Access token generation failed: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Login failed")
                return
        }

        refreshTokenString, err := generateRefreshToken(user.ID, h.config)
        if err != nil {
                h.logger.LogError("Refresh token generation failed: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Login failed")
                return
        }

        // Store refresh token in database
        expiresAt := time.Now().Add(h.config.JWTRefreshTokenTTL)
        _, err = h.db.CreateRefreshToken(user.ID, refreshTokenString, expiresAt)
        if err != nil {
                h.logger.LogError("Refresh token storage failed: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Login failed")
                return
        }

        // Set refresh token cookie
        h.setRefreshTokenCookie(w, refreshTokenString)

        h.logger.LogSuccess("Login successful for user: %s", user.Email)

        response := LoginResponse{
                Success:      true,
                AccessToken:  accessToken,
                RefreshToken: refreshTokenString,
                User: UserResponse{
                        ID:           user.ID,
                        Email:        user.Email,
                        Nickname:     user.Nickname,
                        Money:        user.Money,
                        Topup:        user.Topup,
                        LastTopupAt:  user.LastTopupAt,
                        AuthProvider: user.AuthProvider,
                },
        }

        h.writeJSON(w, http.StatusOK, response)
}

// User info handler
func (h *Handler) userHandler(w http.ResponseWriter, r *http.Request) {
        h.logger.LogAuth("Validating JWT token...")

        // Get access token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
                h.logger.LogAuth("No JWT token found in Authorization header")
                h.writeError(w, http.StatusUnauthorized, "No access token")
                return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // Validate JWT token
        claims, err := validateAccessToken(tokenString, h.config)
        if err != nil {
                h.logger.LogAuth("Invalid JWT token: %s", err.Error())
                h.writeError(w, http.StatusUnauthorized, "Invalid access token")
                return
        }

        // Get user data
        user, err := h.db.GetUserByID(claims.UserID)
        if err != nil {
                h.logger.LogError("Failed to get user data: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "User not found")
                return
        }

        // Get user betting stats
        bets, wonBets, settledBets, avgOdds, _ := h.db.GetUserStats(user.ID)

        h.logger.LogSuccess("Session valid for user: %s", user.Nickname)

        response := LoginResponse{
                Success: true,
                User: UserResponse{
                        ID:           user.ID,
                        Email:        user.Email,
                        Nickname:     user.Nickname,
                        Money:        user.Money,
                        Topup:        user.Topup,
                        LastTopupAt:  user.LastTopupAt,
                        Bets:         bets,
                        WonBets:      wonBets,
                        SettledBets:  settledBets,
                        AvgOdds:      avgOdds,
                        AuthProvider: user.AuthProvider,
                },
        }

        h.writeJSON(w, http.StatusOK, response)
}

// Logout handler
func (h *Handler) logoutHandler(w http.ResponseWriter, r *http.Request) {
        h.logger.LogAuth("Processing logout request")

        cookie, err := r.Cookie(h.config.CookieName)
        if err == nil && cookie.Value != "" {
                // Delete refresh token from database
                h.logger.LogAuth("Deleting refresh token from database")
                h.db.DeleteRefreshToken(cookie.Value)
        }

        // Clear refresh token cookie
        h.clearRefreshTokenCookie(w)

        h.logger.LogSuccess("Logout successful")
        h.writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// Topup handler
func (h *Handler) topupHandler(w http.ResponseWriter, r *http.Request) {
        h.logger.LogAuth("Starting balance top-up process...")

        // Get JWT token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
                h.logger.LogAuth("No JWT token found in Authorization header")
                h.writeError(w, http.StatusUnauthorized, "No access token")
                return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // Validate JWT token
        claims, err := validateAccessToken(tokenString, h.config)
        if err != nil {
                h.logger.LogAuth("Invalid JWT token: %s", err.Error())
                h.writeError(w, http.StatusUnauthorized, "Invalid access token")
                return
        }

        // Get user data
        user, err := h.db.GetUserByID(claims.UserID)
        if err != nil {
                h.logger.LogError("User not found: %s", err.Error())
                h.writeError(w, http.StatusNotFound, "User not found")
                return
        }

        h.logger.LogAuth("Processing top-up for user: %s", user.ID)

        // Check balance
        if user.Money >= h.config.MaxTopupBalance {
                h.logger.LogAuth("Top-up not allowed: balance $%.2f >= $%.2f", user.Money, h.config.MaxTopupBalance)
                h.writeError(w, http.StatusBadRequest, "Top-up not available. Balance must be less than $500.")
                return
        }

        // Check if user has already topped up today
        lastTopupTime, err := h.db.GetUserLastTopupTime(user.ID)
        if err != nil {
                h.logger.LogError("Failed to get last topup time: %s", err.Error())
                // Don't fail the request, just log
        } else if lastTopupTime != nil {
                // Check if last topup was less than 24 hours ago
                timeSinceLastTopup := time.Since(*lastTopupTime)
                if timeSinceLastTopup < 24*time.Hour {
                        hoursRemaining := 24 - int(timeSinceLastTopup.Hours())
                        minutesRemaining := 60 - int(timeSinceLastTopup.Minutes()) % 60
                        h.logger.LogAuth("Top-up not allowed: last topup was %v ago", timeSinceLastTopup)
                        h.writeError(w, http.StatusBadRequest, fmt.Sprintf("You can only top up once per day. Please wait %d hours and %d minutes.", hoursRemaining, minutesRemaining))
                        return
                }
        }

        // Update balance and increment topup counter
        newBalance := user.Money + h.config.TopupAmount
        h.logger.LogAuth("Balance will be updated: $%.2f → $%.2f", user.Money, newBalance)

        if err := h.db.UpdateUserMoney(user.ID, newBalance); err != nil {
                h.logger.LogError("Balance update failed: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Top-up failed")
                return
        }

        if err := h.db.IncrementUserTopup(user.ID); err != nil {
                h.logger.LogError("Topup counter update failed: %s", err.Error())
                // Don't fail the request, just log
        }

        h.logger.LogSuccess("Balance updated successfully: $%.2f → $%.2f", user.Money, newBalance)

        response := TopupResponse{
                Success:    true,
                Message:    "Balance topped up successfully! Added $10,000.",
                NewBalance: newBalance,
        }

        h.writeJSON(w, http.StatusOK, response)
}

// Change password handler
func (h *Handler) changePasswordHandler(w http.ResponseWriter, r *http.Request) {
        h.logger.LogAuth("Starting password change process...")

        // Get JWT token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
                h.logger.LogAuth("No JWT token found in Authorization header")
                h.writeError(w, http.StatusUnauthorized, "No access token")
                return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        // Validate JWT token
        claims, err := validateAccessToken(tokenString, h.config)
        if err != nil {
                h.logger.LogAuth("Invalid JWT token: %s", err.Error())
                h.writeError(w, http.StatusUnauthorized, "Invalid access token")
                return
        }

        // Get user data
        user, err := h.db.GetUserByID(claims.UserID)
        if err != nil {
                h.logger.LogError("User not found: %s", err.Error())
                h.writeError(w, http.StatusNotFound, "User not found")
                return
        }

        h.logger.LogAuth("Processing password change for user: %s", user.ID)

        var req ChangePasswordRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                h.writeError(w, http.StatusBadRequest, "Invalid JSON")
                return
        }

        if req.CurrentPassword == "" || req.NewPassword == "" {
                h.writeError(w, http.StatusBadRequest, "Current password and new password are required")
                return
        }

        if len(req.NewPassword) < h.config.MinPasswordLength {
                h.writeError(w, http.StatusBadRequest, fmt.Sprintf("New password must be at least %d characters long", h.config.MinPasswordLength))
                return
        }

        // Verify current password
        h.logger.LogAuth("Verifying current password...")
        if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(req.CurrentPassword)); err != nil {
                h.logger.LogAuth("Current password is incorrect")
                h.writeError(w, http.StatusBadRequest, "Current password is incorrect")
                return
        }

        // Hash new password
        h.logger.LogAuth("Hashing new password...")
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), h.config.BcryptCost)
        if err != nil {
                h.logger.LogError("Password hashing failed: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Password change failed")
                return
        }

        // Update password
        h.logger.LogAuth("Updating password in database...")
        if err := h.db.UpdateUserPassword(user.ID, string(hashedPassword)); err != nil {
                h.logger.LogError("Password update failed: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Password change failed")
                return
        }

        h.logger.LogSuccess("Password updated successfully for user: %s", user.ID)

        h.writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// BETS HANDLERS

// Get bets handler
func (h *Handler) getBetsHandler(w http.ResponseWriter, r *http.Request) {
        h.logger.LogBets("Getting user bets from PostgreSQL...")

        // Check if requesting bets for another player
        playerParam := r.URL.Query().Get("player")
        var targetUserID string
        var targetUser *User

        if playerParam != "" {
                // Viewing another player's bets - no auth required
                h.logger.LogBets("Requesting bets for player: %s", playerParam)
                var err error
                targetUser, err = h.db.GetUserByNickname(playerParam)
                if err != nil {
                        h.logger.LogBets("Player %s not found", playerParam)
                        h.writeError(w, http.StatusNotFound, "Player not found")
                        return
                }
                targetUserID = targetUser.ID
                h.logger.LogBets("Viewing bets for player: %s (%s)", playerParam, targetUserID)
        } else {
                // Viewing own bets - JWT required
                h.logger.LogBets("Validating JWT for own bets...")

                authHeader := r.Header.Get("Authorization")
                if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
                        h.logger.LogBets("No JWT token found in Authorization header")
                        h.writeError(w, http.StatusUnauthorized, "No access token")
                        return
                }

                tokenString := strings.TrimPrefix(authHeader, "Bearer ")

                claims, err := validateAccessToken(tokenString, h.config)
                if err != nil {
                        h.logger.LogBets("Invalid JWT token: %s", err.Error())
                        h.writeError(w, http.StatusUnauthorized, "Invalid access token")
                        return
                }

                targetUserID = claims.UserID
        }

        // Get bets
        bets, err := h.db.GetUserBets(targetUserID, playerParam)
        if err != nil {
                h.logger.LogError("Failed to get bets: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Failed to get bets")
                return
        }

        h.logger.LogBets("Found %d bets for user", len(bets))

        // If viewing another player's bets, return extended response with player info and stats
        if playerParam != "" && targetUser != nil {
                // Calculate stats
                wonBets := 0
                settledBets := 0
                totalOdds := 0.0
                for _, bet := range bets {
                        if bet.Status == "won" {
                                wonBets++
                                settledBets++
                        } else if bet.Status == "lost" {
                                settledBets++
                        }
                        totalOdds += bet.Odds
                }

                avgOdds := 0.0
                if len(bets) > 0 {
                        avgOdds = totalOdds / float64(len(bets))
                }

                winRate := 0.0
                if settledBets > 0 {
                        winRate = float64(wonBets) / float64(settledBets) * 100
                }

                response := map[string]interface{}{
                        "success": true,
                        "player": map[string]interface{}{
                                "id":       targetUser.ID,
                                "nickname": targetUser.Nickname,
                                "money":    targetUser.Money,
                                "created":  targetUser.CreatedAt,
                        },
                        "bets": bets,
                        "stats": map[string]interface{}{
                                "total_bets":   len(bets),
                                "won_bets":     wonBets,
                                "settled_bets": settledBets,
                                "win_rate":     winRate,
                                "avg_odds":     avgOdds,
                        },
                }

                h.writeJSON(w, http.StatusOK, response)
                return
        }

        // Standard response for own bets
        var betDisplays []BetDisplay
        for _, bet := range bets {
                betDisplays = append(betDisplays, BetDisplay{
                        ID:           bet.BetID,
                        MatchID:      bet.MatchID,
                        BetType:      bet.BetType,
                        BetAmount:    bet.BetAmount,
                        Odds:         bet.Odds,
                        PotentialWin: bet.PotentialWin,
                        Status:       bet.Status,
                        HomeTeam:     bet.HomeTeam,
                        AwayTeam:     bet.AwayTeam,
                        CreatedAt:    bet.CreatedAt,
                        CommenceTime: bet.CommenceTime,
                })
        }

        response := BetsResponse{
                Success: true,
                Bets:    betDisplays,
        }

        h.writeJSON(w, http.StatusOK, response)
}

// Place bet handler
func (h *Handler) placeBetHandler(w http.ResponseWriter, r *http.Request) {
        h.logger.LogBets("Placing a new bet...")

        // Get JWT token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
                h.logger.LogBets("No JWT token found in Authorization header")
                h.writeError(w, http.StatusUnauthorized, "No access token")
                return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")

        h.logger.LogBets("Validating JWT token...")

        // Validate JWT token
        claims, err := validateAccessToken(tokenString, h.config)
        if err != nil {
                h.logger.LogBets("Invalid JWT token: %s", err.Error())
                h.writeError(w, http.StatusUnauthorized, "Invalid access token")
                return
        }

        // Get user data
        user, err := h.db.GetUserByID(claims.UserID)
        if err != nil {
                h.logger.LogError("User not found: %s", err.Error())
                h.writeError(w, http.StatusNotFound, "User not found")
                return
        }

        var req PlaceBetRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
                h.writeError(w, http.StatusBadRequest, "Invalid JSON")
                return
        }

        // Validate request
        if req.MatchID == "" || req.BetType == "" || req.BetAmount <= 0 || req.Odds <= 0 {
                h.writeError(w, http.StatusBadRequest, "Missing required fields")
                return
        }

        if req.BetAmount > user.Money {
                h.writeError(w, http.StatusBadRequest, "Insufficient balance")
                return
        }

        // Validate bet type
        if req.BetType != "home" && req.BetType != "draw" && req.BetType != "away" {
                h.writeError(w, http.StatusBadRequest, "Invalid bet type")
                return
        }

        // Check if match exists and hasn't started
        match, err := h.db.GetMatchByID(req.MatchID)
        if err != nil {
                h.writeError(w, http.StatusNotFound, "Match not found")
                return
        }

        if match.CommenceTime.Before(time.Now()) {
                h.logger.LogBets("Match %s has already started or finished", req.MatchID)
                h.writeError(w, http.StatusBadRequest, "Cannot place bet on a match that has already started")
                return
        }

        // Create bet
        bet := &Bet{
                UserID:       user.ID,
                MatchID:      req.MatchID,
                BetType:      req.BetType,
                BetAmount:    req.BetAmount,
                Odds:         req.Odds,
                PotentialWin: req.BetAmount * req.Odds,
                Status:       "pending",
                HomeTeam:     req.HomeTeam,
                AwayTeam:     req.AwayTeam,
        }

        h.logger.LogBets("Inserting bet into database...")

        // Use transaction-like behavior (simplified)
        placedBet, err := h.db.PlaceBet(bet)
        if err != nil {
                h.logger.LogError("Failed to place bet: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Failed to place bet")
                return
        }

        // Update user balance
        h.logger.LogBets("Updating user balance...")
        newBalance := user.Money - req.BetAmount
        if err := h.db.UpdateUserMoney(user.ID, newBalance); err != nil {
                h.logger.LogError("Failed to update balance: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Failed to update balance")
                return
        }

        h.logger.LogSuccess("Bet placed successfully! User: %s, Amount: $%.2f, New balance: $%.2f",
                user.Nickname, req.BetAmount, newBalance)
        h.logger.LogSuccess("BetID: %s", placedBet.BetID)

        response := BetResponse{
                Success: true,
                Bet: BetInfo{
                        ID:           placedBet.BetID,
                        Amount:       req.BetAmount,
                        Odds:         req.Odds,
                        PotentialWin: req.BetAmount * req.Odds,
                        NewBalance:   newBalance,
                },
        }

        h.writeJSON(w, http.StatusOK, response)
}

// MATCHES HANDLERS

// Get matches handler
func (h *Handler) getMatchesHandler(w http.ResponseWriter, r *http.Request) {
        h.logger.LogSystem("MATCHES", "Getting matches from database...")
        
        matches, err := h.db.GetMatches()
        if err != nil {
                h.logger.LogError("Failed to get matches: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Failed to get matches")
                return
        }

        h.logger.LogSystem("MATCHES", "Found %d matches", len(matches))

        // Convert to response format
        var matchDisplays []MatchDisplay
        for _, match := range matches {
                matchDisplays = append(matchDisplays, MatchDisplay{
                        ID:           match.APIID,
                        HomeTeam:     match.HomeTeam,
                        AwayTeam:     match.AwayTeam,
                        CommenceTime: match.CommenceTime,
                        HomeOdds:     match.HomeOdds,
                        DrawOdds:     match.DrawOdds,
                        AwayOdds:     match.AwayOdds,
                })
        }

        response := MatchesResponse{
                Success: true,
                Matches: matchDisplays,
        }

        h.writeJSON(w, http.StatusOK, response)
}

// PLAYERS HANDLERS

// Get players handler
func (h *Handler) getPlayersHandler(w http.ResponseWriter, r *http.Request) {
        h.logger.LogSystem("PLAYERS", "Getting players list...")

        // Parse pagination parameters
        limit := h.config.DefaultPlayerLimit
        offset := 0

        if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
                if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 && parsedLimit <= h.config.MaxPlayerLimit {
                        limit = parsedLimit
                }
        }

        if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
                if parsedOffset, err := strconv.Atoi(offsetParam); err == nil && parsedOffset >= 0 {
                        offset = parsedOffset
                }
        }

        h.logger.LogSystem("PLAYERS", "Fetching players (limit: %d, offset: %d)", limit, offset)

        // Get players
        players, err := h.db.GetPlayers(limit, offset)
        if err != nil {
                h.logger.LogError("Failed to get players: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Failed to get players")
                return
        }

        // Get total count for pagination
        total, err := h.db.GetTotalPlayers()
        if err != nil {
                h.logger.LogError("Failed to get total count: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Failed to get players")
                return
        }

        h.logger.LogSystem("PLAYERS", "Found %d players (total: %d)", len(players), total)

        response := PlayersResponse{
                Success: true,
                Players: players,
                Pagination: PaginationInfo{
                        Limit:   limit,
                        Offset:  offset,
                        Total:   total,
                        HasMore: offset+limit < total,
                },
        }

        h.writeJSON(w, http.StatusOK, response)
}

// HELPER FUNCTIONS

// Set refresh token cookie
func (h *Handler) setRefreshTokenCookie(w http.ResponseWriter, token string) {
        http.SetCookie(w, &http.Cookie{
                Name:     h.config.CookieName,
                Value:    token,
                Path:     "/",
                HttpOnly: h.config.CookieHTTPOnly,
                Secure:   h.config.CookieSecure,
                SameSite: http.SameSiteLaxMode,
                MaxAge:   int(h.config.JWTRefreshTokenTTL.Seconds()),
        })
}

// Clear refresh token cookie
func (h *Handler) clearRefreshTokenCookie(w http.ResponseWriter) {
        http.SetCookie(w, &http.Cookie{
                Name:     h.config.CookieName,
                Value:    "",
                Path:     "/",
                HttpOnly: h.config.CookieHTTPOnly,
                Secure:   h.config.CookieSecure,
                SameSite: http.SameSiteLaxMode,
                MaxAge:   -1,
        })
}

// Refresh token handler
func (h *Handler) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
        h.logger.LogAuth("Processing token refresh request")

        // Get refresh token from cookie
        cookie, err := r.Cookie(h.config.CookieName)
        if err != nil || cookie.Value == "" {
                h.logger.LogAuth("No refresh token found")
                h.writeError(w, http.StatusUnauthorized, "No refresh token")
                return
        }

        refreshTokenString := cookie.Value

        // Generate new access token
        accessToken, err := refreshAccessToken(refreshTokenString, h.db, h.config)
        if err != nil {
                h.logger.LogAuth("Token refresh failed: %s", err.Error())
                // Clear invalid refresh token
                h.clearRefreshTokenCookie(w)
                h.writeError(w, http.StatusUnauthorized, "Invalid refresh token")
                return
        }

        h.logger.LogSuccess("Token refresh successful")

        response := RefreshResponse{
                Success:     true,
                AccessToken: accessToken,
        }

        h.writeJSON(w, http.StatusOK, response)
}

// Write JSON response
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(status)
        json.NewEncoder(w).Encode(data)
}

// Write error response
func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
        response := APIResponse{
                Success: false,
                Error:   message,
        }
        h.writeJSON(w, status, response)
}

// ADMIN SYNC HANDLERS

// OddsSyncHandler handles POST /api/odds/sync
func (h *Handler) oddsSyncHandler(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Log incoming request details for debugging
        clientIP := h.getClientIP(r)
        h.logger.LogSystem("ODDS_SYNC", "=== ODDS SYNC REQUEST START ===")
        h.logger.LogSystem("ODDS_SYNC", "Client IP: %s, Time: %s", clientIP, start.Format(time.RFC3339))

        admin, ok := getAdminFromContext(r.Context())
        if !ok {
                h.logger.LogSystem("ODDS_SYNC", "=== ODDS SYNC REQUEST END (UNAUTHORIZED) ===")
                h.writeError(w, http.StatusUnauthorized, "Admin authentication required")
                return
        }

        h.logger.LogSystem("ODDS_SYNC", "Starting odds sync by admin: %s", admin.Username)

        // Fetch odds from API
        events, apiStats, err := fetchOddsFromAPI(h.config.OddsAPIKey)
        if err != nil {
                h.logger.LogError("Failed to fetch odds from API: %s", err.Error())
                h.logger.LogSystem("ODDS_SYNC", "=== ODDS SYNC REQUEST END (API ERROR) ===")
                h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch odds: %s", err.Error()))
                return
        }

        if len(events) == 0 {
                h.logger.LogSystem("ODDS_SYNC", "No upcoming matches found")
                h.logger.LogSystem("ODDS_SYNC", "=== ODDS SYNC REQUEST END (NO MATCHES) ===")
                h.writeJSON(w, http.StatusOK, map[string]interface{}{
                        "ok":      true,
                        "task":    "odds:sync",
                        "admin":   admin.Username,
                        "created": 0,
                        "updated": 0,
                        "skipped": 0,
                        "message": "No upcoming matches found",
                        "apiStats": apiStats,
                        "ms":      time.Since(start).Milliseconds(),
                })
                return
        }

        // Process matches
        results := map[string]int{
                "created": 0,
                "updated": 0,
                "skipped": 0,
        }

        for _, event := range events {
                match, err := processOddsEvent(event)
                if err != nil {
                        h.logger.LogError("Failed to process event: %s", err.Error())
                        continue
                }

                // Check if match exists
                existingMatch, err := h.db.GetMatchByAPIID(match.APIID)
                if err == nil && existingMatch != nil {
                        // Update existing match - preserve old odds if new ones are null
                        if match.HomeOdds == nil {
                                match.HomeOdds = existingMatch.HomeOdds
                        }
                        if match.DrawOdds == nil {
                                match.DrawOdds = existingMatch.DrawOdds
                        }
                        if match.AwayOdds == nil {
                                match.AwayOdds = existingMatch.AwayOdds
                        }
                        _, err = h.db.UpdateMatchByAPIID(match.APIID, match)
                        if err != nil {
                                h.logger.LogError("Failed to update match: %s", err.Error())
                                continue
                        }
                        results["updated"]++
                } else {
                        // Create new match - only if has odds
                        if match.HomeOdds == nil || match.DrawOdds == nil || match.AwayOdds == nil {
                                results["skipped"]++
                                continue
                        }
                        _, err = h.db.UpsertMatch(match)
                        if err != nil {
                                h.logger.LogError("Failed to create match: %s", err.Error())
                                continue
                        }
                        results["created"]++
                }
        }

        duration := time.Since(start)
        h.logger.LogSuccess("Odds sync completed: created=%d, updated=%d, skipped=%d in %v", results["created"], results["updated"], results["skipped"], duration)

        h.logger.LogSystem("ODDS_SYNC", "=== ODDS SYNC REQUEST END (SUCCESS) ===")

        h.writeJSON(w, http.StatusOK, map[string]interface{}{
                "ok":       true,
                "task":     "odds:sync",
                "admin":    admin.Username,
                "created":  results["created"],
                "updated":  results["updated"],
                "skipped":  results["skipped"],
                "apiStats": apiStats,
                "ms":       duration.Milliseconds(),
        })
}

// ScoresSyncHandler handles POST /api/scores/sync
func (h *Handler) scoresSyncHandler(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Log incoming request details for debugging
        clientIP := h.getClientIP(r)
        h.logger.LogSystem("SCORES_SYNC", "=== SCORES SYNC REQUEST START ===")
        h.logger.LogSystem("SCORES_SYNC", "Client IP: %s, Time: %s", clientIP, start.Format(time.RFC3339))

        admin, ok := getAdminFromContext(r.Context())
        if !ok {
                h.logger.LogSystem("SCORES_SYNC", "=== SCORES SYNC REQUEST END (UNAUTHORIZED) ===")
                h.writeError(w, http.StatusUnauthorized, "Admin authentication required")
                return
        }

        h.logger.LogSystem("SCORES_SYNC", "Starting scores sync by admin: %s", admin.Username)

        // Fetch scores from API
        scores, apiStats, err := fetchScoresFromAPI(h.config.OddsAPIKey)
        if err != nil {
                h.logger.LogError("Failed to fetch scores from API: %s", err.Error())
                h.logger.LogSystem("SCORES_SYNC", "=== SCORES SYNC REQUEST END (API ERROR) ===")
                h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to fetch scores: %s", err.Error()))
                return
        }

        if len(scores) == 0 {
                h.logger.LogSystem("SCORES_SYNC", "No scores found")
                h.logger.LogSystem("SCORES_SYNC", "=== SCORES SYNC REQUEST END (NO SCORES) ===")
                h.writeJSON(w, http.StatusOK, map[string]interface{}{
                        "ok":      true,
                        "task":    "scores:sync",
                        "admin":   admin.Username,
                        "created": 0,
                        "updated": 0,
                        "message": "No scores found",
                        "apiStats": apiStats,
                        "ms":      time.Since(start).Milliseconds(),
                })
                return
        }

        // Process scores
        results := map[string]int{
                "created": 0,
                "updated": 0,
        }

        for _, score := range scores {
                match, err := processScoreEvent(score)
                if err != nil {
                        h.logger.LogError("Failed to process score: %s", err.Error())
                        continue
                }

                // Check if match exists
                existingMatch, err := h.db.GetMatchByAPIID(match.APIID)
                if err == nil && existingMatch != nil {
                        // Update existing match - don't touch odds
                        match.HomeOdds = existingMatch.HomeOdds
                        match.DrawOdds = existingMatch.DrawOdds
                        match.AwayOdds = existingMatch.AwayOdds
                        _, err = h.db.UpdateMatchByAPIID(match.APIID, match)
                        if err != nil {
                                h.logger.LogError("Failed to update match: %s", err.Error())
                                continue
                        }
                        results["updated"]++
                } else {
                        // Create new match with scores but no odds
                        match.HomeOdds = nil
                        match.DrawOdds = nil
                        match.AwayOdds = nil
                        _, err = h.db.UpsertMatch(match)
                        if err != nil {
                                h.logger.LogError("Failed to create match: %s", err.Error())
                                continue
                        }
                        results["created"]++
                }
        }

        duration := time.Since(start)
        h.logger.LogSuccess("Scores sync completed: created=%d, updated=%d in %v", results["created"], results["updated"], duration)

        h.logger.LogSystem("SCORES_SYNC", "=== SCORES SYNC REQUEST END (SUCCESS) ===")

        h.writeJSON(w, http.StatusOK, map[string]interface{}{
                "ok":       true,
                "task":     "scores:sync",
                "admin":    admin.Username,
                "created":  results["created"],
                "updated":  results["updated"],
                "apiStats": apiStats,
                "ms":       duration.Milliseconds(),
        })
}

// CalcHandler handles POST /api/calc
func (h *Handler) calcHandler(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        admin, ok := getAdminFromContext(r.Context())
        if !ok {
                h.writeError(w, http.StatusUnauthorized, "Admin authentication required")
                return
        }

        h.logger.LogSystem("CALC", "Starting calculation by admin: %s", admin.Username)

        // Get completed uncalculated matches
        matches, err := h.db.GetCompletedUncalculatedMatches()
        if err != nil {
                h.logger.LogError("Failed to get uncalculated matches: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Failed to get matches")
                return
        }

        updatedCount := 0
        calculatedMatches := []map[string]interface{}{}

        if len(matches) == 0 {
                h.logger.LogSystem("CALC", "No matches to calculate")
        } else {
                for _, match := range matches {
                // Determine result
                var result string
                if match.HomeScore == nil || match.AwayScore == nil {
                        continue
                }
                if *match.HomeScore > *match.AwayScore {
                        result = "home"
                } else if *match.HomeScore < *match.AwayScore {
                        result = "away"
                } else {
                        result = "draw"
                }

                // Update bets and user money
                if err := h.db.UpdateBetsStatusAndUserMoney(match.APIID, result); err != nil {
                        h.logger.LogError("Failed to update bets for match %s: %s", match.APIID, err.Error())
                        continue
                }

                // Mark match as calculated
                if err := h.db.UpdateMatchCalculated(match.APIID, result); err != nil {
                        h.logger.LogError("Failed to mark match as calculated: %s", err.Error())
                        continue
                }

                updatedCount++
                calculatedMatches = append(calculatedMatches, map[string]interface{}{
                        "home_team": match.HomeTeam,
                        "away_team": match.AwayTeam,
                        "score":     fmt.Sprintf("%d-%d", *match.HomeScore, *match.AwayScore),
                        "result":    result,
                })

                h.logger.LogSuccess("Match calculated: %s %d-%d %s | Winner: %s",
                        match.HomeTeam, *match.HomeScore, *match.AwayScore, match.AwayTeam, result)
                }
        }

        // Send Telegram notification if configured (always send, even if no matches)
        h.logger.LogSystem("CALC", "Checking Telegram notification: updatedCount=%d, botToken=%s, channelID=%s",
                updatedCount, maskToken(h.config.TelegramBotToken), maskToken(h.config.TelegramChannelID))

        if h.config.TelegramBotToken != "" && h.config.TelegramChannelID != "" {
                h.logger.LogSystem("CALC", "Sending Telegram notification for %d matches", len(calculatedMatches))
                if err := sendTelegramNotification(h.config.TelegramBotToken, h.config.TelegramChannelID, calculatedMatches); err != nil {
                        h.logger.LogError("Failed to send Telegram notification: %s", err.Error())
                } else {
                        h.logger.LogSuccess("Telegram notification sent successfully")
                }
        } else {
                if updatedCount == 0 {
                        h.logger.LogSystem("CALC", "Skipping Telegram notification: no matches were updated")
                }
                if h.config.TelegramBotToken == "" {
                        h.logger.LogSystem("CALC", "Skipping Telegram notification: bot token not configured")
                }
                if h.config.TelegramChannelID == "" {
                        h.logger.LogSystem("CALC", "Skipping Telegram notification: channel ID not configured")
                }
        }

        h.logger.LogSuccess("Calculation completed: %d matches processed", updatedCount)

        message := "Calculation completed"
        if updatedCount == 0 {
                message = "No matches to calculate"
        }

        h.writeJSON(w, http.StatusOK, map[string]interface{}{
                "ok":      true,
                "task":    "calc",
                "admin":   admin.Username,
                "updated": updatedCount,
                "message": message,
                "matches": calculatedMatches,
                "ms":      time.Since(start).Milliseconds(),
        })
}

// AnalyticsHandler returns visitor statistics from Cloudflare Analytics API
// Cloudflare Analytics handler - COMMENTED OUT
/*
func (h *Handler) analyticsHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.LogSystem("ANALYTICS", "Fetching visitor statistics")

	// Cloudflare Analytics API endpoint
	// Requires CF_API_TOKEN and CF_ZONE_ID environment variables
	apiToken := os.Getenv("CF_API_TOKEN")
	zoneID := os.Getenv("CF_ZONE_ID")

	if apiToken == "" || zoneID == "" {
		h.logger.LogError("Cloudflare API credentials not configured")
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"total_visitors": 0,
			"message": "Analytics not configured",
		})
		return
	}

	// Get visitor analytics for the last 30 days
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/analytics/dashboard?since=-30", zoneID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		h.logger.LogError("Failed to create analytics request: %s", err.Error())
		h.writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"total_visitors": 0,
			"error": "Request creation failed",
		})
		return
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.LogError("Failed to fetch analytics: %s", err.Error())
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"total_visitors": 0,
			"message": "API request failed",
		})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.LogError("Cloudflare API returned status: %d", resp.StatusCode)
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"total_visitors": 0,
			"message": "API error",
		})
		return
	}

	var result struct {
		Success bool `json:"success"`
		Result  struct {
			Totals struct {
				Requests struct {
					All  int `json:"all"`
					Cached int `json:"cached"`
					Uncachable int `json:"uncached"`
				} `json:"requests"`
				Pageviews struct {
					All  int `json:"all"`
					Cached int `json:"cached"`
					Uncachable int `json:"uncached"`
				} `json:"pageviews"`
			} `json:"totals"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		h.logger.LogError("Failed to parse analytics response: %s", err.Error())
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"total_visitors": 0,
			"message": "Parse error",
		})
		return
	}

	if !result.Success {
		h.logger.LogError("Cloudflare API returned success=false")
		h.writeJSON(w, http.StatusOK, map[string]interface{}{
			"total_visitors": 0,
			"message": "API returned error",
		})
		return
	}

	// Use pageviews as visitor count (more accurate than requests)
	totalVisitors := result.Result.Totals.Pageviews.All

	h.logger.LogSuccess("Analytics data retrieved: %d total visitors", totalVisitors)

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_visitors": totalVisitors,
	})
}
*/

// getClientIP extracts the real client IP from request headers
func (h *Handler) getClientIP(r *http.Request) string {
        // Check X-Forwarded-For header (can contain multiple IPs)
        xForwardedFor := r.Header.Get("X-Forwarded-For")
        if xForwardedFor != "" {
                // Take the first IP in the chain (original client)
                ips := strings.Split(xForwardedFor, ",")
                if len(ips) > 0 {
                        ip := strings.TrimSpace(ips[0])
                        if ip != "" && ip != "unknown" {
                                return ip
                        }
                }
        }

        // Check X-Real-IP header
        xRealIP := r.Header.Get("X-Real-IP")
        if xRealIP != "" && xRealIP != "unknown" {
                return xRealIP
        }

        // Check CF-Connecting-IP (Cloudflare)
        cfConnectingIP := r.Header.Get("CF-Connecting-IP")
        if cfConnectingIP != "" {
                return cfConnectingIP
        }

        // Check X-Client-IP
        xClientIP := r.Header.Get("X-Client-IP")
        if xClientIP != "" {
                return xClientIP
        }

        // Fallback to RemoteAddr (remove port if present)
        remoteAddr := r.RemoteAddr
        if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
                return host
        }

        return remoteAddr
}

// GOOGLE OAUTH HANDLERS

// Google OAuth login handler - initiates OAuth flow
func (h *Handler) googleLoginHandler(w http.ResponseWriter, r *http.Request) {
        h.logger.LogAuth("Initiating Google OAuth login")

        // Check if Google OAuth is configured
        if h.config.GoogleClientID == "" || h.config.GoogleClientSecret == "" {
                h.logger.LogError("Google OAuth not configured")
                h.writeError(w, http.StatusServiceUnavailable, "Google authentication is not available")
                return
        }

        // Get redirect URL from query parameter (optional)
        redirectURL := r.URL.Query().Get("redirect_url")
        if redirectURL != "" {
                // Basic validation - allow localhost, relative URLs, and our domain
                if parsedURL, err := url.Parse(redirectURL); err != nil {
                        redirectURL = "" // Reset if invalid URL
                } else if parsedURL.IsAbs() {
                        // Allow localhost and freebet.guru domain
                        allowedHosts := []string{"localhost", "127.0.0.1", "freebet.guru"}
                        isAllowed := false
                        for _, host := range allowedHosts {
                                if strings.Contains(parsedURL.Host, host) {
                                        isAllowed = true
                                        break
                                }
                        }
                        if !isAllowed {
                                redirectURL = "" // Reset if not allowed host
                        }
                }
                // Relative URLs are allowed
        }

        // Generate OAuth state
        state, err := generateOAuthState(redirectURL)
        if err != nil {
                h.logger.LogError("Failed to generate OAuth state: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Failed to initiate authentication")
                return
        }

        // Get OAuth config and generate authorization URL
        oauthConfig := getGoogleOAuthConfig(h.config)
        authURL := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)

        h.logger.LogAuth("Redirecting to Google OAuth: %s", authURL)

        // Redirect to Google
        http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// Google OAuth callback handler
func (h *Handler) googleCallbackHandler(w http.ResponseWriter, r *http.Request) {
        h.logger.LogAuth("Processing Google OAuth callback")

        // Get authorization code and state from query parameters
        code := r.URL.Query().Get("code")
        state := r.URL.Query().Get("state")

        if code == "" {
                h.logger.LogAuth("No authorization code received")
                h.writeError(w, http.StatusBadRequest, "Authorization code missing")
                return
        }

        // Validate state parameter
        oauthState, valid := validateOAuthState(state)
        if !valid {
                h.logger.LogAuth("Invalid or expired OAuth state")
                h.writeError(w, http.StatusBadRequest, "Invalid authentication state")
                return
        }

        // Exchange authorization code for access token
        oauthConfig := getGoogleOAuthConfig(h.config)
        token, err := oauthConfig.Exchange(context.Background(), code)
        if err != nil {
                h.logger.LogError("Failed to exchange authorization code: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Authentication failed")
                return
        }

        // Get user info from Google
        googleUser, err := getGoogleUserInfo(token, h.config)
        if err != nil {
                h.logger.LogError("Failed to get Google user info: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Failed to get user information")
                return
        }

        h.logger.LogAuth("Google user authenticated: %s (%s)", googleUser.Email, googleUser.ID)

        // Check if user exists
        user, err := h.db.GetUserByGoogleID(googleUser.ID)
        if err != nil {
                // User doesn't exist, create new user
                h.logger.LogAuth("Creating new user for Google ID: %s", googleUser.ID)

                nickname := generateNicknameFromGoogleEmail(googleUser.Email)
                // Ensure nickname is unique
                if existingUser, _ := h.db.GetUserByNickname(nickname); existingUser != nil {
                        // Add random suffix if nickname exists
                        nickname = fmt.Sprintf("%s%d", nickname, time.Now().Unix()%1000)
                        if len(nickname) > 10 {
                                nickname = nickname[:10]
                        }
                }

                user, err = h.db.CreateUserWithGoogle(googleUser.ID, googleUser.Email, nickname, googleUser.Picture, h.config.InitialBalance)
                if err != nil {
                        h.logger.LogError("Failed to create user: %s", err.Error())
                        h.writeError(w, http.StatusInternalServerError, "User creation failed")
                        return
                }

                h.logger.LogSuccess("Created new user via Google OAuth: %s", user.Email)
        } else {
                h.logger.LogAuth("Existing user logged in via Google: %s", user.Email)

                // Update profile picture if changed
                if user.PictureURL.String != googleUser.Picture {
                        // Note: In production, you might want to add a method to update profile picture
                        h.logger.LogAuth("Profile picture changed for user: %s", user.ID)
                }
        }

        // Generate JWT tokens
        accessToken, err := generateAccessToken(user, h.config)
        if err != nil {
                h.logger.LogError("Access token generation failed: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Authentication failed")
                return
        }

        refreshTokenString, err := generateRefreshToken(user.ID, h.config)
        if err != nil {
                h.logger.LogError("Refresh token generation failed: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Authentication failed")
                return
        }

        // Store refresh token in database
        expiresAt := time.Now().Add(h.config.JWTRefreshTokenTTL)
        _, err = h.db.CreateRefreshToken(user.ID, refreshTokenString, expiresAt)
        if err != nil {
                h.logger.LogError("Refresh token storage failed: %s", err.Error())
                h.writeError(w, http.StatusInternalServerError, "Authentication failed")
                return
        }

        // Set refresh token cookie
        h.setRefreshTokenCookie(w, refreshTokenString)

        h.logger.LogSuccess("Google OAuth authentication successful for user: %s", user.Email)

        // Prepare response
        response := map[string]interface{}{
                "success":       true,
                "message":       "Authentication successful",
                "access_token":  accessToken,
                "refresh_token": refreshTokenString,
                "user": map[string]interface{}{
                        "id":            user.ID,
                        "email":         user.Email,
                        "nickname":      user.Nickname,
                        "picture_url":   user.PictureURL.String,
                        "auth_provider": user.AuthProvider,
                        "money":         user.Money,
                        "topup":         user.Topup,
                        "last_topup_at": user.LastTopupAt,
                },
        }

        // If redirect URL was provided, redirect with tokens as query parameters
        if oauthState.RedirectURL != "" {
                redirectURL := fmt.Sprintf("%s?access_token=%s&refresh_token=%s",
                        oauthState.RedirectURL, accessToken, refreshTokenString)
                http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
                return
        }

        // Return JSON response
        h.writeJSON(w, http.StatusOK, response)
}

// formatUptime formats uptime seconds into a human readable string
func (h *Handler) formatUptime(seconds int64) string {
        days := seconds / 86400
        hours := (seconds % 86400) / 3600
        minutes := (seconds % 3600) / 60
        secs := seconds % 60

        if days > 0 {
                return fmt.Sprintf("%dd %dh", days, hours)
        }
        if hours > 0 {
                return fmt.Sprintf("%dh %dm", hours, minutes)
        }
        if minutes > 0 {
                return fmt.Sprintf("%dm %ds", minutes, secs)
        }
        return fmt.Sprintf("%ds", secs)
}

// maskToken masks sensitive tokens for logging
func maskToken(token string) string {
        if len(token) <= 8 {
                return strings.Repeat("*", len(token))
        }
        return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}