package main

import (
        "context"
        "fmt"
        "net/url"
        "strings"
        "time"

        "github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDB implements the Database interface using PostgreSQL
type PostgresDB struct {
        pool   *pgxpool.Pool
        logger *Logger
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(databaseURL string, dbConfig *Config, logger *Logger) (*PostgresDB, error) {
        logger.LogDB("Creating PostgreSQL connection pool")

        // Parse and modify DATABASE_URL for better connection handling
        parsedURL, err := url.Parse(databaseURL)
        if err != nil {
                return nil, fmt.Errorf("invalid database URL: %w", err)
        }

        // Get password (url.User.Password() returns string, error)
        password, _ := parsedURL.User.Password()

        // Build connection string
        connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
                parsedURL.User.Username(),
                password,
                parsedURL.Hostname(),
                parsedURL.Port(),
                strings.TrimPrefix(parsedURL.Path, "/"),
        )

        logger.LogDB("Connecting to PostgreSQL at %s:%s database: %s",
                parsedURL.Hostname(), parsedURL.Port(), strings.TrimPrefix(parsedURL.Path, "/"))

        // Configure connection pool
        config, err := pgxpool.ParseConfig(connString)
        if err != nil {
                return nil, fmt.Errorf("failed to parse database config: %w", err)
        }

        // Set configurable pool settings
        config.MaxConns = int32(dbConfig.DBMaxConns)
        config.MinConns = int32(dbConfig.DBMinConns)
        config.MaxConnLifetime = time.Duration(dbConfig.DBMaxLifetime) * time.Second
        config.MaxConnIdleTime = time.Duration(dbConfig.DBMaxIdleTime) * time.Second

        // Create connection pool
        pool, err := pgxpool.NewWithConfig(context.Background(), config)
        if err != nil {
                return nil, fmt.Errorf("failed to create connection pool: %w", err)
        }

        // Test connection
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        if err := pool.Ping(ctx); err != nil {
                pool.Close()
                return nil, fmt.Errorf("failed to ping database: %w", err)
        }

        logger.LogDB("PostgreSQL connection established")

        return &PostgresDB{
                pool:   pool,
                logger: logger,
        }, nil
}

// Ping tests the database connection
func (db *PostgresDB) Ping() error {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        return db.pool.Ping(ctx)
}

// Close closes the database connection pool
func (db *PostgresDB) Close() error {
        db.logger.LogDB("Closing PostgreSQL connection pool")
        db.pool.Close()
        return nil
}

// User methods
func (db *PostgresDB) GetUserByEmail(email string) (*User, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("SELECT users by email", []interface{}{email}, time.Since(start))
        }()

        query := `
                SELECT id, email, nickname, password_hash, google_id, picture_url, auth_provider,
                       money, topup, last_topup_at, created_at, updated_at
                FROM users WHERE email = $1`

        var user User
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err := db.pool.QueryRow(ctx, query, email).Scan(
                &user.ID, &user.Email, &user.Nickname, &user.PasswordHash, &user.GoogleID,
                &user.PictureURL, &user.AuthProvider, &user.Money, &user.Topup,
                &user.LastTopupAt, &user.CreatedAt, &user.UpdatedAt,
        )

        if err != nil {
                return nil, err
        }

        return &user, nil
}

func (db *PostgresDB) GetUserByNickname(nickname string) (*User, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("SELECT users by nickname", []interface{}{nickname}, time.Since(start))
        }()

        query := `
                SELECT id, email, nickname, password_hash, google_id, picture_url, auth_provider,
                       money, topup, last_topup_at, created_at, updated_at
                FROM users WHERE nickname = $1`

        var user User
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err := db.pool.QueryRow(ctx, query, nickname).Scan(
                &user.ID, &user.Email, &user.Nickname, &user.PasswordHash, &user.GoogleID,
                &user.PictureURL, &user.AuthProvider, &user.Money, &user.Topup,
                &user.LastTopupAt, &user.CreatedAt, &user.UpdatedAt,
        )

        if err != nil {
                return nil, err
        }

        return &user, nil
}

