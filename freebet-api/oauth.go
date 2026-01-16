package main

import (
        "context"
        "crypto/rand"
        "encoding/base64"
        "encoding/json"
        "fmt"
        "net/http"
        "strings"
        "time"

	"golang.org/x/oauth2"
)

// OAuth state storage (in production, use Redis or database)
var oauthStates = make(map[string]*OAuthState)

// GenerateOAuthState generates a random state parameter for OAuth
func generateOAuthState(redirectURL string) (string, error) {
        // Generate random bytes
        bytes := make([]byte, 32)
        if _, err := rand.Read(bytes); err != nil {
                return "", err
        }

        // Encode to base64 URL-safe string
        state := base64.URLEncoding.EncodeToString(bytes)

        // Store state with expiration
        oauthStates[state] = &OAuthState{
                State:       state,
                RedirectURL: redirectURL,
                CreatedAt:   time.Now(),
                ExpiresAt:   time.Now().Add(10 * time.Minute), // 10 minutes
        }

        return state, nil
}

// ValidateOAuthState validates the OAuth state parameter
func validateOAuthState(state string) (*OAuthState, bool) {
        oauthState, exists := oauthStates[state]
        if !exists {
                return nil, false
        }

        // Check if expired
        if time.Now().After(oauthState.ExpiresAt) {
                delete(oauthStates, state)
                return nil, false
        }

        // Clean up used state
        delete(oauthStates, state)

        return oauthState, true
}

// GetGoogleOAuthConfig returns the Google OAuth2 configuration
func getGoogleOAuthConfig(config *Config) *oauth2.Config {
        return &oauth2.Config{
                ClientID:     config.GoogleClientID,
                ClientSecret: config.GoogleClientSecret,
                RedirectURL:  config.GoogleRedirectURL,
                Scopes:       []string{"openid", "profile", "email"},
                Endpoint: oauth2.Endpoint{
                        AuthURL:  "https://accounts.google.com/o/oauth2/auth",
                        TokenURL: "https://oauth2.googleapis.com/token",
                },
        }
}

// GetGoogleUserInfo fetches user information from Google
func getGoogleUserInfo(token *oauth2.Token, config *Config) (*GoogleUser, error) {
        oauthConfig := getGoogleOAuthConfig(config)

        // Create HTTP client with the token
        client := oauthConfig.Client(context.Background(), token)

        // Fetch user info from Google
        resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
        if err != nil {
                return nil, fmt.Errorf("failed to fetch user info: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                return nil, fmt.Errorf("Google API returned status: %d", resp.StatusCode)
        }

        // Parse the response
        var googleUser GoogleUser
        if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
                return nil, fmt.Errorf("failed to decode user info: %w", err)
        }

        return &googleUser, nil
}

// GenerateNicknameFromGoogleEmail generates a nickname from Google email
func generateNicknameFromGoogleEmail(email string) string {
        // Extract part before @ and clean it
        parts := strings.Split(email, "@")
        if len(parts) == 0 {
                return "user"
        }

        nickname := strings.ToLower(parts[0])
        // Remove special characters and limit length
        nickname = strings.ReplaceAll(nickname, ".", "")
        nickname = strings.ReplaceAll(nickname, "_", "")
        nickname = strings.ReplaceAll(nickname, "-", "")

        // Ensure minimum length and maximum length
        if len(nickname) < 3 {
                nickname = nickname + "user"
        }
        if len(nickname) > 10 {
                nickname = nickname[:10]
        }

        return nickname
}