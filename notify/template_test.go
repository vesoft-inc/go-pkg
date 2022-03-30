package notify

import (
	"bytes"
	"context"
	"testing"
	"text/template"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestNewWithTemplate(t *testing.T) {
	var (
		ast       = assert.New(t)
		notifier  Notifier
		notifiers []Notifier
	)

	notifier = NewWithTemplate(template.New(""), NotifierFunc(nil))
	ast.IsType(&templateNotifier{}, notifier)
	notifier = notifier.(*templateNotifier).notifier
	ast.IsType(NotifierFunc(nil), notifier)

	notifier = NewWithTemplate(template.New(""), NotifierFunc(nil), NotifierFunc(nil))
	ast.IsType(&templateNotifier{}, notifier)
	notifier = notifier.(*templateNotifier).notifier
	ast.IsType(&defaultNotify{}, notifier)
	notifiers = notifier.(*defaultNotify).notifiers
	ast.Len(notifiers, 2)
}

func TestTemplateNotify(t *testing.T) {
	var (
		notifier Notifier
		err      error
	)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	ast := assert.New(t)

	n11 := NewMockNotifier(ctrl)
	n12 := NewMockNotifier(ctrl)
	n13 := NewMockNotifier(ctrl)
	n21 := NewMockNotifier(ctrl)
	n22 := NewMockNotifier(ctrl)
	n23 := NewMockNotifier(ctrl)

	t1 := template.Must(template.New("test-t1").Parse("t1 {{.Data}}"))
	t2 := template.Must(template.New("test-t2").Parse("t2 {{.Data}}"))

	notifier = New(
		New(NewWithTemplate(t1, n11, n12), n13),
		New(NewWithTemplate(t2, n21), n22, n23),
	)

	for _, data := range []interface{}{
		100,
		9.2,
		"str",
		[]byte("bytes"),
		[]interface{}{100, 9.2, "str", []byte("bytes")},
		map[string]interface{}{
			"i":  100,
			"f":  9.2,
			"s":  "str",
			"bs": []byte("bytes"),
		},
	} {
		data = map[string]interface{}{"Data": data}

		var buff bytes.Buffer
		err = t1.Execute(&buff, data)
		ast.NoError(err)
		m1 := buff.String()

		buff.Reset()
		err = t2.Execute(&buff, data)
		ast.NoError(err)
		m2 := buff.String()

		n11.EXPECT().Notify(gomock.Any(), m1).Times(1)
		n12.EXPECT().Notify(gomock.Any(), m1).Times(1)
		n13.EXPECT().Notify(gomock.Any(), data).Times(1)
		n21.EXPECT().Notify(gomock.Any(), m2).Times(1)
		n22.EXPECT().Notify(gomock.Any(), data).Times(1)
		n23.EXPECT().Notify(gomock.Any(), data).Times(1)
		err = notifier.Notify(context.TODO(), data)
		ast.NoError(err)
	}

	n11.EXPECT().Notify(gomock.Any(), gomock.Any()).MaxTimes(1)
	n12.EXPECT().Notify(gomock.Any(), gomock.Any()).MaxTimes(1)
	n13.EXPECT().Notify(gomock.Any(), gomock.Any()).MaxTimes(1)
	n21.EXPECT().Notify(gomock.Any(), gomock.Any()).MaxTimes(1)
	n22.EXPECT().Notify(gomock.Any(), gomock.Any()).MaxTimes(1)
	n23.EXPECT().Notify(gomock.Any(), gomock.Any()).MaxTimes(1)
	err = notifier.Notify(context.TODO(), "")
	ast.ErrorIs(err, ErrNotifyNotification)
}
