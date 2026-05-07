package app

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Danil-Ivonin/WalletTest/internal/adapters/db"
	"github.com/Danil-Ivonin/WalletTest/internal/adapters/httpadapter"
	"github.com/Danil-Ivonin/WalletTest/internal/repository"
	"github.com/Danil-Ivonin/WalletTest/internal/service"
	"github.com/spf13/viper"
)

func Run(ctx context.Context) error {
	//db initialization
	pool, err := db.NewPostgres(ctx, db.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
	})
	if err != nil {
		return fmt.Errorf("failed to initialize db: %w", err)
	}

	repo := repository.NewRepository(pool)
	srv := service.NewService(repo)
	handler := httpadapter.NewHandler(srv)

	server := &http.Server{
		Addr:    "addr",
		Handler: handler.InitRoutes(),
	}

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
