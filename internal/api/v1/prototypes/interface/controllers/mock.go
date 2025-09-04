package controllers

import (
	"common/domain/customctx"
	"common/domain/logger"

	"github.com/gin-gonic/gin"
)

func (c *PrototypesController) Mock(ctx *gin.Context) {

	entry := logger.FromContext(ctx.Request.Context())

	cc := customctx.NewCustomContext(ctx.Request.Context())

	entry.Info("Mocking request")

	response := c.prototypesService.Mock(cc, ctx.Request)

	if response.Error != nil {
		ctx.JSON(response.StatusCode, response.ToMapWithCustomContext(cc))
		return
	}

	ctx.JSON(response.StatusCode, response.Data)

}
