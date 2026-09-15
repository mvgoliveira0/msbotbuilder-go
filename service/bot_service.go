package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/infracloudio/msbotbuilder-go/core"
	"github.com/infracloudio/msbotbuilder-go/core/activity"
	"github.com/infracloudio/msbotbuilder-go/repository"
	"github.com/infracloudio/msbotbuilder-go/schema"
)

var cardJSON = []byte(`{
  "$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
  "type": "AdaptiveCard",
  "version": "1.0",
  "body": [
    {
      "type": "TextBlock",
      "text": "🚨 ALERTA DO SISTEMA",
      "size": "large",
      "weight": "bolder",
      "color": "attention"
    },
    {
      "type": "TextBlock",
      "text": "Notificação proativa enviada com sucesso!",
      "wrap": true
    }
  ]
}`)

// BotService defines business logic for bot activities and proactive messages
type BotService interface {
	ProcessWebhookRequest(req *http.Request) error
	SendProactiveAlert(ctx context.Context, tenantID, userID, message string) error
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

	// Trigger proactive welcome alert after 2s delay
	go s.triggerProactiveWelcome()

	return nil
}

func (s *botService) triggerProactiveWelcome() {
	time.Sleep(2 * time.Second)

	ref, err := s.repo.GetLatest()
	if err != nil {
		fmt.Println("[ProactiveAlert] Warning:", err)
		return
	}

	err = s.sendCardAlert(context.Background(), *ref, "Alerta Proativo Inicial")
	if err != nil {
		fmt.Println("[ProactiveAlert] Error sending proactive message:", err)
		return
	}
	fmt.Println("[ProactiveAlert] Proactive message sent successfully after 2s delay.")
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
	handler := activity.HandlerFuncs{
		OnMessageFunc: func(turn *activity.TurnContext) (schema.Activity, error) {
			var obj map[string]interface{}
			_ = json.Unmarshal(cardJSON, &obj)

			attachments := []schema.Attachment{
				{
					ContentType: "application/vnd.microsoft.card.adaptive",
					Content:     obj,
				},
			}
			return turn.SendActivity(activity.MsgOptionText("🚨 "+text), activity.MsgOptionAttachments(attachments))
		},
	}

	return s.adapter.ProactiveMessage(ctx, ref, handler)
}
