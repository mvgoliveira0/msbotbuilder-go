package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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

// ProcessWebhookRequest parses incoming activity and handles onboarding, commands, and options menu
func (s *botService) ProcessWebhookRequest(req *http.Request) error {
	ctx := req.Context()
	act, err := s.adapter.ParseRequest(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to parse request: %w", err)
	}

	// Log incoming user details
	fmt.Printf("[Incoming Message] User ID: %q | Name: %q | AAD Object ID: %q | Tenant ID: %q | Conversation ID: %q\n",
		act.From.ID, act.From.Name, act.From.AadObjectID, act.Conversation.TenantID, act.Conversation.ID)

	ref := activity.GetCoversationReference(act)
	if act.ServiceURL != "" {
		_ = s.repo.Save(ref)
	}

	handler := activity.HandlerFuncs{
		OnConversationUpdateFunc: func(turn *activity.TurnContext) (schema.Activity, error) {
			welcomeCard := createWelcomeCard()
			attachment := schema.Attachment{
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content:     welcomeCard,
			}
			return turn.SendActivity(activity.MsgOptionAttachments([]schema.Attachment{attachment}))
		},
		OnMessageFunc: func(turn *activity.TurnContext) (schema.Activity, error) {
			cmd := extractCommand(act)

			switch cmd {
			case "/start", "subscribe", "inscrever", "iniciar":
				_ = s.repo.SetSubscription(ref, true)
				card := createSubscribeCard()
				attachment := schema.Attachment{
					ContentType: "application/vnd.microsoft.card.adaptive",
					Content:     card,
				}
				return turn.SendActivity(activity.MsgOptionAttachments([]schema.Attachment{attachment}))

			case "/stop", "unsubscribe", "cancelar", "sair":
				_ = s.repo.SetSubscription(ref, false)
				card := createUnsubscribeCard()
				attachment := schema.Attachment{
					ContentType: "application/vnd.microsoft.card.adaptive",
					Content:     card,
				}
				return turn.SendActivity(activity.MsgOptionAttachments([]schema.Attachment{attachment}))

			case "/status", "status":
				isSub := s.repo.IsSubscribed(act.Conversation.TenantID, act.From.ID)
				card := createMenuCard(isSub)
				attachment := schema.Attachment{
					ContentType: "application/vnd.microsoft.card.adaptive",
					Content:     card,
				}
				return turn.SendActivity(activity.MsgOptionAttachments([]schema.Attachment{attachment}))

			default:
				isSub := s.repo.IsSubscribed(act.Conversation.TenantID, act.From.ID)
				card := createMenuCard(isSub)
				attachment := schema.Attachment{
					ContentType: "application/vnd.microsoft.card.adaptive",
					Content:     card,
				}
				return turn.SendActivity(activity.MsgOptionAttachments([]schema.Attachment{attachment}))
			}
		},
	}

	err = s.adapter.ProcessActivity(ctx, act, handler)
	if err != nil {
		return fmt.Errorf("failed to process activity: %w", err)
	}

	return nil
}

