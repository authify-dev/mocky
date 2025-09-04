package controllers

import (
	"common/domain/customctx"
	"common/domain/logger"
	"common/interface/cdtos"
	"mocky/internal/api/v1/prototypes/interface/dtos"

	"github.com/gin-gonic/gin"
)

func (c *PrototypesController) Create(ctx *gin.Context) {

	entry := logger.FromContext(ctx.Request.Context())

	cc := customctx.NewCustomContext(ctx.Request.Context())

	dto := cdtos.GetDTOWithResponse[dtos.CreatePrototypeDTO](ctx, cc)
	if dto.Error != nil {
		entry.Error(dto.Error.Error())
		ctx.JSON(dto.StatusCode, dto.ToMapWithCustomContext(cc))
		return
	}

	command := dto.Data.ToCommand()

	response := c.prototypesService.Create(cc, command)

	ctx.JSON(response.StatusCode, response.ToMapWithCustomContext(cc))

}