func (db *PostgresDB) GetUserByID(id string) (*User, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("SELECT users by ID", []interface{}{id}, time.Since(start))
        }()

        query := `
                SELECT id, email, nickname, password_hash, google_id, picture_url, auth_provider,
                       money, topup, last_topup_at, created_at, updated_at
                FROM users WHERE id = $1`

        var user User
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err := db.pool.QueryRow(ctx, query, id).Scan(
                &user.ID, &user.Email, &user.Nickname, &user.PasswordHash, &user.GoogleID,
                &user.PictureURL, &user.AuthProvider, &user.Money, &user.Topup,
                &user.LastTopupAt, &user.CreatedAt, &user.UpdatedAt,
        )

        if err != nil {
                return nil, err
        }

        return &user, nil
}

func (db *PostgresDB) CreateUser(email, passwordHash, nickname string, initialBalance float64) (*User, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("INSERT user", []interface{}{email, nickname}, time.Since(start))
        }()

        query := `
                INSERT INTO users (email, nickname, password_hash, auth_provider, money, topup, last_topup_at)
                VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
                RETURNING id, email, nickname, password_hash, google_id, picture_url,
                         auth_provider, money, topup, last_topup_at, created_at, updated_at`

        var user User
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err := db.pool.QueryRow(ctx, query, email, nickname, passwordHash, "email", initialBalance, 1).Scan(
                &user.ID, &user.Email, &user.Nickname, &user.PasswordHash, &user.GoogleID,
                &user.PictureURL, &user.AuthProvider, &user.Money, &user.Topup,
                &user.LastTopupAt, &user.CreatedAt, &user.UpdatedAt,
        )

        if err != nil {
                return nil, err
        }

        return &user, nil
}

func (db *PostgresDB) UpdateUserMoney(userID string, newMoney float64) error {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("UPDATE user money", []interface{}{userID, newMoney}, time.Since(start))
        }()

        query := `UPDATE users SET money = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        _, err := db.pool.Exec(ctx, query, newMoney, userID)
        return err
}

func (db *PostgresDB) IncrementUserTopup(userID string) error {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("UPDATE user topup", []interface{}{userID}, time.Since(start))
        }()

        query := `UPDATE users SET topup = COALESCE(topup, 0) + 1, last_topup_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1`

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        _, err := db.pool.Exec(ctx, query, userID)
        return err
}

func (db *PostgresDB) GetUserLastTopupTime(userID string) (*time.Time, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("SELECT user last_topup_at", []interface{}{userID}, time.Since(start))
        }()

        query := `SELECT last_topup_at FROM users WHERE id = $1`

        var lastTopupAt *time.Time
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err := db.pool.QueryRow(ctx, query, userID).Scan(&lastTopupAt)
        if err != nil {
                return nil, err
        }

        return lastTopupAt, nil
}

func (db *PostgresDB) UpdateUserPassword(userID string, newPasswordHash string) error {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("UPDATE user password", []interface{}{userID}, time.Since(start))
        }()

        query := `UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        _, err := db.pool.Exec(ctx, query, newPasswordHash, userID)
        return err
}


// Google OAuth User methods
func (db *PostgresDB) GetUserByGoogleID(googleID string) (*User, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("SELECT user by google_id", []interface{}{googleID[:10] + "..."}, time.Since(start))
        }()

        query := `
                SELECT u.id, u.email, u.nickname, u.password_hash, u.google_id, u.picture_url,
                       u.auth_provider, u.money, u.topup, u.last_topup_at, u.created_at, u.updated_at
                FROM users u
                WHERE u.google_id = $1`

        var user User
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err := db.pool.QueryRow(ctx, query, googleID).Scan(
                &user.ID, &user.Email, &user.Nickname, &user.PasswordHash, &user.GoogleID,
                &user.PictureURL, &user.AuthProvider, &user.Money, &user.Topup,
                &user.LastTopupAt, &user.CreatedAt, &user.UpdatedAt,
        )

        if err != nil {
                return nil, err
        }

        return &user, nil
}

