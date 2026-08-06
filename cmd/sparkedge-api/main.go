package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/kelwSagashi/sparkedge-go/internal/app"
)

func main() {
	ctx := context.Background()
	application := app.New()
	application.StartScheduler(ctx, 0)

	addr := os.Getenv("SPARKEDGE_HTTP_ADDR")
	if addr == "" {
		addr = ":3009"
	}

	server := application.HTTPServer(addr)

	log.Printf("sparkedge api listening on %s", addr)
	if err := server.ListenAndServe(ctx); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
