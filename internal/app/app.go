package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Danil-Ivonin/WalletTest/internal/config"
	"github.com/Danil-Ivonin/WalletTest/internal/handler"
	"github.com/Danil-Ivonin/WalletTest/internal/repository"
	"github.com/Danil-Ivonin/WalletTest/internal/repository/db"
	"github.com/Danil-Ivonin/WalletTest/internal/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Run(ctx context.Context) error {
	//db initialization
	pool, err := db.NewPostgres(ctx, config.DSN())
	if err != nil {
		return fmt.Errorf("failed to initialize db: %w", err)
	}
	defer pool.Close()
	err = db.RunMigrations(ctx, pool, "migrations")
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	//App initialization
	repo := repository.NewRepository(pool)
	srv := service.NewService(repo)
	h := handler.NewHandler(srv)

	server := &http.Server{
		Addr:    config.HTTPAddr(),
		Handler: h.InitRoutes(),
	}

	//Run server in gorutine
	serverErr := make(chan error, 1)
	go func() {
		logrus.WithField("address", config.HTTPAddr()).Info("wallet api started")
		err = server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	//Handle signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	select {
	case <-ctx.Done():
	case <-stop:
	case err := <-serverErr:
		return err
	}

	//Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), viper.GetDuration("app.shutdown_timeout"))
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}

	return <-serverErr
}
