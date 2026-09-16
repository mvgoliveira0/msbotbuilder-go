package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/infracloudio/msbotbuilder-go/connector/auth"
	"github.com/infracloudio/msbotbuilder-go/connector/client"
	"github.com/infracloudio/msbotbuilder-go/core/activity"
	"github.com/infracloudio/msbotbuilder-go/schema"
	"github.com/pkg/errors"
)

// Adapter is the primary interface for the user program to perform operations with
// the connector service.
type Adapter interface {
	ParseRequest(ctx context.Context, req *http.Request) (schema.Activity, error)
	ProcessActivity(ctx context.Context, req schema.Activity, handler activity.Handler) error
	ProactiveMessage(ctx context.Context, ref schema.ConversationReference, handler activity.Handler) error
	DeleteActivity(ctx context.Context, activityID string, ref schema.ConversationReference) error
	UpdateActivity(ctx context.Context, activity schema.Activity) error
}

// AdapterSetting is the configuration for the Adapter.
type AdapterSetting struct {
	AppID              string
	AppPassword        string
	ChannelAuthTenant  string
	OauthEndpoint      string
	OpenIDMetadata     string
	ChannelService     string
	CredentialProvider auth.CredentialProvider
	AuthClient         *http.Client
	ReplyClient        *http.Client
}

type BotFrameworkAdapter struct {
	AdapterSetting
	auth.TokenValidator
	client.Client
}

func NewBotAdapter(settings AdapterSetting) (Adapter, error) {
	settings.CredentialProvider = auth.SimpleCredentialProvider{
		AppID:    settings.AppID,
		Password: settings.AppPassword,
	}

	if settings.ChannelService == "" {
		settings.ChannelService = auth.ChannelService
	}

	loginURL := auth.ToChannelFromBotLoginURL[0]
	if settings.OauthEndpoint != "" {
		loginURL = settings.OauthEndpoint
	} else if settings.ChannelAuthTenant != "" {
		loginURL = fmt.Sprintf("%s%s%s", auth.ToChannelFromBotLoginURLPrefix, settings.ChannelAuthTenant, auth.ToChannelFromBotTokenEndpointPathTOCHANNELFROMBOTTOKENENDPOINTPATH)
	}

	clientConfig, err := client.NewClientConfig(settings.CredentialProvider, loginURL)
	if err != nil {
		return nil, err
	}

	if settings.AuthClient != nil {
		clientConfig.AuthClient = settings.AuthClient
	}

	if settings.ReplyClient != nil {
		clientConfig.ReplyClient = settings.ReplyClient
	}

	connectorClient, err := client.NewClient(clientConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create Connector Client.")
	}

	return &BotFrameworkAdapter{settings, auth.NewJwtTokenValidator(), connectorClient}, nil
}

func (bf *BotFrameworkAdapter) ProcessActivity(ctx context.Context, req schema.Activity, handler activity.Handler) error {
	turnContext := &activity.TurnContext{
		Activity: req,
	}

	replyActivity, err := activity.PrepareActivityContext(handler, turnContext)
	if err != nil {
		return errors.Wrap(err, "Failed to create Activity context.")
	}

	response, err := activity.NewActivityResponse(bf.Client)
	if err != nil {
		return errors.Wrap(err, "Failed to create response object.")
	}

	return response.SendActivity(ctx, replyActivity)
}

func (bf *BotFrameworkAdapter) ProactiveMessage(ctx context.Context, ref schema.ConversationReference, handler activity.Handler) error {
	activity := activity.ApplyConversationReference(schema.Activity{Type: schema.Message}, ref, true)
	return bf.ProcessActivity(ctx, activity, handler)
}

func (bf *BotFrameworkAdapter) DeleteActivity(ctx context.Context, activityID string, ref schema.ConversationReference) error {
	req := activity.ApplyConversationReference(schema.Activity{Type: schema.Message}, ref, true)
	req.ID = activityID

	response, err := activity.NewActivityResponse(bf.Client)
	if err != nil {
		return errors.Wrap(err, "Failed to create response object.")
	}

	return response.DeleteActivity(ctx, req)
}

func (bf *BotFrameworkAdapter) ParseRequest(ctx context.Context, req *http.Request) (schema.Activity, error) {
	activity := schema.Activity{}

	authHeader := req.Header.Get("Authorization")
	if len(authHeader) == 0 {
		return activity, errors.New("Authentication headers are missing in the request")
	}


	err := json.NewDecoder(req.Body).Decode(&activity)
	if err != nil {
		return activity, errors.Wrap(err, "Error while parsing Bot request")
	}
	return activity, bf.authenticateRequest(ctx, activity, authHeader)
}

func (bf *BotFrameworkAdapter) authenticateRequest(ctx context.Context, req schema.Activity, headers string) error {

	_, err := bf.TokenValidator.AuthenticateRequest(ctx, req, headers, bf.CredentialProvider, bf.ChannelService)

	return errors.Wrap(err, "Authentication failed.")
}

func (bf *BotFrameworkAdapter) UpdateActivity(ctx context.Context, req schema.Activity) error {
	response, err := activity.NewActivityResponse(bf.Client)

	if err != nil {
		return errors.Wrap(err, "Failed to create response object.")
	}
	return response.UpdateActivity(ctx, req)
}
