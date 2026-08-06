package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/kelwSagashi/sparkedge-go/internal/app"
	"github.com/kelwSagashi/sparkedge-go/internal/edge"
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
		runStatus(ctx)
	case "onboarding":
		runOnboarding(ctx, os.Args[2:])
	case "pair":
		runPair(ctx, os.Args[2:])
	case "connect":
		runConnect(ctx, os.Args[2:])
	case "disconnect":
		runDisconnect(ctx)
	case "reconnect":
		runReconnect(ctx)
	case "remove":
		runRemove(ctx)
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

func runStatus(ctx context.Context) {
	application := app.New()
	status, err := application.Edge.Status(ctx)
	mustPrint(status, err)
}

func runOnboarding(ctx context.Context, args []string) {
	flags := flag.NewFlagSet("onboarding", flag.ExitOnError)
	name := flags.String("name", "", "edge display name")
	description := flags.String("description", "", "edge description")
	lat := flags.String("lat", "", "edge latitude")
	lng := flags.String("lng", "", "edge longitude")
	tagsCSV := flags.String("tags", "", "comma-separated edge tags")
	_ = flags.Parse(args)

	application := app.New()
	result, err := application.Edge.SaveOnboarding(ctx, edge.OnboardingRequest{
		Name:        *name,
		Description: *description,
		Lat:         *lat,
		Lng:         *lng,
		Tags:        splitCSV(*tagsCSV),
	})
	mustPrint(result, err)
}

func runPair(ctx context.Context, args []string) {
	flags := flag.NewFlagSet("pair", flag.ExitOnError)
	token := flags.String("token", "", "cloud pairing token")
	name := flags.String("name", "", "optional edge name")
	_ = flags.Parse(args)

	application := app.New()
	result, err := application.Edge.Pair(ctx, edge.PairRequest{Token: *token, Name: *name})
	mustPrint(result, err)
}

func runConnect(ctx context.Context, args []string) {
	flags := flag.NewFlagSet("connect", flag.ExitOnError)
	email := flags.String("email", "", "Spark Cloud account email")
	password := flags.String("password", "", "Spark Cloud account password")
	_ = flags.Parse(args)

	application := app.New()
	result, err := application.Edge.Connect(ctx, edge.ConnectRequest{Email: *email, Password: *password})
	mustPrint(result, err)
}

func runDisconnect(ctx context.Context) {
	application := app.New()
	mustPrint(map[string]any{"success": true}, application.Edge.Disconnect(ctx))
}

func runReconnect(ctx context.Context) {
	application := app.New()
	mustPrint(map[string]any{"success": true}, application.Edge.Reconnect(ctx))
}

func runRemove(ctx context.Context) {
	application := app.New()
	mustPrint(map[string]any{"success": true}, application.Edge.Remove(ctx))
}

func mustPrint(payload any, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func splitCSV(value string) []string {
	if value == "" {
		return []string{}
	}
	var result []string
	current := ""
	for _, char := range value {
		if char == ',' {
			if current != "" {
				result = append(result, current)
			}
			current = ""
			continue
		}
		current += string(char)
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
