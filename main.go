package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"mascot/internal/app"
	"mascot/internal/config"
)

const serviceName = "mascot"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal(err)
	}

	logger = logger.With(zap.String("service", serviceName))

	cfg, err := config.Init(serviceName)
	if err != nil {
		logger.Fatal("init config", zap.Error(err))
	}

	service := app.NewService(logger)
	go service.Start(ctx, cfg)

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := service.Shutdown(ctx); err != nil {
		logger.Fatal("shutdown service", zap.Error(err))
	}
}
