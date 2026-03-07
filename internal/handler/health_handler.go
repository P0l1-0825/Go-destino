package handler

import (
	"net/http"

	"github.com/P0l1-0825/Go-destino/pkg/response"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "godestino-api",
		"version": "1.0.0",
	})
}

func ReadyCheck(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{
		"status": "ready",
	})
}
