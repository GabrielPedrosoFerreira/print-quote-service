// Package command holds use cases that change state.
package command

import (
	"context"
	"fmt"

	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/application/port"
	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/domain"
)

// CreateQuote is the raw input for a new quote, before validation.
type CreateQuote struct {
	Product      string
	WidthInches  int
	HeightInches int
	Quantity     int
}

// CreateQuoteHandler validates the input, prices it and stores the quote.
type CreateQuoteHandler struct {
	quotes port.QuoteRepository
	clock  port.Clock
	ids    port.IDGenerator
}

func NewCreateQuoteHandler(quotes port.QuoteRepository, clock port.Clock, ids port.IDGenerator) CreateQuoteHandler {
	return CreateQuoteHandler{quotes: quotes, clock: clock, ids: ids}
}

// Handle returns the ID of the created quote. Commands return only an ID;
// callers read the full quote through a query, keeping writes and reads separate.
func (h CreateQuoteHandler) Handle(ctx context.Context, cmd CreateQuote) (string, error) {
	spec, err := parseSpec(cmd)
	if err != nil {
		return "", err
	}

	quote := domain.NewQuote(h.ids.NewID(), spec, h.clock.Now())
	if err := h.quotes.Create(ctx, quote); err != nil {
		return "", fmt.Errorf("create quote: %w", err)
	}
	return quote.ID, nil
}

func parseSpec(cmd CreateQuote) (domain.StickerSpec, error) {
	product, err := domain.ParseProduct(cmd.Product)
	if err != nil {
		return domain.StickerSpec{}, err
	}
	size, err := domain.NewSize(cmd.WidthInches, cmd.HeightInches)
	if err != nil {
		return domain.StickerSpec{}, err
	}
	quantity, err := domain.NewQuantity(cmd.Quantity)
	if err != nil {
		return domain.StickerSpec{}, err
	}
	return domain.StickerSpec{Product: product, Size: size, Quantity: quantity}, nil
}
