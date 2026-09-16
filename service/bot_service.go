package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/infracloudio/msbotbuilder-go/core"
	"github.com/infracloudio/msbotbuilder-go/core/activity"
	"github.com/infracloudio/msbotbuilder-go/models"
	"github.com/infracloudio/msbotbuilder-go/repository"
	"github.com/infracloudio/msbotbuilder-go/schema"
)

// BotService defines business logic for bot activities and proactive messages
type BotService interface {
	ProcessWebhookRequest(req *http.Request) error
	SendProactiveAlert(ctx context.Context, tenantID, userID, message string) error
	GetActiveSessions(ctx context.Context) ([]models.SessionResponse, error)
}

type botService struct {
	adapter core.Adapter
	repo    repository.ConversationRepository
}

// NewBotService creates a new BotService instance
func NewBotService(adapter core.Adapter, repo repository.ConversationRepository) BotService {
	return &botService{
		adapter: adapter,
		repo:    repo,
	}
}

// ProcessWebhookRequest parses incoming activity and saves conversation reference
func (s *botService) ProcessWebhookRequest(req *http.Request) error {
	ctx := req.Context()
	act, err := s.adapter.ParseRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to parse request: %w", err)
	}

	// Save conversation reference if valid ServiceURL is present
	if act.ServiceURL != "" {
		ref := activity.GetCoversationReference(act)
		_ = s.repo.Save(ref)
	}

	handler := activity.HandlerFuncs{
		OnMessageFunc: func(turn *activity.TurnContext) (schema.Activity, error) {
			return turn.SendActivity(activity.MsgOptionText("Mensagem recebida com sucesso!"))
		},
		OnConversationUpdateFunc: func(turn *activity.TurnContext) (schema.Activity, error) {
			return schema.Activity{}, nil
		},
	}

	err = s.adapter.ProcessActivity(ctx, act, handler)
	if err != nil {
		return fmt.Errorf("failed to process activity: %w", err)
	}

	return nil
}

// SendProactiveAlert sends a proactive alert message to a specific user/tenant or latest session
func (s *botService) SendProactiveAlert(ctx context.Context, tenantID, userID, message string) error {
	var ref *schema.ConversationReference
	var err error

	if tenantID != "" && userID != "" {
		ref, err = s.repo.Get(tenantID, userID)
	} else {
		ref, err = s.repo.GetLatest()
	}

	if err != nil {
		return fmt.Errorf("cannot send alert: %w", err)
	}

	return s.sendCardAlert(ctx, *ref, message)
}

func (s *botService) sendCardAlert(ctx context.Context, ref schema.ConversationReference, text string) error {
	cardData := map[string]interface{}{
		"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
		"type":    "AdaptiveCard",
		"version": "1.0",
		"body": []map[string]interface{}{
			{
				"type":   "TextBlock",
				"text":   "🚨 ALERTA DO SISTEMA",
				"size":   "large",
				"weight": "bolder",
				"color":  "attention",
			},
			{
				"type": "TextBlock",
				"text": text,
				"wrap": true,
			},
		},
	}

	handler := activity.HandlerFuncs{
		OnMessageFunc: func(turn *activity.TurnContext) (schema.Activity, error) {
			attachments := []schema.Attachment{
				{
					ContentType: "application/vnd.microsoft.card.adaptive",
					Content:     cardData,
				},
			}
			return turn.SendActivity(activity.MsgOptionAttachments(attachments))
		},
	}

	return s.adapter.ProactiveMessage(ctx, ref, handler)
}

// GetActiveSessions returns all active conversation sessions
func (s *botService) GetActiveSessions(ctx context.Context) ([]models.SessionResponse, error) {
	refs, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	sessions := make([]models.SessionResponse, 0, len(refs))
	for _, ref := range refs {
		sessions = append(sessions, models.SessionResponse{
			TenantID:       ref.Conversation.TenantID,
			UserID:         ref.User.ID,
			AadObjectID:    ref.User.AadObjectID,
			UserName:       ref.User.Name,
			ConversationID: ref.Conversation.ID,
			ServiceURL:     ref.ServiceURL,
		})
	}
	return sessions, nil
}
