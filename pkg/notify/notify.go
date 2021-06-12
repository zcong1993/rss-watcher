package notify

import (
	"github.com/pkg/errors"
	"github.com/zcong1993/notifiers/v2"
)

type Notifier interface {
	notifiers.Notifier
	Init(cfg interface{}) error
}

type Component struct {
	Name          string
	FactoryMethod func() Notifier
}

func New(name string, factoryMethod func() Notifier) Component {
	return Component{
		Name:          name,
		FactoryMethod: factoryMethod,
	}
}

type Registry interface {
	Register(components ...Component)
	Create(name string) (Notifier, error)
}

type notifiersRegistry struct {
	notifiers map[string]func() Notifier
}

func NewRegistry() Registry {
	return &notifiersRegistry{
		notifiers: map[string]func() Notifier{},
	}
}

func (s *notifiersRegistry) Register(components ...Component) {
	for _, component := range components {
		s.notifiers[component.Name] = component.FactoryMethod
	}
}

func (s *notifiersRegistry) Create(name string) (Notifier, error) {
	if method, ok := s.notifiers[name]; ok {
		return method(), nil
	}

	return nil, errors.Errorf("couldn't find notifier %s", name)
}
