package command

import (
	"context"
	"fmt"

	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/application/port"
)

// AcceptQuoteHandler locks in the price of a pending quote.
type AcceptQuoteHandler struct {
	quotes port.QuoteRepository
	clock  port.Clock
}

func NewAcceptQuoteHandler(quotes port.QuoteRepository, clock port.Clock) AcceptQuoteHandler {
	return AcceptQuoteHandler{quotes: quotes, clock: clock}
}

// Handle accepts the quote. When two requests race, the repository's version
// check lets exactly one win; the other gets domain.ErrConcurrentUpdate.
func (h AcceptQuoteHandler) Handle(ctx context.Context, quoteID string) error {
	quote, err := h.quotes.FindByID(ctx, quoteID)
	if err != nil {
		return err
	}

	accepted, err := quote.Accept(h.clock.Now())
	if err != nil {
		return err
	}

	if err := h.quotes.Update(ctx, accepted); err != nil {
		return fmt.Errorf("accept quote %s: %w", quoteID, err)
	}
	return nil
}
