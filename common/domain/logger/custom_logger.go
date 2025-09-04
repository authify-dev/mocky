package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

type LoggerSettings struct {
	ENVIRONMENT string
	APP_NAME    string
	LOKI_URL    string
}

var LoggerConfig LoggerSettings

func InitLogger(env string, app_name string, loki_url string) {
	LoggerConfig = LoggerSettings{
		ENVIRONMENT: env,
		APP_NAME:    app_name,
		LOKI_URL:    loki_url,
	}
}

// contextKey es el tipo para la clave del logger en el contexto.
type contextKey string

const loggerKey contextKey = "logger"

// CustomFormatter es un formateador personalizado para Logrus.
type CustomFormatter struct{}

// Format implementa la interfaz Formatter de Logrus.
// Genera un formato: timestamp | level | function:line | fields | message
func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format(time.RFC3339)
	fields, ok := entry.Data["fields"].(LogFields)
	if !ok {
		fields = LogFields{}
	}
	caller := ""
	lineNum := 0
	if entry.Caller != nil {
		caller = entry.Caller.Function
		lineNum = entry.Caller.Line
	}
	line := fmt.Sprintf("%s | %s | %s:%d | %s | %s\n",
		timestamp,
		entry.Level.String(),
		caller,
		lineNum,
		fields.ToString(),
		entry.Message,
	)
	return []byte(line), nil
}

var globalLogger *logrus.Logger

func init() {
	// Configuración del logger global.
	globalLogger = logrus.New()
	globalLogger.SetFormatter(&CustomFormatter{})
	globalLogger.SetLevel(logrus.InfoLevel)
	globalLogger.SetReportCaller(true)

	// Se agrega el hook para que cada log se envíe a Loki.
	// globalLogger.AddHook(&LokiHook{})
	// globalLogger.AddHook(&SentryHook{})
}

// WithFields crea un entry de logger con campos adicionales.
func WithFields(fields LogFields) *logrus.Entry {
	mapFields := map[string]interface{}{
		"fields": fields,
	}
	return globalLogger.WithFields(mapFields)
}

// WithLogger inyecta un entry de logger en el contexto.
func WithLogger(ctx context.Context, entry *logrus.Entry) context.Context {
	return context.WithValue(ctx, loggerKey, entry)
}

// FromContext obtiene el logger desde el contexto; si no hay, devuelve el logger global.
func FromContext(ctx context.Context) *logrus.Entry {
	if entry, ok := ctx.Value(loggerKey).(*logrus.Entry); ok {
		return entry
	}
	return globalLogger.WithFields(logrus.Fields{})
}

func FromContextWithExit(ctx context.Context) (*logrus.Entry, func()) {
	// Identificar el lugar que llamó a esta función
	pc, _, line, ok := runtime.Caller(1) // 1 => calling frame
	caller := "unknown"
	if ok {
		fn := runtime.FuncForPC(pc)
		caller = fmt.Sprintf("%s:%d", fn.Name(), line)
	}

	entry, ok := ctx.Value(loggerKey).(*logrus.Entry)

	if !ok {
		entry = globalLogger.WithFields(logrus.Fields{})
	}
	entry.Infof("===========Iniciando función %s===========", caller)

	// Closure para marcar la salida
	done := func() {
		entry.Infof("===========Terminando función %s===========", caller)
	}

	return entry, done
}

// publishToLoki envía un log a Loki usando el endpoint push.
func publishToLoki(message, level string, fields LogFields) {
	type Stream struct {
		Stream map[string]string `json:"stream"`
		Values [][2]string       `json:"values"`
	}
	type Payload struct {
		Streams []Stream `json:"streams"`
	}

	timestamp := time.Now().UnixNano()
	payload := Payload{
		Streams: []Stream{
			{
				Stream: map[string]string{
					"app":       LoggerConfig.APP_NAME,
					"env":       LoggerConfig.ENVIRONMENT,
					"level":     level,
					"caller_id": fields.CallerID,
					"trace_id":  fields.TraceID,
					"method":    fields.Method,
					"client_ip": fields.ClientIP,
					"user_id":   fields.UserID,
					"path":      fields.Path,
					"call_in":   fields.CallIn,
				},
				Values: [][2]string{
					{fmt.Sprintf("%d", timestamp), message},
				},
			},
		},
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		fmt.Println("Error encoding JSON for Loki payload:", err)
		return
	}

	loki_url := fmt.Sprintf("%s/loki/api/v1/push", LoggerConfig.LOKI_URL)

	resp, err := http.Post(loki_url, "application/json", buf)
	if err != nil {
		fmt.Println("Error sending log to Loki:", err)
		return
	}
	defer resp.Body.Close()
}

// Hook para enviar logs a Loki.
type LokiHook struct{}

func (hook *LokiHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *LokiHook) Fire(entry *logrus.Entry) error {

	fields, ok := entry.Data["fields"].(LogFields)
	if !ok {
		fields = LogFields{}
	}
	fields.CallIn = fmt.Sprintf("%s:%d", entry.Caller.Function, entry.Caller.Line)

	publishToLoki(entry.Message, entry.Level.String(), fields)
	return nil
}

func InfoD(entry *logrus.Entry, args ...interface{}) {
	if LoggerConfig.ENVIRONMENT != "production" {
		entry.Info(args...)
	}
}

// Hook para enviar logs a Sentry.
type SentryHook struct{}

func (hook *SentryHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.ErrorLevel}
}

func (hook *SentryHook) Fire(entry *logrus.Entry) error {
	caller := ""
	if entry.HasCaller() {
		caller = fmt.Sprintf("%v:%d", entry.Caller.Function, entry.Caller.Line)
	}
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	fullMsg := fmt.Sprintf("%s | %s | %s", timestamp, caller, entry.Message)

	sentry.CaptureException(errors.New(fullMsg))
	return nil
}
