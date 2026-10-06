package domain

import "time"

// QuoteValidity is how long a customer has to accept a quote at its price.
const QuoteValidity = 7 * 24 * time.Hour

// QuoteStatus is derived from stored state plus the current time.
type QuoteStatus string

const (
	QuoteStatusPending  QuoteStatus = "pending"
	QuoteStatusAccepted QuoteStatus = "accepted"
	QuoteStatusExpired  QuoteStatus = "expired"
)

// Quote is a price offered for a sticker order, valid until ExpiresAt.
//
// Expiration is not stored: a pending quote becomes expired simply because
// time passes, so no background job is needed to flip a flag.
type Quote struct {
	ID         string
	Spec       StickerSpec
	Price      Price
	CreatedAt  time.Time
	ExpiresAt  time.Time
	AcceptedAt *time.Time

	// Version supports optimistic concurrency: every state change
	// increments it, and repositories reject a stale write.
	Version int
}

// NewQuote prices the spec and opens the validity window at now.
func NewQuote(id string, spec StickerSpec, now time.Time) Quote {
	return Quote{
		ID:        id,
		Spec:      spec,
		Price:     PriceFor(spec),
		CreatedAt: now,
		ExpiresAt: now.Add(QuoteValidity),
		Version:   1,
	}
}

// StatusAt reports the quote status at a given instant.
func (q Quote) StatusAt(now time.Time) QuoteStatus {
	switch {
	case q.AcceptedAt != nil:
		return QuoteStatusAccepted
	case !now.Before(q.ExpiresAt):
		return QuoteStatusExpired
	default:
		return QuoteStatusPending
	}
}

// Accept locks in the quoted price. It returns a new value instead of
// mutating the receiver, so a failed save never leaves a half-applied change.
func (q Quote) Accept(now time.Time) (Quote, error) {
	switch q.StatusAt(now) {
	case QuoteStatusAccepted:
		return Quote{}, ErrQuoteAlreadyAccepted
	case QuoteStatusExpired:
		return Quote{}, ErrQuoteExpired
	}

	acceptedAt := now
	q.AcceptedAt = &acceptedAt
	q.Version++
	return q, nil
}
