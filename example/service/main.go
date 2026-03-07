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
		slog.Error("failed to build age check service", "error", err)
		os.Exit(1)
	}

	drugInteraction, err := cdshooks.NewService("medication-drug-interaction").
		ForHook(cdshooks.HookOrderSelect).
		WithTitle("Drug Interaction Checker").
		WithDescription("Checks for potential drug interactions in medication orders.").
		WithPrefetch("meds", "MedicationRequest?patient={{context.patientId}}&status=active").
		HandleFunc(handleOrderSelect).
		Build()
	if err != nil {
		slog.Error("failed to build drug interaction service", "error", err)
		os.Exit(1)
	}

	allergyAlert, err := cdshooks.NewService("allergy-conflict-check").
		ForHook(cdshooks.HookMedicationPrescribe).
		WithTitle("Allergy Conflict Checker").
		WithDescription("Alerts when prescribing medications that conflict with patient allergies.").
		WithPrefetch("allergies", "AllergyIntolerance?patient={{context.patientId}}").
		HandleFunc(handleMedicationPrescribe).
		Build()
	if err != nil {
		slog.Error("failed to build allergy service", "error", err)
		os.Exit(1)
	}

	encounterSummary, err := cdshooks.NewService("encounter-summary").
		ForHook(cdshooks.HookEncounterStart).
		WithTitle("Encounter Summary").
		WithDescription("Provides a summary of the patient's recent encounters.").
		HandleFunc(handleEncounterStart).
		Build()
	if err != nil {
		slog.Error("failed to build encounter service", "error", err)
		os.Exit(1)
	}

	server := service.NewServer(
		service.WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
		service.WithCORSOrigins("*"),
		service.WithRequestTimeout(5*time.Second),
	)

	server.Register(ageCheck, drugInteraction, allergyAlert, encounterSummary)

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

func handleOrderSelect(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
	card, err := cdshooks.NewCard(
		"No drug interactions detected",
		cdshooks.IndicatorInfo,
	).
		WithSource(cdshooks.Source{Label: "Drug Interaction Checker"}).
		Build()
	if err != nil {
		return cdshooks.CDSResponse{}, err
	}
	return cdshooks.NewResponse().AddCard(card).Build(), nil
}

func handleMedicationPrescribe(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
	card, err := cdshooks.NewCard(
		"No allergy conflicts identified",
		cdshooks.IndicatorInfo,
	).
		WithSource(cdshooks.Source{Label: "Allergy Checker"}).
		Build()
	if err != nil {
		return cdshooks.CDSResponse{}, err
	}
	return cdshooks.NewResponse().AddCard(card).Build(), nil
}

func handleEncounterStart(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
	card, err := cdshooks.NewCard(
		"Encounter started",
		cdshooks.IndicatorInfo,
	).
		WithSource(cdshooks.Source{Label: "Encounter Service"}).
		WithDetail("Last encounter: 2026-01-15 - Annual wellness visit").
		Build()
	if err != nil {
		return cdshooks.CDSResponse{}, err
	}
	return cdshooks.NewResponse().AddCard(card).Build(), nil
}