func (db *PostgresDB) CreateUserWithGoogle(googleID, email, nickname, pictureURL string, initialBalance float64) (*User, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("INSERT user with google", []interface{}{email, nickname}, time.Since(start))
        }()

        query := `
                INSERT INTO users (email, nickname, google_id, picture_url, auth_provider, money, topup, last_topup_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP)
                RETURNING id, email, nickname, password_hash, google_id, picture_url,
                         auth_provider, money, topup, last_topup_at, created_at, updated_at`

        var user User
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err := db.pool.QueryRow(ctx, query, email, nickname, googleID, pictureURL, "google", initialBalance, 1).Scan(
                &user.ID, &user.Email, &user.Nickname, &user.PasswordHash, &user.GoogleID,
                &user.PictureURL, &user.AuthProvider, &user.Money, &user.Topup,
                &user.LastTopupAt, &user.CreatedAt, &user.UpdatedAt,
        )

        if err != nil {
                return nil, err
        }

        return &user, nil
}

// JWT Refresh Token methods
func (db *PostgresDB) CreateRefreshToken(userID string, token string, expiresAt time.Time) (*RefreshToken, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("INSERT refresh_token", []interface{}{userID}, time.Since(start))
        }()

        query := `
                INSERT INTO refresh_tokens (user_id, token, expires_at)
                VALUES ($1, $2, $3)
                RETURNING id, user_id, token, expires_at, created_at`

        var refreshToken RefreshToken
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err := db.pool.QueryRow(ctx, query, userID, token, expiresAt).Scan(
                &refreshToken.ID, &refreshToken.UserID, &refreshToken.Token,
                &refreshToken.ExpiresAt, &refreshToken.CreatedAt,
        )

        if err != nil {
                return nil, err
        }

        return &refreshToken, nil
}

func (db *PostgresDB) GetRefreshTokenByToken(token string) (*RefreshToken, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("SELECT refresh_token by token", []interface{}{token[:10] + "..."}, time.Since(start))
        }()

        query := `
                SELECT rt.id, rt.user_id, rt.token, rt.expires_at, rt.created_at
                FROM refresh_tokens rt
                WHERE rt.token = $1 AND rt.expires_at > CURRENT_TIMESTAMP`

        var refreshToken RefreshToken
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err := db.pool.QueryRow(ctx, query, token).Scan(
                &refreshToken.ID, &refreshToken.UserID, &refreshToken.Token,
                &refreshToken.ExpiresAt, &refreshToken.CreatedAt,
        )

        if err != nil {
                return nil, err
        }

        return &refreshToken, nil
}

func (db *PostgresDB) DeleteRefreshToken(token string) error {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("DELETE refresh_token", []interface{}{token[:10] + "..."}, time.Since(start))
        }()

        query := `DELETE FROM refresh_tokens WHERE token = $1`

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        _, err := db.pool.Exec(ctx, query, token)
        return err
}

func (db *PostgresDB) DeleteAllUserRefreshTokens(userID string) error {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("DELETE all user refresh_tokens", []interface{}{userID}, time.Since(start))
        }()

        query := `DELETE FROM refresh_tokens WHERE user_id = $1`

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        _, err := db.pool.Exec(ctx, query, userID)
        return err
}

// Bet methods
func (db *PostgresDB) GetUserBets(userID string, playerNickname string) ([]Bet, error) {
        start := time.Now()

        var query string
        var args []interface{}

        if playerNickname != "" {
                // Get bets for another player
                query = `
                        SELECT b.bet_id, b.user_id, b.match_id, b.bet_type, b.bet_amount,
                                   b.odds, b.potential_win, b.status, b.home_team, b.away_team, b.created_at,
                                   m.commence_time
                        FROM bets b
                        JOIN users u ON b.user_id = u.id
                        LEFT JOIN epl_matches m ON b.match_id = m.api_id
                        WHERE u.nickname = $1
                        ORDER BY b.created_at DESC`
                args = []interface{}{playerNickname}
        } else {
                // Get bets for current user
                query = `
                        SELECT b.bet_id, b.user_id, b.match_id, b.bet_type, b.bet_amount,
                                   b.odds, b.potential_win, b.status, b.home_team, b.away_team, b.created_at,
                                   m.commence_time
                        FROM bets b
                        LEFT JOIN epl_matches m ON b.match_id = m.api_id
                        WHERE b.user_id = $1
                        ORDER BY b.created_at DESC`
                args = []interface{}{userID}
        }

        defer func() {
                db.logger.LogSQL("SELECT bets", args, time.Since(start))
        }()

        ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
        defer cancel()

        rows, err := db.pool.Query(ctx, query, args...)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        var bets []Bet
        for rows.Next() {
                var bet Bet
                err := rows.Scan(
                        &bet.BetID, &bet.UserID, &bet.MatchID, &bet.BetType,
                        &bet.BetAmount, &bet.Odds, &bet.PotentialWin, &bet.Status,
                        &bet.HomeTeam, &bet.AwayTeam, &bet.CreatedAt, &bet.CommenceTime,
                )
                if err != nil {
                        return nil, err
                }
                bets = append(bets, bet)
        }

        return bets, rows.Err()
}

