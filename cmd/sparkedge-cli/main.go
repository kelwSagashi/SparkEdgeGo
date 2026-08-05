package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/kelwSagashi/sparkedge-go/internal/app"
)

func main() {
	ctx := context.Background()
	command := "start"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "start":
		runStart(ctx, os.Args[2:])
	case "status":
		runStatus()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		os.Exit(2)
	}
}

func runStart(ctx context.Context, args []string) {
	flags := flag.NewFlagSet("start", flag.ExitOnError)
	addr := flags.String("addr", ":3009", "HTTP listen address")
	_ = flags.Parse(args)

	application := app.New()
	if err := application.HTTPServer(*addr).ListenAndServe(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runStatus() {
	fmt.Println("SparkEdge Go: initialized")
}
