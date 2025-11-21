package cmd

import (
	"BackendTraineeAssignmentAutumn2025/internal/config"
	"go.uber.org/zap"
)

func main() {
	// TODO config init
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	// TODO logger init
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	// TODO db init

	// TODO layers init

	// TODO server init
}
