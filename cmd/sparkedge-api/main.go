package main

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/kelwSagashi/sparkedge-go/internal/app"
	"github.com/kelwSagashi/sparkedge-go/internal/config"
)

func main() {
	ctx := context.Background()
	configManager := config.NewManager("")
	_, runtimeCfg, err := configManager.Load()
	if err != nil {
		log.Fatal(err)
	}

	application := app.New(configManager)
	application.StartScheduler(ctx, 0)

	addr := ":" + strconv.Itoa(runtimeCfg.HTTPPort)

	server := application.HTTPServer(addr)

	log.Printf("sparkedge api listening on %s", addr)
	if err := server.ListenAndServe(ctx); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
