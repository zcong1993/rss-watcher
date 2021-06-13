package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/zcong1993/rss-watcher/pkg/logger"

	"github.com/zcong1993/rss-watcher/pkg/notify"

	"github.com/zcong1993/rss-watcher/pkg/runtime"
	"github.com/zcong1993/rss-watcher/pkg/store"
	"github.com/zcong1993/rss-watcher/pkg/store/fauna"
	"github.com/zcong1993/rss-watcher/pkg/store/file"
	"github.com/zcong1993/rss-watcher/pkg/store/mem"
	"gopkg.in/alecthomas/kingpin.v2"
)

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

var (
	app = kingpin.New("rss-watcher", "Watcher rss source and notify new.")

	configFile    = app.Flag("config", "Config file path.").Required().String()
	limitInterval = app.Flag("limit", "If sleep between notify messages.").Duration()

	singleCmd = app.Command("single", "Run single and exit.")
	serialize = singleCmd.Flag("serialize", "If run serialize, only work in single mode.").Bool()

	daemonCmd = app.Command("daemon", "Run as daemon.")
)

func main() {
	optFn := logger.BindKingpinFlags(app)

	app.Version(buildVersion(version, commit, date, builtBy))
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	l, err := logger.NewLogger("rss-watcher", optFn())

	if err != nil {
		log.Fatalf("init logger error: %s", err.Error())
	}

	r := runtime.NewRssWatcherRuntime(l)

	opts := []runtime.Option{
		runtime.WithLimitInterval(*limitInterval),
		runtime.WithConfigFile(*configFile),
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

	if cmd == singleCmd.FullCommand() {
		opts = append(opts, runtime.WithSingle(true), runtime.WithSerialize(*serialize))
	}

	if cmd == daemonCmd.FullCommand() {
		l.Info("run as daemon")
	}

	err = r.Run(opts...)

	if err != nil {
		l.Fatalf("fatal error from runtime: %s", err)
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
