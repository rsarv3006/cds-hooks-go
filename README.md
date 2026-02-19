# CDS Hooks Go SDK

An idiomatic Go SDK for the [CDS Hooks 2.0 specification](https://cds-hooks.org/).

## Overview

This SDK handles all protocol concerns so consumers only write clinical logic. It supports:

- **CDS Service Authors**: Implement decision support services
- **CDS Clients (EHRs)**: Call CDS services from EHR systems

## Requirements

- Go 1.21+
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

## Project Structure

```
cds-hooks-go/
├── cdshooks/          # Main SDK package
│   ├── card.go        # Card, Source, Suggestion, Action, Link types
│   ├── client.go      # Client for EHRs to call CDS services
│   ├── errors.go      # Typed error types
│   ├── hook.go        # Hook constants and context types
│   ├── prefetch.go    # Prefetch decoding helpers
│   ├── request.go     # CDSRequest type
│   ├── response.go    # CDSResponse and Feedback types
│   ├── service.go     # Service definition and Builder
│   └── server.go      # HTTP server with CORS support
│
├── fhir/              # Thin FHIR R4 projections
│   ├── patient.go     # Patient with Age(), DisplayName()
│   ├── bundle.go      # Bundle with Resources[T] decoder
│   ├── coding.go      # CodeableConcept, Coding, Reference
│   ├── medication_request.go
│   └── allergy_intolerance.go
│
└── example/
    ├── service/       # Example CDS service
    └── client/       # Example EHR client
```

## Testing

```bash
go test ./...
```

## License

Apache 2.0
