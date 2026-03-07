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
	"github.com/P0l1-0825/Go-destino/internal/repository"
	"github.com/P0l1-0825/Go-destino/internal/router"
	"github.com/P0l1-0825/Go-destino/internal/service"
)

func main() {
	cfg := config.Load()

	// Database
	db, err := repository.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

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

	// Services
	auditSvc := service.NewAuditService(auditRepo)

	authSvc := service.NewAuthServiceFull(service.AuthServiceConfig{
		UserRepo:       userRepo,
		JWTCfg:         cfg.JWT,
		TokenBlacklist: nil,
		LoginLimiter:   nil,
		ResetStore:     nil,
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
	notifSvc := service.NewNotificationService(notifRepo)
	voucherSvc := service.NewVoucherService(voucherRepo, paymentRepo)
	shiftSvc := service.NewShiftService(shiftRepo)
	flightSvc := service.NewFlightService(bookingRepo, notifSvc)
	safetySvc := service.NewSafetyService(db, notifSvc)

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
	adminH := handler.NewAdminHandler(tenantRepo, userRepo, airportRepo, auditSvc)
	flightH := handler.NewFlightHandler(flightSvc)
	safetyH := handler.NewSafetyHandler(safetySvc)
	wsH := handler.NewWSHandler(fleetSvc)

	// Router
	r := router.New(
		authSvc, authH, routeH, ticketH, bookingH, kioskH,
		fleetH, aiH, analyticsH, notifH, voucherH, shiftH, adminH,
		flightH, safetyH, wsH,
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
		log.Printf("Modules: auth, routes, tickets, bookings, kiosks, fleet, ai, analytics, notifications, vouchers, shifts, admin, flights, safety, tracking")
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
