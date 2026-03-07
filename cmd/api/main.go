package main

import (
	"log"
	"net/http"

	"github.com/P0l1-0825/Go-destino/internal/config"
	"github.com/P0l1-0825/Go-destino/internal/handler"
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

	// Services
	authSvc := service.NewAuthService(userRepo, cfg.JWT)
	routeSvc := service.NewRouteService(routeRepo)
	ticketSvc := service.NewTicketService(ticketRepo, routeRepo, paymentRepo)
	bookingSvc := service.NewBookingService(bookingRepo, paymentRepo)
	kioskSvc := service.NewKioskService(kioskRepo)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	routeH := handler.NewRouteHandler(routeSvc)
	ticketH := handler.NewTicketHandler(ticketSvc)
	bookingH := handler.NewBookingHandler(bookingSvc)
	kioskH := handler.NewKioskHandler(kioskSvc)

	// Router
	r := router.New(authSvc, authH, routeH, ticketH, bookingH, kioskH)

	addr := ":" + cfg.Server.Port
	log.Printf("GoDestino API starting on %s [env=%s]", addr, cfg.Server.Env)
	log.Printf("Endpoints: /health | /ready | /api/v1/*")

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
