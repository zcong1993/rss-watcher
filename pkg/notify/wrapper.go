package notify

import (
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/zcong1993/notifiers/v2"
)

const (
	Ding     = "ding"
	Printer  = "printer"
	Telegram = "tg"
	Mail     = "mail"
)

var (
	ErrInvalidConfig = errors.New("invalid config")
)

type DingConfig struct {
	Webhook string `json:"webhook" validate:"required"`
	Secret  string `json:"secret"`
}

type MailConfig struct {
	Domain     string `json:"domain" validate:"required"`
	PrivateKey string `json:"private_key" validate:"required"`
	From       string `json:"from" validate:"required"`
	To         string `json:"to" validate:"required"`
}

type TelegramConfig struct {
	Token string `json:"token" validate:"required"`
	ToID  int64  `json:"to_id" validate:"required"`
}

type DingNotifier struct {
	*notifiers.Ding
}

func NewDingNotifier() *DingNotifier {
	return &DingNotifier{}
}

func (dn *DingNotifier) Init(cfg interface{}) error {
	config, ok := cfg.(*DingConfig)
	if !ok {
		return ErrInvalidConfig
	}

	err := validator.New().Struct(config)
	if err != nil {
		return validateError(err)
	}

	dn.Ding = notifiers.NewDing(config.Webhook, config.Secret)

	return nil
}

var _ Notifier = (*DingNotifier)(nil)

type PrinterNotifier struct {
	*notifiers.Printer
}

func NewPrinterNotifier() *PrinterNotifier {
	return &PrinterNotifier{}
}

func (n *PrinterNotifier) Init(cfg interface{}) error {
	n.Printer = notifiers.NewPrinter(os.Stderr)

	return nil
}

var _ Notifier = (*PrinterNotifier)(nil)

type TelegramNotifier struct {
	*notifiers.Telegram
}

func NewTelegramNotifier() *TelegramNotifier {
	return &TelegramNotifier{}
}

func (n *TelegramNotifier) Init(cfg interface{}) error {
	config, ok := cfg.(*TelegramConfig)
	if !ok {
		return ErrInvalidConfig
	}

	err := validator.New().Struct(config)
	if err != nil {
		return validateError(err)
	}

	tg, err := notifiers.NewTelegram(config.Token, config.ToID)

	if err != nil {
		return err
	}

	n.Telegram = tg

	return nil
}

var _ Notifier = (*TelegramNotifier)(nil)

type MailerNotifier struct {
	*notifiers.Mailer
}

func NewMailerNotifier() *MailerNotifier {
	return &MailerNotifier{}
}

func (n *MailerNotifier) Init(cfg interface{}) error {
	config, ok := cfg.(*MailConfig)
	if !ok {
		return ErrInvalidConfig
	}

	err := validator.New().Struct(config)
	if err != nil {
		return validateError(err)
	}

	n.Mailer = notifiers.NewMailer(config.Domain, config.PrivateKey, config.To, config.From)

	return nil
}

var _ Notifier = (*MailerNotifier)(nil)

func validateError(err error) error {
	return errors.Wrap(err, "validate error")
}
