package main

import (
        "bytes"
        "encoding/json"
        "fmt"
        "io"
        "net/http"
        "net/url"
        "strconv"
        "time"
)

// OddsAPIEvent represents an event from Odds API
type OddsAPIEvent struct {
        ID           string    `json:"id"`
        SportKey     string    `json:"sport_key"`
        CommenceTime time.Time `json:"commence_time"`
        HomeTeam     string    `json:"home_team"`
        AwayTeam     string    `json:"away_team"`
        Bookmakers   []struct {
                Key         string    `json:"key"`
                Title       string    `json:"title"`
                LastUpdate  time.Time `json:"last_update"`
                Markets     []struct {
                        Key      string `json:"key"`
                        Outcomes []struct {
                                Name  string  `json:"name"`
                                Price float64 `json:"price"`
                        } `json:"outcomes"`
                } `json:"markets"`
        } `json:"bookmakers"`
}

// ScoresAPIEvent represents a score event from Odds API
type ScoresAPIEvent struct {
        ID           string    `json:"id"`
        SportKey     string    `json:"sport_key"`
        CommenceTime time.Time `json:"commence_time"`
        HomeTeam     string    `json:"home_team"`
        AwayTeam     string    `json:"away_team"`
        Completed    bool      `json:"completed"`
        Scores       []struct {
                Name  string `json:"name"`
                Score string `json:"score"`
        } `json:"scores"`
}

// APIStats represents API usage statistics
type APIStats struct {
        RequestsRemaining string `json:"requests_remaining"`
        RequestsUsed      string `json:"requests_used"`
}

// fetchOddsFromAPI fetches odds from The Odds API
func fetchOddsFromAPI(apiKey string) ([]OddsAPIEvent, *APIStats, error) {
        if apiKey == "" {
                return nil, nil, fmt.Errorf("ODDS_API_KEY is not configured")
        }

        baseURL := "https://api.the-odds-api.com/v4/sports/soccer_epl/odds"
        u, err := url.Parse(baseURL)
        if err != nil {
                return nil, nil, err
        }

        q := u.Query()
        q.Set("apiKey", apiKey)
        q.Set("regions", "us")
        q.Set("markets", "h2h")
        q.Set("oddsFormat", "decimal")
        q.Set("dateFormat", "iso")
        q.Set("bookmakers", "marathonbet")
        u.RawQuery = q.Encode()

        fullURL := u.String()
        fmt.Printf("EXTERNAL API REQUEST (ODDS): %s\n", fullURL)

        resp, err := http.Get(fullURL)
        if err != nil {
                return nil, nil, fmt.Errorf("failed to fetch odds: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                body, _ := io.ReadAll(resp.Body)
                return nil, nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
        }

        var events []OddsAPIEvent
        if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
                return nil, nil, fmt.Errorf("failed to decode response: %w", err)
        }

        apiStats := &APIStats{
                RequestsRemaining: resp.Header.Get("x-requests-remaining"),
                RequestsUsed:      resp.Header.Get("x-requests-used"),
        }

        // Log API stats for debugging
        fmt.Printf("ODDS API: requests_used=%s, requests_remaining=%s\n", apiStats.RequestsUsed, apiStats.RequestsRemaining)

        return events, apiStats, nil
}

// fetchScoresFromAPI fetches scores from The Odds API
func fetchScoresFromAPI(apiKey string) ([]ScoresAPIEvent, *APIStats, error) {
        if apiKey == "" {
                return nil, nil, fmt.Errorf("ODDS_API_KEY is not configured")
        }

        baseURL := "https://api.the-odds-api.com/v4/sports/soccer_epl/scores/"
        u, err := url.Parse(baseURL)
        if err != nil {
                return nil, nil, err
        }

        q := u.Query()
        q.Set("daysFrom", "3")
        q.Set("apiKey", apiKey)
        u.RawQuery = q.Encode()

        fullURL := u.String()
        fmt.Printf("EXTERNAL API REQUEST (SCORES): %s\n", fullURL)

        resp, err := http.Get(fullURL)
        if err != nil {
                return nil, nil, fmt.Errorf("failed to fetch scores: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                body, _ := io.ReadAll(resp.Body)
                return nil, nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
        }

        var events []ScoresAPIEvent
        if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
                return nil, nil, fmt.Errorf("failed to decode response: %w", err)
        }

        apiStats := &APIStats{
                RequestsRemaining: resp.Header.Get("x-requests-remaining"),
                RequestsUsed:      resp.Header.Get("x-requests-used"),
        }

        // Log API stats for debugging
        fmt.Printf("SCORES API: requests_used=%s, requests_remaining=%s\n", apiStats.RequestsUsed, apiStats.RequestsRemaining)

        return events, apiStats, nil
}

