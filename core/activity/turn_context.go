package activity

import (
	"github.com/infracloudio/msbotbuilder-go/schema"
	"github.com/pkg/errors"
)
type TurnContext struct {
	Activity schema.Activity
}

func (t *TurnContext) SendActivity(options ...MsgOption) (schema.Activity, error) {
	activity, err := applyMsgOptions(schema.Activity{Type: schema.Message}, options...)
	if err != nil {
		return activity, errors.Wrap(err, "Failed to apply MsgOptions.")
	}
	return ApplyConversationReference(activity, GetCoversationReference(t.Activity), false), nil
}

func applyMsgOptions(activity schema.Activity, options ...MsgOption) (schema.Activity, error) {
	for _, opt := range options {
		if err := opt(&activity); err != nil {
			return activity, err
		}
	}
	return activity, nil
}