func (db *PostgresDB) PlaceBet(bet *Bet) (*Bet, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("INSERT bet", []interface{}{bet.UserID, bet.MatchID}, time.Since(start))
        }()

        query := `
                INSERT INTO bets (user_id, match_id, bet_type, bet_amount, odds, potential_win, status, home_team, away_team, created_at)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
                RETURNING bet_id`

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err := db.pool.QueryRow(ctx, query,
                bet.UserID, bet.MatchID, bet.BetType, bet.BetAmount,
                bet.Odds, bet.PotentialWin, bet.Status, bet.HomeTeam, bet.AwayTeam,
        ).Scan(&bet.BetID)

        if err != nil {
                return nil, err
        }

        return bet, nil
}

func (db *PostgresDB) GetMatchByID(matchID string) (*Match, error) {
        return db.GetMatchByAPIID(matchID)
}

// Match methods
func (db *PostgresDB) GetMatches() ([]Match, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("SELECT matches", nil, time.Since(start))
        }()

        query := `
                SELECT id, api_id, home_team, away_team, commence_time,
                           home_odds, draw_odds, away_odds, completed, home_score, away_score, calculated, result
                FROM epl_matches
                WHERE home_odds IS NOT NULL AND draw_odds IS NOT NULL AND away_odds IS NOT NULL
                        AND home_odds != 0 AND draw_odds != 0 AND away_odds != 0
                        AND commence_time > CURRENT_TIMESTAMP
                ORDER BY commence_time ASC`

        ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
        defer cancel()

        rows, err := db.pool.Query(ctx, query)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        var matches []Match
        for rows.Next() {
                var match Match
                err := rows.Scan(
                        &match.ID, &match.APIID, &match.HomeTeam, &match.AwayTeam,
                        &match.CommenceTime, &match.HomeOdds, &match.DrawOdds,
                        &match.AwayOdds, &match.Completed, &match.HomeScore, &match.AwayScore,
                        &match.Calculated, &match.Result,
                )
                if err != nil {
                        return nil, err
                }
                matches = append(matches, match)
        }

        return matches, rows.Err()
}

// Players methods
func (db *PostgresDB) GetPlayers(limit, offset int) ([]PlayerDisplay, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("SELECT players", []interface{}{limit, offset}, time.Since(start))
        }()

        query := `
                SELECT
                        u.id, u.nickname, u.money, u.topup, u.created_at, u.updated_at,
                        COUNT(b.bet_id) as bets,
                        COALESCE(SUM(CASE WHEN b.status = 'won' THEN 1 ELSE 0 END), 0) as won_bets,
                        COALESCE(SUM(CASE WHEN b.status IN ('won','lost') THEN 1 ELSE 0 END), 0) as settled_bets,
                        AVG(b.odds) as avg_odds
                FROM users u
                LEFT JOIN bets b ON u.id = b.user_id
                GROUP BY u.id, u.nickname, u.money, u.topup, u.created_at, u.updated_at
                ORDER BY bets DESC, u.money DESC
                LIMIT $1 OFFSET $2`

        ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
        defer cancel()

        rows, err := db.pool.Query(ctx, query, limit, offset)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        var players []PlayerDisplay
        for rows.Next() {
                var player PlayerDisplay
                var avgOdds *float64
                var createdAt, updatedAt time.Time

                err := rows.Scan(
                        &player.ID, &player.Nickname, &player.Money, &player.Topup,
                        &createdAt, &updatedAt, &player.Bets, &player.WonBets,
                        &player.SettledBets, &avgOdds,
                )
                if err != nil {
                        return nil, err
                }

                // Convert timestamps to ISO strings
                player.Created = createdAt.Format(time.RFC3339)
                player.Updated = updatedAt.Format(time.RFC3339)

                // Handle nullable avg_odds
                if avgOdds != nil {
                        player.AvgOdds = *avgOdds
                }

                players = append(players, player)
        }

        return players, rows.Err()
}

