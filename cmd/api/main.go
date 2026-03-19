package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/P0l1-0825/Go-destino/internal/config"
	"github.com/P0l1-0825/Go-destino/internal/handler"
	"github.com/P0l1-0825/Go-destino/internal/middleware"
	"github.com/P0l1-0825/Go-destino/internal/migrate"
	"github.com/P0l1-0825/Go-destino/internal/repository"
	"github.com/P0l1-0825/Go-destino/internal/router"
	"github.com/P0l1-0825/Go-destino/internal/security"
	"github.com/P0l1-0825/Go-destino/internal/service"
	"github.com/P0l1-0825/Go-destino/pkg/redisclient"
)

func main() {
	cfg := config.Load()
	cfg.Validate()

	// Database
	db, err := repository.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations (idempotent — safe to run on every startup)
	// Handles "already exists" errors gracefully (e.g., from docker-entrypoint-initdb.d)
	if err := migrate.Run(db, os.DirFS("migrations")); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Redis (graceful fallback — in-memory security components if unavailable)
	redisClient, err := redisclient.New(cfg.Redis)
	if err != nil {
		log.Printf("WARNING: Redis unavailable (%v) — using in-memory security components", err)
	} else {
		log.Printf("Redis connected at %s", redisClient.Addr())
		defer redisClient.Close()
	}

	// Security components — Redis-backed when available, in-memory fallback
	var tokenBlacklist security.TokenBlacklistStore
	var loginLimiter security.LoginLimiterStore
	var resetStore security.PasswordResetTokenStore

	if redisClient != nil {
		rdb := redisClient.Unwrap()
		tokenBlacklist = security.NewRedisTokenBlacklist(rdb)
		loginLimiter = security.NewRedisLoginLimiter(rdb, 5, 15*time.Minute, 30*time.Minute)
		resetStore = security.NewRedisPasswordResetStore(rdb)
		log.Println("Security components: Redis-backed (token blacklist, login limiter, password reset)")
	} else {
		tokenBlacklist = security.NewTokenBlacklist()
		loginLimiter = security.NewLoginLimiter(5, 15*time.Minute, 30*time.Minute)
		resetStore = security.NewPasswordResetStore()
		log.Println("Security components: in-memory (single-instance only)")
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	routeRepo := repository.NewRouteRepository(db)
	ticketRepo := repository.NewTicketRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	bookingRepo := repository.NewBookingRepository(db)
	kioskRepo := repository.NewKioskRepository(db)
	tenantRepo := repository.NewTenantRepository(db)
	driverRepo := repository.NewDriverRepository(db)
	vehicleRepo := repository.NewVehicleRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	notifRepo := repository.NewNotificationRepository(db)
	voucherRepo := repository.NewVoucherRepository(db)
	shiftRepo := repository.NewShiftRepository(db)
	airportRepo := repository.NewAirportRepository(db)
	cardRepo := repository.NewTransportCardRepository(db)
	sessionRepo := repository.NewKioskSessionRepository(db)
	monitorRepo := repository.NewKioskMonitorRepository(db)

	// Services
	auditSvc := service.NewAuditService(auditRepo)

	authSvc := service.NewAuthServiceFull(service.AuthServiceConfig{
		UserRepo:       userRepo,
		JWTCfg:         cfg.JWT,
		TokenBlacklist: tokenBlacklist,
		LoginLimiter:   loginLimiter,
		ResetStore:     resetStore,
		AuditFn: func(tenantID, userID, action, resource, resourceID, details, ip, ua string) {
			auditSvc.Log(context.Background(), tenantID, userID, action, resource, resourceID, details, ip, ua)
		},
	})
	routeSvc := service.NewRouteService(routeRepo)
	ticketSvc := service.NewTicketService(ticketRepo, routeRepo, paymentRepo)
	bookingSvc := service.NewBookingService(bookingRepo, paymentRepo)
	kioskSvc := service.NewKioskService(kioskRepo)
	fleetSvc := service.NewFleetService(driverRepo, vehicleRepo)
	aiSvc := service.NewAIService(bookingRepo)
	analyticsSvc := service.NewAnalyticsService(db)

	// Messaging infrastructure
	smtpSvc := service.NewSMTPService(cfg.SMTP)
	twilioSvc := service.NewTwilioService(cfg.Twilio)
	emailTemplateSvc := service.NewEmailTemplateService()

	// Notification service with full delivery (SMTP + Twilio + templates)
	notifSvc := service.NewNotificationServiceFull(notifRepo, userRepo, smtpSvc, twilioSvc, emailTemplateSvc)

	// Payment service with notifications + audit
	paymentSvc := service.NewPaymentService(paymentRepo, notifSvc, auditSvc)

	// Wire notification + payment into booking/ticket services
	bookingSvc.SetNotificationService(notifSvc)
	bookingSvc.SetPaymentService(paymentSvc)
	ticketSvc.SetNotificationService(notifSvc)

	voucherSvc := service.NewVoucherService(voucherRepo, paymentRepo)
	shiftSvc := service.NewShiftService(shiftRepo)
	flightSvc := service.NewFlightService(bookingRepo, notifSvc)
	safetySvc := service.NewSafetyService(db, notifSvc)
	kioskUXSvc := service.NewKioskUXService(bookingSvc, kioskRepo, bookingRepo, paymentRepo, routeRepo, cardRepo, sessionRepo)
	kioskMonSvc := service.NewKioskMonitorService(monitorRepo, kioskRepo, notifSvc)
	concesionRepo := repository.NewConcesionRepository(db)
	concesionSvc := service.NewConcesionService(concesionRepo, auditSvc)

	log.Printf("Messaging: SMTP=%v Twilio=%v", smtpSvc.IsEnabled(), twilioSvc.IsEnabled())

	// CORS configuration
	corsCfg := middleware.CORSConfig{
		AllowedOrigins: cfg.CORSOrigins,
	}

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	routeH := handler.NewRouteHandler(routeSvc)
	ticketH := handler.NewTicketHandler(ticketSvc)
	bookingH := handler.NewBookingHandler(bookingSvc)
	kioskH := handler.NewKioskHandler(kioskSvc)
	fleetH := handler.NewFleetHandler(fleetSvc)
	aiH := handler.NewAIHandler(aiSvc)
	analyticsH := handler.NewAnalyticsHandler(analyticsSvc)
	notifH := handler.NewNotificationHandler(notifSvc)
	voucherH := handler.NewVoucherHandler(voucherSvc)
	shiftH := handler.NewShiftHandler(shiftSvc)
	paymentH := handler.NewPaymentHandler(paymentSvc)
	adminH := handler.NewAdminHandler(tenantRepo, userRepo, airportRepo, auditSvc)
	flightH := handler.NewFlightHandler(flightSvc)
	safetyH := handler.NewSafetyHandler(safetySvc)
	wsH := handler.NewWSHandler(fleetSvc)
	kioskUXH := handler.NewKioskUXHandler(kioskUXSvc)
	kioskMonH := handler.NewKioskMonitorHandler(kioskMonSvc)
	qrH := handler.NewQRHandler(bookingSvc, ticketSvc)
	concesionH := handler.NewConcesionHandler(concesionSvc)

	// Router
	r := router.New(
		authSvc, authH, routeH, ticketH, bookingH, kioskH,
		fleetH, aiH, analyticsH, notifH, voucherH, shiftH, adminH,
		flightH, safetyH, wsH, kioskUXH, kioskMonH, paymentH, qrH,
		concesionH,
		corsCfg,
	)

	// Graceful shutdown
	addr := ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("GoDestino API starting on %s [env=%s]", addr, cfg.Server.Env)
		log.Printf("Modules: auth, routes, tickets, bookings, kiosks, kiosk-ux, kiosk-monitor, fleet, ai, analytics, notifications, payments, vouchers, shifts, admin, flights, safety, tracking, qr, concesiones")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
