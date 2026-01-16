package main

import (
	"net"
	"net/http"
	"regexp"
	"strings"
)

// WAFMiddleware - веб-брандмауэр на уровне приложения
func WAFMiddleware(logger *Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем заголовки на подозрительные паттерны
			if isThreatInHeaders(r.Header) {
				logger.LogWarning("[WAF] Suspicious headers detected from IP: %s", getClientIP(r))
				http.Error(w, `{"success": false, "error": "Request blocked by WAF"}`, http.StatusForbidden)
				return
			}

			// Проверяем URL-параметры
			if isThreatInURL(r.URL.RawQuery) {
				logger.LogWarning("[WAF] Suspicious URL parameters detected from IP: %s", getClientIP(r))
				http.Error(w, `{"success": false, "error": "Request blocked by WAF"}`, http.StatusForbidden)
				return
			}

			// Проверяем тело запроса (если есть)
			if r.ContentLength > 0 {
				bodyThreat := isThreatInBody(r)
				if bodyThreat {
					logger.LogWarning("[WAF] Suspicious content in request body detected from IP: %s", getClientIP(r))
					http.Error(w, `{"success": false, "error": "Request blocked by WAF"}`, http.StatusForbidden)
					return
				}
			}

			// Проверяем User-Agent
			userAgent := r.Header.Get("User-Agent")
			if isThreatInUserAgent(userAgent) {
				logger.LogWarning("[WAF] Suspicious User-Agent detected from IP: %s", getClientIP(r))
				http.Error(w, `{"success": false, "error": "Request blocked by WAF"}`, http.StatusForbidden)
				return
			}

			// Все проверки пройдены, передаем запрос дальше
			next.ServeHTTP(w, r)
		})
	}
}

