package main

import (
	"context"
	"log"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/app"
)

func main() {
	ctx := context.Background()

	application, err := app.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
