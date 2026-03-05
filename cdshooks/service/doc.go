// Package service provides an HTTP server for exposing CDS Hooks services.
// It is built on top of the cdshooks SDK and uses chi for routing.
//
// The service package handles all HTTP concerns including:
//   - Discovery endpoint (/cds-services)
//   - Service invocation (/cds-services/{id})
//   - Feedback endpoint (/cds-services/{id}/feedback)
//   - CORS configuration
//   - Request validation
//   - Logging and error handling
//
// For a complete example, see example/service/main.go.
//
// # Usage
//
//	svc := cdshooks.NewService("patient-view-example").
//	    ForHook(cdshooks.HookPatientView).
//	    WithTitle("Example Service").
//	    HandleFunc(func(ctx context.Context, req cdshooks.CDSRequest) (cdshooks.CDSResponse, error) {
//	        return cdshooks.EmptyResponse(), nil
//	    }).Build()
//
//	server := service.NewServer(
//	    service.WithLogger(slog.New(slog.NewJSONHandler(os.Stdout, nil))),
//	    service.WithCORSOrigins("*"),
//	)
//	server.Register(svc)
//	server.ListenAndServe(":8080")
package service
