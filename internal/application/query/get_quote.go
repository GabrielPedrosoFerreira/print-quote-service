package query

import (
	"context"

	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/application/port"
)

// GetQuoteHandler returns one quote by ID.
type GetQuoteHandler struct {
	quotes port.QuoteRepository
	clock  port.Clock
}

func NewGetQuoteHandler(quotes port.QuoteRepository, clock port.Clock) GetQuoteHandler {
	return GetQuoteHandler{quotes: quotes, clock: clock}
}

func (h GetQuoteHandler) Handle(ctx context.Context, quoteID string) (QuoteView, error) {
	quote, err := h.quotes.FindByID(ctx, quoteID)
	if err != nil {
		return QuoteView{}, err
	}
	return newQuoteView(quote, h.clock.Now()), nil
}
