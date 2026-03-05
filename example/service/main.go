package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	cdshooks "github.com/your-org/cds-hooks-go/cdshooks"
	"github.com/your-org/cds-hooks-go/cdshooks/service"
)

func main() {
	ageCheck, err := cdshooks.NewService("patient-view-age-check").
		ForHook(cdshooks.HookPatientView).
		WithTitle("Patient Age Medication Review").
		WithDescription("Flags patients 65+ for STOPP/START criteria review.").
		WithPrefetch("patient", "Patient/{{context.patientId}}").
		WithPrefetch("meds", "MedicationRequest?subject={{context.patientId}}&status=active").
		HandleFunc(handlePatientView).
		Build()
	if err != nil {
		slog.Error("failed to build service", "error", err)
		os.Exit(1)
	}

	server := service.NewServer(
		service.WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
		service.WithCORSOrigins("*"),
		service.WithRequestTimeout(5*time.Second),
	)

	server.Register(ageCheck)

	if err := server.ListenAndServe(":8080"); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func handlePatientView(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
	_, err := cdshooks.DecodeContext[cdshooks.PatientViewContext](req.Context)
	if err != nil {
		return cdshooks.CDSResponse{}, err
	}

	patient, err := req.Prefetch.Patient("patient")
	if err != nil {
		return cdshooks.EmptyResponse(), nil
	}

	age, err := cdshooks.PatientAge(patient)
	if err != nil || age < 65 {
		return cdshooks.EmptyResponse(), nil
	}

	meds, _ := req.Prefetch.Bundle("meds")
	medCount := cdshooks.BundleEntryCount(meds)

	indicator := cdshooks.IndicatorInfo
	if medCount >= 5 {
		indicator = cdshooks.IndicatorWarning
	}

	url := "https://www.ncbi.nlm.nih.gov/pmc/articles/PMC4339726/"
	card, err := cdshooks.NewCard(
		fmt.Sprintf("Medication review recommended — patient aged %d (%d active medications)", age, medCount),
		indicator,
	).
		WithSource(cdshooks.Source{
			Label: "STOPP/START Criteria v3",
			URL:   &url,
		}).
		WithDetail("Consider a pharmacist-led medication review per STOPP/START criteria.").
		AddLink(cdshooks.Link{
			Label: "Open Medication Review App",
			URL:   "https://apps.example.org/med-review",
			Type:  cdshooks.LinkSmart,
		}).
		Build()
	if err != nil {
		return cdshooks.CDSResponse{}, err
	}

	return cdshooks.NewResponse().AddCard(card).Build(), nil
}
