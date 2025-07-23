package middlewares

import (
	"bytes"
	"encoding/json"
	"friendship/utils"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestResponseLogger - расширенный middleware для логирования
func RequestResponseLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Базовая информация о запросе
		logEntry := map[string]interface{}{
			"timestamp":   param.TimeStamp.Format("2006-01-02 15:04:05"),
			"method":      param.Method,
			"path":        param.Path,
			"status_code": param.StatusCode,
			"latency":     param.Latency.String(),
			"client_ip":   param.ClientIP,
			"user_agent":  param.Request.UserAgent(),
			"request_id":  param.Request.Header.Get("X-Request-ID"),
		}

		// Добавляем query параметры если есть
		if param.Request.URL.RawQuery != "" {
			logEntry["query_params"] = param.Request.URL.RawQuery
		}

		// Добавляем размер ответа
		logEntry["response_size"] = param.BodySize

		// Определяем уровень лога по статус коду
		var level string
		switch {
		case param.StatusCode >= 500:
			level = "ERROR"
		case param.StatusCode >= 400:
			level = "WARN"
		case param.StatusCode >= 300:
			level = "INFO"
		default:
			level = "INFO"
		}
		logEntry["level"] = level

		// Добавляем ошибку если есть
		if param.ErrorMessage != "" {
			logEntry["error"] = param.ErrorMessage
		}

		// Форматируем в JSON для структурированного логирования
		jsonBytes, _ := json.Marshal(logEntry)

		return string(jsonBytes) + "\n"
	})
}

// DetailedRequestLogger - более детальный middleware с телом запроса/ответа
func DetailedRequestLogger() gin.HandlerFunc {
	// Определяем уровень детализации из переменной окружения
	logBodies := os.Getenv("LOG_BODIES") == "true"
	logHeaders := os.Getenv("LOG_HEADERS") == "true"

	return func(c *gin.Context) {
		start := time.Now()

		// Генерируем ID запроса для трассировки
		requestID := generateRequestID()
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)

		// Читаем тело запроса
		var requestBody []byte
		if c.Request.Body != nil && logBodies && shouldLogBody(c.Request.Header.Get("Content-Type"), c.Request.Method) {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Создаем writer для перехвата ответа
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// Логируем начало запроса
		utils.Log.Printf("REQUEST_START [%s]: %s %s from %s",
			requestID, c.Request.Method, c.Request.URL.Path, c.ClientIP())

		// Обрабатываем запрос
		c.Next()

		// Логируем результат
		duration := time.Since(start)

		logData := map[string]interface{}{
			"request_id":    requestID,
			"method":        c.Request.Method,
			"path":          c.Request.URL.Path,
			"status_code":   c.Writer.Status(),
			"duration_ms":   duration.Milliseconds(),
			"client_ip":     c.ClientIP(),
			"user_agent":    c.Request.UserAgent(),
			"response_size": len(responseWriter.body.Bytes()),
		}

		// Добавляем query параметры
		if c.Request.URL.RawQuery != "" {
			logData["query_params"] = c.Request.URL.RawQuery
		}

		// Добавляем важные заголовки (очищенные от чувствительных данных)
		if logHeaders && len(c.Request.Header) > 0 {
			importantHeaders := make(map[string][]string)
			for key, values := range c.Request.Header {
				// Логируем только важные заголовки
				lowerKey := strings.ToLower(key)
				if lowerKey == "content-type" || lowerKey == "user-agent" ||
					lowerKey == "accept" || lowerKey == "authorization" ||
					lowerKey == "x-forwarded-for" || lowerKey == "x-real-ip" {
					importantHeaders[key] = values
				}
			}
			if len(importantHeaders) > 0 {
				logData["headers"] = sanitizeHeaders(importantHeaders)
			}
		}

		// Добавляем тело запроса (только для определенных типов)
		if logBodies && len(requestBody) > 0 && len(requestBody) < 1024 { // Ограничиваем размер
			if isJSON(requestBody) {
				var jsonBody interface{}
				if json.Unmarshal(requestBody, &jsonBody) == nil {
					// Очищаем чувствительные данные перед логированием
					sanitizedBody := sanitizeSensitiveData(jsonBody)
					logData["request_body"] = sanitizedBody
				}
			} else {
				logData["request_body"] = "[non-json content]"
			}
		}

		// Добавляем ответ для ошибок
		if logBodies && c.Writer.Status() >= 400 {
			responseBody := responseWriter.body.String()
			if len(responseBody) > 0 && len(responseBody) < 1024 {
				if isJSON([]byte(responseBody)) {
					var jsonBody interface{}
					if json.Unmarshal([]byte(responseBody), &jsonBody) == nil {
						// Очищаем чувствительные данные из ответа тоже
						sanitizedResponse := sanitizeSensitiveData(jsonBody)
						logData["response_body"] = sanitizedResponse
					}
				} else {
					logData["response_body"] = "[non-json response]"
				}
			}
		}

		// Добавляем ошибки из контекста
		if len(c.Errors) > 0 {
			logData["errors"] = c.Errors.Errors()
		}

		// Определяем уровень лога
		var level string
		switch {
		case c.Writer.Status() >= 500:
			level = "ERROR"
		case c.Writer.Status() >= 400:
			level = "WARN"
		default:
			level = "INFO"
		}

		jsonBytes, _ := json.Marshal(logData)
		utils.Log.Printf("%s: %s", level, string(jsonBytes))
	}
}

