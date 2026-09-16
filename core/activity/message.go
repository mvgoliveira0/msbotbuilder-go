package activity

import (
	"github.com/infracloudio/msbotbuilder-go/schema"
)

type MsgOption func(*schema.Activity) error

func MsgOptionText(text string) MsgOption {
	return func(activity *schema.Activity) error {
		activity.Text = text
		return nil
	}
}

func MsgOptionAttachments(attachments []schema.Attachment) MsgOption {
	return func(activity *schema.Activity) error {
		activity.Attachments = attachments
		return nil
	}
}
