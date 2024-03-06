package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"report-generator/config"
	"report-generator/internal/delivery"
	"report-generator/internal/repository"
	"report-generator/internal/service"
	"syscall"

	"go.uber.org/zap"
)

func Run() error {
	var logger *zap.Logger
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		return err
	}
	//nolint
	defer logger.Sync()

	sugar := logger.Sugar()
	cfg, err := config.New()
	if err != nil {
		sugar.Error(err.Error())
		return err
	}
	repo := repository.NewRepository(sugar, cfg.Generator, nil)
	service := service.NewService(repo, sugar, cfg)
	handler := delivery.NewHandler(service, sugar)
	srv := http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.App.Port),
		Handler:           handler.InitRoutes(),
		ReadTimeout:       cfg.App.Timeout.Read,
		WriteTimeout:      cfg.App.Timeout.Write,
		ReadHeaderTimeout: cfg.App.Timeout.ReadHeader,
	}
	errChan := make(chan error, 1)
	go func() {
		sugar.Infof("starting server on: %d", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
			return
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-quit:
	case <-errChan:
		sugar.Error("server runtime error: ", err)
	}

	sugar.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.App.Timeout.Shutdown)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		sugar.Errorf("Server forced to shutdown: %s", err)
		return err
	}
	return nil
}