// Проверяет заголовки на наличие подозрительных паттернов
func isThreatInHeaders(headers http.Header) bool {
	suspiciousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union\s+select|insert\s+into|drop\s+table|exec\s*\(|script|<script|onerror|onload)`),
		regexp.MustCompile(`(?i)(\.\./|\.\.\\|%2e%2e%2f|\.\.\/)`), // Path traversal
		regexp.MustCompile(`(?i)(eval\(|expression\(|javascript:|vbscript:)`),
	}

	for name, values := range headers {
		if strings.ToLower(name) == "authorization" || strings.ToLower(name) == "cookie" {
			continue // Эти заголовки проверяются отдельно
		}
		
		for _, value := range values {
			for _, pattern := range suspiciousPatterns {
				if pattern.MatchString(value) {
					return true
				}
			}
		}
	}
	return false
}

// Проверяет URL-параметры на наличие подозрительных паттернов
func isThreatInURL(rawQuery string) bool {
	if rawQuery == "" {
		return false
	}

	suspiciousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union\s+select|insert\s+into|drop\s+table|exec\s*\(|script|<script|onerror|onload)`),
		regexp.MustCompile(`(?i)(\.\./|\.\.\\|%2e%2e%2f|\.\.\/)`), // Path traversal
		regexp.MustCompile(`(?i)(eval\(|expression\(|javascript:|vbscript:)`),
		regexp.MustCompile(`(?i)(\b(select|update|delete|insert|drop|create|alter|exec|execute)\b)`),
		// Добавляем проверки на SQL-инъекции
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+[\d=']+\s*(--|#|\/\*|{))`), // OR/AND 1=1
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+['"][\d\s]=[\d\s]['"]\s*(--|#|\/\*))`), // OR/AND '1'='1'
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+\d+\s*[=<>]\s*\d+)`), // OR/AND 1=1
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+['"` + "`" + `][^'"` + "`" + `]*['"` + "`" + `]\s*[=<>]\s*['"` + "`" + `][^'"` + "`" + `]*['"` + "`" + `])`), // OR/AND 'a'='a'
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+0x)`), // OR/AND 0xHEX
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+true\b)`), // OR/AND true
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+false\b)`), // OR/AND false
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+NULL\b)`), // OR/AND NULL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+IS\s+NULL\b)`), // OR/AND IS NULL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+IS\s+NOT\s+NULL\b)`), // OR/AND IS NOT NULL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+EXISTS\s*\()`), // OR/AND EXISTS()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+IN\s*\()`), // OR/AND IN()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+BETWEEN\s+)`), // OR/AND BETWEEN
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+LIKE\s+)`), // OR/AND LIKE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+RLIKE\s+)`), // OR/AND RLIKE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+SOUNDS\s+LIKE\b)`), // OR/AND SOUNDS LIKE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+REGEXP\b)`), // OR/AND REGEXP
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+MATCH\s+\()`), // OR/AND MATCH()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+AGAINST\s+\()`), // OR/AND AGAINST()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+BINARY\b)`), // OR/AND BINARY
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+INTERVAL\b)`), // OR/AND INTERVAL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+CAST\s*\()`), // OR/AND CAST()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+CONVERT\s*\()`), // OR/AND CONVERT()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+CASE\s+)`), // OR/AND CASE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WHEN\s+)`), // OR/AND WHEN
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+THEN\s+)`), // OR/AND THEN
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+ELSE\s+)`), // OR/AND ELSE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+END\b)`), // OR/AND END
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+IF\s*\()`), // OR/AND IF()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+IFNULL\s*\()`), // OR/AND IFNULL()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+COALESCE\s*\()`), // OR/AND COALESCE()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+ISNULL\s*\()`), // OR/AND ISNULL()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+NULLIF\s*\()`), // OR/AND NULLIF()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+LEAST\s*\()`), // OR/AND LEAST()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+GREATEST\s*\()`), // OR/AND GREATEST()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+VALUES\s*\()`), // OR/AND VALUES()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+ROW\s*\()`), // OR/AND ROW()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+ROW_NUMBER\s*\()`), // OR/AND ROW_NUMBER()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+RANK\s*\()`), // OR/AND RANK()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+DENSE_RANK\s*\()`), // OR/AND DENSE_RANK()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+NTILE\s*\()`), // OR/AND NTILE()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+PERCENT_RANK\s*\()`), // OR/AND PERCENT_RANK()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+CUME_DIST\s*\()`), // OR/AND CUME_DIST()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+FIRST_VALUE\s*\()`), // OR/AND FIRST_VALUE()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+LAST_VALUE\s*\()`), // OR/AND LAST_VALUE()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+LAG\s*\()`), // OR/AND LAG()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+LEAD\s*\()`), // OR/AND LEAD()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+NTH_VALUE\s*\()`), // OR/AND NTH_VALUE()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+OVER\s*\()`), // OR/AND OVER()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+PARTITION\s+BY\b)`), // OR/AND PARTITION BY
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+ORDER\s+BY\b)`), // OR/AND ORDER BY
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+GROUP\s+BY\b)`), // OR/AND GROUP BY
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+HAVING\b)`), // OR/AND HAVING
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+LIMIT\b)`), // OR/AND LIMIT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+OFFSET\b)`), // OR/AND OFFSET
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+UNION\s+ALL\b)`), // OR/AND UNION ALL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+INTERSECT\b)`), // OR/AND INTERSECT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+EXCEPT\b)`), // OR/AND EXCEPT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+MINUS\b)`), // OR/AND MINUS
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+INTERSECT\s+ALL\b)`), // OR/AND INTERSECT ALL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+EXCEPT\s+ALL\b)`), // OR/AND EXCEPT ALL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+MINUS\s+ALL\b)`), // OR/AND MINUS ALL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+ROLLUP\b)`), // OR/AND WITH ROLLUP
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+CUBE\b)`), // OR/AND WITH CUBE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+MAX)\b`), // OR/AND WITH MAX
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+MIN)\b`), // OR/AND WITH MIN
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+AVG)\b`), // OR/AND WITH AVG
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+SUM)\b`), // OR/AND WITH SUM
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+COUNT)\b`), // OR/AND WITH COUNT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+STDDEV)\b`), // OR/AND WITH STDDEV
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+VARIANCE)\b`), // OR/AND WITH VARIANCE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+GROUPING)\b`), // OR/AND WITH GROUPING
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+GROUPING_ID)\b`), // OR/AND WITH GROUPING_ID
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+GROUPING_SETS)\b`), // OR/AND WITH GROUPING_SETS
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+SET)\b`), // OR/AND WITH SET
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+MEMBER)\b`), // OR/AND WITH MEMBER
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+HIERARCHY)\b`), // OR/AND WITH HIERARCHY
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+LEVEL)\b`), // OR/AND WITH LEVEL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+CONNECT)\b`), // OR/AND WITH CONNECT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+START)\b`), // OR/AND WITH START
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+PRIOR)\b`), // OR/AND WITH PRIOR
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+PARENT)\b`), // OR/AND WITH PARENT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+CHILD)\b`), // OR/AND WITH CHILD
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+ANCESTOR)\b`), // OR/AND WITH ANCESTOR
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+DESCENDANT)\b`), // OR/AND WITH DESCENDANT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+RELATION)\b`), // OR/AND WITH RELATION
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+REFERENCE)\b`), // OR/AND WITH REFERENCE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+FOREIGN)\b`), // OR/AND WITH FOREIGN
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+PRIMARY)\b`), // OR/AND WITH PRIMARY
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+UNIQUE)\b`), // OR/AND WITH UNIQUE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+CHECK)\b`), // OR/AND WITH CHECK
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+TRIGGER)\b`), // OR/AND WITH TRIGGER
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+PROCEDURE)\b`), // OR/AND WITH PROCEDURE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+FUNCTION)\b`), // OR/AND WITH FUNCTION
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+PACKAGE)\b`), // OR/AND WITH PACKAGE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+TYPE)\b`), // OR/AND WITH TYPE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+OBJECT)\b`), // OR/AND WITH OBJECT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+CLASS)\b`), // OR/AND WITH CLASS
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+INSTANCE)\b`), // OR/AND WITH INSTANCE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+CONSTRUCTOR)\b`), // OR/AND WITH CONSTRUCTOR
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+DESTRUCTOR)\b`), // OR/AND WITH DESTRUCTOR
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+METHOD)\b`), // OR/AND WITH METHOD
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+ATTRIBUTE)\b`), // OR/AND WITH ATTRIBUTE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+PARAMETER)\b`), // OR/AND WITH PARAMETER
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+VARIABLE)\b`), // OR/AND WITH VARIABLE
	}

	// Декодируем URL и проверяем
	decodedQuery := rawQuery
	// Простая замена URL-кодированных символов для проверки
	decodedQuery = strings.ReplaceAll(decodedQuery, "%3C", "<")
	decodedQuery = strings.ReplaceAll(decodedQuery, "%3E", ">")
	decodedQuery = strings.ReplaceAll(decodedQuery, "%27", "'")
	decodedQuery = strings.ReplaceAll(decodedQuery, "%22", "\"")
	decodedQuery = strings.ReplaceAll(decodedQuery, "%3B", ";")
	decodedQuery = strings.ReplaceAll(decodedQuery, "%2D", "-")
	decodedQuery = strings.ReplaceAll(decodedQuery, "%28", "(")
	decodedQuery = strings.ReplaceAll(decodedQuery, "%29", ")")
	decodedQuery = strings.ReplaceAll(decodedQuery, "%20", " ")

	for _, pattern := range suspiciousPatterns {
		if pattern.MatchString(decodedQuery) {
			return true
		}
	}
	return false
}

// Проверяет тело запроса на наличие подозрительных паттернов
func isThreatInBody(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "application/json") && 
	   !strings.Contains(strings.ToLower(contentType), "application/x-www-form-urlencoded") &&
	   !strings.Contains(strings.ToLower(contentType), "multipart/form-data") {
		return false
	}

	// Ограничиваем размер тела запроса для анализа (1MB)
	const maxSize = 1024 * 1024
	r.Body = http.MaxBytesReader(nil, r.Body, maxSize)

	// Для безопасности, читаем только часть тела для проверки
	buf := make([]byte, 1024)
	n, err := r.Body.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return false
	}
	
	// Возвращаем тело обратно в request (для дальнейшей обработки)
	bodyStr := string(buf[:n])
	r.Body = http.MaxBytesReader(nil, r.Body, maxSize)

	suspiciousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union\s+select|insert\s+into|drop\s+table|exec\s*\(|script|<script|onerror|onload)`),
		regexp.MustCompile(`(?i)(\.\./|\.\.\\|%2e%2e%2f|\.\.\/)`), // Path traversal
		regexp.MustCompile(`(?i)(eval\(|expression\(|javascript:|vbscript:)`),
		regexp.MustCompile(`(?i)(\b(select|update|delete|insert|drop|create|alter|exec|execute)\b)`),
		// Добавляем проверки на SQL-инъекции
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+[\d=']+\s*(--|#|\/\*|{))`), // OR/AND 1=1
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+['"][\d\s]=[\d\s]['"]\s*(--|#|\/\*))`), // OR/AND '1'='1'
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+\d+\s*[=<>]\s*\d+)`), // OR/AND 1=1
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+['"` + "`" + `][^'"` + "`" + `]*['"` + "`" + `]\s*[=<>]\s*['"` + "`" + `][^'"` + "`" + `]*['"` + "`" + `])`), // OR/AND 'a'='a'
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+0x)`), // OR/AND 0xHEX
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+true\b)`), // OR/AND true
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+false\b)`), // OR/AND false
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+NULL\b)`), // OR/AND NULL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+IS\s+NULL\b)`), // OR/AND IS NULL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+IS\s+NOT\s+NULL\b)`), // OR/AND IS NOT NULL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+EXISTS\s*\()`), // OR/AND EXISTS()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+IN\s*\()`), // OR/AND IN()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+BETWEEN\s+)`), // OR/AND BETWEEN
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+LIKE\s+)`), // OR/AND LIKE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+RLIKE\s+)`), // OR/AND RLIKE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+SOUNDS\s+LIKE\b)`), // OR/AND SOUNDS LIKE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+REGEXP\b)`), // OR/AND REGEXP
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+MATCH\s+\()`), // OR/AND MATCH()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+AGAINST\s+\()`), // OR/AND AGAINST()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+BINARY\b)`), // OR/AND BINARY
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+INTERVAL\b)`), // OR/AND INTERVAL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+CAST\s*\()`), // OR/AND CAST()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+CONVERT\s*\()`), // OR/AND CONVERT()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+CASE\s+)`), // OR/AND CASE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WHEN\s+)`), // OR/AND WHEN
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+THEN\s+)`), // OR/AND THEN
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+ELSE\s+)`), // OR/AND ELSE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+END\b)`), // OR/AND END
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+IF\s*\()`), // OR/AND IF()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+IFNULL\s*\()`), // OR/AND IFNULL()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+COALESCE\s*\()`), // OR/AND COALESCE()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+ISNULL\s*\()`), // OR/AND ISNULL()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+NULLIF\s*\()`), // OR/AND NULLIF()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+LEAST\s*\()`), // OR/AND LEAST()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+GREATEST\s*\()`), // OR/AND GREATEST()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+VALUES\s*\()`), // OR/AND VALUES()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+ROW\s*\()`), // OR/AND ROW()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+ROW_NUMBER\s*\()`), // OR/AND ROW_NUMBER()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+RANK\s*\()`), // OR/AND RANK()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+DENSE_RANK\s*\()`), // OR/AND DENSE_RANK()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+NTILE\s*\()`), // OR/AND NTILE()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+PERCENT_RANK\s*\()`), // OR/AND PERCENT_RANK()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+CUME_DIST\s*\()`), // OR/AND CUME_DIST()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+FIRST_VALUE\s*\()`), // OR/AND FIRST_VALUE()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+LAST_VALUE\s*\()`), // OR/AND LAST_VALUE()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+LAG\s*\()`), // OR/AND LAG()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+LEAD\s*\()`), // OR/AND LEAD()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+NTH_VALUE\s*\()`), // OR/AND NTH_VALUE()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+OVER\s*\()`), // OR/AND OVER()
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+PARTITION\s+BY\b)`), // OR/AND PARTITION BY
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+ORDER\s+BY\b)`), // OR/AND ORDER BY
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+GROUP\s+BY\b)`), // OR/AND GROUP BY
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+HAVING\b)`), // OR/AND HAVING
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+LIMIT\b)`), // OR/AND LIMIT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+OFFSET\b)`), // OR/AND OFFSET
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+UNION\s+ALL\b)`), // OR/AND UNION ALL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+INTERSECT\b)`), // OR/AND INTERSECT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+EXCEPT\b)`), // OR/AND EXCEPT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+MINUS\b)`), // OR/AND MINUS
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+INTERSECT\s+ALL\b)`), // OR/AND INTERSECT ALL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+EXCEPT\s+ALL\b)`), // OR/AND EXCEPT ALL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+MINUS\s+ALL\b)`), // OR/AND MINUS ALL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+ROLLUP\b)`), // OR/AND WITH ROLLUP
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+CUBE\b)`), // OR/AND WITH CUBE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+MAX)\b`), // OR/AND WITH MAX
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+MIN)\b`), // OR/AND WITH MIN
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+AVG)\b`), // OR/AND WITH AVG
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+SUM)\b`), // OR/AND WITH SUM
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+COUNT)\b`), // OR/AND WITH COUNT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+STDDEV)\b`), // OR/AND WITH STDDEV
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+VARIANCE)\b`), // OR/AND WITH VARIANCE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+GROUPING)\b`), // OR/AND WITH GROUPING
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+GROUPING_ID)\b`), // OR/AND WITH GROUPING_ID
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+GROUPING_SETS)\b`), // OR/AND WITH GROUPING_SETS
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+SET)\b`), // OR/AND WITH SET
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+MEMBER)\b`), // OR/AND WITH MEMBER
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+HIERARCHY)\b`), // OR/AND WITH HIERARCHY
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+LEVEL)\b`), // OR/AND WITH LEVEL
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+CONNECT)\b`), // OR/AND WITH CONNECT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+START)\b`), // OR/AND WITH START
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+PRIOR)\b`), // OR/AND WITH PRIOR
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+PARENT)\b`), // OR/AND WITH PARENT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+CHILD)\b`), // OR/AND WITH CHILD
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+ANCESTOR)\b`), // OR/AND WITH ANCESTOR
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+DESCENDANT)\b`), // OR/AND WITH DESCENDANT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+RELATION)\b`), // OR/AND WITH RELATION
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+REFERENCE)\b`), // OR/AND WITH REFERENCE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+FOREIGN)\b`), // OR/AND WITH FOREIGN
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+PRIMARY)\b`), // OR/AND WITH PRIMARY
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+UNIQUE)\b`), // OR/AND WITH UNIQUE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+CHECK)\b`), // OR/AND WITH CHECK
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+TRIGGER)\b`), // OR/AND WITH TRIGGER
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+PROCEDURE)\b`), // OR/AND WITH PROCEDURE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+FUNCTION)\b`), // OR/AND WITH FUNCTION
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+PACKAGE)\b`), // OR/AND WITH PACKAGE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+TYPE)\b`), // OR/AND WITH TYPE
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+OBJECT)\b`), // OR/AND WITH OBJECT
		regexp.MustCompile(`(?i)(\b(OR|AND)\s+WITH\s+CLASS)\b`), // OR/AND WITH CLASS
	}

	for _, pattern := range suspiciousPatterns {
		if pattern.MatchString(bodyStr) {
			return true
		}
	}
	return false
}

// Проверяет User-Agent на подозрительные паттерны
func isThreatInUserAgent(userAgent string) bool {
	if userAgent == "" {
		return false
	}

	// Проверяем на подозрительные боты и сканеры
	suspiciousAgents := []string{
		"sqlmap",
		"nikto",
		"nessus",
		"acunetix",
		"netsparker",
		"dirbuster",
		"w3af",
		"skipfish",
		"grabber",
		"zaproxy",
		"burp",
		"paros",
		"webinspect",
		"appscan",
		"fiddler",
		"charles",
		"crawler",
		"scanner",
		"bot",
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, agent := range suspiciousAgents {
		if strings.Contains(userAgentLower, agent) {
			return true
		}
	}

	// Проверяем на SQL-инъекции в User-Agent
	suspiciousPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union\s+select|insert\s+into|drop\s+table|exec\s*\(|'|\")`),
		regexp.MustCompile(`(?i)(\b(select|update|delete|insert|drop|create|alter|exec|execute)\b)`),
	}

	for _, pattern := range suspiciousPatterns {
		if pattern.MatchString(userAgent) {
			return true
		}
	}

	return false
}

// getClientIP extracts the real client IP from request headers
func getClientIP(r *http.Request) string {
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
	host, _, err := net.SplitHostPort(remoteAddr)
	if err == nil {
		return host
	}

	return remoteAddr
}