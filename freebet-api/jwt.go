package main

import (
        "crypto/rand"
        "encoding/hex"
        "time"

        "github.com/golang-jwt/jwt/v5"
)

// generateAccessToken generates a new JWT access token
func generateAccessToken(user *User, config *Config) (string, error) {
        now := time.Now()
        claims := AccessTokenClaims{
                UserID:   user.ID,
                Email:    user.Email,
                Nickname: user.Nickname,
                RegisteredClaims: jwt.RegisteredClaims{
                        IssuedAt:  jwt.NewNumericDate(now),
                        ExpiresAt: jwt.NewNumericDate(now.Add(config.JWTAccessTokenTTL)),
                        NotBefore: jwt.NewNumericDate(now),
                        Issuer:    "freebet-api",
                        Subject:   user.ID,
                },
        }

        token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
        return token.SignedString([]byte(config.JWTSecret))
}

// generateRefreshToken generates a new JWT refresh token
func generateRefreshToken(userID string, config *Config) (string, error) {
        now := time.Now()
        claims := RefreshTokenClaims{
                UserID: userID,
                RegisteredClaims: jwt.RegisteredClaims{
                        IssuedAt:  jwt.NewNumericDate(now),
                        ExpiresAt: jwt.NewNumericDate(now.Add(config.JWTRefreshTokenTTL)),
                        NotBefore: jwt.NewNumericDate(now),
                        Issuer:    "freebet-api",
                        Subject:   userID,
                        ID:        generateTokenID(),
                },
        }

        token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
        return token.SignedString([]byte(config.JWTSecret))
}

// validateAccessToken validates and parses an access token
func validateAccessToken(tokenString string, config *Config) (*AccessTokenClaims, error) {
        claims := &AccessTokenClaims{}

        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                        return nil, jwt.ErrSignatureInvalid
                }
                return []byte(config.JWTSecret), nil
        })

        if err != nil {
                return nil, err
        }

        if !token.Valid {
                return nil, jwt.ErrTokenMalformed
        }

        return claims, nil
}

// validateRefreshToken validates and parses a refresh token
func validateRefreshToken(tokenString string, config *Config) (*RefreshTokenClaims, error) {
        claims := &RefreshTokenClaims{}

        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                        return nil, jwt.ErrSignatureInvalid
                }
                return []byte(config.JWTSecret), nil
        })

        if err != nil {
                return nil, err
        }

        if !token.Valid {
                return nil, jwt.ErrTokenMalformed
        }

        return claims, nil
}

// generateTokenID generates a random token ID for refresh tokens
func generateTokenID() string {
        bytes := make([]byte, 16)
        rand.Read(bytes)
        return hex.EncodeToString(bytes)
}

// refreshAccessToken refreshes an access token using a valid refresh token
func refreshAccessToken(refreshTokenString string, db Database, config *Config) (string, error) {
        // Validate refresh token
        refreshClaims, err := validateRefreshToken(refreshTokenString, config)
        if err != nil {
                return "", err
        }

        // Check if refresh token exists in database (optional, but good practice)
        storedToken, err := db.GetRefreshTokenByToken(refreshTokenString)
        if err != nil || storedToken == nil {
                return "", jwt.ErrTokenNotValidYet // Token not found or expired
        }

        // Get user data
        user, err := db.GetUserByID(refreshClaims.UserID)
        if err != nil {
                return "", err
        }

        // Generate new access token
        return generateAccessToken(user, config)
}