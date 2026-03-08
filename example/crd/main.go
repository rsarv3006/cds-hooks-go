package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	fhir "github.com/samply/golang-fhir-models/fhir-models/fhir"
	cdshooks "github.com/your-org/cds-hooks-go/cdshooks"
	"github.com/your-org/cds-hooks-go/cdshooks/service"
)

func main() {
	orderSign, err := cdshooks.NewService("crd-order-sign").
		ForHook(cdshooks.HookOrderSign).
		WithTitle("CRD Coverage Requirements Discovery").
		WithDescription("Returns coverage requirements and documentation needs for medication orders.").
		WithPrefetch("patient", "Patient/{{context.patientId}}").
		WithPrefetch("coverage", "Coverage?patient={{context.patientId}}&status=active").
		WithPrefetch("medications", "MedicationRequest?patient={{context.patientId}}&status=active").
		HandleFunc(handleOrderSign).
		Build()
	if err != nil {
		slog.Error("failed to build order-sign service", "error", err)
		os.Exit(1)
	}

	orderSelect, err := cdshooks.NewService("crd-order-select").
		ForHook(cdshooks.HookOrderSelect).
		WithTitle("CRD Order Selection").
		WithDescription("Evaluates coverage requirements when orders are selected.").
		WithPrefetch("patient", "Patient/{{context.patientId}}").
		WithPrefetch("coverage", "Coverage?patient={{context.patientId}}&status=active").
		HandleFunc(handleOrderSelect).
		Build()
	if err != nil {
		slog.Error("failed to build order-select service", "error", err)
		os.Exit(1)
	}

	appointmentBook, err := cdshooks.NewService("crd-appointment-book").
		ForHook(cdshooks.HookAppointmentBook).
		WithTitle("CRD Appointment Coverage").
		WithDescription("Checks coverage for scheduled appointments.").
		WithPrefetch("patient", "Patient/{{context.patientId}}").
		WithPrefetch("coverage", "Coverage?patient={{context.patientId}}&status=active").
		HandleFunc(handleAppointmentBook).
		Build()
	if err != nil {
		slog.Error("failed to build appointment-book service", "error", err)
		os.Exit(1)
	}

	encounterStart, err := cdshooks.NewService("crd-encounter-start").
		ForHook(cdshooks.HookEncounterStart).
		WithTitle("CRD Encounter Start").
		WithDescription("Provides coverage information at encounter start.").
		WithPrefetch("patient", "Patient/{{context.patientId}}").
		WithPrefetch("coverage", "Coverage?patient={{context.patientId}}&status=active").
		HandleFunc(handleEncounterStart).
		Build()
	if err != nil {
		slog.Error("failed to build encounter-start service", "error", err)
		os.Exit(1)
	}

	encounterDischarge, err := cdshooks.NewService("crd-encounter-discharge").
		ForHook(cdshooks.HookEncounterDischarge).
		WithTitle("CRD Encounter Discharge").
		WithDescription("Provides discharge instructions and coverage information.").
		WithPrefetch("patient", "Patient/{{context.patientId}}").
		WithPrefetch("coverage", "Coverage?patient={{context.patientId}}&status=active").
		HandleFunc(handleEncounterDischarge).
		Build()
	if err != nil {
		slog.Error("failed to build encounter-discharge service", "error", err)
		os.Exit(1)
	}

	orderDispatch, err := cdshooks.NewService("crd-order-dispatch").
		ForHook(cdshooks.HookOrderDispatch).
		WithTitle("CRD Order Dispatch").
		WithDescription("Validates coverage before dispatching orders to fulfillment.").
		WithPrefetch("patient", "Patient/{{context.patientId}}").
		WithPrefetch("coverage", "Coverage?patient={{context.patientId}}&status=active").
		HandleFunc(handleOrderDispatch).
		Build()
	if err != nil {
		slog.Error("failed to build order-dispatch service", "error", err)
		os.Exit(1)
	}

	server := service.NewServer(
		service.WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
		service.WithCORSOrigins("*"),
		service.WithRequestTimeout(5*time.Second),
	)

	server.Register(orderSign, orderSelect, appointmentBook, encounterStart, encounterDischarge, orderDispatch)

	slog.Info("CRD example server starting on :8080")
	if err := server.ListenAndServe(":8080"); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func handleOrderSign(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
	var hookCtx cdshooks.OrderSignContext
	if err := json.Unmarshal(req.Context, &hookCtx); err != nil {
		return cdshooks.CDSResponse{}, err
	}

	patient, err := req.Prefetch.Patient("patient")
	if err != nil {
		return createInstructionsCard("Patient data required", "Unable to fetch patient information.")
	}

	age, _ := cdshooks.PatientAge(patient)
	_ = age

	covBundle, _ := req.Prefetch.Bundle("coverage")
	coverages := extractCoverages(covBundle)

	if len(coverages) == 0 {
		card, _ := cdshooks.NewCard(
			"No active coverage found",
			cdshooks.IndicatorWarning,
		).
			WithSource(cdshooks.Source{Label: "CRD Service"}).
			WithDetail("The patient does not have active insurance coverage on file. Coverage verification is recommended prior to signing the order.").
			Build()
		return cdshooks.NewResponse().AddCard(card).Build(), nil
	}

	coverageInfo := buildCoverageInformation(coverages)

	patientName := getPatientName(patient)
	detail := fmt.Sprintf("Patient: %s\n\nActive Coverage:\n%s\n\nDocumentation may be required for this order.", patientName, coverageInfo)

	card, _ := cdshooks.NewCard(
		"Coverage requirements checked",
		cdshooks.IndicatorInfo,
	).
		WithSource(cdshooks.Source{
			Label: "CRD Coverage Service",
			URL:   stringPtr("https://example.org/crd"),
		}).
		WithDetail(detail).
		AddExtension(cdshooks.ExtCoverageInformation, map[string]any{
			"requirementsMet": true,
		}).
		Build()

	return cdshooks.NewResponse().AddCard(card).Build(), nil
}

func handleOrderSelect(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
	var hookCtx cdshooks.OrderSelectContext
	if err := json.Unmarshal(req.Context, &hookCtx); err != nil {
		return cdshooks.CDSResponse{}, err
	}

	card, _ := cdshooks.NewCard(
		"Order selection noted",
		cdshooks.IndicatorInfo,
	).
		WithSource(cdshooks.Source{Label: "CRD Service"}).
		WithDetail("Coverage requirements will be evaluated when the order is signed.").
		Build()

	return cdshooks.NewResponse().AddCard(card).Build(), nil
}

func handleAppointmentBook(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
	var hookCtx cdshooks.AppointmentBookContext
	if err := json.Unmarshal(req.Context, &hookCtx); err != nil {
		return cdshooks.CDSResponse{}, err
	}

	covBundle, _ := req.Prefetch.Bundle("coverage")
	coverages := extractCoverages(covBundle)

	if len(coverages) == 0 {
		card, _ := cdshooks.NewCard(
			"Coverage verification needed",
			cdshooks.IndicatorWarning,
		).
			WithSource(cdshooks.Source{Label: "CRD Service"}).
			WithDetail("No active coverage found. Please verify insurance before scheduling.").
			Build()
		return cdshooks.NewResponse().AddCard(card).Build(), nil
	}

	coverageInfo := buildCoverageInformation(coverages)
	detail := fmt.Sprintf("Active coverage found:\n\n%s", coverageInfo)

	card, _ := cdshooks.NewCard(
		"Appointment coverage verified",
		cdshooks.IndicatorInfo,
	).
		WithSource(cdshooks.Source{Label: "CRD Service"}).
		WithDetail(detail).
		AddExtension(cdshooks.ExtCoverageInformation, map[string]any{}).
		Build()

	return cdshooks.NewResponse().AddCard(card).Build(), nil
}

func handleEncounterStart(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
	var hookCtx cdshooks.EncounterStartContext
	if err := json.Unmarshal(req.Context, &hookCtx); err != nil {
		return cdshooks.CDSResponse{}, err
	}

	covBundle, _ := req.Prefetch.Bundle("coverage")
	coverages := extractCoverages(covBundle)

	if len(coverages) == 0 {
		card, _ := cdshooks.NewCard(
			"No active coverage",
			cdshooks.IndicatorWarning,
		).
			WithSource(cdshooks.Source{Label: "CRD Service"}).
			WithDetail("No active insurance coverage found for this patient.").
			Build()
		return cdshooks.NewResponse().AddCard(card).Build(), nil
	}

	coverageInfo := buildCoverageInformation(coverages)

	card, _ := cdshooks.NewCard(
		"Encounter coverage information",
		cdshooks.IndicatorInfo,
	).
		WithSource(cdshooks.Source{Label: "CRD Service"}).
		WithDetail(coverageInfo).
		AddLink(cdshooks.Link{
			Label: "View Coverage Details",
			URL:   "https://example.org/coverage-portal",
			Type:  cdshooks.LinkSmart,
		}).
		Build()

	return cdshooks.NewResponse().AddCard(card).Build(), nil
}

func handleEncounterDischarge(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
	var hookCtx cdshooks.EncounterDischargeContext
	if err := json.Unmarshal(req.Context, &hookCtx); err != nil {
		return cdshooks.CDSResponse{}, err
	}

	card, _ := cdshooks.NewCard(
		"Discharge instructions available",
		cdshooks.IndicatorInfo,
	).
		WithSource(cdshooks.Source{Label: "CRD Service"}).
		WithDetail("Coverage information has been documented. Patient may contact insurance for coverage questions.").
		AddExtension(cdshooks.ExtInstructions, map[string]any{
			"text": "Follow up with your primary care provider within 7 days. Contact insurance for coverage questions.",
		}).
		Build()

	return cdshooks.NewResponse().AddCard(card).Build(), nil
}

func handleOrderDispatch(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
	var hookCtx cdshooks.OrderDispatchContext
	if err := json.Unmarshal(req.Context, &hookCtx); err != nil {
		return cdshooks.CDSResponse{}, err
	}

	covBundle, _ := req.Prefetch.Bundle("coverage")
	coverages := extractCoverages(covBundle)

	if len(coverages) == 0 {
		card, _ := cdshooks.NewCard(
			"Cannot dispatch - no coverage",
			cdshooks.IndicatorCritical,
		).
			WithSource(cdshooks.Source{Label: "CRD Service"}).
			WithDetail("Order cannot be dispatched without active insurance coverage. Please verify coverage before proceeding.").
			Build()
		return cdshooks.NewResponse().AddCard(card).Build(), nil
	}

	card, _ := cdshooks.NewCard(
		"Order ready for dispatch",
		cdshooks.IndicatorSuccess,
	).
		WithSource(cdshooks.Source{Label: "CRD Service"}).
		WithDetail("Coverage verified. Order is ready for dispatch to fulfillment.").
		AddExtension(cdshooks.ExtCoverageInformation, map[string]any{
			"requirementsMet": true,
			"dispatchAllowed": true,
		}).
		Build()

	return cdshooks.NewResponse().AddCard(card).Build(), nil
}

func extractCoverages(bundle fhir.Bundle) []fhir.Coverage {
	if bundle.Entry == nil {
		return nil
	}
	var coverages []fhir.Coverage
	for _, entry := range bundle.Entry {
		if entry.Resource == nil {
			continue
		}
		var cov fhir.Coverage
		if err := json.Unmarshal(entry.Resource, &cov); err == nil {
			coverages = append(coverages, cov)
		}
	}
	return coverages
}

func buildCoverageInformation(coverages []fhir.Coverage) string {
	info := ""
	for _, cov := range coverages {
		info += fmt.Sprintf("- Status: %s\n", cov.Status)
		for _, payor := range cov.Payor {
			if payor.Display != nil {
				info += fmt.Sprintf("  Payor: %s\n", *payor.Display)
			}
		}
		for _, class := range cov.Class {
			classType := ""
			if len(class.Type.Coding) > 0 && class.Type.Coding[0].Display != nil {
				classType = *class.Type.Coding[0].Display
			}
			if classType != "" {
				info += fmt.Sprintf("  Class: %s - %s\n", classType, class.Value)
			}
		}
	}
	return info
}

func getPatientName(patient fhir.Patient) string {
	for _, name := range patient.Name {
		if name.Family != nil {
			given := ""
			if len(name.Given) > 0 {
				given = name.Given[0]
			}
			if given != "" {
				return fmt.Sprintf("%s, %s", *name.Family, given)
			}
			return *name.Family
		}
	}
	return "Unknown"
}

func createInstructionsCard(summary, detail string) (cdshooks.CDSResponse, error) {
	card, err := cdshooks.NewCard(summary, cdshooks.IndicatorInfo).
		WithSource(cdshooks.Source{Label: "CRD Service"}).
		WithDetail(detail).
		AddExtension(cdshooks.ExtInstructions, map[string]any{
			"text": detail,
		}).
		Build()
	if err != nil {
		return cdshooks.CDSResponse{}, err
	}
	return cdshooks.NewResponse().AddCard(card).Build(), nil
}

func stringPtr(s string) *string {
	return &s
}
