package controllers

import (
	"common/domain/customctx"
	"common/domain/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

func extractPath(c *gin.Context) map[string]string {
	out := make(map[string]string)
	for _, p := range c.Params {
		out[p.Key] = p.Value
	}
	return out
}

func extractHeaders(r *http.Request) map[string]string {
	out := make(map[string]string)
	for k, vals := range r.Header {
		if len(vals) > 0 {
			out[k] = vals[0] // primer valor
		}
	}
	return out
}

func extractQuery(r *http.Request) map[string]string {
	out := make(map[string]string)
	for k, vals := range r.URL.Query() {
		if len(vals) > 0 {
			out[k] = vals[0] // solo el primer valor
		}
	}
	return out
}

func (c *PrototypesController) Mock(ctx *gin.Context) {

	entry := logger.FromContext(ctx.Request.Context())

	cc := customctx.NewCustomContext(ctx.Request.Context())

	entry.Info("Mocking request")

	pathParams := extractPath(ctx)
	headers := extractHeaders(ctx.Request)
	query := extractQuery(ctx.Request)

	response := c.prototypesService.Mock(cc, ctx.Request, pathParams, headers, query)

	if response.Error != nil {
		ctx.JSON(response.StatusCode, response.ToMapWithCustomContext(cc))
		return
	}

	ctx.JSON(response.StatusCode, response.Data)

}
