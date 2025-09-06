package controllers

import (
	"common/domain/customctx"
	"common/domain/logger"

	"github.com/gin-gonic/gin"
)

func (c *PrototypesController) List(ctx *gin.Context) {

	entry := logger.FromContext(ctx.Request.Context())

	entry.Info("List prototypes")

	cc := customctx.NewCustomContext(ctx.Request.Context())

	prototypes := c.prototypesService.List(cc)

	ctx.JSON(prototypes.StatusCode, prototypes.ToMapWithCustomContext(cc))
}
