package domain

import (
	"errors"
	"testing"
	"time"
)

var createdAt = time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)

func newTestQuote(t *testing.T) Quote {
	t.Helper()
	return NewQuote("qt_test", mustSpec(t, ProductDieCut, 3, 3, 100), createdAt)
}

func TestQuoteStatusAt(t *testing.T) {
	quote := newTestQuote(t)

	tests := []struct {
		name string
		at   time.Time
		want QuoteStatus
	}{
		{"just created", createdAt, QuoteStatusPending},
		{"one nanosecond before expiry", quote.ExpiresAt.Add(-time.Nanosecond), QuoteStatusPending},
		{"exactly at expiry", quote.ExpiresAt, QuoteStatusExpired},
		{"after expiry", quote.ExpiresAt.Add(time.Hour), QuoteStatusExpired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := quote.StatusAt(tt.at); got != tt.want {
				t.Errorf("StatusAt() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestQuoteAccept(t *testing.T) {
	t.Run("pending quote is accepted and versioned", func(t *testing.T) {
		quote := newTestQuote(t)
		acceptAt := createdAt.Add(24 * time.Hour)

		accepted, err := quote.Accept(acceptAt)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if accepted.StatusAt(acceptAt) != QuoteStatusAccepted {
			t.Errorf("status = %s, want accepted", accepted.StatusAt(acceptAt))
		}
		if accepted.Version != quote.Version+1 {
			t.Errorf("version = %d, want %d", accepted.Version, quote.Version+1)
		}
		if quote.AcceptedAt != nil {
			t.Error("Accept mutated the original quote")
		}
	})

	t.Run("accepted quote stays accepted after the validity window", func(t *testing.T) {
		accepted, err := newTestQuote(t).Accept(createdAt)
		if err != nil {
			t.Fatal(err)
		}
		if got := accepted.StatusAt(accepted.ExpiresAt.Add(time.Hour)); got != QuoteStatusAccepted {
			t.Errorf("status = %s, want accepted", got)
		}
	})

	t.Run("cannot accept twice", func(t *testing.T) {
		accepted, err := newTestQuote(t).Accept(createdAt)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := accepted.Accept(createdAt); !errors.Is(err, ErrQuoteAlreadyAccepted) {
			t.Errorf("err = %v, want ErrQuoteAlreadyAccepted", err)
		}
	})

	t.Run("cannot accept an expired quote", func(t *testing.T) {
		quote := newTestQuote(t)
		if _, err := quote.Accept(quote.ExpiresAt); !errors.Is(err, ErrQuoteExpired) {
			t.Errorf("err = %v, want ErrQuoteExpired", err)
		}
	})
}
