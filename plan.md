# CDS Hooks Go SDK — Design Plan

## Goal

A idiomatic, publishable Go SDK for the [CDS Hooks 2.0 specification](https://cds-hooks.org/specification/current/).
Intended to be used both by **CDS service authors** (clinical decision support providers) and
**CDS clients** (EHRs or middleware calling CDS services). The SDK handles all protocol concerns
so consumers only write clinical logic.

**Target import path:** `github.com/your-org/cds-hooks-go`  
**Go version:** 1.21+  
**No generated code. No external framework dependencies. Standard library + chi for HTTP.**

---

## Package Structure

```
cds-hooks-go/
├── go.mod
├── go.sum
├── README.md
│
├── cdshooks/                  ← primary public package
│   ├── doc.go                 — package doc + overview
│   ├── errors.go              — typed error types
│   ├── hook.go                — hook name constants + typed context structs
│   ├── card.go                — Card, Source, Suggestion, Link, SystemAction + fluent builders
│   ├── request.go             — CDSRequest, typed prefetch access
│   ├── response.go            — CDSResponse, FeedbackRequest/Response
│   ├── prefetch.go            — Prefetch map + FHIR resource helpers
│   ├── service.go             — Service definition, ServiceHandler interface, ServiceBuilder
│   ├── server.go              — HTTP server: discovery + dispatch
│   └── client.go              — HTTP client for EHRs calling CDS services
│
├── fhir/                      ← thin FHIR R4 type projections (no full model dependency)
│   ├── doc.go
│   ├── patient.go             — Patient projection
│   ├── bundle.go              — Bundle + entry helpers
│   ├── medication_request.go  — MedicationRequest projection
│   ├── allergy_intolerance.go — AllergyIntolerance projection
│   └── coding.go              — CodeableConcept, Coding, Identifier
│
└── example/
    ├── service/               — runnable example CDS service
    │   └── main.go
    └── client/                — runnable example EHR client
        └── main.go
```

---

## Package: `cdshooks`

### `hook.go` — Hook Constants & Typed Contexts

Define string constants for all standard CDS Hooks hook names.

```go
const (
    HookPatientView   Hook = "patient-view"
    HookOrderSelect   Hook = "order-select"
    HookOrderSign     Hook = "order-sign"
    HookAppointmentBook Hook = "appointment-book"
    HookEncounterStart  Hook = "encounter-start"
    HookEncounterDischarge Hook = "encounter-discharge"
)
```

Define a typed context struct for **each** standard hook so callers get compile-time safety
instead of `map[string]interface{}` fumbling:

```go
type PatientViewContext struct {
    UserID      string // FHIR Practitioner or Patient id
    PatientID   string
    EncounterID string // optional
}

type OrderSelectContext struct {
    UserID      string
    PatientID   string
    EncounterID string
    Selections  []string         // draft order resource ids selected by clinician
    DraftOrders fhir.Bundle      // FHIR Bundle of draft MedicationRequest / ServiceRequest etc.
}

type OrderSignContext struct {
    UserID      string
    PatientID   string
    EncounterID string
    DraftOrders fhir.Bundle
}

type AppointmentBookContext struct {
    UserID        string
    PatientID     string
    EncounterID   string
    Appointments  fhir.Bundle
}

type EncounterStartContext struct {
    UserID      string
    PatientID   string
    EncounterID string
}

type EncounterDischargeContext struct {
    UserID      string
    PatientID   string
    EncounterID string
}
```

Each typed context struct implements a private `hookContext` interface so the server can
marshal/unmarshal them without reflection on the consumer side.

Also provide `ParseContext[T any](raw json.RawMessage) (T, error)` — a generic helper for
decoding raw context JSON into any of the above structs.

---

### `card.go` — Cards, Sources, Suggestions, Links, System Actions

The `Card` struct must be complete per spec. Provide a fluent builder so consumers never
have to deal with pointer fields directly.

**Types:**

```go
type CardIndicator string
const (
    IndicatorInfo     CardIndicator = "info"
    IndicatorWarning  CardIndicator = "warning"
    IndicatorCritical CardIndicator = "critical"
    IndicatorSuccess  CardIndicator = "success"  // CDS Hooks 2.0 addition
)

type Card struct {
    UUID           string         // auto-generated if not set
    Summary        string         // required, max 140 chars
    Detail         string         // optional, markdown
    Indicator      CardIndicator  // required
    Source         Source         // required
    Suggestions    []Suggestion
    SelectionBehavior string      // "at-most-one" | "any" — required if Suggestions non-empty
    OverrideReasons []OverrideReason
    Links          []Link
}

type Source struct {
    Label    string
    URL      string
    Icon     string  // absolute URL to 100x100 img
    Topic    Coding  // fhir Coding
}

type Suggestion struct {
    Label   string
    UUID    string   // auto-generated if not set
    IsRecommended bool
    Actions []Action
}

type Action struct {
    Type        ActionType      // "create" | "update" | "delete"
    Description string
    Resource    json.RawMessage // FHIR resource
    ResourceID  string          // for delete actions
}

type ActionType string
const (
    ActionCreate ActionType = "create"
    ActionUpdate ActionType = "update"
    ActionDelete ActionType = "delete"
)

type Link struct {
    Label    string
    URL      string
    Type     LinkType  // "absolute" | "smart"
    AppContext string  // only for smart links
}

type LinkType string
const (
    LinkAbsolute LinkType = "absolute"
    LinkSmart    LinkType = "smart"
)

type OverrideReason struct {
    ReasonCode Coding
}
```

**Fluent builder:**

```go
// Usage:
card := cdshooks.NewCard("Medication review recommended", cdshooks.IndicatorWarning).
    WithSource(cdshooks.Source{Label: "STOPP/START Criteria", URL: "https://..."}).
    WithDetail("Patient is 67 years old with 6 active medications...").
    AddSuggestion(
        cdshooks.NewSuggestion("Deprescribe omeprazole").
            AddAction(cdshooks.NewDeleteAction("MedicationRequest/789", "Remove PPI")),
    ).
    AddLink(cdshooks.Link{
        Label: "Open Medication Review App",
        URL:   "https://apps.example.org/med-review",
        Type:  cdshooks.LinkSmart,
    }).
    Build()
```

`Build()` validates the card (summary length, required fields, SelectionBehavior present when
Suggestions non-empty) and returns `(Card, error)`.

Also provide `MustBuild()` which panics on invalid input — useful for static cards defined
at startup.

---

### `prefetch.go` — Prefetch Decoding

The `Prefetch` type wraps the raw JSON map from the CDS request and provides typed accessors
so handlers don't have to do their own unmarshalling.

```go
type Prefetch struct {
    raw map[string]json.RawMessage
}

// Get returns the raw JSON for a prefetch key.
func (p Prefetch) Get(key string) (json.RawMessage, bool)

// Decode unmarshals a prefetch key into any target type.
func (p Prefetch) Decode(key string, target any) error

// Patient decodes the prefetch key into a fhir.Patient.
func (p Prefetch) Patient(key string) (fhir.Patient, error)

// Bundle decodes the prefetch key into a fhir.Bundle.
func (p Prefetch) Bundle(key string) (fhir.Bundle, error)

// Missing returns all keys that are declared but absent from the prefetch map.
// Useful for deciding whether to fall back to a FHIR server call.
func (p Prefetch) Missing(declared map[string]string) []string
```

---

### `request.go` — CDSRequest

```go
type CDSRequest struct {
    Hook         string          `json:"hook"`
    HookInstance string          `json:"hookInstance"`  // UUID, unique per invocation
    FHIRServer   string          `json:"fhirServer"`    // base URL, may be empty
    FHIRAuth     *FHIRAuth       `json:"fhirAuthorization"` // Bearer token if SMART
    Context      json.RawMessage `json:"context"`       // decoded per hook type by handler
    Prefetch     Prefetch        `json:"prefetch"`
}

type FHIRAuth struct {
    AccessToken string `json:"access_token"`
    TokenType   string `json:"token_type"`
    ExpiresIn   int    `json:"expires_in"`
    Scope       string `json:"scope"`
    Subject     string `json:"subject"`
}
```

Provide `request.DecodeContext[T any]() (T, error)` — generic helper that unmarshals
`Context` into the appropriate typed context struct.

---

### `response.go` — CDSResponse & Feedback

```go
type CDSResponse struct {
    Cards         []Card         `json:"cards"`          // required, may be empty array
    SystemActions []Action       `json:"systemActions"`  // optional
}

// FeedbackRequest is sent by the EHR back to the CDS service after the clinician acts.
type FeedbackRequest struct {
    Card          string         `json:"card"`           // card UUID
    Outcome       FeedbackOutcome `json:"outcome"`
    AcceptedSuggestions []AcceptedSuggestion `json:"acceptedSuggestions"`
    OverrideReason *OverrideReason `json:"overrideReason"`
    OutcomeTimestamp string      `json:"outcomeTimestamp"` // ISO 8601
}

type FeedbackOutcome string
const (
    OutcomeAccepted     FeedbackOutcome = "accepted"
    OutcomeOverridden   FeedbackOutcome = "overridden"
    OutcomeNoActionTaken FeedbackOutcome = "noActionTaken"
)

type AcceptedSuggestion struct {
    ID string `json:"id"` // suggestion UUID
}
```

---

### `service.go` — Service Definition & Handler Interface

A `Service` is the unit of registration — one per CDS hook endpoint.

```go
type Service struct {
    ID                string
    Hook              Hook
    Title             string
    Description       string
    Prefetch          map[string]string // FHIR query templates
    UsageRequirements string            // optional plain text
}

// Handler is what the SDK consumer implements — just this one method.
type Handler interface {
    Handle(ctx context.Context, req CDSRequest) (CDSResponse, error)
}

// HandlerFunc is a function adapter for Handler.
type HandlerFunc func(ctx context.Context, req CDSRequest) (CDSResponse, error)

// ServiceEntry binds a Service definition to its Handler.
type ServiceEntry struct {
    Service Service
    Handler Handler
}
```

**ServiceBuilder** — fluent API for constructing a ServiceEntry:

```go
// Usage:
entry := cdshooks.NewService("patient-view-age-check").
    ForHook(cdshooks.HookPatientView).
    WithTitle("Patient Age Medication Review").
    WithDescription("Flags patients 65+ for STOPP/START medication review.").
    WithPrefetch("patient",    "Patient/{{context.patientId}}").
    WithPrefetch("medications","MedicationRequest?subject={{context.patientId}}&status=active").
    HandleFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
        var hookCtx cdshooks.PatientViewContext
        if err := req.DecodeContext(&hookCtx); err != nil {
            return cdshooks.CDSResponse{}, err
        }

        patient, err := req.Prefetch.Patient("patient")
        if err != nil {
            return cdshooks.CDSResponse{}, err
        }

        age := patient.Age()
        if age < 65 {
            return cdshooks.EmptyResponse(), nil
        }

        card, err := cdshooks.NewCard(
            fmt.Sprintf("Medication review recommended — patient aged %d", age),
            cdshooks.IndicatorWarning,
        ).
            WithSource(cdshooks.Source{Label: "STOPP/START Criteria"}).
            Build()
        if err != nil {
            return cdshooks.CDSResponse{}, err
        }

        return cdshooks.NewResponse().AddCard(card).Build(), nil
    }).
    Build()
```

`Build()` validates that ID, Hook, Title, and Handler are all set.

Also provide `cdshooks.EmptyResponse()` as a convenience for returning no cards.

---

### `server.go` — HTTP Server

The server handles all CDS Hooks HTTP protocol concerns. The consumer only registers
services and starts listening.

```go
type Server struct {
    // unexported fields: registry, logger, options
}

type ServerOption func(*Server)

func WithLogger(l *slog.Logger) ServerOption
func WithCORSOrigins(origins ...string) ServerOption        // CDS Hooks requires CORS
func WithRequestTimeout(d time.Duration) ServerOption
func WithFeedbackHandler(h FeedbackHandler) ServerOption    // optional feedback endpoint

func NewServer(opts ...ServerOption) *Server

func (s *Server) Register(entries ...ServiceEntry) *Server

// Handler returns the http.Handler — for embedding in an existing router.
func (s *Server) Handler() http.Handler

// ListenAndServe starts the server with graceful shutdown on SIGINT/SIGTERM.
func (s *Server) ListenAndServe(addr string) error
```

**Endpoints implemented by the server:**

| Method | Path                          | Description                               |
| ------ | ----------------------------- | ----------------------------------------- |
| `GET`  | `/cds-services`               | Discovery — lists all registered services |
| `POST` | `/cds-services/{id}`          | Hook invocation — dispatches to handler   |
| `POST` | `/cds-services/{id}/feedback` | Feedback — if FeedbackHandler registered  |

**Server behaviour:**

- Returns `Content-Type: application/json` on all responses
- Sets CORS headers (`Access-Control-Allow-Origin`, `Access-Control-Allow-Headers`) — required by the spec for browser-based EHRs
- Validates `hookInstance` UUID is present on invocation requests
- Ensures `cards` in response is always a JSON array, never `null`
- Returns structured JSON errors (not plain text) with appropriate HTTP status codes
- Logs each request with hook name, service id, card count, and latency using `slog`
- Recovers from handler panics and returns 500 rather than crashing

**FeedbackHandler interface:**

```go
type FeedbackHandler interface {
    Feedback(ctx context.Context, serviceID string, feedback FeedbackRequest) error
}
```

---

### `client.go` — CDS Client (EHR Side)

For EHRs or middleware that need to call CDS services.

```go
type Client struct {
    // unexported: baseURL, httpClient, auth
}

type ClientOption func(*Client)

func WithHTTPClient(c *http.Client) ClientOption
func WithBearerToken(token string) ClientOption
func WithFHIRServer(baseURL string, auth *FHIRAuth) ClientOption
func WithTimeout(d time.Duration) ClientOption

func NewClient(baseURL string, opts ...ClientOption) *Client

// Discover fetches the /cds-services discovery document.
func (c *Client) Discover(ctx context.Context) ([]Service, error)

// Call invokes a specific CDS service.
// hookContext must be one of the typed context structs (PatientViewContext, etc.)
// or any JSON-serialisable struct for custom hooks.
func (c *Client) Call(
    ctx       context.Context,
    serviceID string,
    hookCtx   any,
    prefetch  map[string]any,
) (CDSResponse, error)

// Feedback sends clinician outcome feedback to a CDS service.
func (c *Client) Feedback(
    ctx       context.Context,
    serviceID string,
    feedback  FeedbackRequest,
) error
```

---

### `errors.go` — Typed Errors

```go
// ErrUnknownService is returned by the server when no handler is registered for the id.
type ErrUnknownService struct{ ID string }

// ErrMissingPrefetch is returned when a required prefetch key is absent and
// no FHIR server fallback is configured.
type ErrMissingPrefetch struct{ Key string }

// ErrInvalidContext is returned when context JSON cannot be decoded into the expected type.
type ErrInvalidContext struct{ Hook Hook; Cause error }

// ErrInvalidCard is returned by Card.Build() when validation fails.
type ErrInvalidCard struct{ Field string; Reason string }

// ErrFHIRRequest is returned when a FHIR server call fails.
type ErrFHIRRequest struct{ URL string; StatusCode int; Body string }
```

All error types implement `error` and `errors.Is`/`errors.As` correctly.

---

## Package: `fhir`

Thin, hand-written FHIR R4 projections. **Not** a full FHIR model library — only the
fields relevant to CDS Hooks prefetch decoding. The goal is zero mandatory dependencies
on heavy FHIR libraries, while still giving consumers useful typed access.

### `patient.go`

```go
type Patient struct {
    ResourceType string
    ID           string
    Active       *bool
    Name         []HumanName
    Gender       string         // "male" | "female" | "other" | "unknown"
    BirthDate    string         // "YYYY-MM-DD"
    Deceased     *bool
}

// Age returns the patient's age in whole years calculated from BirthDate.
// Returns 0 and an error if BirthDate is absent or unparseable.
func (p Patient) Age() (int, error)

// DisplayName returns the first official name as "Family, Given" or falls back
// to the first name of any use.
func (p Patient) DisplayName() string
```

### `bundle.go`

```go
type Bundle struct {
    ResourceType string
    ID           string
    Type         string   // "searchset" | "collection" etc.
    Total        *int
    Entry        []BundleEntry
}

type BundleEntry struct {
    FullURL  string
    Resource json.RawMessage
}

// Resources returns all entry resources decoded into T.
// e.g. bundle.Resources[fhir.MedicationRequest]()
func Resources[T any](b Bundle) ([]T, error)

// Count returns len(Entry) — useful when you only need the count, not the resources.
func (b Bundle) Count() int
```

### `medication_request.go`

```go
type MedicationRequest struct {
    ResourceType string
    ID           string
    Status       string   // "active" | "stopped" | "cancelled" etc.
    Intent       string
    MedicationCodeableConcept *CodeableConcept
    Subject      Reference
    AuthoredOn   string
}
```

### `allergy_intolerance.go`

```go
type AllergyIntolerance struct {
    ResourceType    string
    ID              string
    ClinicalStatus  CodeableConcept
    Code            CodeableConcept
    Patient         Reference
    Reaction        []AllergyReaction
}

type AllergyReaction struct {
    Substance    *CodeableConcept
    Manifestation []CodeableConcept
    Severity     string  // "mild" | "moderate" | "severe"
}
```

### `coding.go`

```go
type CodeableConcept struct {
    Coding []Coding
    Text   string
}

type Coding struct {
    System  string
    Version string
    Code    string
    Display string
}

type Reference struct {
    Reference string
    Display   string
}

type HumanName struct {
    Use    string
    Family string
    Given  []string
}

// HasCode returns true if any Coding in the concept matches the given system+code.
func (c CodeableConcept) HasCode(system, code string) bool

// DisplayText returns the first non-empty Display from Coding, falling back to Text.
func (c CodeableConcept) DisplayText() string
```

---

## Dependencies

| Dependency                 | Purpose                                 | Justification                             |
| -------------------------- | --------------------------------------- | ----------------------------------------- |
| `github.com/go-chi/chi/v5` | HTTP routing                            | Lightweight, stdlib-compatible, idiomatic |
| `github.com/google/uuid`   | UUID generation for card/suggestion IDs | Spec requires UUIDs                       |

No FHIR library dependency. No codegen runtime dependency. No ORM. No framework.

`go.mod` minimum Go version: **1.21** (for `log/slog` and generics stabilisation).

---

## Validation Rules

Enforce these at `Build()` time and on incoming requests in the server:

| Rule                                                           | Where enforced           |
| -------------------------------------------------------------- | ------------------------ |
| `Card.Summary` ≤ 140 characters                                | `Card.Build()`           |
| `Card.Indicator` must be a known value                         | `Card.Build()`           |
| `Card.SelectionBehavior` required when `Suggestions` non-empty | `Card.Build()`           |
| `Card.Source.Label` required                                   | `Card.Build()`           |
| `Suggestion.Label` required                                    | `Suggestion.Build()`     |
| `Link.Type` must be `"absolute"` or `"smart"`                  | `Link` validation        |
| `Link.AppContext` only allowed when `Type == "smart"`          | `Link` validation        |
| `Service.ID` must be URL-safe (no spaces, slashes)             | `ServiceBuilder.Build()` |
| `CDSRequest.HookInstance` must be a valid UUID                 | server middleware        |
| `CDSResponse.Cards` must be a JSON array (not null)            | server response writer   |

---

## Testing Strategy

All packages should have unit tests. Key areas:

- **`card_test.go`** — builder validation: summary too long, missing source, SelectionBehavior rules
- **`prefetch_test.go`** — decode Patient, Bundle, missing key handling
- **`request_test.go`** — context decoding for each hook type; malformed JSON handling
- **`server_test.go`** — discovery endpoint shape, dispatch to correct handler, unknown service 404, panic recovery, CORS headers
- **`client_test.go`** — Discover, Call, Feedback against an `httptest.Server`
- **`fhir/patient_test.go`** — Age calculation including leap years and year-boundary cases
- **`fhir/bundle_test.go`** — `Resources[T]` generic decoding, empty bundle

Use `net/http/httptest` throughout. No test framework beyond `testing` + `github.com/stretchr/testify`.

---

## Example Usage (Consumer-Facing API)

### Authoring a CDS Service

```go
package main

import (
    "context"
    "fmt"
    "log/slog"
    "os"

    "github.com/your-org/cds-hooks-go/cdshooks"
    "github.com/your-org/cds-hooks-go/fhir"
)

func main() {
    ageCheck := cdshooks.NewService("patient-view-age-check").
        ForHook(cdshooks.HookPatientView).
        WithTitle("Patient Age Medication Review").
        WithDescription("Flags patients 65+ for STOPP/START criteria review.").
        WithPrefetch("patient", "Patient/{{context.patientId}}").
        WithPrefetch("meds", "MedicationRequest?subject={{context.patientId}}&status=active").
        HandleFunc(handlePatientView).
        Build()

    server := cdshooks.NewServer(
        cdshooks.WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
        cdshooks.WithCORSOrigins("*"),
        cdshooks.WithRequestTimeout(5 * time.Second),
    )

    server.Register(ageCheck)

    if err := server.ListenAndServe(":8080"); err != nil {
        slog.Error("server failed", "err", err)
        os.Exit(1)
    }
}

func handlePatientView(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
    var hookCtx cdshooks.PatientViewContext
    if err := req.DecodeContext(&hookCtx); err != nil {
        return cdshooks.CDSResponse{}, err
    }

    patient, err := req.Prefetch.Patient("patient")
    if err != nil {
        // Prefetch missing — could fall back to req.FHIRServer here
        return cdshooks.EmptyResponse(), nil
    }

    age, err := patient.Age()
    if err != nil || age < 65 {
        return cdshooks.EmptyResponse(), nil
    }

    meds, _ := req.Prefetch.Bundle("meds")
    medCount := meds.Count()

    indicator := cdshooks.IndicatorInfo
    if medCount >= 5 {
        indicator = cdshooks.IndicatorWarning
    }

    card, err := cdshooks.NewCard(
        fmt.Sprintf("Medication review recommended — patient aged %d (%d active medications)", age, medCount),
        indicator,
    ).
        WithSource(cdshooks.Source{
            Label: "STOPP/START Criteria v3",
            URL:   "https://www.ncbi.nlm.nih.gov/pmc/articles/PMC4339726/",
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
```

### Calling a CDS Service (EHR Side)

```go
client := cdshooks.NewClient("https://cds.example.org",
    cdshooks.WithTimeout(3 * time.Second),
    cdshooks.WithBearerToken(ehrToken),
)

// Discover available services
services, err := client.Discover(ctx)

// Invoke a hook
response, err := client.Call(ctx, "patient-view-age-check",
    cdshooks.PatientViewContext{
        UserID:    "Practitioner/abc",
        PatientID: "Patient/123",
    },
    map[string]any{
        "patient": patientResource, // pre-fetched FHIR Patient
    },
)

for _, card := range response.Cards {
    fmt.Println(card.Summary, card.Indicator)
}
```

---

## What to Tell the AI Tool

When handing this to another AI to implement, include these constraints:

1. **Idiomatic Go** — no `interface{}` where generics or typed structs can be used; unexported fields with exported methods; constructors over struct literals
2. **No magic** — no `reflect` usage except where unavoidable (JSON decoding); no `init()` side effects
3. **Errors are values** — always return typed errors from the `errors.go` file; never `fmt.Errorf` at the boundary
4. **The `fhir` package is projection-only** — do not pull in `google/fhir` or `samply/golang-fhir-models`; hand-write only the fields listed
5. **The server must set CORS headers** — the spec explicitly requires it for browser-based EHRs; `Access-Control-Allow-Origin: *` and `Access-Control-Allow-Headers: Content-Type, X-Requested-With` at minimum
6. **Cards array must never be JSON null** — initialise as `[]Card{}` not `nil` before encoding
7. **UUIDs** — use `github.com/google/uuid` to auto-generate `Card.UUID` and `Suggestion.UUID` in `Build()` if not supplied by the caller
8. **Graceful shutdown** — `ListenAndServe` must listen for `SIGINT`/`SIGTERM` and call `http.Server.Shutdown` with a 15s context
9. **All exported types get godoc comments** — every exported symbol must have a Go doc comment
10. **Tests use `net/http/httptest`** — no integration test dependencies; mock the FHIR server with `httptest.NewServer`
