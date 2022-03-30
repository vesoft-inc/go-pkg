package notify

import (
	"context"
	"fmt"

	"github.com/vesoft-inc/go-pkg/httpclient"

	"github.com/pkg/errors"
)

// docs: https://open.dingtalk.com/document/group/custom-robot-access

const (
	dingTalkRobotSendAddr = "https://oapi.dingtalk.com/robot/send"

	DingDingMsgText     DingDingMsgType = "text"
	DingDingMsgMarkdown DingDingMsgType = "markdown"
)

var _ StringNotifier = (*dingTalkNotifier)(nil)

type (
	DingDingMsgType string

	DingTalkConfig struct {
		AccessToken string
		MsgType     DingDingMsgType
		AtMobiles   []string
		IsAtAll     bool
		Title       string
	}

	dingTalkNotifier struct {
		client httpclient.ObjectClient
		config DingTalkConfig
	}

	dingTalkMessage struct {
		MsgType  string                   `json:"msgtype"`
		At       dingTalkAtInfo           `json:"at"`
		Text     *dingTalkMessageText     `json:"text,omitempty"`
		Markdown *dingTalkMessageMarkdown `json:"markdown,omitempty"`
	}

	dingTalkAtInfo struct {
		AtMobiles []string `json:"atMobiles"`
		IsAtAll   bool     `json:"isAtAll"`
	}

	dingTalkMessageText struct {
		Content string `json:"content"`
	}

	dingTalkMessageMarkdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	}
)

// NewWithDingTalks creates Notifier for many  ding talk robots.
func NewWithDingTalks(configs ...DingTalkConfig) Notifier {
	stringNotifiers := make([]StringNotifier, len(configs))
	for i := range configs {
		stringNotifiers[i] = newDingTalkNotifier(configs[i])
	}
	return NewWithStringNotifiers(stringNotifiers...)
}

func newDingTalkNotifier(config DingTalkConfig) StringNotifier { // nolint:gocritic
	return &dingTalkNotifier{
		client: httpclient.NewObjectClient(dingTalkRobotSendAddr, httpclient.WithQueryParam("access_token", config.AccessToken)),
		config: config,
	}
}

func (n *dingTalkNotifier) Notify(_ context.Context, message string) error {
	messageBody := &dingTalkMessage{
		MsgType: string(n.config.MsgType),
		At: dingTalkAtInfo{
			AtMobiles: n.config.AtMobiles,
			IsAtAll:   n.config.IsAtAll,
		},
	}
	if n.config.MsgType == DingDingMsgMarkdown {
		messageBody.Markdown = &dingTalkMessageMarkdown{
			Title: n.config.Title,
			Text:  message,
		}
	} else {
		messageBody.Text = &dingTalkMessageText{
			Content: fmt.Sprintf("%s\n%s", n.config.Title, message),
		}
	}

	var responseObj struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}

	if err := n.client.Post("", messageBody, &responseObj); err != nil {
		return err
	}
	if responseObj.ErrCode != 0 {
		return errors.Errorf("%d:%s", responseObj.ErrCode, responseObj.ErrMsg)
	}
	return nil
}
