package router

import (
	"net/http"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/handler"
	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/service"
)

func New(
	authSvc *service.AuthService,
	authH *handler.AuthHandler,
	routeH *handler.RouteHandler,
	ticketH *handler.TicketHandler,
	bookingH *handler.BookingHandler,
	kioskH *handler.KioskHandler,
) http.Handler {
	mux := http.NewServeMux()

	// Health & readiness probes
	mux.HandleFunc("GET /health", handler.HealthCheck)
	mux.HandleFunc("GET /ready", handler.ReadyCheck)

	// Auth (public)
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)

	// Auth (protected)
	mux.Handle("GET /api/v1/auth/me", applyAuth(authSvc, http.HandlerFunc(authH.Me)))

	// Routes
	mux.Handle("POST /api/v1/routes", applyAuthPerm(authSvc, domain.PermSysSettingsEdit, http.HandlerFunc(routeH.Create)))
	mux.Handle("GET /api/v1/routes/{id}", applyAuth(authSvc, http.HandlerFunc(routeH.GetByID)))
	mux.Handle("GET /api/v1/routes", applyAuth(authSvc, http.HandlerFunc(routeH.List)))

	// Tickets (kiosk operations)
	mux.Handle("POST /api/v1/tickets/purchase", applyAuthPerm(authSvc, domain.PermKioskBookCreate, http.HandlerFunc(ticketH.Purchase)))
	mux.Handle("POST /api/v1/tickets/validate", applyAuth(authSvc, http.HandlerFunc(ticketH.Validate)))
	mux.Handle("GET /api/v1/tickets/{id}", applyAuth(authSvc, http.HandlerFunc(ticketH.GetByID)))

	// Bookings
	mux.Handle("POST /api/v1/bookings", applyAuthPerm(authSvc, domain.PermResCreateWeb, http.HandlerFunc(bookingH.Create)))
	mux.Handle("GET /api/v1/bookings", applyAuth(authSvc, http.HandlerFunc(bookingH.List)))
	mux.Handle("GET /api/v1/bookings/{id}", applyAuth(authSvc, http.HandlerFunc(bookingH.GetByID)))
	mux.Handle("GET /api/v1/bookings/number/{number}", applyAuth(authSvc, http.HandlerFunc(bookingH.GetByNumber)))
	mux.Handle("POST /api/v1/bookings/{id}/cancel", applyAuthPerm(authSvc, domain.PermResCancelOwn, http.HandlerFunc(bookingH.Cancel)))
	mux.Handle("PUT /api/v1/bookings/{id}/status", applyAuthPerm(authSvc, domain.PermResAssignDriver, http.HandlerFunc(bookingH.UpdateStatus)))
	mux.Handle("POST /api/v1/bookings/estimate", applyAuthPerm(authSvc, domain.PermResPriceEstimate, http.HandlerFunc(bookingH.Estimate)))

	// Kiosks
	mux.Handle("POST /api/v1/kiosks/register", applyAuthPerm(authSvc, domain.PermSysKioskManage, http.HandlerFunc(kioskH.Register)))
	mux.Handle("GET /api/v1/kiosks/{id}", applyAuthPerm(authSvc, domain.PermSysKioskView, http.HandlerFunc(kioskH.GetByID)))
	mux.Handle("PUT /api/v1/kiosks/{id}/heartbeat", applyAuth(authSvc, http.HandlerFunc(kioskH.Heartbeat)))
	mux.Handle("PUT /api/v1/kiosks/{id}/status", applyAuthPerm(authSvc, domain.PermSysKioskManage, http.HandlerFunc(kioskH.UpdateStatus)))
	mux.Handle("GET /api/v1/kiosks", applyAuthPerm(authSvc, domain.PermSysKioskView, http.HandlerFunc(kioskH.List)))

	// Apply global middleware
	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.CORS(h)

	return h
}

func applyAuth(authSvc *service.AuthService, h http.Handler) http.Handler {
	return middleware.Auth(authSvc)(h)
}

func applyAuthPerm(authSvc *service.AuthService, perm domain.Permission, h http.Handler) http.Handler {
	return middleware.Auth(authSvc)(middleware.RequirePermission(perm)(h))
}
