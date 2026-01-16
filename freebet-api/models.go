package main

import (
        "database/sql"
        "time"

        "github.com/golang-jwt/jwt/v5"
)

// User represents a user in the system
type User struct {
        ID            string         `json:"id" db:"id"`
        Email         string         `json:"email" db:"email"`
        Nickname      string         `json:"nickname" db:"nickname"`
        PasswordHash  sql.NullString `json:"-" db:"password_hash"` // Never expose in JSON (legacy)
        GoogleID      sql.NullString `json:"-" db:"google_id"`      // Google OAuth ID
        PictureURL    sql.NullString `json:"picture_url" db:"picture_url"` // Profile picture URL
        AuthProvider  string         `json:"auth_provider" db:"auth_provider"` // 'email' or 'google'
        Money         float64        `json:"money" db:"money"`
        Topup         int            `json:"topup" db:"topup"`
        LastTopupAt   *time.Time     `json:"last_topup_at,omitempty" db:"last_topup_at"`
        CreatedAt     time.Time      `json:"created_at" db:"created_at"`
        UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}

// RefreshToken represents a stored refresh token (for logout functionality)
type RefreshToken struct {
        ID          string    `json:"id" db:"id"`
        UserID      string    `json:"user_id" db:"user_id"`
        Token       string    `json:"token" db:"token"`
        ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
        CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// JWT Claims structures
type AccessTokenClaims struct {
        UserID   string `json:"user_id"`
        Email    string `json:"email"`
        Nickname string `json:"nickname"`
        jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
        UserID string `json:"user_id"`
        jwt.RegisteredClaims
}

// Google OAuth structures
type GoogleUser struct {
        ID            string `json:"id"`
        Email         string `json:"email"`
        VerifiedEmail bool   `json:"verified_email"`
        Name          string `json:"name"`
        GivenName     string `json:"given_name"`
        FamilyName    string `json:"family_name"`
        Picture       string `json:"picture"`
        Locale        string `json:"locale"`
}

type OAuthState struct {
        State       string    `json:"state"`
        RedirectURL string    `json:"redirect_url"`
        CreatedAt   time.Time `json:"created_at"`
        ExpiresAt   time.Time `json:"expires_at"`
}


// Admin represents an admin user
type Admin struct {
        ID        string    `json:"id" db:"id"`
        Username  string    `json:"username" db:"username"`
        Email     string    `json:"email" db:"email"`
        PasswordHash string `json:"-" db:"password_hash"`
        IsActive  bool      `json:"is_active" db:"is_active"`
        LastLogin *time.Time `json:"last_login" db:"last_login"`
        CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Bet represents a betting transaction
type Bet struct {
        BetID        string     `json:"bet_id" db:"bet_id"`
        UserID       string     `json:"user_id" db:"user_id"`
        MatchID      string     `json:"match_id" db:"match_id"`
        BetType      string     `json:"bet_type" db:"bet_type"` // "home", "draw", "away"
        BetAmount    float64    `json:"bet_amount" db:"bet_amount"`
        Odds         float64    `json:"odds" db:"odds"`
        PotentialWin float64    `json:"potential_win" db:"potential_win"`
        Status       string     `json:"status" db:"status"` // "pending", "won", "lost"
        HomeTeam     string     `json:"home_team" db:"home_team"`
        AwayTeam     string     `json:"away_team" db:"away_team"`
        CreatedAt    time.Time  `json:"created_at" db:"created_at"`
        CommenceTime *time.Time `json:"commence_time,omitempty" db:"commence_time"`
}

// Match represents a football match with odds
type Match struct {
        ID          string    `json:"id" db:"id"`
        APIID       string    `json:"api_id" db:"api_id"`
        HomeTeam    string    `json:"home_team" db:"home_team"`
        AwayTeam    string    `json:"away_team" db:"away_team"`
        CommenceTime time.Time `json:"commence_time" db:"commence_time"`
        HomeOdds    *float64  `json:"home_odds" db:"home_odds"`
        DrawOdds    *float64  `json:"draw_odds" db:"draw_odds"`
        AwayOdds    *float64  `json:"away_odds" db:"away_odds"`
        Completed   bool      `json:"completed" db:"completed"`
        HomeScore   *int      `json:"home_score" db:"home_score"`
        AwayScore   *int      `json:"away_score" db:"away_score"`
        Calculated  bool      `json:"calculated" db:"calculated"`
        Result      *string   `json:"result" db:"result"` // "home", "draw", "away"
}

// API Response DTOs (Data Transfer Objects)

// Auth responses
type RegisterResponse struct {
        Success      bool         `json:"success"`
        Message      string       `json:"message"`
        AccessToken  string       `json:"access_token"`
        RefreshToken string       `json:"refresh_token"`
        User         UserResponse `json:"user"`
}

type LoginResponse struct {
        Success      bool         `json:"success"`
        AccessToken  string       `json:"access_token"`
        RefreshToken string       `json:"refresh_token"`
        User         UserResponse `json:"user"`
}

// Refresh token response
type RefreshResponse struct {
        Success     bool   `json:"success"`
        AccessToken string `json:"access_token"`
}

type UserResponse struct {
        ID           string     `json:"id"`
        Email        string     `json:"email"`
        Nickname     string     `json:"nickname"`
        Money        float64    `json:"money"`
        Topup        int        `json:"topup"`
        LastTopupAt  *time.Time `json:"last_topup_at,omitempty"`
        Bets         int        `json:"bets"`
        WonBets      int        `json:"won_bets"`
        SettledBets  int        `json:"settled_bets"`
        AvgOdds      float64    `json:"avg_odds"`
        AuthProvider string     `json:"auth_provider,omitempty"`
}

type TopupResponse struct {
        Success    bool    `json:"success"`
        Message    string  `json:"message"`
        NewBalance float64 `json:"new_balance"`
}

// Bet responses
type BetResponse struct {
        Success bool `json:"success"`
        Bet     BetInfo `json:"bet"`
}

type BetInfo struct {
        ID           string  `json:"id"`
        Amount       float64 `json:"amount"`
        Odds         float64 `json:"odds"`
        PotentialWin float64 `json:"potential_win"`
        NewBalance   float64 `json:"new_balance"`
}

type BetsResponse struct {
        Success bool  `json:"success"`
        Bets    []BetDisplay `json:"bets"`
}

type BetDisplay struct {
        ID           string    `json:"bet_id"`
        MatchID      string    `json:"match_id"`
        BetType      string    `json:"bet_type"`
        BetAmount    float64   `json:"bet_amount"`
        Odds         float64   `json:"odds"`
        PotentialWin float64   `json:"potential_win"`
        Status       string    `json:"status"`
        HomeTeam     string    `json:"home_team"`
        AwayTeam     string    `json:"away_team"`
        CreatedAt    time.Time `json:"created_at"`
        CommenceTime *time.Time `json:"commence_time,omitempty"`
}

// Match responses
type MatchesResponse struct {
        Success bool           `json:"success"`
        Matches []MatchDisplay `json:"matches"`
}

type MatchDisplay struct {
        ID           string    `json:"id"` // Uses api_id as id
        HomeTeam     string    `json:"home_team"`
        AwayTeam     string    `json:"away_team"`
        CommenceTime time.Time `json:"commence_time"`
        HomeOdds     *float64  `json:"home_odds"`
        DrawOdds     *float64  `json:"draw_odds"`
        AwayOdds     *float64  `json:"away_odds"`
}

// Players responses
type PlayersResponse struct {
        Success    bool            `json:"success"`
        Players    []PlayerDisplay `json:"players"`
        Pagination PaginationInfo  `json:"pagination"`
}

type PlayerDisplay struct {
        ID           string  `json:"id"`
        Nickname     string  `json:"nickname"`
        Money        float64 `json:"money"`
        Bets         int     `json:"bets"`
        WonBets      int     `json:"won_bets"`
        SettledBets  int     `json:"settled_bets"`
        AvgOdds      float64 `json:"avg_odds"`
        Topup        int     `json:"topup"`
        Created      string  `json:"created"` // ISO string
        Updated      string  `json:"updated"` // ISO string
}

type PaginationInfo struct {
        Limit    int  `json:"limit"`
        Offset   int  `json:"offset"`
        Total    int  `json:"total"`
        HasMore  bool `json:"has_more"`
}

// Request DTOs
type RegisterRequest struct {
        Email        string `json:"email"`
        Password     string `json:"password"`
        Nickname     string `json:"nickname"`
        AgeConfirmed bool   `json:"age_confirmed"`
}

type LoginRequest struct {
        Identifier string `json:"identifier"` // email or nickname
        Password   string `json:"password"`
}

type ChangePasswordRequest struct {
        CurrentPassword string `json:"current_password"`
        NewPassword     string `json:"new_password"`
}

type PlaceBetRequest struct {
        MatchID    string  `json:"match_id"`
        BetType    string  `json:"bet_type"` // "home", "draw", "away"
        BetAmount  float64 `json:"bet_amount"`
        Odds       float64 `json:"odds"`
        HomeTeam   string  `json:"home_team"`
        AwayTeam   string  `json:"away_team"`
}

// Generic API response
type APIResponse struct {
        Success bool        `json:"success"`
        Data    interface{} `json:"data,omitempty"`
        Error   string      `json:"error,omitempty"`
}

// Health check response
type HealthResponse struct {
        // Mobile app format (основной формат)
        Ok            bool   `json:"ok"`
        Status        string `json:"status"`        // Для совместимости
        UptimeSeconds int64  `json:"uptime"`        // в секундах
        ClientIP      string `json:"client_ip"`
        Time          string `json:"time"`          // ISO 8601
        Version       string `json:"version"`
        UsersCount    int    `json:"users_count"`
        BetsCount     int    `json:"bets_count"`
        MatchesCount  int    `json:"matches_count"`
        DatabaseStatus string `json:"database_status"`
        Port          int    `json:"port"`          // Для информации
}

// Root endpoint response
type RootResponse struct {
        Message   string            `json:"message"`
        Endpoints map[string]string `json:"endpoints"`
}

// Database connection interface for dependency injection
type Database interface {
        // User management
        GetUserByEmail(email string) (*User, error)
        GetUserByNickname(nickname string) (*User, error)
        GetUserByGoogleID(googleID string) (*User, error)
        GetUserByID(id string) (*User, error)
        CreateUser(email, passwordHash, nickname string, initialBalance float64) (*User, error)
        CreateUserWithGoogle(googleID, email, nickname, pictureURL string, initialBalance float64) (*User, error)
        UpdateUserMoney(userID string, newMoney float64) error
        IncrementUserTopup(userID string) error
        GetUserLastTopupTime(userID string) (*time.Time, error)
        UpdateUserPassword(userID string, newPasswordHash string) error

        // JWT refresh token methods
        CreateRefreshToken(userID string, token string, expiresAt time.Time) (*RefreshToken, error)
        GetRefreshTokenByToken(token string) (*RefreshToken, error)
        DeleteRefreshToken(token string) error
        DeleteAllUserRefreshTokens(userID string) error // For logout from all devices

        GetUserBets(userID string, playerNickname string) ([]Bet, error)
        PlaceBet(bet *Bet) (*Bet, error)
        GetMatchByID(matchID string) (*Match, error)
        GetMatchByAPIID(apiID string) (*Match, error)

        GetMatches() ([]Match, error)
        GetPlayers(limit, offset int) ([]PlayerDisplay, error)
        GetTotalPlayers() (int, error)
        GetUserStats(userID string) (bets int, wonBets int, settledBets int, avgOdds float64, err error)

        GetDatabaseStats() (map[string]int, error)

        // Admin methods
        GetAdminByUsername(username string) (*Admin, error)
        UpdateAdminLastLogin(adminID string) error

        // Match sync methods
        UpsertMatch(match *Match) (*Match, error)
        UpdateMatchByAPIID(apiID string, match *Match) (*Match, error)
        GetCompletedUncalculatedMatches() ([]Match, error)
        UpdateMatchCalculated(apiID string, result string) error
        UpdateBetsStatusAndUserMoney(matchAPIID string, result string) error

        Ping() error
        Close() error
}

// PlayerStats is used for the database query result
type PlayerStats struct {
        ID           sql.NullString  `db:"id"`
        Nickname     sql.NullString  `db:"nickname"`
        Money        sql.NullFloat64 `db:"money"`
        Topup        sql.NullInt64   `db:"topup"`
        CreatedAt    sql.NullTime    `db:"created_at"`
        UpdatedAt    sql.NullTime    `db:"updated_at"`
        Bets         sql.NullInt64   `db:"bets"`
        WonBets      sql.NullInt64   `db:"won_bets"`
        SettledBets  sql.NullInt64   `db:"settled_bets"`
        AvgOdds      sql.NullFloat64 `db:"avg_odds"`
}
