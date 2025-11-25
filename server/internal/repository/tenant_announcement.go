package repository

import (
	"context"
	"sync"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/domain"
)

// TenantAnnouncementRepository defines the interface for tenant announcements
type TenantAnnouncementRepository interface {
	// CRUD operations
	Create(ctx context.Context, announcement *domain.TenantAnnouncement) error
	GetByID(ctx context.Context, id string) (*domain.TenantAnnouncement, error)
	Update(ctx context.Context, announcement *domain.TenantAnnouncement) error
	Delete(ctx context.Context, id string) error

	// List operations
	ListByTenant(ctx context.Context, tenantID string, pinnedOnly bool, limit int, cursor string) ([]*domain.TenantAnnouncement, string, error)

	// Pin operations
	SetPinned(ctx context.Context, id string, pinned bool) error
}

// InMemoryTenantAnnouncementRepository is an in-memory implementation
type InMemoryTenantAnnouncementRepository struct {
	mu            sync.RWMutex
	announcements map[string]*domain.TenantAnnouncement // id -> announcement
}

// NewInMemoryTenantAnnouncementRepository creates a new in-memory repository
func NewInMemoryTenantAnnouncementRepository() *InMemoryTenantAnnouncementRepository {
	return &InMemoryTenantAnnouncementRepository{
		announcements: make(map[string]*domain.TenantAnnouncement),
	}
}

func (r *InMemoryTenantAnnouncementRepository) Create(ctx context.Context, announcement *domain.TenantAnnouncement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if announcement.ID == "" {
		announcement.ID = domain.GenerateAnnouncementID()
	}
	announcement.CreatedAt = time.Now()
	announcement.UpdatedAt = time.Now()
	r.announcements[announcement.ID] = announcement
	return nil
}

func (r *InMemoryTenantAnnouncementRepository) GetByID(ctx context.Context, id string) (*domain.TenantAnnouncement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ann, exists := r.announcements[id]
	if !exists {
		return nil, domain.NewError(domain.ErrNotFound, "Announcement not found", nil)
	}
	return ann, nil
}

func (r *InMemoryTenantAnnouncementRepository) Update(ctx context.Context, announcement *domain.TenantAnnouncement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.announcements[announcement.ID]; !exists {
		return domain.NewError(domain.ErrNotFound, "Announcement not found", nil)
	}
	announcement.UpdatedAt = time.Now()
	r.announcements[announcement.ID] = announcement
	return nil
}

func (r *InMemoryTenantAnnouncementRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.announcements[id]; !exists {
		return domain.NewError(domain.ErrNotFound, "Announcement not found", nil)
	}
	delete(r.announcements, id)
	return nil
}

func (r *InMemoryTenantAnnouncementRepository) ListByTenant(ctx context.Context, tenantID string, pinnedOnly bool, limit int, cursor string) ([]*domain.TenantAnnouncement, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*domain.TenantAnnouncement
	for _, a := range r.announcements {
		if a.TenantID == tenantID {
			if !pinnedOnly || a.IsPinned {
				result = append(result, a)
			}
		}
	}

	// Sort by created_at desc (pinned first)
	// Simple bubble sort for in-memory (production uses SQL ORDER BY)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			// Pinned items first, then by date
			if (!result[i].IsPinned && result[j].IsPinned) ||
				(result[i].IsPinned == result[j].IsPinned && result[i].CreatedAt.Before(result[j].CreatedAt)) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result, "", nil
}

func (r *InMemoryTenantAnnouncementRepository) SetPinned(ctx context.Context, id string, pinned bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ann, exists := r.announcements[id]
	if !exists {
		return domain.NewError(domain.ErrNotFound, "Announcement not found", nil)
	}
	ann.IsPinned = pinned
	ann.UpdatedAt = time.Now()
	return nil
}
