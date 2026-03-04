package cdshooks

import "context"

type Service struct {
	ID                string
	Hook              Hook
	Title             *string
	Description       string
	Prefetch          map[string]string
	UsageRequirements string
	Extension         map[string]any
}

type Handler interface {
	Handle(ctx context.Context, req CDSRequest) (CDSResponse, error)
}

type HandlerFunc func(ctx context.Context, req CDSRequest) (CDSResponse, error)

func (f HandlerFunc) Handle(ctx context.Context, req CDSRequest) (CDSResponse, error) {
	return f(ctx, req)
}

type ServiceEntry struct {
	Service Service
	Handler Handler
}

type ServiceBuilder struct {
	service Service
	handler Handler
	err     error
}

func NewService(id string) *ServiceBuilder {
	return &ServiceBuilder{
		service: Service{
			ID:       id,
			Prefetch: make(map[string]string),
		},
	}
}

func (b *ServiceBuilder) ForHook(hook Hook) *ServiceBuilder {
	if b.err != nil {
		return b
	}
	b.service.Hook = hook
	return b
}

func (b *ServiceBuilder) WithTitle(title string) *ServiceBuilder {
	if b.err != nil {
		return b
	}
	b.service.Title = &title
	return b
}

func (b *ServiceBuilder) WithDescription(desc string) *ServiceBuilder {
	if b.err != nil {
		return b
	}
	b.service.Description = desc
	return b
}

func (b *ServiceBuilder) WithPrefetch(key, query string) *ServiceBuilder {
	if b.err != nil {
		return b
	}
	b.service.Prefetch[key] = query
	return b
}

func (b *ServiceBuilder) WithUsageRequirements(req string) *ServiceBuilder {
	if b.err != nil {
		return b
	}
	b.service.UsageRequirements = req
	return b
}

func (b *ServiceBuilder) Handle(handler Handler) *ServiceBuilder {
	if b.err != nil {
		return b
	}
	b.handler = handler
	return b
}

func (b *ServiceBuilder) HandleFunc(fn HandlerFunc) *ServiceBuilder {
	if b.err != nil {
		return b
	}
	b.handler = fn
	return b
}

func (b *ServiceBuilder) Build() (ServiceEntry, error) {
	if b.err != nil {
		return ServiceEntry{}, b.err
	}

	if b.service.ID == "" {
		return ServiceEntry{}, &ErrInvalidCard{Field: "Service.ID", Reason: "required"}
	}

	if b.service.Hook == "" {
		return ServiceEntry{}, &ErrInvalidCard{Field: "Service.Hook", Reason: "required"}
	}

	if b.service.Title == nil || *b.service.Title == "" {
		return ServiceEntry{}, &ErrInvalidCard{Field: "Service.Title", Reason: "required"}
	}

	if b.handler == nil {
		return ServiceEntry{}, &ErrInvalidCard{Field: "Service.Handler", Reason: "required"}
	}

	return ServiceEntry{
		Service: b.service,
		Handler: b.handler,
	}, nil
}
