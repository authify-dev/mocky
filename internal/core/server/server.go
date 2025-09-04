package server

import (
	"fmt"
	"mocky/internal/api/health"
	"mocky/internal/api/v1/prototypes"
	"mocky/internal/core/router"
	"mocky/internal/core/settings"

	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

func Run() {
	r := setUpRouter()
	r.Run(fmt.Sprintf(":%d", settings.Settings.PORT))
}

func RunLambda() {

	r := setUpRouter()

	// Adaptar Gin a Lambda
	ginLambda = ginadapter.New(r)

	// Iniciar Lambda
	lambda.Start(ginLambda.Proxy)
}

func setUpRouter() *gin.Engine {

	r := router.NewRouter()

	// Rutas de health
	health.SetupHealthModule(r)

	// Rutas de prototypes
	prototypes.SetupPrototypesModule(r)

	return r
}
