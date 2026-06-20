package main

import (
	"context"
	"log"

	"github.com/pavestack/service-template-api/internal/app"
)

func main() {
	a := app.New()
	if err := a.Run(context.Background()); err != nil {
		log.Fatalf("app run failed: %v", err)
	}
}
