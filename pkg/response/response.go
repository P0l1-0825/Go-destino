package response

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	TotalCount int `json:"total_count,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: status >= 200 && status < 300,
		Data:    data,
	})
}

func JSONWithMeta(w http.ResponseWriter, status int, data interface{}, total, limit, offset int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	page := 1
	if limit > 0 {
		page = (offset / limit) + 1
	}
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
		Meta:    &Meta{Page: page, PerPage: limit, TotalCount: total},
	})
}

func Error(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error:   msg,
	})
}

// CachedJSON serialises data as JSON and adds HTTP caching headers.
//
// maxAge controls the Cache-Control max-age directive (how long CDN/browser
// may serve a stale copy without revalidating).  An ETag derived from the
// JSON payload is also set so clients can issue conditional GET requests
// (If-None-Match) and receive a 304 Not Modified when the data has not changed.
//
// Use this helper for reference data that changes infrequently:
// routes, airports, emergency numbers, concesiones list.
func CachedJSON(w http.ResponseWriter, r *http.Request, status int, data interface{}, maxAge time.Duration) {
	// Serialise first so we can compute the ETag before writing headers.
	payload, err := json.Marshal(APIResponse{
		Success: status >= 200 && status < 300,
		Data:    data,
	})
	if err != nil {
		Error(w, http.StatusInternalServerError, "serialization error")
		return
	}

	etag := fmt.Sprintf(`"%x"`, sha256.Sum256(payload))

	// Conditional GET support: return 304 if client already has current data.
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d, stale-while-revalidate=60", int(maxAge.Seconds())))
	w.Header().Set("ETag", etag)
	w.WriteHeader(status)
	w.Write(payload) //nolint:errcheck
}
