//go:generate mockgen -package mail -destination mail_mock.go -source mail.go Client
package mail

import (
	"gopkg.in/gomail.v2"
)

const (
	ContentTypePlain = "text/plain"
	ContentTypeHTML  = "text/html"
)

// Stubbed out for tests.
var newDialerSender = func(config Config) dialerSender {
	return gomail.NewDialer(config.Host, config.Port, config.User, config.Password)
}

type (
	Client interface {
		Send(to []string, subject, contentType, body string) error
	}

	Config struct {
		Host     string
		Port     int
		User     string
		Password string
	}

	defaultClient struct {
		config Config
	}

	dialerSender interface {
		DialAndSend(m ...*gomail.Message) error
	}
)

func NewClient(config Config) Client {
	return &defaultClient{
		config: config,
	}
}

func (c *defaultClient) Send(to []string, subject, contentType, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", c.config.User)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", subject)
	m.SetBody(contentType, body)

	return newDialerSender(c.config).DialAndSend(m)
}