// SendProactiveAlert sends a proactive alert message ONLY if the target user is subscribed
func (s *botService) SendProactiveAlert(ctx context.Context, tenantID, userID, message string) error {
	var ref *schema.ConversationReference
	var err error

	if tenantID != "" && userID != "" {
		ref, err = s.repo.GetSubscribed(tenantID, userID)
	} else if userID != "" {
		ref, err = s.repo.GetSubscribed("", userID)
	} else {
		ref, err = s.repo.GetLatestSubscribed()
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

// GetActiveSessions returns all active conversation sessions with subscription status
func (s *botService) GetActiveSessions(ctx context.Context) ([]models.SessionResponse, error) {
	return s.repo.GetAllSessions()
}

func extractCommand(act schema.Activity) string {
	text := strings.TrimSpace(strings.ToLower(act.Text))
	if text != "" {
		return text
	}

	if act.Value != nil {
		if action, ok := act.Value["action"].(string); ok {
			return strings.TrimSpace(strings.ToLower(action))
		}
	}
	return ""
}

func createWelcomeCard() map[string]interface{} {
	return map[string]interface{}{
		"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
		"type":    "AdaptiveCard",
		"version": "1.0",
		"body": []map[string]interface{}{
			{
				"type":   "TextBlock",
				"text":   "👋 Bem-vindo ao Bot de Alertas!",
				"size":   "large",
				"weight": "bolder",
				"color":  "accent",
			},
			{
				"type": "TextBlock",
				"text": "Para receber notificações proativas e alertas importantes diretamente neste chat, por favor clique no botão abaixo para se inscrever ou digite /start.",
				"wrap": true,
			},
		},
		"actions": []map[string]interface{}{
			{
				"type":  "Action.Submit",
				"title": "🔔 Inscrever para Alertas",
				"data": map[string]interface{}{
					"action": "subscribe",
				},
			},
		},
	}
}

func createSubscribeCard() map[string]interface{} {
	return map[string]interface{}{
		"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
		"type":    "AdaptiveCard",
		"version": "1.0",
		"body": []map[string]interface{}{
			{
				"type":   "TextBlock",
				"text":   "✅ Inscrição Realizada!",
				"size":   "large",
				"weight": "bolder",
				"color":  "good",
			},
			{
				"type": "TextBlock",
				"text": "Você foi cadastrado com sucesso! A partir de agora você receberá alertas proativos e notificações neste chat.",
				"wrap": true,
			},
		},
		"actions": []map[string]interface{}{
			{
				"type":  "Action.Submit",
				"title": "🔕 Cancelar Inscrição",
				"data": map[string]interface{}{
					"action": "unsubscribe",
				},
			},
		},
	}
}

func createUnsubscribeCard() map[string]interface{} {
	return map[string]interface{}{
		"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
		"type":    "AdaptiveCard",
		"version": "1.0",
		"body": []map[string]interface{}{
			{
				"type":   "TextBlock",
				"text":   "🔕 Inscrição Cancelada",
				"size":   "large",
				"weight": "bolder",
				"color":  "warning",
			},
			{
				"type": "TextBlock",
				"text": "Sua inscrição para alertas proativos foi cancelada. Você não receberá mais notificações neste chat.",
				"wrap": true,
			},
		},
		"actions": []map[string]interface{}{
			{
				"type":  "Action.Submit",
				"title": "🔔 Reativar Inscrição",
				"data": map[string]interface{}{
					"action": "subscribe",
				},
			},
		},
	}
}

func createMenuCard(isSubscribed bool) map[string]interface{} {
	statusText := "Status atual: 🔴 Não Inscrito"
	if isSubscribed {
		statusText = "Status atual: 🟢 Inscrito para Alertas"
	}

	return map[string]interface{}{
		"$schema": "http://adaptivecards.io/schemas/adaptive-card.json",
		"type":    "AdaptiveCard",
		"version": "1.0",
		"body": []map[string]interface{}{
			{
				"type":   "TextBlock",
				"text":   "🤖 Menu de Opções",
				"size":   "large",
				"weight": "bolder",
			},
			{
				"type":   "TextBlock",
				"text":   statusText,
				"weight": "bolder",
				"wrap":   true,
			},
			{
				"type": "TextBlock",
				"text": "Selecione uma das opções abaixo ou digite um dos comandos disponíveis (/start, /stop, /status):",
				"wrap": true,
			},
		},
		"actions": []map[string]interface{}{
			{
				"type":  "Action.Submit",
				"title": "🔔 Inscrever (/start)",
				"data": map[string]interface{}{
					"action": "subscribe",
				},
			},
			{
				"type":  "Action.Submit",
				"title": "🔕 Cancelar Inscrição (/stop)",
				"data": map[string]interface{}{
					"action": "unsubscribe",
				},
			},
			{
				"type":  "Action.Submit",
				"title": "ℹ️ Verificar Status (/status)",
				"data": map[string]interface{}{
					"action": "status",
				},
			},
		},
	}
}
