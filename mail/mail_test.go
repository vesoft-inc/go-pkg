package mail

import (
	"errors"
	"testing"

	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"gopkg.in/gomail.v2"
)

type (
	dialerSenderMock struct {
		send func(...*gomail.Message) error
	}
)

func (d dialerSenderMock) DialAndSend(m ...*gomail.Message) error {
	return d.send(m...)
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		to          []string
		subject     string
		contentType string
		body        string
		err         error
	}{{
		name:        "to:1",
		config:      Config{User: "u"},
		to:          []string{"1"},
		subject:     "subject",
		contentType: ContentTypePlain,
		body:        "body",
	}, {
		name:        "to:2",
		config:      Config{User: "u"},
		to:          []string{"1"},
		subject:     "subject",
		contentType: ContentTypeHTML,
		body:        "body",
	}, {
		name:        "error",
		config:      Config{User: "u"},
		to:          []string{"1"},
		subject:     "subject",
		contentType: ContentTypePlain,
		body:        "body",
		err:         errors.New("send error"),
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ast := assert.New(t)
			stubs := gostub.StubFunc(&newDialerSender, dialerSenderMock{
				send: func(msg ...*gomail.Message) error {
					if ast.Len(msg, 1) {
						ast.Equal([]string{test.config.User}, msg[0].GetHeader("From"))
						ast.Equal(test.to, msg[0].GetHeader("To"))
						ast.Equal([]string{test.subject}, msg[0].GetHeader("Subject"))
					}
					return test.err
				},
			})
			defer stubs.Reset()
			c := NewClient(test.config)
			err := c.Send(test.to, test.subject, test.contentType, test.body)
			ast.Equal(test.err, err)
		})
	}
}

func Test_newDialerSender(t *testing.T) {
	d := newDialerSender(Config{})
	assert.IsType(t, (*gomail.Dialer)(nil), d)
}
