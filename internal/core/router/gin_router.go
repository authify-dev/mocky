package router

import (
	middleware "mocky/internal/api/middlewares"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	r.Use(middleware.RequestLogMiddleware())
	r.Use(middleware.LoggerMiddleware())

	return r
}
