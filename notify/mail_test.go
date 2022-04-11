package notify

import (
	"context"
	"testing"

	"github.com/vesoft-inc/go-pkg/mail"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewWithMails(t *testing.T) {
	var (
		ast      = assert.New(t)
		notifier Notifier
	)

	notifier = NewWithMails(MailConfig{})
	ast.IsType(NotifierFunc(nil), notifier)

	notifier = NewWithMails(MailConfig{}, MailConfig{})
	ast.IsType(&defaultNotify{}, notifier)
	ast.Len(notifier.(*defaultNotify).notifiers, 2)
}

func Test_newMailNotifier(t *testing.T) {
	ast := assert.New(t)

	n := newMailNotifier(MailConfig{})
	ast.IsType(&mailNotifier{}, n)
	ast.NotNil(n.(*mailNotifier).client)
}

func TestMailNotify(t *testing.T) {
	var (
		notifier Notifier
		err      error
	)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ast := assert.New(t)

	client := mail.NewMockClient(ctrl)

	n1 := newMailNotifier(MailConfig{})
	n1.(*mailNotifier).client = client
	n2 := newMailNotifier(MailConfig{})
	n2.(*mailNotifier).client = client

	notifier = NewWithStringNotifiers(n1, n2)

	client.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(2)
	err = notifier.Notify(context.TODO(), "Message")
	ast.NoError(err)

	client.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1)
	client.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("sendError"))
	err = notifier.Notify(context.TODO(), "Message")
	if ast.ErrorIs(err, ErrNotifyNotification) {
		ast.Contains(err.Error(), "sendError")
	}

	client.EXPECT().Send(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).MaxTimes(2).Return(errors.New("sendError"))
	err = notifier.Notify(context.TODO(), "Message")
	if ast.ErrorIs(err, ErrNotifyNotification) {
		ast.Contains(err.Error(), "sendError")
	}
}
