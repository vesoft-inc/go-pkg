package notify

import (
	"context"

	"github.com/vesoft-inc/go-pkg/mail"
)

var _ StringNotifier = (*mailNotifier)(nil)

type (
	mailNotifier struct {
		client mail.Client
		config MailConfig
	}

	MailConfig struct {
		Mail        mail.Config
		To          []string
		ContentType string
		Subject     string
	}
)

// NewWithMails creates Notifier for many mails.
func NewWithMails(configs ...MailConfig) Notifier {
	stringNotifiers := make([]StringNotifier, len(configs))
	for i := range configs {
		stringNotifiers[i] = newMailNotifier(configs[i])
	}
	return NewWithStringNotifiers(stringNotifiers...)
}

func newMailNotifier(config MailConfig) StringNotifier { //nolint:gocritic
	return &mailNotifier{
		client: mail.NewClient(config.Mail),
		config: config,
	}
}

func (n *mailNotifier) Notify(_ context.Context, message string) error {
	return n.client.Send(n.config.To, n.config.Subject, n.config.ContentType, message)
}