// processOddsEvent converts OddsAPIEvent to Match
func processOddsEvent(event OddsAPIEvent) (*Match, error) {
        match := &Match{
                APIID:       event.ID,
                HomeTeam:    event.HomeTeam,
                AwayTeam:    event.AwayTeam,
                CommenceTime: event.CommenceTime,
                Completed:   false,
                Calculated:  false,
        }

        // Extract odds from bookmaker
        if len(event.Bookmakers) > 0 && len(event.Bookmakers[0].Markets) > 0 {
                outcomes := event.Bookmakers[0].Markets[0].Outcomes
                for _, outcome := range outcomes {
                        if outcome.Name == event.HomeTeam {
                                match.HomeOdds = &outcome.Price
                        } else if outcome.Name == event.AwayTeam {
                                match.AwayOdds = &outcome.Price
                        } else if outcome.Name == "Draw" {
                                match.DrawOdds = &outcome.Price
                        }
                }
        }

        return match, nil
}

// processScoreEvent converts ScoresAPIEvent to Match
func processScoreEvent(event ScoresAPIEvent) (*Match, error) {
        match := &Match{
                APIID:        event.ID,
                HomeTeam:     event.HomeTeam,
                AwayTeam:     event.AwayTeam,
                CommenceTime: event.CommenceTime,
                Completed:    event.Completed,
                Calculated:   false,
        }

        // Extract scores
        homeScore := -1
        awayScore := -1
        for _, score := range event.Scores {
                if score.Name == event.HomeTeam {
                        if score.Score != "" {
                                if s, err := strconv.Atoi(score.Score); err == nil {
                                        homeScore = s
                                }
                        }
                } else if score.Name == event.AwayTeam {
                        if score.Score != "" {
                                if s, err := strconv.Atoi(score.Score); err == nil {
                                        awayScore = s
                                }
                        }
                }
        }

        if homeScore != -1 {
                match.HomeScore = &homeScore
        }
        if awayScore != -1 {
                match.AwayScore = &awayScore
        }

        return match, nil
}

// sendTelegramNotification sends a notification to Telegram
func sendTelegramNotification(botToken, channelID string, matches []map[string]interface{}) error {
        if botToken == "" || channelID == "" {
                return fmt.Errorf("Telegram credentials not configured")
        }

        // Log attempt to send notification
        fmt.Printf("TELEGRAM: Attempting to send notification to channel %s with %d matches\n", channelID, len(matches))

        now := time.Now()
        dateTime := now.Format("02/01/2006 15:04:05")

        message := fmt.Sprintf("🎯 <b>Matches Calculated!</b>\n\n📅 %s\n\n⚽ <b>Match Results:</b>\n", dateTime)

        for i, match := range matches {
                message += fmt.Sprintf("%d. %s %s %s\n", i+1, match["home_team"], match["score"], match["away_team"])
        }

        message += "\n💰 <i>Dear clients, bets have been calculated automatically!</i>"

        apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)
        fmt.Printf("EXTERNAL API REQUEST (TELEGRAM): %s\n", apiURL)

        payload := map[string]interface{}{
                "chat_id":    channelID,
                "text":       message,
                "parse_mode": "HTML",
        }

        jsonData, err := json.Marshal(payload)
        if err != nil {
                return fmt.Errorf("failed to marshal payload: %w", err)
        }

        resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
        if err != nil {
                return fmt.Errorf("failed to send request: %w", err)
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                body, _ := io.ReadAll(resp.Body)
                return fmt.Errorf("Telegram API returned status %d: %s", resp.StatusCode, string(body))
        }

        // Log successful send
        fmt.Printf("TELEGRAM: Notification sent successfully to channel %s\n", channelID)
        return nil
}

