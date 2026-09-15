package models

import "github.com/infracloudio/msbotbuilder-go/schema"

// AlertRequest represents the payload to trigger a proactive alert via API
type AlertRequest struct {
	TenantID string `json:"tenant_id,omitempty"`
	UserID   string `json:"user_id,omitempty"`
	Message  string `json:"message"`
}

// StoredReference wraps ConversationReference with metadata
type StoredReference struct {
	Reference schema.ConversationReference `json:"reference"`
	TenantID  string                      `json:"tenant_id"`
	UserID    string                      `json:"user_id"`
}
