// Package response writes the JSON envelope shared by every API endpoint.
package response

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type Envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorItem `json:"error,omitempty"`
}

type ErrorItem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func write(w http.ResponseWriter, status int, body Envelope) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("failed to write response body: %v", err)
	}
}

func OK(w http.ResponseWriter, data any) {
	write(w, http.StatusOK, Envelope{Success: true, Data: data})
}

func Created(w http.ResponseWriter, data any) {
	write(w, http.StatusCreated, Envelope{Success: true, Data: data})
}

func Error(w http.ResponseWriter, status int, code, message string) {
	write(w, status, Envelope{Error: &ErrorItem{Code: code, Message: message}})
}

// OKWithETag sends a successful response that browsers must revalidate on every use. The ETag is a hash of the
// body, so an unchanged resource costs one small 304 instead of re-downloading base64 artwork, while edits show up immediately.
func OKWithETag(w http.ResponseWriter, r *http.Request, data any) {
	body, err := json.Marshal(Envelope{Success: true, Data: data})
	if err != nil {
		log.Printf("failed to encode response body: %v", err)
		Error(w, http.StatusInternalServerError, "ENCODE_FAILED", "Failed to encode response")
		return
	}

	sum := sha256.Sum256(body)
	etag := `"` + hex.EncodeToString(sum[:16]) + `"`
	w.Header().Set("ETag", etag)
	w.Header().Set("Cache-Control", "no-cache")

	if etagMatches(r.Header.Get("If-None-Match"), etag) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(append(body, '\n')); err != nil {
		log.Printf("failed to write response body: %v", err)
	}
}

// etagMatches implements weak comparison: a proxy may turn our strong ETag into a weak one (W/"...") when it compresses the body.
func etagMatches(header, etag string) bool {
	for _, candidate := range strings.Split(header, ",") {
		candidate = strings.TrimPrefix(strings.TrimSpace(candidate), "W/")
		if candidate == etag || candidate == "*" {
			return true
		}
	}
	return false
}
