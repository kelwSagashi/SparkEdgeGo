package providers

import "context"

type Adapter interface {
	Send(ctx context.Context, payload map[string]any) error
	Test(ctx context.Context, payload map[string]any) error
	Discover(ctx context.Context) ([]Resource, error)
}

type Factory func(Config) (Adapter, error)

type Config struct {
	Server      map[string]any
	Resource    map[string]any
	Operation   map[string]any
	Credentials map[string]any
}

type Resource struct {
	Name   string
	Type   string
	Config map[string]any
	Fields []Field
}

type Field struct {
	Name string
	Type string
}

type Registry struct {
	factories map[string]Factory
}

func NewRegistry() *Registry {
	return &Registry{factories: map[string]Factory{}}
}

func (r *Registry) Register(key string, factory Factory) {
	r.factories[key] = factory
}

func (r *Registry) Create(key string, config Config) (Adapter, bool, error) {
	factory, ok := r.factories[key]
	if !ok {
		return nil, false, nil
	}
	adapter, err := factory(config)
	return adapter, true, err
}
