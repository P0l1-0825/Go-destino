package router

import (
	"net/http"
	"time"

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
	flightH *handler.FlightHandler,
	safetyH *handler.SafetyHandler,
	wsH *handler.WSHandler,
	kioskUXH *handler.KioskUXHandler,
	kioskMonH *handler.KioskMonitorHandler,
	paymentH *handler.PaymentHandler,
	corsConfig ...middleware.CORSConfig,
) http.Handler {
	mux := http.NewServeMux()

	// Health & readiness probes
	mux.HandleFunc("GET /health", handler.HealthCheck)
	mux.HandleFunc("GET /ready", handler.ReadyCheck)

	// Auth (public)
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authH.RefreshToken)
	mux.HandleFunc("POST /api/v1/auth/forgot-password", authH.RequestPasswordReset)
	mux.HandleFunc("POST /api/v1/auth/reset-password", authH.ResetPassword)

	// Auth (protected)
	mux.Handle("GET /api/v1/auth/me", applyAuth(authSvc, http.HandlerFunc(authH.Me)))
	mux.Handle("POST /api/v1/auth/change-password", applyAuth(authSvc, http.HandlerFunc(authH.ChangePassword)))
	mux.Handle("POST /api/v1/auth/logout", applyAuth(authSvc, http.HandlerFunc(authH.Logout)))

	// Routes
	mux.Handle("POST /api/v1/routes", applyAuthPerm(authSvc, domain.PermSysSettingsEdit, http.HandlerFunc(routeH.Create)))
	mux.Handle("GET /api/v1/routes/{id}", applyAuth(authSvc, http.HandlerFunc(routeH.GetByID)))
	mux.Handle("GET /api/v1/routes", applyAuth(authSvc, http.HandlerFunc(routeH.List)))
	mux.Handle("PUT /api/v1/routes/{id}", applyAuthPerm(authSvc, domain.PermSysSettingsEdit, http.HandlerFunc(routeH.Update)))
	mux.Handle("DELETE /api/v1/routes/{id}", applyAuthPerm(authSvc, domain.PermSysSettingsEdit, http.HandlerFunc(routeH.Delete)))

	// Tickets (kiosk operations)
	mux.Handle("POST /api/v1/tickets/purchase", applyAuthPerm(authSvc, domain.PermKioskBookCreate, http.HandlerFunc(ticketH.Purchase)))
	mux.Handle("POST /api/v1/tickets/validate", applyAuth(authSvc, http.HandlerFunc(ticketH.Validate)))
	mux.Handle("GET /api/v1/tickets/{id}", applyAuth(authSvc, http.HandlerFunc(ticketH.GetByID)))
	mux.Handle("PUT /api/v1/tickets/{id}/cancel", applyAuth(authSvc, http.HandlerFunc(ticketH.Cancel)))

	// Bookings
	mux.Handle("POST /api/v1/bookings", applyAuthPerm(authSvc, domain.PermResCreateWeb, http.HandlerFunc(bookingH.Create)))
	mux.Handle("GET /api/v1/bookings", applyAuth(authSvc, http.HandlerFunc(bookingH.List)))
	mux.Handle("GET /api/v1/bookings/{id}", applyAuth(authSvc, http.HandlerFunc(bookingH.GetByID)))
	mux.Handle("GET /api/v1/bookings/number/{number}", applyAuth(authSvc, http.HandlerFunc(bookingH.GetByNumber)))
	mux.Handle("POST /api/v1/bookings/{id}/cancel", applyAuthPerm(authSvc, domain.PermResCancelOwn, http.HandlerFunc(bookingH.Cancel)))
	mux.Handle("PUT /api/v1/bookings/{id}/status", applyAuthPerm(authSvc, domain.PermResAssignDriver, http.HandlerFunc(bookingH.UpdateStatus)))
	mux.Handle("POST /api/v1/bookings/{id}/assign", applyAuthPerm(authSvc, domain.PermResAssignDriver, http.HandlerFunc(bookingH.AssignDriver)))
	mux.Handle("POST /api/v1/bookings/estimate", applyAuthPerm(authSvc, domain.PermResPriceEstimate, http.HandlerFunc(bookingH.Estimate)))
	mux.Handle("PUT /api/v1/bookings/{id}/start", applyAuthPerm(authSvc, domain.PermResAssignDriver, http.HandlerFunc(bookingH.StartTrip)))

	// Kiosks
	mux.Handle("POST /api/v1/kiosks/register", applyAuthPerm(authSvc, domain.PermSysKioskManage, http.HandlerFunc(kioskH.Register)))
	mux.Handle("GET /api/v1/kiosks/{id}", applyAuthPerm(authSvc, domain.PermSysKioskView, http.HandlerFunc(kioskH.GetByID)))
	mux.Handle("PUT /api/v1/kiosks/{id}/heartbeat", applyAuth(authSvc, http.HandlerFunc(kioskH.Heartbeat)))
	mux.Handle("PUT /api/v1/kiosks/{id}/status", applyAuthPerm(authSvc, domain.PermSysKioskManage, http.HandlerFunc(kioskH.UpdateStatus)))
	mux.Handle("GET /api/v1/kiosks", applyAuthPerm(authSvc, domain.PermSysKioskView, http.HandlerFunc(kioskH.List)))

	// Kiosk UX (AI-powered fast booking experience)
	mux.Handle("GET /api/v1/kiosk/suggestions", applyAuthPerm(authSvc, domain.PermKioskBookCreate, http.HandlerFunc(kioskUXH.Suggestions)))
	mux.Handle("GET /api/v1/kiosk/flights/{number}", applyAuthPerm(authSvc, domain.PermKioskBookCreate, http.HandlerFunc(kioskUXH.FlightLookup)))
	mux.Handle("POST /api/v1/kiosk/recommend", applyAuthPerm(authSvc, domain.PermKioskBookCreate, http.HandlerFunc(kioskUXH.RecommendService)))
	mux.Handle("POST /api/v1/kiosk/quick-book", applyAuthPerm(authSvc, domain.PermKioskBookCreate, http.HandlerFunc(kioskUXH.QuickBook)))
	mux.Handle("POST /api/v1/kiosk/estimate", applyAuthPerm(authSvc, domain.PermKioskBookCreate, http.HandlerFunc(kioskUXH.Estimate)))
	mux.Handle("GET /api/v1/kiosk/receipts/{bookingId}", applyAuthPerm(authSvc, domain.PermKioskPrintTicket, http.HandlerFunc(kioskUXH.GetReceipt)))
	mux.Handle("POST /api/v1/kiosk/sessions", applyAuth(authSvc, http.HandlerFunc(kioskUXH.StartSession)))
	mux.Handle("PUT /api/v1/kiosk/sessions/{id}/end", applyAuth(authSvc, http.HandlerFunc(kioskUXH.EndSession)))
	mux.Handle("GET /api/v1/kiosk/cards/{number}/balance", applyAuth(authSvc, http.HandlerFunc(kioskUXH.CardBalance)))
	mux.Handle("POST /api/v1/kiosk/cards/recharge", applyAuthPerm(authSvc, domain.PermPayCharge, http.HandlerFunc(kioskUXH.RechargeCard)))
	mux.Handle("POST /api/v1/kiosk/cards/issue", applyAuthPerm(authSvc, domain.PermKioskBookCreate, http.HandlerFunc(kioskUXH.IssueCard)))

	// Kiosk Monitoring & Remote Support
	mux.Handle("PUT /api/v1/monitor/kiosks/{id}/heartbeat", applyAuth(authSvc, http.HandlerFunc(kioskMonH.HeartbeatFull)))
	mux.Handle("GET /api/v1/monitor/dashboard", applyAuthPerm(authSvc, domain.PermSysKioskView, http.HandlerFunc(kioskMonH.FleetDashboard)))
	mux.Handle("GET /api/v1/monitor/kiosks/{id}/diagnostics", applyAuthPerm(authSvc, domain.PermSysKioskView, http.HandlerFunc(kioskMonH.Diagnostics)))
	mux.Handle("GET /api/v1/monitor/kiosks/{id}/telemetry", applyAuthPerm(authSvc, domain.PermSysKioskView, http.HandlerFunc(kioskMonH.TelemetryHistory)))
	mux.Handle("GET /api/v1/monitor/kiosks/{id}/events", applyAuthPerm(authSvc, domain.PermSysKioskView, http.HandlerFunc(kioskMonH.Events)))
	mux.Handle("POST /api/v1/monitor/kiosks/{id}/commands", applyAuthPerm(authSvc, domain.PermSysKioskManage, http.HandlerFunc(kioskMonH.SendCommand)))
	mux.Handle("GET /api/v1/monitor/kiosks/{id}/commands", applyAuthPerm(authSvc, domain.PermSysKioskView, http.HandlerFunc(kioskMonH.CommandHistory)))
	mux.Handle("GET /api/v1/monitor/alerts", applyAuthPerm(authSvc, domain.PermSysKioskView, http.HandlerFunc(kioskMonH.Alerts)))
	mux.Handle("PUT /api/v1/monitor/alerts/{id}/ack", applyAuthPerm(authSvc, domain.PermSysKioskManage, http.HandlerFunc(kioskMonH.AckAlert)))
	mux.Handle("PUT /api/v1/monitor/alerts/{id}/resolve", applyAuthPerm(authSvc, domain.PermSysKioskManage, http.HandlerFunc(kioskMonH.ResolveAlert)))
	mux.Handle("GET /api/v1/monitor/events", applyAuthPerm(authSvc, domain.PermSysKioskView, http.HandlerFunc(kioskMonH.EventsByTenant)))
	mux.Handle("PUT /api/v1/monitor/commands/{id}/result", applyAuth(authSvc, http.HandlerFunc(kioskMonH.CommandResult)))
	mux.Handle("GET /api/v1/monitor/stream", applyAuthPerm(authSvc, domain.PermSysKioskView, http.HandlerFunc(kioskMonH.StreamMonitor)))

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

	// Payments
	mux.Handle("GET /api/v1/payments", applyAuthPerm(authSvc, domain.PermPayCharge, http.HandlerFunc(paymentH.List)))
	mux.Handle("GET /api/v1/payments/{id}", applyAuthPerm(authSvc, domain.PermPayCharge, http.HandlerFunc(paymentH.GetByID)))
	mux.Handle("GET /api/v1/payments/booking/{bookingId}", applyAuth(authSvc, http.HandlerFunc(paymentH.GetByBooking)))
	mux.Handle("POST /api/v1/payments/{id}/refund", applyAuthPerm(authSvc, domain.PermPayRefundAny, http.HandlerFunc(paymentH.Refund)))

	// Notifications
	mux.Handle("POST /api/v1/notifications", applyAuthPerm(authSvc, domain.PermSysUsersManage, http.HandlerFunc(notifH.Send)))
	mux.Handle("GET /api/v1/notifications/user/{id}", applyAuth(authSvc, http.HandlerFunc(notifH.GetUserNotifications)))
	mux.Handle("PUT /api/v1/notifications/{id}/read", applyAuth(authSvc, http.HandlerFunc(notifH.MarkRead)))
	mux.Handle("PUT /api/v1/notifications/read-all", applyAuth(authSvc, http.HandlerFunc(notifH.MarkAllRead)))

	// Vouchers
	mux.Handle("POST /api/v1/vouchers", applyAuthPerm(authSvc, domain.PermPayVoucherCreate, http.HandlerFunc(voucherH.Create)))
	mux.Handle("POST /api/v1/vouchers/redeem", applyAuthPerm(authSvc, domain.PermPayVoucherRedeem, http.HandlerFunc(voucherH.Redeem)))
	mux.Handle("GET /api/v1/vouchers", applyAuth(authSvc, http.HandlerFunc(voucherH.List)))
	mux.Handle("GET /api/v1/vouchers/{id}", applyAuth(authSvc, http.HandlerFunc(voucherH.GetByID)))
	mux.Handle("GET /api/v1/vouchers/code/{code}", applyAuth(authSvc, http.HandlerFunc(voucherH.GetByCode)))

	// Shifts (POS)
	mux.Handle("POST /api/v1/shifts", applyAuthPerm(authSvc, domain.PermKioskShiftOpen, http.HandlerFunc(shiftH.Open)))
	mux.Handle("PUT /api/v1/shifts/{id}/close", applyAuthPerm(authSvc, domain.PermKioskShiftClose, http.HandlerFunc(shiftH.Close)))
	mux.Handle("GET /api/v1/shifts/active", applyAuth(authSvc, http.HandlerFunc(shiftH.GetActive)))
	mux.Handle("GET /api/v1/shifts", applyAuth(authSvc, http.HandlerFunc(shiftH.List)))
	mux.Handle("GET /api/v1/shifts/{id}", applyAuth(authSvc, http.HandlerFunc(shiftH.GetByID)))
	mux.Handle("GET /api/v1/shifts/kiosk/{kioskId}", applyAuthPerm(authSvc, domain.PermSysKioskView, http.HandlerFunc(shiftH.ListByKiosk)))

	// Flights & IROPS
	mux.Handle("GET /api/v1/flights/{number}", applyAuth(authSvc, http.HandlerFunc(flightH.GetFlightStatus)))
	mux.Handle("GET /api/v1/flights/arrivals/{code}", applyAuth(authSvc, http.HandlerFunc(flightH.ListArrivals)))
	mux.Handle("POST /api/v1/flights/irops", applyAuthPerm(authSvc, domain.PermResOverrideAI, http.HandlerFunc(flightH.ReportIROPS)))

	// Safety & SOS
	mux.Handle("POST /api/v1/safety/incidents", applyAuth(authSvc, http.HandlerFunc(safetyH.ReportIncident)))
	mux.Handle("GET /api/v1/safety/incidents", applyAuthPerm(authSvc, domain.PermSysAuditLog, http.HandlerFunc(safetyH.ListIncidents)))
	mux.Handle("GET /api/v1/safety/incidents/{id}", applyAuth(authSvc, http.HandlerFunc(safetyH.GetIncident)))
	mux.Handle("POST /api/v1/safety/sos", applyAuth(authSvc, http.HandlerFunc(safetyH.TriggerSOS)))
	mux.Handle("PUT /api/v1/safety/sos/{id}/resolve", applyAuthPerm(authSvc, domain.PermResOverrideAI, http.HandlerFunc(safetyH.ResolveSOS)))
	mux.HandleFunc("GET /api/v1/safety/emergency-numbers", safetyH.GetEmergencyNumbers)

	// Real-time tracking (SSE)
	mux.Handle("GET /api/v1/track/driver/{id}", applyAuth(authSvc, http.HandlerFunc(wsH.TrackDriver)))
	mux.Handle("POST /api/v1/track/publish", applyAuthPerm(authSvc, domain.PermFleetLocationOwn, http.HandlerFunc(wsH.PublishLocation)))
	mux.Handle("GET /api/v1/track/drivers", applyAuthPerm(authSvc, domain.PermFleetLocationView, http.HandlerFunc(wsH.DriverLocations)))

	// Admin
	mux.Handle("POST /api/v1/admin/tenants", applyAuthPerm(authSvc, domain.PermSysSettingsEdit, http.HandlerFunc(adminH.CreateTenant)))
	mux.Handle("GET /api/v1/admin/tenants", applyAuthPerm(authSvc, domain.PermSysSettingsView, http.HandlerFunc(adminH.ListTenants)))
	mux.Handle("GET /api/v1/admin/tenants/{id}", applyAuthPerm(authSvc, domain.PermSysSettingsView, http.HandlerFunc(adminH.GetTenant)))
	mux.Handle("GET /api/v1/admin/users", applyAuthPerm(authSvc, domain.PermSysUsersRead, http.HandlerFunc(adminH.ListUsers)))
	mux.Handle("GET /api/v1/admin/users/{id}", applyAuthPerm(authSvc, domain.PermSysUsersRead, http.HandlerFunc(adminH.GetUser)))
	mux.Handle("POST /api/v1/admin/users", applyAuthPerm(authSvc, domain.PermSysUsersManage, http.HandlerFunc(adminH.CreateUser)))
	mux.Handle("PUT /api/v1/admin/users/{id}", applyAuthPerm(authSvc, domain.PermSysUsersManage, http.HandlerFunc(adminH.UpdateUser)))
	mux.Handle("PUT /api/v1/admin/users/{id}/activate", applyAuthPerm(authSvc, domain.PermSysUsersManage, http.HandlerFunc(adminH.ActivateUser)))
	mux.Handle("PUT /api/v1/admin/users/{id}/deactivate", applyAuthPerm(authSvc, domain.PermSysUsersManage, http.HandlerFunc(adminH.DeactivateUser)))
	mux.Handle("PUT /api/v1/admin/users/{id}/role", applyAuthPerm(authSvc, domain.PermSysRolesAssign, http.HandlerFunc(adminH.UpdateUserRole)))
	mux.Handle("POST /api/v1/admin/airports", applyAuthPerm(authSvc, domain.PermSysAirportsManage, http.HandlerFunc(adminH.CreateAirport)))
	mux.Handle("GET /api/v1/admin/airports", applyAuthPerm(authSvc, domain.PermSysAirportsRead, http.HandlerFunc(adminH.ListAirports)))
	mux.Handle("GET /api/v1/admin/airports/{id}", applyAuthPerm(authSvc, domain.PermSysAirportsRead, http.HandlerFunc(adminH.GetAirport)))
	mux.Handle("GET /api/v1/admin/audit", applyAuthPerm(authSvc, domain.PermSysAuditLog, http.HandlerFunc(adminH.AuditLog)))
	mux.Handle("GET /api/v1/admin/roles", applyAuthPerm(authSvc, domain.PermSysRolesAssign, http.HandlerFunc(adminH.ListRoles)))
	mux.Handle("GET /api/v1/admin/permissions", applyAuthPerm(authSvc, domain.PermSysRolesAssign, http.HandlerFunc(adminH.ListPermissions)))

	// Apply global middleware (order matters: outermost runs first)
	var h http.Handler = mux
	h = middleware.SecurityHeaders(h)
	h = middleware.RateLimit(200, time.Minute)(h)
	h = middleware.Logging(h)
	h = middleware.RequestID(h)
	h = middleware.Recovery(h)
	if len(corsConfig) > 0 {
		h = middleware.CORSWithConfig(corsConfig[0])(h)
	} else {
		h = middleware.CORS(h)
	}

	return h
}

func applyAuth(authSvc *service.AuthService, h http.Handler) http.Handler {
	return middleware.Auth(authSvc)(h)
}

func applyAuthPerm(authSvc *service.AuthService, perm domain.Permission, h http.Handler) http.Handler {
	return middleware.Auth(authSvc)(middleware.RequirePermission(perm)(h))
}
