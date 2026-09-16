package repository

import (
	"fmt"
	"sync"

	"github.com/infracloudio/msbotbuilder-go/schema"
)

// ConversationRepository interface for storing and retrieving conversation references
type ConversationRepository interface {
	Save(ref schema.ConversationReference) error
	Get(tenantID, userID string) (*schema.ConversationReference, error)
	GetLatest() (*schema.ConversationReference, error)
	GetAll() ([]schema.ConversationReference, error)
}

// InMemoryConversationRepository is a thread-safe in-memory store for conversation references
type InMemoryConversationRepository struct {
	mu     sync.RWMutex
	data   map[string]schema.ConversationReference
	latest *schema.ConversationReference
}

// NewInMemoryConversationRepository creates a new InMemoryConversationRepository instance
func NewInMemoryConversationRepository() *InMemoryConversationRepository {
	return &InMemoryConversationRepository{
		data: make(map[string]schema.ConversationReference),
	}
}

// Save stores or updates a conversation reference
func (r *InMemoryConversationRepository) Save(ref schema.ConversationReference) error {
	if ref.ServiceURL == "" {
		return fmt.Errorf("invalid reference: empty ServiceURL")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := ref.Conversation.ID
	r.data[key] = ref

	if ref.User.ID != "" {
		r.data[ref.User.ID] = ref
		if ref.Conversation.TenantID != "" {
			keyUser := fmt.Sprintf("%s_%s", ref.Conversation.TenantID, ref.User.ID)
			r.data[keyUser] = ref
		}
	}
	if ref.User.AadObjectID != "" {
		r.data[ref.User.AadObjectID] = ref
		if ref.Conversation.TenantID != "" {
			keyAAD := fmt.Sprintf("%s_%s", ref.Conversation.TenantID, ref.User.AadObjectID)
			r.data[keyAAD] = ref
		}
	}

	r.latest = &ref
	return nil
}

// Get retrieves a reference by tenant ID and user ID
func (r *InMemoryConversationRepository) Get(tenantID, userID string) (*schema.ConversationReference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 1. Try tenantID_userID key if tenantID is provided
	if tenantID != "" && userID != "" {
		key := fmt.Sprintf("%s_%s", tenantID, userID)
		if ref, exists := r.data[key]; exists {
			return &ref, nil
		}
	}

	// 2. Try direct userID lookup (works for WebChat/Emulator or direct user IDs)
	if userID != "" {
		if ref, exists := r.data[userID]; exists {
			return &ref, nil
		}
	}

	return nil, fmt.Errorf("conversation reference not found for tenant: %q, user: %q", tenantID, userID)
}

// GetLatest retrieves the most recently stored conversation reference
func (r *InMemoryConversationRepository) GetLatest() (*schema.ConversationReference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.latest == nil || r.latest.ServiceURL == "" {
		return nil, fmt.Errorf("no conversation reference available yet")
	}
	return r.latest, nil
}

// GetAll returns all stored conversation references deduplicated
func (r *InMemoryConversationRepository) GetAll() ([]schema.ConversationReference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]bool)
	refs := make([]schema.ConversationReference, 0)
	for _, ref := range r.data {
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
