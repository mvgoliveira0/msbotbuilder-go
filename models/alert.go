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

// SessionResponse represents an active conversation session for GET /api/sessions
type SessionResponse struct {
	TenantID       string `json:"tenant_id"`
	UserID         string `json:"user_id"`
	AadObjectID    string `json:"aad_object_id,omitempty"`
	UserName       string `json:"user_name,omitempty"`
	ConversationID string `json:"conversation_id"`
	ServiceURL     string `json:"service_url"`
}
