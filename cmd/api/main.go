package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	infrastructurepostgres "example.com/taskservice/internal/infrastructure/postgres"
	taskrepository "example.com/taskservice/internal/repository/postgres/task"
	taskgeneratorrepository "example.com/taskservice/internal/repository/postgres/task_generator"
	transporthttp "example.com/taskservice/internal/transport/http"
	swaggerdocs "example.com/taskservice/internal/transport/http/docs"
	taskhandler "example.com/taskservice/internal/transport/http/handlers/task"
	taskgeneratorhandler "example.com/taskservice/internal/transport/http/handlers/task_generator"
	usecasetask "example.com/taskservice/internal/usecase/task"
	usecasetaskgenerator "example.com/taskservice/internal/usecase/task_generator"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("loading config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := infrastructurepostgres.Open(ctx, cfg.DatabaseDSN)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	taskRepo := taskrepository.NewTaskRepository(pool)
	taskUsecase := usecasetask.NewTaskService(taskRepo)
	taskHandler := taskhandler.NewTaskHandler(taskUsecase)

	taskGeneratorRepo := taskgeneratorrepository.NewTaskGeneratorRepository(pool)
	taskGeneratorUsecase := usecasetaskgenerator.NewTaskGeneratorService(taskGeneratorRepo, taskUsecase, cfg.RunTaskGeneratorParams)
	taskGeneratorUsecase.Run(ctx)
	taskGeneratorHandler := taskgeneratorhandler.NewTaskGeneratorHandler(taskGeneratorUsecase)

	docsHandler := swaggerdocs.NewHandler()

	router := transporthttp.NewRouter(taskHandler, taskGeneratorHandler, docsHandler)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("shutdown http server", "error", err)
		}
	}()

	logger.Info("http server started", "addr", cfg.HTTPAddr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("listen and serve", "error", err)
		os.Exit(1)
	}
}

type config struct {
	RunTaskGeneratorParams usecasetaskgenerator.RunTaskGeneratorParams
	HTTPAddr               string
	DatabaseDSN            string
}

func loadConfig() (*config, error) {
	reprocessingCooldown, err := strconv.ParseInt(envOrDefault("REPROCESSING_COOLDOWN", "1"), 10, 64)
	if err != nil {
		return nil, err
	} else if reprocessingCooldown < 1 {
		return nil, errors.New("invalid run task generator params")
	}
	everyNMinutes, err := strconv.ParseInt(envOrDefault("EVERY_N_MINUTES", "1"), 10, 64)
	if err != nil {
		return nil, err
	} else if everyNMinutes < 1 {
		return nil, errors.New("invalid run task generator params")
	}
	daysAhead, err := strconv.ParseInt(envOrDefault("DAYS_AHEAD", "1"), 10, 64)
	if err != nil {
		return nil, err
	} else if daysAhead < 1 {
		return nil, errors.New("invalid run task generator params")
	}
	numWorkers, err := strconv.ParseInt(envOrDefault("NUM_WORKERS", "1"), 10, 64)
	if err != nil {
		return nil, err
	} else if numWorkers < 1 {
		return nil, errors.New("invalid run task generator params")
	}
	runTaskGeneratorParams := usecasetaskgenerator.RunTaskGeneratorParams{
		EveryNMinutes: everyNMinutes,
		DaysAhead:     daysAhead,
		NumWorkers:    numWorkers,
	}
	cfg := config{
		RunTaskGeneratorParams: runTaskGeneratorParams,
		HTTPAddr:               envOrDefault("HTTP_ADDR", ":8080"),
		DatabaseDSN:            envOrDefault("DATABASE_DSN", "postgres://postgres:postgres@localhost:5432/taskservice?sslmode=disable"),
	}

	if cfg.DatabaseDSN == "" {
		panic(fmt.Errorf("DATABASE_DSN is required"))
	}

	return &cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
