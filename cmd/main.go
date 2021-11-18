package main

import (
	"fmt"
	"log"
	"runtime/debug"
	"time"

	"github.com/zcong1993/rss-watcher/pkg/store/pg"

	"github.com/spf13/cobra"
	"github.com/zcong1993/rss-watcher/pkg/logger"
	"github.com/zcong1993/rss-watcher/pkg/notify"
	"github.com/zcong1993/rss-watcher/pkg/runtime"
	"github.com/zcong1993/rss-watcher/pkg/store"
	"github.com/zcong1993/rss-watcher/pkg/store/fauna"
	"github.com/zcong1993/rss-watcher/pkg/store/file"
	"github.com/zcong1993/rss-watcher/pkg/store/mem"
)

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

func main() {
	appName := "rss-watcher"

	var (
		configFile    string
		limitInterval time.Duration
		serialize     bool
	)

	app := &cobra.Command{
		Use:     appName,
		Short:   "Watcher rss source and notify new.",
		Version: buildVersion(version, commit, date, builtBy),
	}

	app.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config file path.")
	_ = app.MarkPersistentFlagRequired("config")
	app.PersistentFlags().DurationVarP(&limitInterval, "limit", "l", 0, "If sleep between notify messages.")
	loggerFactory := logger.BindCobraFlags(app)

	var optsFactory = func() []runtime.Option {
		return []runtime.Option{
			runtime.WithLimitInterval(limitInterval),
			runtime.WithConfigFile(configFile),
			runtime.WithStores(
				store.New(file.Name, func() store.Store {
					return file.NewFileStore()
				}),
				store.New(mem.Name, func() store.Store {
					return mem.NewMemStore()
				}),
				store.New(fauna.Name, func() store.Store {
					return fauna.NewFanuaStore()
				}),
				store.New(pg.Name, func() store.Store {
					return pg.NewPg()
				}),
			),
			runtime.WithNotifiers(
				notify.New(notify.Printer, func() notify.Notifier {
					return notify.NewPrinterNotifier()
				}),
				notify.New(notify.Telegram, func() notify.Notifier {
					return notify.NewTelegramNotifier()
				}),
				notify.New(notify.Ding, func() notify.Notifier {
					return notify.NewDingNotifier()
				}),
				notify.New(notify.Mail, func() notify.Notifier {
					return notify.NewMailerNotifier()
				}),
			),
		}
	}

	singleCmd := &cobra.Command{
		Use:   "single",
		Short: "Run single and exit.",
		Run: cmdRun(func(cmd *cobra.Command, args []string) error {
			l, err := loggerFactory(appName)
			if err != nil {
				return err
			}

			r := runtime.NewRssWatcherRuntime(l)

			opts := optsFactory()

			opts = append(opts, runtime.WithSingle(true), runtime.WithSerialize(serialize))

			return r.Run(opts...)
		}),
	}

	singleCmd.PersistentFlags().BoolVarP(&serialize, "serialize", "s", false, "If run serialize, only work in single mode.")

	daemonCmd := &cobra.Command{
		Use:   "daemon",
		Short: "Run as daemon.",
		Run: cmdRun(func(cmd *cobra.Command, args []string) error {
			l, err := loggerFactory(appName)
			if err != nil {
				return err
			}

			l.Info("run as daemon")

			r := runtime.NewRssWatcherRuntime(l)

			return r.Run(optsFactory()...)
		}),
	}

	app.AddCommand(singleCmd, daemonCmd)

	if err := app.Execute(); err != nil {
		log.Fatalf("fatal error: %s", err)
	}
}

func cmdRun(f func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		err := f(cmd, args)
		if err != nil {
			log.Fatalf("fatal error from runtime: %s", err)
		}
	}
}

func buildVersion(version, commit, date, builtBy string) string {
	result := version
	if commit != "" {
		result = fmt.Sprintf("%s\ncommit: %s", result, commit)
	}

	if date != "" {
		result = fmt.Sprintf("%s\nbuilt at: %s", result, date)
	}

	if builtBy != "" {
		result = fmt.Sprintf("%s\nbuilt by: %s", result, builtBy)
	}

	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Sum != "" {
		result = fmt.Sprintf("%s\nmodule version: %s, checksum: %s", result, info.Main.Version, info.Main.Sum)
	}

	return result
}
