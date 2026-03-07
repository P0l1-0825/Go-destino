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
	fleetH *handler.FleetHandler,
	aiH *handler.AIHandler,
	analyticsH *handler.AnalyticsHandler,
	notifH *handler.NotificationHandler,
	voucherH *handler.VoucherHandler,
	shiftH *handler.ShiftHandler,
	adminH *handler.AdminHandler,
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

	// Fleet management
	mux.Handle("POST /api/v1/fleet/drivers", applyAuthPerm(authSvc, domain.PermFleetDriverOnboard, http.HandlerFunc(fleetH.RegisterDriver)))
	mux.Handle("GET /api/v1/fleet/drivers", applyAuthPerm(authSvc, domain.PermFleetDriverRead, http.HandlerFunc(fleetH.ListDrivers)))
	mux.Handle("GET /api/v1/fleet/drivers/{id}", applyAuthPerm(authSvc, domain.PermFleetDriverRead, http.HandlerFunc(fleetH.GetDriver)))
	mux.Handle("PUT /api/v1/fleet/drivers/{id}/status", applyAuthPerm(authSvc, domain.PermFleetStatusOwn, http.HandlerFunc(fleetH.UpdateStatus)))
	mux.Handle("PUT /api/v1/fleet/drivers/{id}/location", applyAuthPerm(authSvc, domain.PermFleetLocationOwn, http.HandlerFunc(fleetH.UpdateLocation)))
	mux.Handle("POST /api/v1/fleet/drivers/{id}/rate", applyAuthPerm(authSvc, domain.PermFleetDriverRate, http.HandlerFunc(fleetH.RateDriver)))
	mux.Handle("PUT /api/v1/fleet/drivers/{id}/verify", applyAuthPerm(authSvc, domain.PermFleetDriverVerify, http.HandlerFunc(fleetH.VerifyDocs)))
	mux.Handle("POST /api/v1/fleet/drivers/nearby", applyAuthPerm(authSvc, domain.PermFleetDispatchMap, http.HandlerFunc(fleetH.NearbyDrivers)))
	mux.Handle("POST /api/v1/fleet/vehicles", applyAuthPerm(authSvc, domain.PermFleetVehicleOwn, http.HandlerFunc(fleetH.RegisterVehicle)))
	mux.Handle("GET /api/v1/fleet/vehicles", applyAuthPerm(authSvc, domain.PermFleetVehicleAll, http.HandlerFunc(fleetH.ListVehicles)))
	mux.Handle("GET /api/v1/fleet/vehicles/{id}", applyAuthPerm(authSvc, domain.PermFleetVehicleOwn, http.HandlerFunc(fleetH.GetVehicle)))

	// AI services
	mux.Handle("GET /api/v1/ai/demand", applyAuthPerm(authSvc, domain.PermAIDemandForecast, http.HandlerFunc(aiH.DemandForecast)))
	mux.Handle("POST /api/v1/ai/pricing", applyAuthPerm(authSvc, domain.PermAIPricingView, http.HandlerFunc(aiH.DynamicPricing)))
	mux.Handle("POST /api/v1/ai/fraud", applyAuthPerm(authSvc, domain.PermAIFraudAlerts, http.HandlerFunc(aiH.FraudCheck)))
	mux.Handle("POST /api/v1/ai/chat", applyAuthPerm(authSvc, domain.PermAIChat, http.HandlerFunc(aiH.Chat)))
	mux.Handle("POST /api/v1/ai/biometric", applyAuth(authSvc, http.HandlerFunc(aiH.VerifyBiometric)))
	mux.Handle("POST /api/v1/ai/optimize-routes", applyAuthPerm(authSvc, domain.PermResAssignDriver, http.HandlerFunc(aiH.OptimizeRoutes)))

	// Analytics
	mux.Handle("GET /api/v1/analytics/dashboard", applyAuthPerm(authSvc, domain.PermAnalyticsKPIBasic, http.HandlerFunc(analyticsH.Dashboard)))
	mux.Handle("GET /api/v1/analytics/revenue", applyAuthPerm(authSvc, domain.PermAnalyticsReports, http.HandlerFunc(analyticsH.Revenue)))
	mux.Handle("GET /api/v1/analytics/funnel", applyAuthPerm(authSvc, domain.PermAnalyticsReports, http.HandlerFunc(analyticsH.BookingFunnel)))
	mux.Handle("GET /api/v1/analytics/slo", applyAuthPerm(authSvc, domain.PermAnalyticsSLO, http.HandlerFunc(analyticsH.SLO)))

	// Notifications
	mux.Handle("POST /api/v1/notifications", applyAuthPerm(authSvc, domain.PermSysUsersManage, http.HandlerFunc(notifH.Send)))
	mux.Handle("GET /api/v1/notifications/user/{id}", applyAuth(authSvc, http.HandlerFunc(notifH.GetUserNotifications)))

	// Vouchers
	mux.Handle("POST /api/v1/vouchers", applyAuthPerm(authSvc, domain.PermPayVoucherCreate, http.HandlerFunc(voucherH.Create)))
	mux.Handle("POST /api/v1/vouchers/redeem", applyAuthPerm(authSvc, domain.PermPayVoucherRedeem, http.HandlerFunc(voucherH.Redeem)))

	// Shifts (POS)
	mux.Handle("POST /api/v1/shifts", applyAuthPerm(authSvc, domain.PermKioskShiftOpen, http.HandlerFunc(shiftH.Open)))
	mux.Handle("PUT /api/v1/shifts/{id}/close", applyAuthPerm(authSvc, domain.PermKioskShiftClose, http.HandlerFunc(shiftH.Close)))
	mux.Handle("GET /api/v1/shifts/active", applyAuth(authSvc, http.HandlerFunc(shiftH.GetActive)))
	mux.Handle("GET /api/v1/shifts", applyAuth(authSvc, http.HandlerFunc(shiftH.List)))

	// Admin
	mux.Handle("POST /api/v1/admin/tenants", applyAuthPerm(authSvc, domain.PermSysSettingsEdit, http.HandlerFunc(adminH.CreateTenant)))
	mux.Handle("GET /api/v1/admin/tenants", applyAuthPerm(authSvc, domain.PermSysSettingsView, http.HandlerFunc(adminH.ListTenants)))
	mux.Handle("GET /api/v1/admin/tenants/{id}", applyAuthPerm(authSvc, domain.PermSysSettingsView, http.HandlerFunc(adminH.GetTenant)))
	mux.Handle("GET /api/v1/admin/users", applyAuthPerm(authSvc, domain.PermSysUsersRead, http.HandlerFunc(adminH.ListUsers)))
	mux.Handle("POST /api/v1/admin/airports", applyAuthPerm(authSvc, domain.PermSysAirportsManage, http.HandlerFunc(adminH.CreateAirport)))
	mux.Handle("GET /api/v1/admin/airports", applyAuthPerm(authSvc, domain.PermSysAirportsRead, http.HandlerFunc(adminH.ListAirports)))
	mux.Handle("GET /api/v1/admin/airports/{id}", applyAuthPerm(authSvc, domain.PermSysAirportsRead, http.HandlerFunc(adminH.GetAirport)))
	mux.Handle("GET /api/v1/admin/audit", applyAuthPerm(authSvc, domain.PermSysAuditLog, http.HandlerFunc(adminH.AuditLog)))
	mux.Handle("GET /api/v1/admin/roles", applyAuthPerm(authSvc, domain.PermSysRolesAssign, http.HandlerFunc(adminH.ListRoles)))
	mux.Handle("GET /api/v1/admin/permissions", applyAuthPerm(authSvc, domain.PermSysRolesAssign, http.HandlerFunc(adminH.ListPermissions)))

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
