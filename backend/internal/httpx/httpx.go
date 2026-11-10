package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Go Concept - Interfaces:
// `http.ResponseWriter` is an interface. Any type that implements `Header()`, `Write()`,
// and `WriteHeader()` satisfies it. In Express, you call `res.status().json()`; in Go,
// you write headers and encode data into the response writer stream.

// Response is the standard JSON error envelope required across the Spendle API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// WriteJSON sets the Content-Type header to application/json, writes the HTTP status code,
// and encodes the provided data as JSON to the response writer.
// Go Concept - 'any' type:
// `any` is an alias for `interface{}`, meaning it can accept any data type (like 'unknown' in TypeScript).
func WriteJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		return fmt.Errorf("httpx: failed to encode JSON response: %w", err)
	}
	return nil
}

// WriteError sends a standard JSON error response with the format: {"error": "message"}.
func WriteError(w http.ResponseWriter, status int, message string) error {
	return WriteJSON(w, status, ErrorResponse{Error: message})
}

// DecodeJSON decodes the JSON request body into the target pointer `dst`.
// It limits request payload size to 1MB to guard against memory exhaustion.
func DecodeJSON(r *http.Request, dst any) error {
	// Limit request body to 1MB
	maxBytes := int64(1048576)
	r.Body = http.MaxBytesReader(nil, r.Body, maxBytes)

	decoder := json.NewDecoder(r.Body)
	// Disallow unknown fields to catch typo bugs early
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("malformed JSON at byte offset %d", syntaxError.Offset)
		case errors.As(err, &unmarshalTypeError):
			return fmt.Errorf("invalid type for field %q (expected %s)", unmarshalTypeError.Field, unmarshalTypeError.Type)
		case errors.Is(err, io.EOF):
			return errors.New("request body cannot be empty")
		default:
			return fmt.Errorf("cannot parse JSON body: %w", err)
		}
	}

	// Ensure there is only one JSON object in the stream
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain only a single JSON object")
	}

	return nil
}

