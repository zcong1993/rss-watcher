package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"syscall"

	"github.com/zcong1993/notifiers/adapters/printer"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/oklog/run"

	"github.com/zcong1993/notifiers/adapters/tg"

	"github.com/zcong1993/notifiers/adapters/ding"
	"github.com/zcong1993/notifiers/adapters/mail"
	"github.com/zcong1993/notifiers/types"
	"github.com/zcong1993/rss-watcher/pkg/config"
	"github.com/zcong1993/rss-watcher/pkg/kv"
	"github.com/zcong1993/rss-watcher/pkg/watcher"
)

var (
	version = "master"
	commit  = ""
	date    = ""
	builtBy = ""
)

func main() {
	versionFlag := flag.Bool("v", false, "Show version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(buildVersion(version, commit, date, builtBy))
		os.Exit(0)
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "./config.json"
	}

	cfg, err := config.LoadConfigFromFile(configFile)

	if err != nil {
		level.Error(logger).Log("msg", "load config", "error", err.Error())
		os.Exit(1)
	}

	level.Info(logger).Log("msg", fmt.Sprintf("use kv store: %s", cfg.KvStore))

	var (
		kvStore      kv.Store
		dingNotifier types.Notifier
		mailNotifier types.Notifier
		tgNotifier   types.Notifier
	)
	switch cfg.KvStore {
	case "mem":
		kvStore = kv.NewMemStore()
	case "file":
		kvStore = kv.NewFileStore(cfg.FileStoreConfigPath)
	case "fauna":
		kvStore, err = kv.NewFanua(cfg.FaunaConfig)
		if err != nil {
			level.Error(logger).Log("msg", "init fauna", "error", err.Error())
			os.Exit(1)
		}
	}

	if cfg.DingConfig != nil {
		dd := ding.NewClient(cfg.DingConfig.Token)
		dd.SetLogger(log.WithPrefix(logger, "component", "dingding"))
		dingNotifier = dd
	}
	if cfg.MailConfig != nil {
		mc := mail.NewClient(cfg.MailConfig.Domain, cfg.MailConfig.PrivateKey, cfg.MailConfig.To, cfg.MailConfig.From)
		mc.SetLogger(log.WithPrefix(logger, "component", "mailgun"))
		mailNotifier = mc
	}
	if cfg.TelegramConfig != nil {
		tgNotifier, err = tg.NewClient(cfg.TelegramConfig.Token, cfg.TelegramConfig.ToID, log.WithPrefix(logger, "component", "tg"))
		if err != nil {
			level.Error(logger).Log("msg", "init tg", "error", err.Error())
			os.Exit(1)
		}
	}

	// run single
	if cfg.Single {
		level.Info(logger).Log("msg", "run single and exit")
		wg := sync.WaitGroup{}
		for _, rw := range cfg.WatcherConfigs {
			notifiers := make([]types.Notifier, 0)
			for _, t := range rw.Notifiers {
				switch t {
				case "mail":
					notifiers = append(notifiers, mailNotifier)
				case "ding":
					notifiers = append(notifiers, dingNotifier)
				case "tg":
					notifiers = append(notifiers, tgNotifier)
				case "printer":
					notifiers = append(notifiers, printer.NewPrinter(os.Stderr))
				}
			}

			rw := rw

			wg.Add(1)
			go func() {
				watcherClient := watcher.NewRSSWatcher(logger, rw.Source, rw.Interval.Duration, kvStore, notifiers, rw.Skip)
				_ = watcherClient.Single(context.Background())
				wg.Done()
			}()
		}

		wg.Wait()
		level.Info(logger).Log("msg", "done")
		os.Exit(0)
	} else {
		// run as daemon
		var g run.Group

		for _, rw := range cfg.WatcherConfigs {
			notifiers := make([]types.Notifier, 0)
			for _, t := range rw.Notifiers {
				switch t {
				case "mail":
					notifiers = append(notifiers, mailNotifier)
				case "ding":
					notifiers = append(notifiers, dingNotifier)
				case "tg":
					notifiers = append(notifiers, tgNotifier)
				}
			}

			rw := rw

			watcherClient := watcher.NewRSSWatcher(logger, rw.Source, rw.Interval.Duration, kvStore, notifiers, rw.Skip)
			g.Add(func() error {
				watcherClient.Run(context.Background())
				return nil
			}, func(err error) {
				if err != nil {
					level.Error(logger).Log("msg", fmt.Sprintf("exit cause: %s", err))
				}
				watcherClient.Close()
			})
		}

		e, i := run.SignalHandler(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		g.Add(e, i)

		err := g.Run()
		if err != nil {
			level.Error(logger).Log("msg", "running failed", "err", err)
			os.Exit(1)
		}
		level.Info(logger).Log("msg", "exiting")
	}
}

func buildVersion(version, commit, date, builtBy string) string {
	var result = version
	if commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, commit)
	}
	if date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, date)
	}
	if builtBy != "" {
		result = fmt.Sprintf("%s\nbuilt by: %s", result, builtBy)
	}
	return result
}
