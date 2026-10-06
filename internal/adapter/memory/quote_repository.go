// Package memory is an in-process QuoteRepository. It keeps the service
// runnable with zero infrastructure; a Postgres adapter would implement the
// same port with the version check done in a single UPDATE ... WHERE version = $n.
package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/domain"
)

type QuoteRepository struct {
	mu     sync.RWMutex
	quotes map[string]domain.Quote
}

func NewQuoteRepository() *QuoteRepository {
	return &QuoteRepository{quotes: make(map[string]domain.Quote)}
}

func (r *QuoteRepository) Create(_ context.Context, quote domain.Quote) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.quotes[quote.ID]; exists {
		return fmt.Errorf("quote %s already exists", quote.ID)
	}
	r.quotes[quote.ID] = quote
	return nil
}

func (r *QuoteRepository) Update(_ context.Context, quote domain.Quote) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	stored, exists := r.quotes[quote.ID]
	if !exists {
		return domain.ErrQuoteNotFound
	}
	if stored.Version != quote.Version-1 {
		return domain.ErrConcurrentUpdate
	}
	r.quotes[quote.ID] = quote
	return nil
}

func (r *QuoteRepository) FindByID(_ context.Context, id string) (domain.Quote, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	quote, exists := r.quotes[id]
	if !exists {
		return domain.Quote{}, domain.ErrQuoteNotFound
	}
	return quote, nil
}

func (r *QuoteRepository) ListNewestFirst(_ context.Context, limit int) ([]domain.Quote, error) {
	r.mu.RLock()
	quotes := make([]domain.Quote, 0, len(r.quotes))
	for _, quote := range r.quotes {
		quotes = append(quotes, quote)
	}
	r.mu.RUnlock()

	sort.Slice(quotes, func(i, j int) bool {
		if quotes[i].CreatedAt.Equal(quotes[j].CreatedAt) {
			return quotes[i].ID > quotes[j].ID // stable order for equal timestamps
		}
		return quotes[i].CreatedAt.After(quotes[j].CreatedAt)
	})

	if len(quotes) > limit {
		quotes = quotes[:limit]
	}
	return quotes, nil
}
