package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Danil-Ivonin/WalletTest/internal/app"
)

func main() {
	if err := app.Run(context.Background()); err != nil {
		log.Fatal(fmt.Errorf("wallet api stopped %w", err))
	}
}
