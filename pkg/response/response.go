package response

import (
	"encoding/json"
	"net/http"
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
