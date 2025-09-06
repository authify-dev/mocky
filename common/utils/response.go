package utils

import (
	"bytes"
	"common/domain/customctx"
	"common/domain/logger"
	"common/utils/cerrs"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
)

type Alert struct {
	Message string `json:"message,omitempty"`
	Title   string `json:"title,omitempty"`
	Icon    string `json:"icon,omitempty"`
	Code    uint   `json:"code,omitempty"`
	Scope   string `json:"scope,omitempty"`
}

type Response[R any] struct {
	Error      cerrs.CustomErrorInterface `json:"error,omitempty"`
	StatusCode int                        `json:"status_code" default:"200"`

	Data    R                            `json:"data,omitempty"`
	Results []R                          `json:"results,omitempty"`
	Alert   *Alert                       `json:"alert,omitempty"`
	TraceID string                       `json:"trace_id,omitempty"`
	Success bool                         `json:"success" default:"true"`
	Errors  []cerrs.CustomErrorInterface `json:"errors,omitempty"`
}

func (r Response[R]) ToMapWithCustomContext(ctx *customctx.CustomContext) map[string]interface{} {

	if ctx == nil {
		return r.ToMap()
	}

	fields := GetFieldsOfLogger(ctx.Context())

	r.TraceID = fields.TraceID

	res := r.ToMap()

	if len(ctx.Errors()) > 0 {

		if logger.LoggerConfig.ENVIRONMENT != "production" {

			res["errors"] = ctx.Errors()
			delete(res, "data")
		}

		//r.ReportErrorToSentry(fields.Path + " - " + fields.Method)
		r.ReportToLoki(ctx, res)

	}
	if r.Results != nil {
		delete(res, "data")
	}

	return res
}

func (r Response[R]) ToMap() map[string]interface{} {

	if r.StatusCode == 0 {
		r.StatusCode = 200
	}

	data, err := json.Marshal(r)
	if err != nil {
		log.Printf("error marshaling OfferEntity: %v", err)
		return nil
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		log.Printf("error unmarshaling to map: %v", err)
		return nil
	}

	return result
}

func (r Response[R]) ReportToLoki(ctx *customctx.CustomContext, payloadErr map[string]interface{}) {

	payloadRes, err := json.Marshal(payloadErr)
	if err != nil {
		// Si falla el marshal, reporta ese error mínimo
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("trace_id", r.TraceID)
			sentry.CaptureException(err)
		})
		sentry.Flush(2 * time.Second)
		return
	}

	fields := GetFieldsOfLogger(ctx)

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
					"app":       logger.LoggerConfig.APP_NAME,
					"env":       logger.LoggerConfig.ENVIRONMENT,
					"level":     "error",
					"caller_id": fields.CallerID,
					"trace_id":  fields.TraceID,
					"method":    fields.Method,
					"client_ip": fields.ClientIP,
					"user_id":   fields.UserID,
					"path":      fields.Path,
					"call_in":   fields.CallIn,
					"api_error": "true",
				},
				Values: [][2]string{
					{fmt.Sprintf("%d", timestamp), string(payloadRes)},
				},
			},
		},
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(payload); err != nil {
		fmt.Println("Error encoding JSON for Loki payload:", err)
		return
	}

	loki_url := fmt.Sprintf("%s/loki/api/v1/push", logger.LoggerConfig.LOKI_URL)

	resp, err := http.Post(loki_url, "application/json", buf)
	if err != nil {

		fmt.Println("Error sending log to Loki:", err)
		return
	}
	defer resp.Body.Close()

}

func (r Response[R]) ReportErrorToSentry(resource string) {
	// Generar un ID de correlación
	errorID := r.TraceID

	// Serializar la respuesta completa
	payload, err := json.Marshal(r.Error)
	if err != nil {
		// Si falla el marshal, reporta ese error mínimo
		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetTag("trace_id", errorID)
			sentry.CaptureException(err)
		})
		sentry.Flush(2 * time.Second)
		return
	}

	// Enviar a Sentry con scope: tag y extras
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetTag("trace_id", errorID)
		scope.SetExtra("api_response", string(payload))
		scope.SetExtra("environment", logger.LoggerConfig.ENVIRONMENT)
		scope.SetExtra("app_name", logger.LoggerConfig.APP_NAME)
		scope.SetExtra("resource", resource)
		scope.SetExtra("trace_id", errorID)
		sentry.CaptureMessage("API error: " + errorID)
	})
	sentry.Flush(2 * time.Second)
}
