package domain

import "errors"

// Validation errors: the request itself is wrong and retrying it will not help.
var (
	ErrUnknownProduct  = errors.New("unknown product")
	ErrInvalidSize     = errors.New("invalid size")
	ErrInvalidQuantity = errors.New("invalid quantity")
)

// State errors: the request is well formed but conflicts with the quote's current state.
var (
	ErrQuoteNotFound        = errors.New("quote not found")
	ErrQuoteExpired         = errors.New("quote expired")
	ErrQuoteAlreadyAccepted = errors.New("quote already accepted")
	ErrConcurrentUpdate     = errors.New("quote was modified concurrently")
)
