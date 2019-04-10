package main

import (
	"os"

	"github.com/zcong1993/notifiers/adapters/ding"
	"github.com/zcong1993/notifiers/adapters/mail"
	"github.com/zcong1993/notifiers/types"
	"github.com/zcong1993/rss-watcher/config"
	"github.com/zcong1993/rss-watcher/kv"
	"github.com/zcong1993/rss-watcher/watcher"
)

func main() {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "./config.json"
	}
	cfg := config.LoadConfigFromFile(configFile)
	var (
		kvStore      kv.Store
		dingNotifier types.Notifier
		mailNotifier types.Notifier
	)
	kvStore = kv.NewFireStore(cfg.FireStoreConfig.ProjectID, cfg.FireStoreConfig.Collection)
	if cfg.DingConfig != nil {
		dingNotifier = ding.NewClient(cfg.DingConfig.Token)
	}
	if cfg.MailConfig != nil {
		mailNotifier = mail.NewClient(cfg.MailConfig.Domain, cfg.MailConfig.PrivateKey, cfg.MailConfig.To, cfg.MailConfig.From)
	}

	for _, rw := range cfg.WatcherConfigs {
		notifiers := make([]types.Notifier, 0)
		for _, t := range rw.Notifiers {
			switch t {
			case "mail":
				notifiers = append(notifiers, mailNotifier)
			case "ding":
				notifiers = append(notifiers, dingNotifier)
			}
		}

		rw := rw

		go func() {
			watcherClient := watcher.NewRSSWatcher(rw.Source, rw.Interval.Duration, kvStore, notifiers, rw.Skip)
			watcherClient.Run()
		}()
	}

	forever := make(chan struct{})
	<-forever
}
