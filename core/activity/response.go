package activity

import (
	"context"
	"fmt"
	"net/url"
	"path"

	"github.com/infracloudio/msbotbuilder-go/connector/client"
	"github.com/infracloudio/msbotbuilder-go/schema"
	"github.com/pkg/errors"
)
type Response interface {
	SendActivity(ctx context.Context, activity schema.Activity) error
	DeleteActivity(ctx context.Context, activity schema.Activity) error
	UpdateActivity(ctx context.Context, activity schema.Activity) error
}

const (
	APIVersion = "v3"

	sendToConversationURL = "/%s/conversations/%s/activities"
	activityResourceURL   = "/%s/conversations/%s/activities/%s"
)
type DefaultResponse struct {
	Client client.Client
}
func (response *DefaultResponse) DeleteActivity(ctx context.Context, activity schema.Activity) error {
	u, err := url.Parse(activity.ServiceURL)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse ServiceURL %s.", activity.ServiceURL)
	}

	respPath := fmt.Sprintf(activityResourceURL, APIVersion, activity.Conversation.ID, activity.ID)

	u.Path = path.Join(u.Path, respPath)
	err = response.Client.Delete(ctx, *u)
	return errors.Wrap(err, "Failed to delete response.")
}
func (response *DefaultResponse) SendActivity(ctx context.Context, activity schema.Activity) error {
	u, err := url.Parse(activity.ServiceURL)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse ServiceURL %s.", activity.ServiceURL)
	}

	respPath := fmt.Sprintf(sendToConversationURL, APIVersion, activity.Conversation.ID)

	if activity.ReplyToID != "" {
		respPath = fmt.Sprintf(activityResourceURL, APIVersion, activity.Conversation.ID, activity.ReplyToID)
	}

	u.Path = path.Join(u.Path, respPath)
	err = response.Client.Post(ctx, *u, activity)
	return errors.Wrap(err, "Failed to send response.")
}
func (response *DefaultResponse) UpdateActivity(ctx context.Context, activity schema.Activity) error {
	u, err := url.Parse(activity.ServiceURL)
	if err != nil {
		return errors.Wrapf(err, "Failed to parse ServiceURL %s.", activity.ServiceURL)
	}

	respPath := fmt.Sprintf(activityResourceURL, APIVersion, activity.Conversation.ID, activity.ID)

	u.Path = path.Join(u.Path, respPath)
	err = response.Client.Put(ctx, *u, activity)
	return errors.Wrap(err, "Failed to update response.")
}
func NewActivityResponse(connectorClient client.Client) (Response, error) {
	if connectorClient == nil {
		return nil, errors.New("Invalid connector client for ActivityResponse")
	}

	return &DefaultResponse{connectorClient}, nil
}
