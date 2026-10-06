// Package query holds read-only use cases. They return views shaped for
// API consumers, so the domain model can change without breaking clients.
package query

import (
	"time"

	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/domain"
)

// QuoteView is the read model of a quote.
type QuoteView struct {
	ID            string     `json:"id"`
	Product       string     `json:"product"`
	WidthInches   int        `json:"widthInches"`
	HeightInches  int        `json:"heightInches"`
	Quantity      int        `json:"quantity"`
	SubtotalCents int64      `json:"subtotalCents"`
	DiscountCents int64      `json:"discountCents"`
	TotalCents    int64      `json:"totalCents"`
	TotalDisplay  string     `json:"totalDisplay"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	ExpiresAt     time.Time  `json:"expiresAt"`
	AcceptedAt    *time.Time `json:"acceptedAt,omitempty"`
}

func newQuoteView(quote domain.Quote, now time.Time) QuoteView {
	return QuoteView{
		ID:            quote.ID,
		Product:       string(quote.Spec.Product),
		WidthInches:   quote.Spec.Size.WidthInches,
		HeightInches:  quote.Spec.Size.HeightInches,
		Quantity:      int(quote.Spec.Quantity),
		SubtotalCents: int64(quote.Price.Subtotal),
		DiscountCents: int64(quote.Price.Discount),
		TotalCents:    int64(quote.Price.Total),
		TotalDisplay:  quote.Price.Total.String(),
		Status:        string(quote.StatusAt(now)),
		CreatedAt:     quote.CreatedAt,
		ExpiresAt:     quote.ExpiresAt,
		AcceptedAt:    quote.AcceptedAt,
	}
}
