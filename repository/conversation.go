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
	if ref.User.ID != "" {
		key = fmt.Sprintf("%s_%s", ref.Conversation.TenantID, ref.User.ID)
	}

	r.data[key] = ref
	r.latest = &ref
	return nil
}

// Get retrieves a reference by tenant ID and user ID
func (r *InMemoryConversationRepository) Get(tenantID, userID string) (*schema.ConversationReference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	key := fmt.Sprintf("%s_%s", tenantID, userID)
	ref, exists := r.data[key]
	if !exists {
		return nil, fmt.Errorf("conversation reference not found for tenant: %s, user: %s", tenantID, userID)
	}
	return &ref, nil
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

// GetAll returns all stored conversation references
func (r *InMemoryConversationRepository) GetAll() ([]schema.ConversationReference, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	refs := make([]schema.ConversationReference, 0, len(r.data))
	for _, ref := range r.data {
		refs = append(refs, ref)
	}
	return refs, nil
}
