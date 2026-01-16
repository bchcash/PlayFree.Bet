package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Logger represents a structured logger
type Logger struct {
	level    string
	startTime time.Time
}

// NewLogger creates a new logger instance
func NewLogger(level string) *Logger {
	return &Logger{
		level:     strings.ToUpper(level),
		startTime: time.Now(),
	}
}

// shouldLog checks if the current log level allows logging this message
func (l *Logger) shouldLog(level string) bool {
	levels := map[string]int{
		"DEBUG": 0,
		"INFO":  1,
		"WARN":  2,
		"ERROR": 3,
	}

	currentLevel, exists := levels[l.level]
	if !exists {
		currentLevel = levels["INFO"]
	}

	msgLevel, exists := levels[strings.ToUpper(level)]
	if !exists {
		msgLevel = levels["INFO"]
	}

	return msgLevel >= currentLevel
}

// formatTimestamp returns a formatted timestamp
func (l *Logger) formatTimestamp() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// formatMessage formats a log message with proper structure
func (l *Logger) formatMessage(level, category, message string, args ...interface{}) string {
	timestamp := l.formatTimestamp()

	var categoryStr string
	if category != "" {
		categoryStr = fmt.Sprintf("[%s] ", category)
	}

	msg := message
	if len(args) > 0 {
		msg = fmt.Sprintf(message, args...)
	}

	// Break long messages into multiple lines if needed
	if len(msg) > 120 {
		words := strings.Fields(msg)
		var lines []string
		currentLine := ""

		for _, word := range words {
			if len(currentLine)+len(word)+1 > 100 {
				if currentLine != "" {
					lines = append(lines, currentLine)
				}
				currentLine = word
			} else {
				if currentLine != "" {
					currentLine += " "
				}
				currentLine += word
			}
		}
		if currentLine != "" {
			lines = append(lines, currentLine)
		}

		result := fmt.Sprintf("%s %-5s %s%s", timestamp, level, categoryStr, lines[0])
		for i := 1; i < len(lines); i++ {
			result += fmt.Sprintf("\n%s       %s%s", timestamp, categoryStr, lines[i])
		}
		return result
	}

	return fmt.Sprintf("%s %-5s %s%s", timestamp, level, categoryStr, msg)
}

// LogInfo logs an info message
func (l *Logger) LogInfo(message string, args ...interface{}) {
	if l.shouldLog("INFO") {
		fmt.Println(l.formatMessage("INFO", "", message, args...))
	}
}

// LogError logs an error message
func (l *Logger) LogError(message string, args ...interface{}) {
	if l.shouldLog("ERROR") {
		fmt.Println(l.formatMessage("ERROR", "", message, args...))
	}
}

// LogWarning logs a warning message
func (l *Logger) LogWarning(message string, args ...interface{}) {
	if l.shouldLog("WARN") {
		fmt.Println(l.formatMessage("WARN", "", message, args...))
	}
}

// LogSuccess logs a success message
func (l *Logger) LogSuccess(message string, args ...interface{}) {
	if l.shouldLog("INFO") {
		fmt.Println(l.formatMessage("INFO", "", message, args...))
	}
}

// LogSystem logs a system message with category
func (l *Logger) LogSystem(category, message string, args ...interface{}) {
	if l.shouldLog("INFO") {
		fmt.Println(l.formatMessage("INFO", category, message, args...))
	}
}

// LogDB logs a database-related message
func (l *Logger) LogDB(message string, args ...interface{}) {
	if l.shouldLog("INFO") {
		fmt.Println(l.formatMessage("INFO", "DB", message, args...))
	}
}

// LogAuth logs an authentication-related message
func (l *Logger) LogAuth(message string, args ...interface{}) {
	if l.shouldLog("INFO") {
		fmt.Println(l.formatMessage("INFO", "AUTH", message, args...))
	}
}

// LogBets logs a bets-related message
func (l *Logger) LogBets(message string, args ...interface{}) {
	if l.shouldLog("INFO") {
		fmt.Println(l.formatMessage("INFO", "BETS", message, args...))
	}
}

// LogSQL logs SQL query information
func (l *Logger) LogSQL(operation string, params []interface{}, duration time.Duration) {
	if l.shouldLog("DEBUG") {
		paramStr := "none"
		if len(params) > 0 {
			// Truncate long parameter lists
			paramStr = fmt.Sprintf("%v", params)
			if len(paramStr) > 50 {
				paramStr = paramStr[:47] + "..."
			}
		}
		fmt.Println(l.formatMessage("DEBUG", "SQL", "%s | params: %s | %v", operation, paramStr, duration.Round(time.Millisecond)))
	}
}

// LogStartup logs application startup information
func (l *Logger) LogStartup(name, port string) {
	if l.shouldLog("INFO") {
		fmt.Println(l.formatMessage("INFO", "STARTUP", "Starting %s on port %s", name, port))
	}
}

// LogShutdown logs application shutdown information
func (l *Logger) LogShutdown() {
	if l.shouldLog("INFO") {
		uptime := time.Since(l.startTime)
		fmt.Println(l.formatMessage("INFO", "SHUTDOWN", "Application uptime: %v", uptime.Round(time.Second)))
	}
}

// LogMetrics logs application metrics
func (l *Logger) LogMetrics() {
	if l.shouldLog("INFO") {
		uptime := time.Since(l.startTime)
		fmt.Println(l.formatMessage("INFO", "METRICS", "Metrics - Uptime: %v", uptime.Round(time.Second)))
	}
}

// Middleware returns HTTP middleware for request logging
func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(wrapper, r)

		// Log the request
		duration := time.Since(start)
		status := wrapper.statusCode
		method := r.Method
		path := r.URL.Path
		ip := r.RemoteAddr

		// Extract real IP if behind proxy
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			ip = forwarded
		} else if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
			ip = realIP
		}

		// Color code status (simple text indicators)
		var statusIndicator string
		if status >= 200 && status < 300 {
			statusIndicator = "OK"
		} else if status >= 300 && status < 400 {
			statusIndicator = "REDIRECT"
		} else if status == 401 {
			statusIndicator = "NotAuthorised try to access site"
		} else if status >= 400 && status < 500 {
			statusIndicator = "CLIENT_ERROR"
		} else {
			statusIndicator = "SERVER_ERROR"
		}

		if l.shouldLog("INFO") {
			fmt.Println(l.formatMessage("INFO", "HTTP",
				"%s %s | %d %s | %v | %s",
				method, path, status, statusIndicator, duration.Round(time.Millisecond), ip))
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}