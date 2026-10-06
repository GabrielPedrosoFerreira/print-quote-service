// Package http exposes the use cases over a JSON API. It only translates:
// HTTP in, command/query call, HTTP out. No business rules live here.
package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/application/command"
	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/application/query"
	"github.com/GabrielPedrosoFerreira/print-quote-service/internal/domain"
)

const maxRequestBodyBytes = 1 << 20 // 1 MiB

// UseCases groups the handlers the API depends on.
type UseCases struct {
	CreateQuote command.CreateQuoteHandler
	AcceptQuote command.AcceptQuoteHandler
	GetQuote    query.GetQuoteHandler
	ListQuotes  query.ListQuotesHandler
}

// NewRouter wires routes using the method and path patterns of net/http (Go 1.22+).
func NewRouter(useCases UseCases, logger *slog.Logger) http.Handler {
	api := quoteAPI{useCases: useCases, logger: logger}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /v1/quotes", api.createQuote)
	mux.HandleFunc("GET /v1/quotes", api.listQuotes)
	mux.HandleFunc("GET /v1/quotes/{id}", api.getQuote)
	mux.HandleFunc("POST /v1/quotes/{id}/accept", api.acceptQuote)
	return mux
}

type quoteAPI struct {
	useCases UseCases
	logger   *slog.Logger
}

type createQuoteRequest struct {
	Product      string `json:"product"`
	WidthInches  int    `json:"widthInches"`
	HeightInches int    `json:"heightInches"`
	Quantity     int    `json:"quantity"`
}

func (a quoteAPI) createQuote(w http.ResponseWriter, r *http.Request) {
	var req createQuoteRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxRequestBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeProblem(w, http.StatusBadRequest, "Malformed request body", err.Error())
		return
	}

	id, err := a.useCases.CreateQuote.Handle(r.Context(), command.CreateQuote(req))
	if err != nil {
		a.writeError(w, r, err)
		return
	}

	view, err := a.useCases.GetQuote.Handle(r.Context(), id)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	w.Header().Set("Location", "/v1/quotes/"+id)
	writeJSON(w, http.StatusCreated, view)
}

func (a quoteAPI) getQuote(w http.ResponseWriter, r *http.Request) {
	view, err := a.useCases.GetQuote.Handle(r.Context(), r.PathValue("id"))
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func (a quoteAPI) listQuotes(w http.ResponseWriter, r *http.Request) {
	limit := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			writeProblem(w, http.StatusBadRequest, "Invalid limit", "limit must be an integer")
			return
		}
		limit = parsed
	}

	views, err := a.useCases.ListQuotes.Handle(r.Context(), limit)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": views})
}

func (a quoteAPI) acceptQuote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := a.useCases.AcceptQuote.Handle(r.Context(), id); err != nil {
		a.writeError(w, r, err)
		return
	}

	view, err := a.useCases.GetQuote.Handle(r.Context(), id)
	if err != nil {
		a.writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

// writeError maps domain errors to HTTP status codes in one place.
// Unknown errors become a generic 500 so internal details never reach the client.
func (a quoteAPI) writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, domain.ErrUnknownProduct),
		errors.Is(err, domain.ErrInvalidSize),
		errors.Is(err, domain.ErrInvalidQuantity):
		writeProblem(w, http.StatusUnprocessableEntity, "Invalid quote request", err.Error())
	case errors.Is(err, domain.ErrQuoteNotFound):
		writeProblem(w, http.StatusNotFound, "Quote not found", "")
	case errors.Is(err, domain.ErrQuoteExpired),
		errors.Is(err, domain.ErrQuoteAlreadyAccepted),
		errors.Is(err, domain.ErrConcurrentUpdate):
		writeProblem(w, http.StatusConflict, "Quote cannot be accepted", err.Error())
	default:
		a.logger.ErrorContext(r.Context(), "unhandled error", "method", r.Method, "path", r.URL.Path, "error", err)
		writeProblem(w, http.StatusInternalServerError, "Internal server error", "")
	}
}

// problem follows RFC 9457 (Problem Details for HTTP APIs).
type problem struct {
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

func writeProblem(w http.ResponseWriter, status int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(problem{Title: title, Status: status, Detail: detail})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
