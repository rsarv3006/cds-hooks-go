// Package cdshooks provides an idiomatic Go SDK for the CDS Hooks 2.0 specification.
// It handles all protocol concerns so consumers only write clinical logic.
//
// The SDK supports two use cases:
//   - CDS Service Authors: Implement decision support services
//   - CDS Clients (EHRs): Call CDS services from EHR systems
//
// # Key Types
//
// Service: Defines a CDS service with its hook, title, and prefetch requirements.
// Handler: Interface for implementing CDS service logic.
// Server: HTTP server for exposing CDS services with CORS support.
// Client: HTTP client for calling CDS services from EHR systems.
//
// # Example
//
//	svc := cdshooks.NewService("patient-view-example").
//	    ForHook(cdshooks.HookPatientView).
//	    WithTitle("Example Service").
//	    HandleFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
//	        return cdshooks.EmptyResponse(), nil
//	    }).Build()
//
//	server := cdshooks.NewServer()
//	server.Register(svc)
//	server.ListenAndServe(":8080")
package cdshooks
