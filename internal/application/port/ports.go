// Package port declares what the application needs from the outside world.
// Adapters implement these interfaces; the application never imports an adapter.
package port

import (
	"context"
	"time"

	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/domain"
)

// QuoteRepository persists quotes.
type QuoteRepository interface {
	// Create stores a new quote.
	Create(ctx context.Context, quote domain.Quote) error
	// Update stores a changed quote. It returns domain.ErrConcurrentUpdate when
	// the stored version is not quote.Version-1, meaning someone else wrote first.
	Update(ctx context.Context, quote domain.Quote) error
	// FindByID returns domain.ErrQuoteNotFound when no quote has the given ID.
	FindByID(ctx context.Context, id string) (domain.Quote, error)
	// ListNewestFirst returns up to limit quotes ordered by CreatedAt descending.
	ListNewestFirst(ctx context.Context, limit int) ([]domain.Quote, error)
}

// Clock makes time an explicit dependency so expiration rules are testable.
type Clock interface {
	Now() time.Time
}

// IDGenerator produces unique quote IDs.
type IDGenerator interface {
	NewID() string
}
