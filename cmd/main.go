package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Danil-Ivonin/WalletTest/internal/app"
	"github.com/Danil-Ivonin/WalletTest/internal/config"
	"github.com/Danil-Ivonin/WalletTest/internal/logger"
	"github.com/sirupsen/logrus"
)

func main() {
	if err := logger.SetupLogger("info"); err != nil {
		panic(err)
	}
	if err := config.Load("config.env"); err != nil {
		logrus.Fatal(err)
	}
	if err := app.Run(context.Background()); err != nil {
		log.Fatal(fmt.Errorf("wallet api stopped %w", err))
	}
}
