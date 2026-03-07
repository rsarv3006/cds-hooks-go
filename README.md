# CDS Hooks Go SDK

An idiomatic Go SDK for the [CDS Hooks 2.0 specification](https://cds-hooks.org/).

## Overview

This SDK handles all protocol concerns so consumers only write clinical logic. It supports:

- **CDS Service Authors**: Implement decision support services that receive clinical context and return recommendations via cards
- **CDS Clients (EHRs)**: Call CDS services from EHR systems to get real-time decision support

## Requirements

- Go 1.25+
- chi router (included as dependency)

## Installation

```bash
go get github.com/your-org/cds-hooks-go
```

## Quick Start

### Running a CDS Service

```bash
cd example/service
go run main.go
```

The server starts on `http://localhost:8080`.

#### Test the service

```bash
# Discovery endpoint
curl http://localhost:8080/cds-services

# Invoke a hook
curl -X POST http://localhost:8080/cds-services/patient-view-age-check \
  -H "Content-Type: application/json" \
  -d '{
    "hook": "patient-view",
    "hookInstance": "550e8400-e29b-41d4-a716-446655440000",
    "context": {
      "userId": "Practitioner/example",
      "patientId": "Patient/123"
    },
    "prefetch": {
      "patient": {"resourceType": "Patient", "id": "123", "birthDate": "1955-03-15"},
      "meds": {"resourceType": "Bundle", "type": "searchset", "entry": []}
    }
  }'
```

### Running an EHR Client

```bash
cd example/client
go run main.go
```

## Creating a CDS Service

### Basic Service

```go
import (
    "context"
    cdshooks "github.com/your-org/cds-hooks-go/cdshooks"
    "github.com/your-org/cds-hooks-go/cdshooks/service"
)

svc, err := cdshooks.NewService("my-service").
    ForHook(cdshooks.HookPatientView).
    WithTitle("My Service").
    WithDescription("Does something useful").
    HandleFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
        // Your clinical logic here
        return cdshooks.EmptyResponse(), nil
    }).Build()

server := service.NewServer()
server.Register(svc)
server.ListenAndServe(":8080")
```

### Service Builder Options

| Method | Description |
|--------|-------------|
| `NewService(id)` | Create a new service with the given ID |
| `ForHook(hook)` | Set the hook this service responds to |
| `WithTitle(title)` | Human-readable title |
| `WithDescription(desc)` | Detailed description |
| `WithPrefetch(key, query)` | Request prefetch data using FHIR queries with template variables |
| `WithUsageRequirements(req)` | Describe service usage requirements |
| `Handle(handler)` | Set a Handler implementation |
| `HandleFunc(fn)` | Set a handler function directly |

### Supported Hooks

```go
cdshooks.HookPatientView        // patient-view
cdshooks.HookOrderSelect        // order-select
cdshooks.HookOrderSign          // order-sign
cdshooks.HookAppointmentBook    // appointment-book
cdshooks.HookEncounterStart     // encounter-start
cdshooks.HookEncounterDischarge // encounter-discharge
```

### Context Types

Each hook has a corresponding context type:

```go
cdshooks.PatientViewContext{
    UserID: "Practitioner/123",
    PatientID: "Patient/456",
    EncounterID: "Encounter/789",  // optional
}

cdshooks.OrderSelectContext{
    UserID: "Practitioner/123",
    PatientID: "Patient/456",
    Selections: []string{"MedicationRequest/1"},
}
```

## Creating Cards

Cards are the primary way services return recommendations to the EHR.

### Basic Card

```go
card, err := cdshooks.NewCard("Patient is due for screening", cdshooks.IndicatorInfo).
    WithSource(cdshooks.Source{Label: "My Service"}).
    Build()
```

### Card with Details

```go
card, err := cdshooks.NewCard("Medication review recommended", cdshooks.IndicatorWarning).
    WithSource(cdshooks.Source{Label: "Pharmacy CDS"}).
    WithDetail("Patient is on 5+ medications. Consider a pharmacist review.").
    Build()
```

### Card Indicators

```go
cdshooks.IndicatorInfo     // Informational
cdshooks.IndicatorWarning  // Warning - may need attention
cdshooks.IndicatorCritical // Critical - urgent action needed
```

### Cards with Suggestions

Suggestions allow users to select from predefined actions:

```go
suggestion := cdshooks.NewSuggestion("Order screening mammogram").
    WithRecommended(true).
    AddAction(cdshooks.NewAction(cdshooks.ActionCreate, "Create mammogram order").
        WithResource(resourceJSON).
        Build()).
    Build()

card, err := cdshooks.NewCard("Screening due", cdshooks.IndicatorInfo).
    WithSource(cdshooks.Source{Label: "CDS"}).
    WithSelectionBehavior("at-most-one").
    AddSuggestion(suggestion).
    Build()
```

### Action Types

```go
cdshooks.ActionCreate // Create a new resource
cdshooks.ActionUpdate // Update an existing resource
cdshooks.ActionDelete // Delete a resource
```

### Cards with Links

Links provide navigation to external applications:

```go
card, err := cdshooks.NewCard("Open medication review", cdshooks.IndicatorInfo).
    WithSource(cdshooks.Source{Label: "CDS"}).
    AddLink(cdshooks.Link{
        Label: "Open Review App",
        URL:   "https://apps.example.org/med-review?patient={{context.patientId}}",
        Type:  cdshooks.LinkSmart,
    }).
    Build()
```

### Link Types

```go
cdshooks.LinkAbsolute // Opens in new window/tab
cdshooks.LinkSmart   // Uses SMART on FHIR launch
```

### Building Responses

```go
// Return multiple cards
resp := cdshooks.NewResponse().
    AddCard(card1).
    AddCard(card2).
    Build()

// Return system-wide actions
resp := cdshooks.NewResponse().
    AddCard(card).
    WithSystemAction(cdshooks.Action{
        Type:        cdshooks.ActionCreate,
        Description: "Create notification",
        Resource:    resourceJSON,
    }).
    Build()

// Empty response (no cards)
return cdshooks.EmptyResponse(), nil
```

## Working with Prefetch

Prefetch allows the EHR to pre-load data before calling your service.

### Declaring Prefetch Requirements

```go
svc, _ := cdshooks.NewService("my-service").
    ForHook(cdshooks.HookPatientView).
    WithTitle("My Service").
    WithPrefetch("patient", "Patient/{{context.patientId}}").
    WithPrefetch("meds", "MedicationRequest?patient={{context.patientId}}&status=active").
    HandleFunc(handleRequest).
    Build()
```

### Accessing Prefetch Data

```go
func handleRequest(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
    // Decode prefetched patient
    patient, err := req.Prefetch.Patient("patient")
    if err != nil {
        // Handle missing prefetch
        return cdshooks.EmptyResponse(), nil
    }
    
    // Calculate age
    age, err := cdshooks.PatientAge(patient)
    
    // Decode prefetched medications
    meds, err := req.Prefetch.Bundle("meds")
    medCount := cdshooks.BundleEntryCount(meds)
    
    // Check for missing prefetch
    missing := req.Prefetch.Missing(map[string]string{
        "patient": "Patient/{{context.patientId}}",
        "meds":    "MedicationRequest?patient={{context.patientId}}",
    })
    if len(missing) > 0 {
        // Handle missing data
    }
    
    return response, nil
}
```

### Prefetch Helper Functions

| Function | Description |
|----------|-------------|
| `req.Prefetch.Patient(key)` | Decode prefetch into FHIR Patient |
| `req.Prefetch.Bundle(key)` | Decode prefetch into FHIR Bundle |
| `req.Prefetch.Decode(key, &target)` | Decode prefetch into custom type |
| `req.Prefetch.Get(key)` | Get raw JSON for custom parsing |
| `req.Prefetch.Missing(declared)` | Get list of missing prefetch keys |
| `cdshooks.PatientAge(patient)` | Calculate patient age from birthDate |
| `cdshooks.BundleEntryCount(bundle)` | Count entries in a Bundle |

## EHR Client Usage

### Creating a Client

```go
client := cdshooks.NewClient("https://cds.example.org")

// With options
client := cdshooks.NewClient("https://cds.example.org",
    cdshooks.WithTimeout(10*time.Second),
    cdshooks.WithBearerToken("your-token"),
    cdshooks.WithFHIRServer("https://fhir.example.org", auth),
)
```

### Client Options

| Option | Description |
|--------|-------------|
| `WithTimeout(d)` | Set request timeout (default 30s) |
| `WithBearerToken(token)` | Set bearer token for auth |
| `WithFHIRServer(url, auth)` | Set FHIR server URL and auth |
| `WithHTTPClient(httpClient)` | Provide custom HTTP client |

### Discovering Services

```go
services, err := client.Discover(ctx)
for _, svc := range services {
    fmt.Printf("Service: %s (%s)\n", svc.Title, svc.Hook)
}
```

### Calling a Service

```go
resp, err := client.Call(ctx, "patient-view-age-check",
    cdshooks.PatientViewContext{
        UserID:    "Practitioner/abc",
        PatientID: "Patient/123",
    },
    map[string]any{
        "patient": patientResource,
    },
)

for _, card := range resp.Cards {
    fmt.Printf("Card: %s [%s]\n", card.Summary, card.Indicator)
}
```

### Sending Feedback

```go
err := client.Feedback(ctx, "service-id", cdshooks.FeedbackRequest{
    HookInstance: "original-hook-instance",
    Outcomes: []cdshooks.FeedbackOutcome{{
        Outcome: "accepted",
    }},
})
```

## Server Configuration

### Server Options

```go
server := service.NewServer(
    service.WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
    service.WithCORSOrigins("*"),                    // Or specific origins
    service.WithRequestTimeout(5*time.Second),
    service.WithFeedbackHandler(myHandler),
)
```

| Option | Description |
|--------|-------------|
| `WithLogger(l)` | Set structured logger |
| `WithCORSOrigins(origins...)` | Configure allowed CORS origins |
| `WithRequestTimeout(d)` | Set request timeout |
| `WithFeedbackHandler(h)` | Enable feedback endpoint |

### CORS Configuration

```go
// Allow all origins
service.WithCORSOrigins("*")

// Allow specific origins
service.WithCORSOrigins("https://ehr.example.com", "https://app.example.com")
```

### Feedback Handler

Implement the FeedbackHandler interface to receive usage feedback:

```go
type FeedbackHandler interface {
    Feedback(ctx context.Context, serviceID string, feedback cdshooks.FeedbackRequest) error
}

type FeedbackRequest struct {
    HookInstance string
    Outcomes     []FeedbackOutcome
}

type FeedbackOutcome struct {
    Outcome     string
    Description string
}
```

## Error Handling

### SDK Errors

The SDK provides typed errors for common issues:

```go
card, err := cdshooks.NewCard("test", cdshooks.IndicatorInfo).Build()
if err != nil {
    // Handle error - e.g., ErrInvalidCard
}
```

### Server Errors

The server handles:
- Missing hookInstance (400 Bad Request)
- Invalid UUID format (400 Bad Request)
- Service not found (404 Not Found)
- Handler panics (500 Internal Server Error)
- Request timeouts (500 Internal Server Error)

## Project Structure

```
cds-hooks-go/
├── cdshooks/              # Main SDK package
│   ├── card.go             # Card, Source, Suggestion, Action, Link types
│   ├── client.go          # Client for EHRs to call CDS services
│   ├── errors.go           # Typed error types
│   ├── hook.go             # Hook constants and context types
│   ├── prefetch.go         # Prefetch decoding helpers
│   ├── request.go          # CDSRequest type
│   ├── response.go         # CDSResponse and Feedback types
│   ├── service.go          # Service definition and Builder
│   └── service/
│       └── server.go       # HTTP server with CORS support
│
└── example/
    ├── service/            # Example CDS service
    └── client/            # Example EHR client
```

## Testing

```bash
go test ./...
```

## License

Apache 2.0
