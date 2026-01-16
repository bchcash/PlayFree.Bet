-- FreeBet Guru Database Schema
-- PostgreSQL initialization script for betting platform
-- Modern JWT authentication with Google OAuth support
--
-- Usage:
--   psql -U your_username -d your_database -f postgres_init.sql
--
-- After running this script:
-- 1. Set up Google OAuth credentials in your .env file
-- 2. Configure JWT_SECRET for token signing
-- 3. Start the API server

-- Drop all tables in correct order (respecting foreign keys)
DROP TABLE IF EXISTS bets CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS epl_matches CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Users table - supports both email/password and Google OAuth authentication
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) UNIQUE NOT NULL,
  nickname VARCHAR(10) UNIQUE NOT NULL,
  password_hash VARCHAR(255),                    -- NULL for OAuth users
  google_id VARCHAR(255) UNIQUE,                 -- Google OAuth ID
  picture_url VARCHAR(500),                      -- Profile picture URL
  auth_provider VARCHAR(20) DEFAULT 'email',     -- 'email' or 'google'
  money DECIMAL(15, 2) DEFAULT 0,               -- Virtual currency balance
  topup INTEGER DEFAULT 0,                       -- Number of balance top-ups
  last_topup_at TIMESTAMP,                       -- Last top-up timestamp
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Refresh tokens table for JWT authentication
CREATE TABLE refresh_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token VARCHAR(512) UNIQUE NOT NULL,           -- JWT refresh token
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Football matches table - stores match data and betting odds
CREATE TABLE epl_matches (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  api_id VARCHAR(255) UNIQUE,              -- External API identifier
  home_team VARCHAR(255) NOT NULL,         -- Home team name
  away_team VARCHAR(255) NOT NULL,         -- Away team name
  commence_time TIMESTAMP NOT NULL,        -- Match start time
  home_odds DECIMAL(10, 2),               -- Betting odds for home win
  draw_odds DECIMAL(10, 2),               -- Betting odds for draw
  away_odds DECIMAL(10, 2),               -- Betting odds for away win
  completed BOOLEAN DEFAULT FALSE,         -- Whether match has finished
  calculated BOOLEAN DEFAULT FALSE,        -- Whether bets have been processed
  result VARCHAR(10),                      -- 'home', 'draw', 'away' - match outcome
  home_score INTEGER,                      -- Final score for home team
  away_score INTEGER,                      -- Final score for away team
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User bets table - stores all betting transactions
CREATE TABLE bets (
  bet_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  match_id VARCHAR(255) NOT NULL,           -- Reference to epl_matches.api_id
  bet_type VARCHAR(50) NOT NULL,            -- 'home', 'draw', 'away'
  bet_amount DECIMAL(15, 2) NOT NULL,       -- Amount bet by user
  odds DECIMAL(10, 2) NOT NULL,             -- Odds at time of bet
  potential_win DECIMAL(15, 2) NOT NULL,    -- Potential payout
  status VARCHAR(50) DEFAULT 'pending',     -- 'pending', 'won', 'lost'
  home_team VARCHAR(255),                   -- Cached team names
  away_team VARCHAR(255),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_users_email ON users(email);
CREATE UNIQUE INDEX idx_users_nickname ON users(nickname);
CREATE UNIQUE INDEX idx_users_google_id ON users(google_id);
CREATE INDEX idx_users_auth_provider ON users(auth_provider);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_bets_user_id ON bets(user_id);
CREATE INDEX idx_bets_match_id ON bets(match_id);
CREATE INDEX idx_bets_status ON bets(status);
CREATE INDEX idx_epl_matches_api_id ON epl_matches(api_id);
CREATE INDEX idx_epl_matches_commence_time ON epl_matches(commence_time);
CREATE INDEX idx_epl_matches_result ON epl_matches(result);
CREATE INDEX idx_epl_matches_completed ON epl_matches(completed);
CREATE INDEX idx_epl_matches_calculated ON epl_matches(calculated);

-- Database initialization complete
-- Ready for user registration via email/password or Google OAuth