// ErrorLogger - middleware для логирования панике и ошибок
func ErrorLogger() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = generateRequestID()
		}

		errorData := map[string]interface{}{
			"level":      "PANIC",
			"request_id": requestID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"client_ip":  c.ClientIP(),
			"panic":      recovered,
			"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
		}

		jsonBytes, _ := json.Marshal(errorData)
		utils.Log.Printf("PANIC: %s", string(jsonBytes))

		c.AbortWithStatus(500)
	})
}

// SlowRequestLogger - логирует медленные запросы
func SlowRequestLogger(threshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		if duration > threshold {
			slowLogData := map[string]interface{}{
				"level":     "WARN",
				"type":      "SLOW_REQUEST",
				"method":    c.Request.Method,
				"path":      c.Request.URL.Path,
				"duration":  duration.String(),
				"threshold": threshold.String(),
				"client_ip": c.ClientIP(),
				"timestamp": time.Now().Format("2006-01-02 15:04:05"),
			}

			jsonBytes, _ := json.Marshal(slowLogData)
			utils.Log.Printf("SLOW_REQUEST: %s", string(jsonBytes))
		}
	}
}

// Вспомогательные структуры и функции

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" +
		string(rune(time.Now().UnixNano()%1000))
}

func shouldLogBody(contentType, method string) bool {
	if method == "GET" || method == "DELETE" {
		return false
	}
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "application/x-www-form-urlencoded")
}

func isJSON(data []byte) bool {
	var js interface{}
	return json.Unmarshal(data, &js) == nil
}

// sanitizeSensitiveData - удаляет или маскирует чувствительные данные
func sanitizeSensitiveData(data interface{}) interface{} {
	// Список полей, которые нужно очистить (можно расширить)
	sensitiveFields := map[string]bool{
		"password":         true,
		"Password":         true,
		"PASSWORD":         true,
		"passwd":           true,
		"pwd":              true,
		"secret":           true,
		"Secret":           true,
		"SECRET":           true,
		"token":            true,
		"Token":            true,
		"TOKEN":            true,
		"access_token":     true,
		"accessToken":      true,
		"AccessToken":      true,
		"refresh_token":    true,
		"refreshToken":     true,
		"RefreshToken":     true,
		"authorization":    true,
		"Authorization":    true,
		"AUTHORIZATION":    true,
		"api_key":          true,
		"apiKey":           true,
		"ApiKey":           true,
		"API_KEY":          true,
		"private_key":      true,
		"privateKey":       true,
		"PrivateKey":       true,
		"PRIVATE_KEY":      true,
		"credit_card":      true,
		"creditCard":       true,
		"card_number":      true,
		"cardNumber":       true,
		"ssn":              true,
		"social_security":  true,
		"socialSecurity":   true,
		"phone":            true,  // Если считаете телефон чувствительным
		"email":            false, // email обычно можно логировать, но можете изменить
		"current_password": true,
		"currentPassword":  true,
		"new_password":     true,
		"newPassword":      true,
		"confirm_password": true,
		"confirmPassword":  true,
		"old_password":     true,
		"oldPassword":      true,
	}

	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			if sensitiveFields[key] {
				// Разные стратегии маскирования
				if str, ok := value.(string); ok && len(str) > 0 {
					if len(str) <= 4 {
						result[key] = "***"
					} else {
						// Показываем только первые 2 символа
						result[key] = str[:2] + strings.Repeat("*", len(str)-2)
					}
				} else {
					result[key] = "[REDACTED]"
				}
			} else {
				// Рекурсивно обрабатываем вложенные объекты
				result[key] = sanitizeSensitiveData(value)
			}
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = sanitizeSensitiveData(item)
		}
		return result

	default:
		return v
	}
}

// sanitizeHeaders - очищает чувствительные заголовки
func sanitizeHeaders(headers map[string][]string) map[string][]string {
	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"Authorization": true,
		"AUTHORIZATION": true,
		"x-api-key":     true,
		"X-Api-Key":     true,
		"X-API-KEY":     true,
		"cookie":        true,
		"Cookie":        true,
		"COOKIE":        true,
		"set-cookie":    true,
		"Set-Cookie":    true,
		"SET-COOKIE":    true,
	}

	result := make(map[string][]string)
	for key, values := range headers {
		if sensitiveHeaders[key] {
			maskedValues := make([]string, len(values))
			for i, value := range values {
				if len(value) > 10 {
					maskedValues[i] = value[:6] + "...***"
				} else {
					maskedValues[i] = "[REDACTED]"
				}
			}
			result[key] = maskedValues
		} else {
			result[key] = values
		}
	}
	return result
}
