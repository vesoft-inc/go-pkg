package notify

import (
	"bytes"
	"context"
	"text/template"
)

var _ Notifier = (*templateNotifier)(nil)

type (
	templateNotifier struct {
		notifier Notifier
		template *template.Template
	}
)

// NewWithTemplate creates Notifier for notifiers with a template.
func NewWithTemplate(tmpl *template.Template, notifiers ...Notifier) Notifier {
	return &templateNotifier{
		notifier: combineNotifiers(notifiers...),
		template: tmpl,
	}
}

func (n *templateNotifier) Notify(ctx context.Context, data interface{}) error {
	if n.template != nil {
		var buff bytes.Buffer
		if err := n.template.Execute(&buff, data); err != nil {
			return err
		}
		// if already set template then those notifiers use message instead.
		data = buff.String()
	}

	return n.notifier.Notify(ctx, data)
}
