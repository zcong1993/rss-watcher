package main

import (
	"fmt"
	"os"
	"sync"

	"github.com/zcong1993/rss-watcher/notifiers/printer"

	"github.com/zcong1993/notifiers/adapters/tg"

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

	fmt.Printf("use kv store: %s\n", cfg.KvStore)

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
	case "firestore":
		kvStore = kv.NewFireStore(cfg.FireStoreConfig.ProjectID, cfg.FireStoreConfig.Collection)
	case "dynamo-kv":
		kvStore = kv.NewDynamoKvClient(cfg.DynamoConfig.Namespace, cfg.DynamoConfig.Token)
	}

	if cfg.DingConfig != nil {
		dingNotifier = ding.NewClient(cfg.DingConfig.Token)
	}
	if cfg.MailConfig != nil {
		mailNotifier = mail.NewClient(cfg.MailConfig.Domain, cfg.MailConfig.PrivateKey, cfg.MailConfig.To, cfg.MailConfig.From)
	}
	if cfg.TelegramConfig != nil {
		tgNotifier = tg.NewClient(cfg.TelegramConfig.Token, cfg.TelegramConfig.ToID)
	}

	// run single
	if cfg.Single {
		fmt.Println("run single and exit.")
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
				watcherClient := watcher.NewRSSWatcher(rw.Source, rw.Interval.Duration, kvStore, notifiers, rw.Skip)
				_ = watcherClient.Single()
				wg.Done()
			}()
		}

		wg.Wait()
		fmt.Println("\nDone!")
	} else {
		// run as daemon
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

			go func() {
				watcherClient := watcher.NewRSSWatcher(rw.Source, rw.Interval.Duration, kvStore, notifiers, rw.Skip)
				watcherClient.Run()
			}()
		}

		forever := make(chan struct{})
		<-forever
	}
}