func (db *PostgresDB) GetTotalPlayers() (int, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("SELECT COUNT players", nil, time.Since(start))
        }()

        query := `SELECT COUNT(*) as total FROM users`

        var total int
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err := db.pool.QueryRow(ctx, query).Scan(&total)
        return total, err
}

// GetUserStats returns betting statistics for a user
func (db *PostgresDB) GetUserStats(userID string) (bets int, wonBets int, settledBets int, avgOdds float64, err error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("SELECT user stats", []interface{}{userID}, time.Since(start))
        }()

        query := `
                SELECT 
                        COUNT(*) as bets,
                        COALESCE(SUM(CASE WHEN status = 'won' THEN 1 ELSE 0 END), 0) as won_bets,
                        COALESCE(SUM(CASE WHEN status IN ('won','lost') THEN 1 ELSE 0 END), 0) as settled_bets,
                        COALESCE(AVG(odds), 0) as avg_odds
                FROM bets WHERE user_id = $1`

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err = db.pool.QueryRow(ctx, query, userID).Scan(&bets, &wonBets, &settledBets, &avgOdds)
        return
}

// GetDatabaseStats returns database statistics
func (db *PostgresDB) GetDatabaseStats() (map[string]int, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("SELECT database stats", nil, time.Since(start))
        }()

        stats := make(map[string]int)

        ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
        defer cancel()

        var count int

        // Get users count
        err := db.pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
        if err != nil {
                return nil, fmt.Errorf("failed to get users count: %w", err)
        }
        stats["users"] = count

        // Get refresh tokens count (replaces sessions)
        err = db.pool.QueryRow(ctx, "SELECT COUNT(*) FROM refresh_tokens").Scan(&count)
        if err != nil {
                return nil, fmt.Errorf("failed to get refresh tokens count: %w", err)
        }
        stats["sessions"] = count // Keep "sessions" key for backward compatibility

        // Get bets count
        err = db.pool.QueryRow(ctx, "SELECT COUNT(*) FROM bets").Scan(&count)
        if err != nil {
                return nil, fmt.Errorf("failed to get bets count: %w", err)
        }
        stats["bets"] = count

        // Get matches count (assuming epl_matches table)
        err = db.pool.QueryRow(ctx, "SELECT COUNT(*) FROM epl_matches").Scan(&count)
        if err != nil {
                // If epl_matches doesn't exist, try matches table or set to 0
                stats["matches"] = 0
                db.logger.LogWarning("epl_matches table not found, matches count set to 0")
        } else {
                stats["matches"] = count
        }

        return stats, nil
}

// Admin methods
func (db *PostgresDB) GetAdminByUsername(username string) (*Admin, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("SELECT admin by username", []interface{}{username}, time.Since(start))
        }()

        query := `SELECT id, username, email, password_hash, is_active, last_login, created_at
                FROM admins WHERE username = $1 AND is_active = true`

        var admin Admin
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err := db.pool.QueryRow(ctx, query, username).Scan(
                &admin.ID, &admin.Username, &admin.Email, &admin.PasswordHash,
                &admin.IsActive, &admin.LastLogin, &admin.CreatedAt,
        )

        if err != nil {
                return nil, err
        }

        return &admin, nil
}

func (db *PostgresDB) UpdateAdminLastLogin(adminID string) error {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("UPDATE admin last_login", []interface{}{adminID}, time.Since(start))
        }()

        query := `UPDATE admins SET last_login = CURRENT_TIMESTAMP WHERE id = $1`

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        _, err := db.pool.Exec(ctx, query, adminID)
        return err
}

