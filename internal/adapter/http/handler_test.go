package http_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	httpadapter "github.com/GabrielPedrosoFerreira/print-quote-service/internal/adapter/http"
	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/adapter/memory"
	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/application/apptest"
	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/application/command"
	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/application/query"
)

// newServer wires the real router, use cases and repository; only time and IDs are faked.
func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	quotes := memory.NewQuoteRepository()
	clock := apptest.NewClock(time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC))
	ids := &apptest.SequentialIDs{}
	router := httpadapter.NewRouter(httpadapter.UseCases{
		CreateQuote: command.NewCreateQuoteHandler(quotes, clock, ids),
		AcceptQuote: command.NewAcceptQuoteHandler(quotes, clock),
		GetQuote:    query.NewGetQuoteHandler(quotes, clock),
		ListQuotes:  query.NewListQuotesHandler(quotes, clock),
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)
	return server
}

func do(t *testing.T, method, url, body string) (*http.Response, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var decoded map[string]any
	if resp.ContentLength != 0 {
		_ = json.NewDecoder(resp.Body).Decode(&decoded)
	}
	return resp, decoded
}

const validBody = `{"product":"circle","widthInches":2,"heightInches":2,"quantity":100}`

func TestQuoteLifecycle(t *testing.T) {
	server := newServer(t)

	resp, created := do(t, http.MethodPost, server.URL+"/v1/quotes", validBody)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want 201", resp.StatusCode)
	}
	if got := resp.Header.Get("Location"); got != "/v1/quotes/qt_1" {
		t.Errorf("Location = %q, want /v1/quotes/qt_1", got)
	}
	if created["totalDisplay"] != "$30.00" {
		t.Errorf("totalDisplay = %v, want $30.00", created["totalDisplay"])
	}

	resp, accepted := do(t, http.MethodPost, server.URL+"/v1/quotes/qt_1/accept", "")
	if resp.StatusCode != http.StatusOK || accepted["status"] != "accepted" {
		t.Fatalf("accept status = %d body = %v, want 200 accepted", resp.StatusCode, accepted)
	}

	resp, problem := do(t, http.MethodPost, server.URL+"/v1/quotes/qt_1/accept", "")
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("second accept status = %d, want 409", resp.StatusCode)
	}
	if resp.Header.Get("Content-Type") != "application/problem+json" || problem["title"] == nil {
		t.Errorf("expected a problem+json body, got %v", problem)
	}
}

func TestErrorMapping(t *testing.T) {
	server := newServer(t)

	tests := []struct {
		name   string
		method string
		path   string
		body   string
		want   int
	}{
		{"malformed json", http.MethodPost, "/v1/quotes", `{"product":`, http.StatusBadRequest},
		{"unknown field", http.MethodPost, "/v1/quotes", `{"product":"circle","color":"red"}`, http.StatusBadRequest},
		{"invalid size", http.MethodPost, "/v1/quotes", `{"product":"circle","widthInches":40,"heightInches":2,"quantity":1}`, http.StatusUnprocessableEntity},
		{"unknown product", http.MethodPost, "/v1/quotes", `{"product":"glitter","widthInches":2,"heightInches":2,"quantity":1}`, http.StatusUnprocessableEntity},
		{"missing quote", http.MethodGet, "/v1/quotes/qt_missing", "", http.StatusNotFound},
		{"accept missing quote", http.MethodPost, "/v1/quotes/qt_missing/accept", "", http.StatusNotFound},
		{"non numeric limit", http.MethodGet, "/v1/quotes?limit=abc", "", http.StatusBadRequest},
		{"wrong method", http.MethodDelete, "/v1/quotes/qt_1", "", http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := do(t, tt.method, server.URL+tt.path, tt.body)
			if resp.StatusCode != tt.want {
				t.Errorf("status = %d, want %d", resp.StatusCode, tt.want)
			}
		})
	}
}

func TestListQuotesEndpoint(t *testing.T) {
	server := newServer(t)
	for range 3 {
		do(t, http.MethodPost, server.URL+"/v1/quotes", validBody)
	}

	resp, body := do(t, http.MethodGet, server.URL+"/v1/quotes?limit=2", "")

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if items, _ := body["items"].([]any); len(items) != 2 {
		t.Errorf("items = %d, want 2", len(items))
	}
}

func TestHealthz(t *testing.T) {
	resp, _ := do(t, http.MethodGet, newServer(t).URL+"/healthz", "")
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204", resp.StatusCode)
	}
}
