package application_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/adapter/memory"
	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/application/apptest"
	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/application/command"
	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/application/query"
	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/domain"
)

type testApp struct {
	clock  *apptest.Clock
	create command.CreateQuoteHandler
	accept command.AcceptQuoteHandler
	get    query.GetQuoteHandler
	list   query.ListQuotesHandler
}

func newTestApp() testApp {
	quotes := memory.NewQuoteRepository()
	clock := apptest.NewClock(time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC))
	ids := &apptest.SequentialIDs{}
	return testApp{
		clock:  clock,
		create: command.NewCreateQuoteHandler(quotes, clock, ids),
		accept: command.NewAcceptQuoteHandler(quotes, clock),
		get:    query.NewGetQuoteHandler(quotes, clock),
		list:   query.NewListQuotesHandler(quotes, clock),
	}
}

var validRequest = command.CreateQuote{Product: "die-cut", WidthInches: 3, HeightInches: 3, Quantity: 100}

func TestCreateThenGetQuote(t *testing.T) {
	app := newTestApp()
	ctx := context.Background()

	id, err := app.create.Handle(ctx, validRequest)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	view, err := app.get.Handle(ctx, id)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if view.TotalCents != 8_100 || view.TotalDisplay != "$81.00" {
		t.Errorf("total = %d (%s), want 8100 ($81.00)", view.TotalCents, view.TotalDisplay)
	}
	if view.Status != string(domain.QuoteStatusPending) {
		t.Errorf("status = %s, want pending", view.Status)
	}
}

func TestCreateQuoteRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*command.CreateQuote)
		wantErr error
	}{
		{"unknown product", func(c *command.CreateQuote) { c.Product = "glitter" }, domain.ErrUnknownProduct},
		{"zero width", func(c *command.CreateQuote) { c.WidthInches = 0 }, domain.ErrInvalidSize},
		{"zero quantity", func(c *command.CreateQuote) { c.Quantity = 0 }, domain.ErrInvalidQuantity},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := validRequest
			tt.mutate(&request)

			_, err := newTestApp().create.Handle(context.Background(), request)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestAcceptQuote(t *testing.T) {
	ctx := context.Background()

	t.Run("pending quote becomes accepted", func(t *testing.T) {
		app := newTestApp()
		id, _ := app.create.Handle(ctx, validRequest)

		if err := app.accept.Handle(ctx, id); err != nil {
			t.Fatalf("accept: %v", err)
		}
		view, _ := app.get.Handle(ctx, id)
		if view.Status != string(domain.QuoteStatusAccepted) || view.AcceptedAt == nil {
			t.Errorf("status = %s, acceptedAt = %v; want accepted with timestamp", view.Status, view.AcceptedAt)
		}
	})

	t.Run("expired quote is rejected", func(t *testing.T) {
		app := newTestApp()
		id, _ := app.create.Handle(ctx, validRequest)
		app.clock.Advance(domain.QuoteValidity)

		if err := app.accept.Handle(ctx, id); !errors.Is(err, domain.ErrQuoteExpired) {
			t.Errorf("err = %v, want ErrQuoteExpired", err)
		}
	})

	t.Run("missing quote", func(t *testing.T) {
		if err := newTestApp().accept.Handle(ctx, "qt_missing"); !errors.Is(err, domain.ErrQuoteNotFound) {
			t.Errorf("err = %v, want ErrQuoteNotFound", err)
		}
	})

	t.Run("concurrent accepts let exactly one win", func(t *testing.T) {
		app := newTestApp()
		id, _ := app.create.Handle(ctx, validRequest)

		const attempts = 50
		var wg sync.WaitGroup
		results := make(chan error, attempts)
		for range attempts {
			wg.Add(1)
			go func() {
				defer wg.Done()
				results <- app.accept.Handle(ctx, id)
			}()
		}
		wg.Wait()
		close(results)

		succeeded := 0
		for err := range results {
			switch {
			case err == nil:
				succeeded++
			case errors.Is(err, domain.ErrConcurrentUpdate), errors.Is(err, domain.ErrQuoteAlreadyAccepted):
				// expected for every loser
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}
		if succeeded != 1 {
			t.Errorf("%d accepts succeeded, want exactly 1", succeeded)
		}
	})
}

func TestListQuotes(t *testing.T) {
	app := newTestApp()
	ctx := context.Background()
	var ids []string
	for range 3 {
		id, _ := app.create.Handle(ctx, validRequest)
		ids = append(ids, id)
		app.clock.Advance(time.Minute)
	}

	t.Run("newest first", func(t *testing.T) {
		views, err := app.list.Handle(ctx, 0)
		if err != nil {
			t.Fatal(err)
		}
		if len(views) != 3 || views[0].ID != ids[2] || views[2].ID != ids[0] {
			t.Errorf("unexpected order: %v", viewIDs(views))
		}
	})

	t.Run("limit is respected", func(t *testing.T) {
		views, _ := app.list.Handle(ctx, 2)
		if len(views) != 2 {
			t.Errorf("len = %d, want 2", len(views))
		}
	})
}

func viewIDs(views []query.QuoteView) []string {
	ids := make([]string, 0, len(views))
	for _, view := range views {
		ids = append(ids, view.ID)
	}
	return ids
}
