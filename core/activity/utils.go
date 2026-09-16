package activity

import (
	"github.com/infracloudio/msbotbuilder-go/schema"
)

func GetCoversationReference(activity schema.Activity) schema.ConversationReference {
	return schema.ConversationReference{
		ActivityID:   activity.ID,
		User:         activity.From,
		Bot:          activity.Recipient,
		Conversation: activity.Conversation,
		ChannelID:    activity.ChannelID,
		ServiceURL:   activity.ServiceURL,
	}
}

func ApplyConversationReference(activity schema.Activity, reference schema.ConversationReference, isIncoming bool) schema.Activity {
	activity.ChannelID = reference.ChannelID
	activity.ServiceURL = reference.ServiceURL
	activity.Conversation = reference.Conversation
	if isIncoming {
		activity.From = reference.User
		activity.Recipient = reference.Bot
		if reference.ActivityID != "" {
			activity.ID = reference.ActivityID
		}
		return activity
	}
	activity.From = reference.Bot
	activity.Recipient = reference.User
	if reference.ActivityID != "" {
		activity.ReplyToID = reference.ActivityID
	}
	return activity
}
