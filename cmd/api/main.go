package main

import (
	"common/domain/logger"
	"log"
	"mocky/internal/core/server"
	"mocky/internal/core/settings"
)

func main() {
	settings.LoadDotEnv()

	settings.LoadEnvs()

	logger.InitLogger(settings.Settings.ENVIRONMENT, "mocky", settings.Settings.LOKI_URL)

	switch settings.Settings.DEPLOY_MODE {
	case settings.DeployModeAPI:
		server.Run()
	case settings.DeployModeLambda:
		server.RunLambda()
	default:
		log.Fatalf("Invalid deploy mode: %s", settings.Settings.DEPLOY_MODE)
	}
}
