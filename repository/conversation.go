package repository

import (
	"fmt"
	"sync"

	"github.com/infracloudio/msbotbuilder-go/models"
	"github.com/infracloudio/msbotbuilder-go/schema"
)

// StoredSession wraps a ConversationReference with its subscription state
type StoredSession struct {
	Reference    schema.ConversationReference
	IsSubscribed bool
}

// ConversationRepository interface for storing and retrieving conversation references
type ConversationRepository interface {
	Save(ref schema.ConversationReference) error
	SetSubscription(ref schema.ConversationReference, isSubscribed bool) error
	Get(tenantID, userID string) (*schema.ConversationReference, error)
	GetSubscribed(tenantID, userID string) (*schema.ConversationReference, error)
	GetLatest() (*schema.ConversationReference, error)
	GetLatestSubscribed() (*schema.ConversationReference, error)
	GetAllSessions() ([]models.SessionResponse, error)
	GetAll() ([]schema.ConversationReference, error)
	IsSubscribed(tenantID, userID string) bool
}

// InMemoryConversationRepository is a thread-safe in-memory store for conversation references
type InMemoryConversationRepository struct {
	mu     sync.RWMutex
	data   map[string]*StoredSession
	latest *StoredSession
}

// NewInMemoryConversationRepository creates a new InMemoryConversationRepository instance
func NewInMemoryConversationRepository() *InMemoryConversationRepository {
	return &InMemoryConversationRepository{
		data: make(map[string]*StoredSession),
	}
}

// Save stores or updates a conversation reference without altering existing subscription state unless new
func (r *InMemoryConversationRepository) Save(ref schema.ConversationReference) error {
	if ref.ServiceURL == "" {
		return fmt.Errorf("invalid reference: empty ServiceURL")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if session exists to preserve subscription state
	existingSub := false
	convKey := ref.Conversation.ID
	if existing, ok := r.data[convKey]; ok && existing != nil {
		existingSub = existing.IsSubscribed
	}

	session := &StoredSession{
		Reference:    ref,
		IsSubscribed: existingSub,
	}

	r.storeKeys(session)
	return nil
}

// SetSubscription updates or stores a conversation reference and sets its subscription state
func (r *InMemoryConversationRepository) SetSubscription(ref schema.ConversationReference, isSubscribed bool) error {
	if ref.ServiceURL == "" {
		return fmt.Errorf("invalid reference: empty ServiceURL")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	session := &StoredSession{
		Reference:    ref,
		IsSubscribed: isSubscribed,
	}

	r.storeKeys(session)
	return nil
}

func (r *InMemoryConversationRepository) storeKeys(session *StoredSession) {
	ref := session.Reference
	key := ref.Conversation.ID
	r.data[key] = session

	if ref.User.ID != "" {
		r.data[ref.User.ID] = session
		if ref.Conversation.TenantID != "" {
			keyUser := fmt.Sprintf("%s_%s", ref.Conversation.TenantID, ref.User.ID)
			r.data[keyUser] = session
		}
	}
	if ref.User.AadObjectID != "" {
		r.data[ref.User.AadObjectID] = session
		if ref.Conversation.TenantID != "" {
			keyAAD := fmt.Sprintf("%s_%s", ref.Conversation.TenantID, ref.User.AadObjectID)
			r.data[keyAAD] = session
		}
	}

	r.latest = session
}

// Get retrieves a reference by tenant ID and user ID regardless of subscription
func (r *InMemoryConversationRepository) Get(tenantID, userID string) (*schema.ConversationReference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session := r.findSession(tenantID, userID)
	if session == nil {
		return nil, fmt.Errorf("conversation reference not found for tenant: %q, user: %q", tenantID, userID)
	}
	return &session.Reference, nil
}

// GetSubscribed retrieves a reference by tenant ID and user ID ONLY if subscribed
func (r *InMemoryConversationRepository) GetSubscribed(tenantID, userID string) (*schema.ConversationReference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session := r.findSession(tenantID, userID)
	if session == nil || !session.IsSubscribed {
		return nil, fmt.Errorf("user %q (tenant %q) is not subscribed to proactive alerts", userID, tenantID)
	}
	return &session.Reference, nil
}

func (r *InMemoryConversationRepository) findSession(tenantID, userID string) *StoredSession {
	if tenantID != "" && userID != "" {
		key := fmt.Sprintf("%s_%s", tenantID, userID)
		if session, exists := r.data[key]; exists {
			return session
		}
	}
	if userID != "" {
		if session, exists := r.data[userID]; exists {
			return session
		}
	}
	return nil
}

// IsSubscribed checks if a user is subscribed
func (r *InMemoryConversationRepository) IsSubscribed(tenantID, userID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session := r.findSession(tenantID, userID)
	return session != nil && session.IsSubscribed
}

// GetLatest retrieves the most recently stored conversation reference
func (r *InMemoryConversationRepository) GetLatest() (*schema.ConversationReference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.latest == nil || r.latest.Reference.ServiceURL == "" {
		return nil, fmt.Errorf("no conversation reference available yet")
	}
	return &r.latest.Reference, nil
}

// GetLatestSubscribed retrieves the most recently stored reference that is subscribed
func (r *InMemoryConversationRepository) GetLatestSubscribed() (*schema.ConversationReference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.latest != nil && r.latest.IsSubscribed && r.latest.Reference.ServiceURL != "" {
		return &r.latest.Reference, nil
	}

	// Search for any subscribed session in data
	for _, session := range r.data {
		if session.IsSubscribed && session.Reference.ServiceURL != "" {
			return &session.Reference, nil
		}
	}

	return nil, fmt.Errorf("no subscribed conversation reference available")
}

// GetAllSessions returns all stored active sessions with metadata
func (r *InMemoryConversationRepository) GetAllSessions() ([]models.SessionResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]bool)
	sessions := make([]models.SessionResponse, 0)
	for _, session := range r.data {
		ref := session.Reference
		key := ref.Conversation.ID
		if key == "" {
			key = fmt.Sprintf("%s_%s", ref.Conversation.TenantID, ref.User.ID)
		}
		if !seen[key] {
			seen[key] = true
			sessions = append(sessions, models.SessionResponse{
				TenantID:       ref.Conversation.TenantID,
				UserID:         ref.User.ID,
				AadObjectID:    ref.User.AadObjectID,
				UserName:       ref.User.Name,
				ConversationID: ref.Conversation.ID,
				ServiceURL:     ref.ServiceURL,
				IsSubscribed:   session.IsSubscribed,
			})
		}
	}
	return sessions, nil
}

// GetAll returns all stored conversation references deduplicated
func (r *InMemoryConversationRepository) GetAll() ([]schema.ConversationReference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]bool)
	refs := make([]schema.ConversationReference, 0)
	for _, session := range r.data {
		ref := session.Reference
		key := ref.Conversation.ID
		if key == "" {
			key = fmt.Sprintf("%s_%s", ref.Conversation.TenantID, ref.User.ID)
		}
		if !seen[key] {
			seen[key] = true
			refs = append(refs, ref)
		}
	}
	return refs, nil
}
