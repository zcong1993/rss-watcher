package runtime

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"sync"
	"syscall"

	"github.com/oklog/run"

	"github.com/dapr/kit/logger"
	"github.com/zcong1993/notifiers/v2"
	"github.com/zcong1993/rss-watcher/pkg/notify"
	"github.com/zcong1993/rss-watcher/pkg/watcher"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/zcong1993/rss-watcher/pkg/store"
	"github.com/zcong1993/rss-watcher/pkg/store/fauna"
	"github.com/zcong1993/rss-watcher/pkg/store/file"
	"github.com/zcong1993/rss-watcher/pkg/store/mem"
)

type RssWatcherRuntime struct {
	storeRegistry     store.Registry
	stores            map[string]store.Store
	notifiersRegistry notify.Registry
	notifiers         map[string]notify.Notifier
	config            *Config
	logger            logger.Logger
	watchers          []*watcher.Watcher
	runtimeOpts       *runtimeOpts
}

func NewRssWatcherRuntime(logger logger.Logger) *RssWatcherRuntime {
	return &RssWatcherRuntime{
		storeRegistry:     store.NewRegistry(),
		stores:            map[string]store.Store{},
		notifiersRegistry: notify.NewRegistry(),
		notifiers:         map[string]notify.Notifier{},
		watchers:          make([]*watcher.Watcher, 0),
		logger:            logger,
	}
}

func (r *RssWatcherRuntime) Run(opts ...Option) error {
	err := r.initRuntime(opts...)
	if err != nil {
		return err
	}

	if r.runtimeOpts.single {
		r.single()

		return nil
	}

	return r.daemon()
}

func (r *RssWatcherRuntime) initRuntime(opts ...Option) error {
	var o runtimeOpts

	for _, opt := range opts {
		opt(&o)
	}

	r.runtimeOpts = &o

	r.storeRegistry.Register(o.stores...)
	r.notifiersRegistry.Register(o.notifiers...)

	err := r.loadConfig(o.configFile)
	if err != nil {
		return err
	}

	err = r.loadStore()
	if err != nil {
		return err
	}

	err = r.loadNotifiers()
	if err != nil {
		return err
	}

	r.initWatchers()

	return nil
}

func (r *RssWatcherRuntime) loadConfig(path string) error {
	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "read config file")
	}

	var config Config
	err = json.Unmarshal(configBytes, &config)

	if err != nil {
		return errors.Wrap(err, "unmarshal config file")
	}

	err = validator.New().Struct(&config)
	if err != nil {
		return errors.Wrap(err, "validate error")
	}

	r.config = &config

	return nil
}

func (r *RssWatcherRuntime) loadStore() error {
	storeType := r.config.KvStore
	s, err := r.storeRegistry.Create(storeType)

	if err != nil {
		return err
	}

	if s != nil {
		var config interface{}

		switch storeType {
		case mem.Name:
			config = nil
		case file.Name:
			config = r.config.FileConfig
			if r.config.FileConfig == nil {
				return errors.New("file_config is required")
			}
		case fauna.Name:
			config = r.config.FaunaConfig
			if r.config.FaunaConfig == nil {
				return errors.New("fauna_config is required")
			}
		}

		err = s.Init(config)

		if err != nil {
			return err
		}

		r.stores[storeType] = s
	}

	return nil
}

func (r *RssWatcherRuntime) loadNotifiers() error {
	usedNotifiers := make(map[string]struct{})

	for _, wc := range r.config.WatcherConfigs {
		for _, v := range wc.Notifiers {
			usedNotifiers[v] = struct{}{}
		}
	}

	for n := range usedNotifiers {
		nn, err := r.notifiersRegistry.Create(n)

		if err != nil {
			return err
		}

		if nn != nil {
			var config interface{}

			switch n {
			case notify.Printer:
				config = nil
			case notify.Mail:
				config = r.config.MailConfig

				if r.config.MailConfig == nil {
					return errors.New("mail_config is required")
				}
			case notify.Ding:
				config = r.config.DingConfig

				if r.config.DingConfig == nil {
					return errors.New("ding_config is required")
				}
			case notify.Telegram:
				config = r.config.TelegramConfig

				if r.config.TelegramConfig == nil {
					return errors.New("telegram_config is required")
				}
			}

			err := nn.Init(config)

			if err != nil {
				return errors.Wrapf(err, "error init notifier: %s", n)
			}

			r.notifiers[n] = nn
		}
	}

	return nil
}

func (r *RssWatcherRuntime) initWatchers() {
	kvStore := r.stores[r.config.KvStore]

	for _, w := range r.config.WatcherConfigs {
		ntfs := make([]notifiers.Notifier, 0)

		for _, t := range w.Notifiers {
			ntf, ok := r.notifiers[t]
			if ok {
				ntfs = append(ntfs, ntf)
			}
		}

		var finalNtfs notifiers.Notifier

		finalNtfs = notifiers.NewCombine(ntfs...)

		limitInterval := r.runtimeOpts.limitInterval

		if limitInterval > 0 {
			r.logger.Infof("use limiter notifiers, duration: %s", limitInterval.String())
			limiter := notifiers.NewLimiter(finalNtfs, limitInterval, 10)
			finalNtfs = limiter

			go func() {
				for err := range limiter.GetErrorCh() {
					r.logger.Errorf("name: %s, error: %s", limiter.GetName(), err.Error())
				}
			}()
		}

		watcherClient := watcher.NewWatcher(r.logger, w.Source, kvStore, finalNtfs, w.Interval.Duration)

		r.watchers = append(r.watchers, watcherClient)
	}
}

func (r *RssWatcherRuntime) single() {
	r.logger.Info("run single and exit")

	if r.runtimeOpts.serialize {
		for _, w := range r.watchers {
			_ = w.Single(context.Background())
		}

		return
	}

	var wg sync.WaitGroup

	for _, w := range r.watchers {
		w := w

		wg.Add(1)

		go func() {
			_ = w.Single(context.Background())

			wg.Done()
		}()
	}

	wg.Wait()

	r.logger.Info("done")

	os.Exit(0)
}

func (r *RssWatcherRuntime) daemon() error {
	var g run.Group

	signals := []os.Signal{syscall.SIGINT, syscall.SIGTERM}
	e, c := run.SignalHandler(context.Background(), signals...)

	g.Add(e, c)

	for _, w := range r.watchers {
		w := w

		g.Add(func() error {
			w.Run()

			return nil
		}, func(err error) {
			if err != nil {
				r.logger.Errorf("exit cause: %s", err.Error())
			}

			w.Close()
		})
	}

	return g.Run()
}
