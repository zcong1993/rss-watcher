package store

import (
	"context"

	"github.com/pkg/errors"
)

// nolint
var ErrNotFound = errors.New("NotFound")

type Store interface {
	Init(cfg interface{}) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string) error
	Close() error
}

type Component struct {
	Name          string
	FactoryMethod func() Store
}

func New(name string, factoryMethod func() Store) Component {
	return Component{
		Name:          name,
		FactoryMethod: factoryMethod,
	}
}

// Registry is an interface for a component that returns registered state store implementations.
type Registry interface {
	Register(components ...Component)
	Create(name string) (Store, error)
}

type storeRegistry struct {
	stores map[string]func() Store
}

// NewRegistry is used to create state store registry.
func NewRegistry() Registry {
	return &storeRegistry{
		stores: map[string]func() Store{},
	}
}

// // Register registers a new factory method that creates an instance of a StateStore.
// // The key is the name of the state store, eg. redis.
func (s *storeRegistry) Register(components ...Component) {
	for _, component := range components {
		s.stores[component.Name] = component.FactoryMethod
	}
}

func (s *storeRegistry) Create(name string) (Store, error) {
	if method, ok := s.stores[name]; ok {
		return method(), nil
	}

	return nil, errors.Errorf("couldn't find store %s", name)
}