// Match sync methods
func (db *PostgresDB) UpsertMatch(match *Match) (*Match, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("UPSERT match", []interface{}{match.APIID}, time.Since(start))
        }()

        // Check if match exists
        existingMatch, err := db.GetMatchByAPIID(match.APIID)
        if err == nil && existingMatch != nil {
                // Update existing match
                return db.UpdateMatchByAPIID(match.APIID, match)
        }

        // Create new match
        query := `
                INSERT INTO epl_matches (
                        api_id, home_team, away_team, commence_time,
                        home_score, away_score, home_odds, draw_odds, away_odds,
                        completed, calculated, result
                )
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
                RETURNING id, api_id, home_team, away_team, commence_time,
                          home_odds, draw_odds, away_odds, completed, home_score, away_score, calculated, result`

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        var resultMatch Match
        homeScore := -1
        awayScore := -1
        if match.HomeScore != nil {
                homeScore = *match.HomeScore
        }
        if match.AwayScore != nil {
                awayScore = *match.AwayScore
        }

        err = db.pool.QueryRow(ctx, query,
                match.APIID, match.HomeTeam, match.AwayTeam, match.CommenceTime,
                homeScore, awayScore, match.HomeOdds, match.DrawOdds, match.AwayOdds,
                match.Completed, match.Calculated, match.Result,
        ).Scan(
                &resultMatch.ID, &resultMatch.APIID, &resultMatch.HomeTeam, &resultMatch.AwayTeam,
                &resultMatch.CommenceTime, &resultMatch.HomeOdds, &resultMatch.DrawOdds,
                &resultMatch.AwayOdds, &resultMatch.Completed, &resultMatch.HomeScore,
                &resultMatch.AwayScore, &resultMatch.Calculated, &resultMatch.Result,
        )

        if err != nil {
                return nil, err
        }

        return &resultMatch, nil
}

func (db *PostgresDB) GetMatchByAPIID(apiID string) (*Match, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("SELECT match by API ID", []interface{}{apiID}, time.Since(start))
        }()

        query := `SELECT id, api_id, home_team, away_team, commence_time,
                         home_odds, draw_odds, away_odds, completed, home_score, away_score, calculated, result
                  FROM epl_matches WHERE api_id = $1`

        var match Match
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        err := db.pool.QueryRow(ctx, query, apiID).Scan(
                &match.ID, &match.APIID, &match.HomeTeam, &match.AwayTeam,
                &match.CommenceTime, &match.HomeOdds, &match.DrawOdds,
                &match.AwayOdds, &match.Completed, &match.HomeScore, &match.AwayScore,
                &match.Calculated, &match.Result,
        )

        if err != nil {
                return nil, err
        }

        return &match, nil
}

func (db *PostgresDB) UpdateMatchByAPIID(apiID string, match *Match) (*Match, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("UPDATE match by API ID", []interface{}{apiID}, time.Since(start))
        }()

        // Build dynamic update query
        updates := []string{}
        values := []interface{}{}
        paramCount := 1

        if match.HomeTeam != "" {
                updates = append(updates, fmt.Sprintf("home_team = $%d", paramCount))
                values = append(values, match.HomeTeam)
                paramCount++
        }
        if match.AwayTeam != "" {
                updates = append(updates, fmt.Sprintf("away_team = $%d", paramCount))
                values = append(values, match.AwayTeam)
                paramCount++
        }
        if !match.CommenceTime.IsZero() {
                updates = append(updates, fmt.Sprintf("commence_time = $%d", paramCount))
                values = append(values, match.CommenceTime)
                paramCount++
        }
        if match.HomeOdds != nil {
                updates = append(updates, fmt.Sprintf("home_odds = $%d", paramCount))
                values = append(values, *match.HomeOdds)
                paramCount++
        }
        if match.DrawOdds != nil {
                updates = append(updates, fmt.Sprintf("draw_odds = $%d", paramCount))
                values = append(values, *match.DrawOdds)
                paramCount++
        }
        if match.AwayOdds != nil {
                updates = append(updates, fmt.Sprintf("away_odds = $%d", paramCount))
                values = append(values, *match.AwayOdds)
                paramCount++
        }
        if match.HomeScore != nil {
                updates = append(updates, fmt.Sprintf("home_score = $%d", paramCount))
                values = append(values, *match.HomeScore)
                paramCount++
        }
        if match.AwayScore != nil {
                updates = append(updates, fmt.Sprintf("away_score = $%d", paramCount))
                values = append(values, *match.AwayScore)
                paramCount++
        }
        updates = append(updates, fmt.Sprintf("completed = $%d", paramCount))
        values = append(values, match.Completed)
        paramCount++

        updates = append(updates, "updated_at = CURRENT_TIMESTAMP")

        query := fmt.Sprintf(`
                UPDATE epl_matches
                SET %s
                WHERE api_id = $%d
                RETURNING id, api_id, home_team, away_team, commence_time,
                          home_odds, draw_odds, away_odds, completed, home_score, away_score, calculated, result`,
                strings.Join(updates, ", "), paramCount)

        values = append(values, apiID)

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        var resultMatch Match
        err := db.pool.QueryRow(ctx, query, values...).Scan(
                &resultMatch.ID, &resultMatch.APIID, &resultMatch.HomeTeam, &resultMatch.AwayTeam,
                &resultMatch.CommenceTime, &resultMatch.HomeOdds, &resultMatch.DrawOdds,
                &resultMatch.AwayOdds, &resultMatch.Completed, &resultMatch.HomeScore,
                &resultMatch.AwayScore, &resultMatch.Calculated, &resultMatch.Result,
        )

        if err != nil {
                return nil, err
        }

        return &resultMatch, nil
}

