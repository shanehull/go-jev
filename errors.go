package jev

import (
	"errors"

	"github.com/shanehull/go-jev/internal"
)

// APIError represents a non-success response from the TypeSafe API.
// It is a type alias, so callers use errors.As.
type APIError = internal.APIError

// ErrNoQuestions is returned when a request has no questions.
var ErrNoQuestions = errors.New("jev: at least one question is required")
