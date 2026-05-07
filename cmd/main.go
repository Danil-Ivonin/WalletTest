package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Danil-Ivonin/WalletTest/internal/app"
	"github.com/Danil-Ivonin/WalletTest/internal/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func main() {
	if err := logger.SetupLogger("info"); err != nil {
		panic(err)
	}
	if err := initConfig(); err != nil {
		logrus.Fatal(err)
	}
	if err := app.Run(context.Background()); err != nil {
		log.Fatal(fmt.Errorf("wallet api stopped %w", err))
	}
}

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
