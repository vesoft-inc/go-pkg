package notify

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vesoft-inc/go-pkg/httpclient"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewWithDingTalks(t *testing.T) {
	var (
		ast      = assert.New(t)
		notifier Notifier
	)

	notifier = NewWithDingTalks(DingTalkConfig{})
	ast.IsType(NotifierFunc(nil), notifier)

	notifier = NewWithDingTalks(DingTalkConfig{}, DingTalkConfig{})
	ast.IsType(&defaultNotify{}, notifier)
	ast.Len(notifier.(*defaultNotify).notifiers, 2)
}

func Test_newDingTalkNotifier(t *testing.T) {
	ast := assert.New(t)

	n := newDingTalkNotifier(DingTalkConfig{})
	ast.IsType(&dingTalkNotifier{}, n)
	ast.NotNil(n.(*dingTalkNotifier).client)
}

func TestDingTalkNotify(t *testing.T) {
	var (
		notifier     Notifier
		err          error
		httpResponse struct {
			status int
			body   []byte
		}
	)

	ast := assert.New(t)

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == resty.MethodPost {
			body, err := io.ReadAll(r.Body) //nolint:govet
			ast.NoError(err)
			var requestBody dingTalkMessage
			err = json.Unmarshal(body, &requestBody)
			ast.NoError(err)

			if requestBody.MsgType != string(DingDingMsgText) &&
				requestBody.MsgType != string(DingDingMsgMarkdown) {
				ast.Fail("unsupported MsgType")
			}
			w.WriteHeader(httpResponse.status)
			_, _ = w.Write(httpResponse.body)
			return
		}
	}))

	client1 := httpclient.NewObjectClientRaw(httpclient.NewClient(testServer.URL))
	client2 := httpclient.NewObjectClientRaw(httpclient.NewClient(testServer.URL))

	n1 := newDingTalkNotifier(DingTalkConfig{MsgType: DingDingMsgText})
	n1.(*dingTalkNotifier).client = client1
	n2 := newDingTalkNotifier(DingTalkConfig{MsgType: DingDingMsgMarkdown})
	n2.(*dingTalkNotifier).client = client2

	notifier = NewWithStringNotifiers(n1, n2)

	httpResponse.status = 200
	httpResponse.body = []byte(`{"errcode": 0,"errmsg": "msg"}`)
	err = notifier.Notify(context.TODO(), "Message")
	ast.NoError(err)

	httpResponse.status = 200
	httpResponse.body = []byte(`{"errcode": 100001,"errmsg": "msg"}`)
	err = notifier.Notify(context.TODO(), "Message")
	if ast.ErrorIs(err, ErrNotifyNotification) {
		ast.Contains(err.Error(), "100001:msg")
	}

	httpResponse.status = 500
	err = notifier.Notify(context.TODO(), "Message")
	if ast.ErrorIs(err, ErrNotifyNotification) {
		ast.Contains(err.Error(), "500 Internal Server Error")
	}
}
