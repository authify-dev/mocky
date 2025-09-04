package middleware

import (
	"common/domain/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestLoggerMiddleware crea un logrus.Entry con los campos de la petición y lo guarda en el contexto.
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		requestID := uuid.New().String()

		method := c.Request.Method
		clientIP := c.ClientIP()

		userID := c.GetString("user_id")

		// Creamos una instancia del struct con los valores correspondientes.
		logFields := logger.LogFields{
			TraceID:  requestID,
			Method:   method,
			ClientIP: clientIP,
			UserID:   userID,
		}

		entry := logger.WithFields(logFields)

		// Inyectamos el logger en el contexto de la petición.
		ctx := logger.WithLogger(c.Request.Context(), entry)
		c.Request = c.Request.WithContext(ctx)

		// Logueamos el inicio de la petición.
		entry.Info("Inicio de la petición " + c.Request.RequestURI)

		c.Next()

		elapsedTime := time.Since(startTime)

		entry.Infof("Fin de la petición %s Tiempo total: %s", c.Request.RequestURI, elapsedTime)
	}
}
