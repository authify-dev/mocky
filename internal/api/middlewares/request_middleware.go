package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// CustomResponseWriter intercepta el output
type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b) // captura la respuesta
	return w.ResponseWriter.Write(b)
}

func RequestLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 1. Lee y clona el body del request
		var requestBody []byte
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = bodyBytes
				// Reemplaza el body para que Gin pueda leerlo después
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// 2. Captura la respuesta
		bodyWriter := &CustomResponseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bodyWriter

		// 3. Continúa con el handler
		c.Next()

		// 4. Logging
		duration := time.Since(start)

		requestBodyJSON := jsonToPrettyString(requestBodyToJSON(requestBody))
		responseBodyJSON := jsonToPrettyString(requestBodyToJSON(bodyWriter.body.Bytes()))

		log.Printf(`[GIN] %s %s | %d | %s | IP: %s
Request: %s
Response: %s`,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
			c.ClientIP(),
			requestBodyJSON,
			responseBodyJSON,
		)
	}
}

func requestBodyToJSON(body []byte) map[string]interface{} {

	var jsonBody map[string]interface{}
	err := json.Unmarshal(body, &jsonBody)
	if err != nil {
		return nil
	}
	return jsonBody
}

func jsonToPrettyString(jsonBodyMap map[string]interface{}) string {
	jsonBody, err := json.MarshalIndent(jsonBodyMap, "", "  ")
	if err != nil {
		return ""
	}
	return string(jsonBody)
}