func (db *PostgresDB) GetCompletedUncalculatedMatches() ([]Match, error) {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("SELECT completed uncalculated matches", nil, time.Since(start))
        }()

        query := `SELECT id, api_id, home_team, away_team, commence_time,
                         home_odds, draw_odds, away_odds, completed, home_score, away_score, calculated, result
                  FROM epl_matches
                  WHERE completed = TRUE AND calculated = FALSE
                        AND home_score IS NOT NULL AND away_score IS NOT NULL
                        AND home_score != -1 AND away_score != -1`

        ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
        defer cancel()

        rows, err := db.pool.Query(ctx, query)
        if err != nil {
                return nil, err
        }
        defer rows.Close()

        var matches []Match
        for rows.Next() {
                var match Match
                err := rows.Scan(
                        &match.ID, &match.APIID, &match.HomeTeam, &match.AwayTeam,
                        &match.CommenceTime, &match.HomeOdds, &match.DrawOdds,
                        &match.AwayOdds, &match.Completed, &match.HomeScore, &match.AwayScore,
                        &match.Calculated, &match.Result,
                )
                if err != nil {
                        return nil, err
                }
                matches = append(matches, match)
        }

        return matches, rows.Err()
}

func (db *PostgresDB) UpdateMatchCalculated(apiID string, result string) error {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("UPDATE match calculated", []interface{}{apiID, result}, time.Since(start))
        }()

        query := `UPDATE epl_matches SET calculated = TRUE, result = $1, updated_at = NOW() WHERE api_id = $2`

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        _, err := db.pool.Exec(ctx, query, result, apiID)
        return err
}

func (db *PostgresDB) UpdateBetsStatusAndUserMoney(matchAPIID string, result string) error {
        start := time.Now()
        defer func() {
                db.logger.LogSQL("UPDATE bets status and user money", []interface{}{matchAPIID, result}, time.Since(start))
        }()

        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        // Start transaction
        tx, err := db.pool.Begin(ctx)
        if err != nil {
                return err
        }
        defer tx.Rollback(ctx)

        // Update bets status
        updateBetsQuery := `
                UPDATE bets
                SET status = CASE WHEN bet_type = $1 THEN 'won' ELSE 'lost' END
                WHERE match_id = $2 AND status = 'pending'
                RETURNING user_id, potential_win, status`

        rows, err := tx.Query(ctx, updateBetsQuery, result, matchAPIID)
        if err != nil {
                return err
        }
        defer rows.Close()

        // Collect winning bets
        type winningBet struct {
                userID      string
                potentialWin float64
        }
        var winningBets []winningBet

        for rows.Next() {
                var userID string
                var potentialWin float64
                var status string
                if err := rows.Scan(&userID, &potentialWin, &status); err != nil {
                        return err
                }
                if status == "won" {
                        winningBets = append(winningBets, winningBet{userID: userID, potentialWin: potentialWin})
                }
        }

        // Update user money for winners
        for _, bet := range winningBets {
                updateMoneyQuery := `UPDATE users SET money = money + $1 WHERE id = $2`
                if _, err := tx.Exec(ctx, updateMoneyQuery, bet.potentialWin, bet.userID); err != nil {
                        return err
                }
        }

        // Commit transaction
        if err := tx.Commit(ctx); err != nil {
                return err
        }

        return nil
}