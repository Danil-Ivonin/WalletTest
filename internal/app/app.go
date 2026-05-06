package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Danil-Ivonin/WalletTest/internal/adapters/db"
	"github.com/Danil-Ivonin/WalletTest/internal/adapters/httpadapter"
	"github.com/Danil-Ivonin/WalletTest/internal/repository"
	"github.com/Danil-Ivonin/WalletTest/internal/service"
)

func Run(ctx context.Context) error {
	//db initialization
	pool, err := db.NewPostgres(ctx, "url")
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
