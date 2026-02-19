package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/your-org/cds-hooks-go/cdshooks"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	client := cdshooks.NewClient("https://cds.example.org",
		cdshooks.WithTimeout(3*time.Second),
		cdshooks.WithBearerToken("your-ehr-token"),
	)

	ctx := context.Background()

	services, err := client.Discover(ctx)
	if err != nil {
		slog.Error("failed to discover services", "error", err)
		os.Exit(1)
	}

	slog.Info("discovered services", "count", len(services))
	for _, svc := range services {
		slog.Info("service", "id", svc.ID, "hook", svc.Hook, "title", svc.Title)
	}

	response, err := client.Call(ctx, "patient-view-age-check",
		cdshooks.PatientViewContext{
			UserID:    "Practitioner/abc",
			PatientID: "Patient/123",
		},
		map[string]any{
			"patient": map[string]any{
				"resourceType": "Patient",
				"id":           "123",
				"birthDate":    "1955-03-15",
			},
		},
	)
	if err != nil {
		slog.Error("failed to call service", "error", err)
		os.Exit(1)
	}

	for _, card := range response.Cards {
		fmt.Printf("Card: %s [%s]\n", card.Summary, card.Indicator)
		if card.Detail != "" {
			fmt.Printf("  Detail: %s\n", card.Detail)
		}
	}
}
