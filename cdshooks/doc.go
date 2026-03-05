// Package cdshooks provides an idiomatic Go SDK for the CDS Hooks 2.0 specification.
// It handles all protocol concerns so consumers only write clinical logic.
//
// The SDK supports two use cases:
//   - CDS Service Authors: Implement decision support services
//   - CDS Clients (EHRs): Call CDS services from EHR systems
//
// For an HTTP server implementation, see the service subpackage:
//
//	import "github.com/your-org/cds-hooks-go/cdshooks/service"
//
// # Key Types
//
// Service: Defines a CDS service with its hook, title, and prefetch requirements.
// Handler: Interface for implementing CDS service logic.
// Client: HTTP client for calling CDS services from EHR systems.
// Builder: Fluent builders for creating Cards, Suggestions, and Responses.
//
// # Example
//
//	svc := cdshooks.NewService("patient-view-example").
//	    ForHook(cdshooks.HookPatientView).
//	    WithTitle("Example Service").
//	    HandleFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
//	        return cdshooks.EmptyResponse(), nil
//	    }).Build()
package cdshooks
