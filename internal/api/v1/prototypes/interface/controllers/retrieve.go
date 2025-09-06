package controllers

import (
	"common/domain/customctx"
	"common/domain/logger"

	"github.com/gin-gonic/gin"
)

func (c *PrototypesController) Retrieve(ctx *gin.Context) {
	entry := logger.FromContext(ctx.Request.Context())

	entry.Info("Retrieving prototype")

	cc := customctx.NewCustomContext(ctx.Request.Context())

	prototype := c.prototypesService.Retrieve(cc, ctx.Param("id"))

	ctx.JSON(prototype.StatusCode, prototype.ToMapWithCustomContext(cc))
}
