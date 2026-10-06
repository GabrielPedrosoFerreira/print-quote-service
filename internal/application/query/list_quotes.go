package query

import (
	"context"

	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/application/port"
)

const (
	DefaultListLimit = 20
	MaxListLimit     = 100
)

// ListQuotesHandler returns the most recent quotes.
type ListQuotesHandler struct {
	quotes port.QuoteRepository
	clock  port.Clock
}

func NewListQuotesHandler(quotes port.QuoteRepository, clock port.Clock) ListQuotesHandler {
	return ListQuotesHandler{quotes: quotes, clock: clock}
}

// Handle clamps limit to [1, MaxListLimit]; zero or negative means the default.
func (h ListQuotesHandler) Handle(ctx context.Context, limit int) ([]QuoteView, error) {
	switch {
	case limit <= 0:
		limit = DefaultListLimit
	case limit > MaxListLimit:
		limit = MaxListLimit
	}

	quotes, err := h.quotes.ListNewestFirst(ctx, limit)
	if err != nil {
		return nil, err
	}

	now := h.clock.Now()
	views := make([]QuoteView, 0, len(quotes))
	for _, quote := range quotes {
		views = append(views, newQuoteView(quote, now))
	}
	return views, nil
}
