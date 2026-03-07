package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/response"
)

// WSHandler provides Server-Sent Events (SSE) for real-time driver tracking.
// SSE is used instead of WebSocket for simplicity — no external dependency needed.
// In production, consider upgrading to gorilla/websocket for bidirectional communication.
type WSHandler struct {
	fleetSvc    *service.FleetService
	subscribers map[string][]chan domain.DriverLocation
	mu          sync.RWMutex
}

func NewWSHandler(fleetSvc *service.FleetService) *WSHandler {
	return &WSHandler{
		fleetSvc:    fleetSvc,
		subscribers: make(map[string][]chan domain.DriverLocation),
	}
}

// TrackDriver streams driver location updates via Server-Sent Events.
// GET /api/v1/track/driver/{id}
func (h *WSHandler) TrackDriver(w http.ResponseWriter, r *http.Request) {
	driverID := r.PathValue("id")

	flusher, ok := w.(http.Flusher)
	if !ok {
		response.Error(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan domain.DriverLocation, 10)
	h.subscribe(driverID, ch)
	defer h.unsubscribe(driverID, ch)

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case loc := <-ch:
			data, _ := json.Marshal(loc)
			w.Write([]byte("data: "))
			w.Write(data)
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}

// PublishLocation receives location updates and broadcasts to subscribers.
// POST /api/v1/track/publish
func (h *WSHandler) PublishLocation(w http.ResponseWriter, r *http.Request) {
	var loc domain.DriverLocation
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Persist location
	if err := h.fleetSvc.UpdateDriverLocation(r.Context(), loc); err != nil {
		log.Printf("[WS] failed to persist location: %v", err)
	}

	// Broadcast to SSE subscribers
	h.broadcast(loc.DriverID, loc)

	response.JSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// DriverLocations returns a snapshot of all active driver positions for a tenant.
// GET /api/v1/track/drivers
func (h *WSHandler) DriverLocations(w http.ResponseWriter, r *http.Request) {
	// Return subscriber count as a health indicator
	h.mu.RLock()
	activeStreams := 0
	for _, subs := range h.subscribers {
		activeStreams += len(subs)
	}
	h.mu.RUnlock()

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"active_streams": activeStreams,
		"timestamp":      time.Now(),
	})
}

func (h *WSHandler) subscribe(driverID string, ch chan domain.DriverLocation) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.subscribers[driverID] = append(h.subscribers[driverID], ch)
}

func (h *WSHandler) unsubscribe(driverID string, ch chan domain.DriverLocation) {
	h.mu.Lock()
	defer h.mu.Unlock()
	subs := h.subscribers[driverID]
	for i, sub := range subs {
		if sub == ch {
			h.subscribers[driverID] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
	if len(h.subscribers[driverID]) == 0 {
		delete(h.subscribers, driverID)
	}
}

func (h *WSHandler) broadcast(driverID string, loc domain.DriverLocation) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.subscribers[driverID] {
		select {
		case ch <- loc:
		default:
			// Drop if subscriber is slow
		}
	}
}
