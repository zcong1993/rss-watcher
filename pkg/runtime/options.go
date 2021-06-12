package runtime

import (
	"time"

	"github.com/zcong1993/rss-watcher/pkg/notify"
	"github.com/zcong1993/rss-watcher/pkg/store"
)

type (
	// runtimeOpts encapsulates the components to include in the runtime.
	runtimeOpts struct {
		configFile    string
		stores        []store.Component
		notifiers     []notify.Component
		limitInterval time.Duration
		single        bool
		serialize     bool
	}

	// Option is a function that customizes the runtime.
	Option func(o *runtimeOpts)
)

func WithStores(stores ...store.Component) Option {
	return func(o *runtimeOpts) {
		o.stores = append(o.stores, stores...)
	}
}

func WithNotifiers(notifiers ...notify.Component) Option {
	return func(o *runtimeOpts) {
		o.notifiers = append(o.notifiers, notifiers...)
	}
}

func WithConfigFile(path string) Option {
	return func(o *runtimeOpts) {
		o.configFile = path
	}
}

func WithSingle(single bool) Option {
	return func(o *runtimeOpts) {
		o.single = single
	}
}

func WithLimitInterval(limitInterval time.Duration) Option {
	return func(o *runtimeOpts) {
		o.limitInterval = limitInterval
	}
}

func WithSerialize(serialize bool) Option {
	return func(o *runtimeOpts) {
		o.serialize = serialize
	}
}
