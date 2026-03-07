package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/P0l1-0825/Go-destino/internal/domain"
	"github.com/P0l1-0825/Go-destino/internal/repository"
	"github.com/P0l1-0825/Go-destino/pkg/geo"
)

// KioskUXService provides AI-enhanced kiosk experience: smart suggestions,
// flight lookups, quick-booking, receipt generation, and session tracking.
type KioskUXService struct {
	bookingSvc  *BookingService
	kioskRepo   *repository.KioskRepository
	bookingRepo *repository.BookingRepository
	paymentRepo *repository.PaymentRepository
	routeRepo   *repository.RouteRepository
	cardRepo    *repository.TransportCardRepository
	sessionRepo *repository.KioskSessionRepository
}

func NewKioskUXService(
	bookingSvc *BookingService,
	kioskRepo *repository.KioskRepository,
	bookingRepo *repository.BookingRepository,
	paymentRepo *repository.PaymentRepository,
	routeRepo *repository.RouteRepository,
	cardRepo *repository.TransportCardRepository,
	sessionRepo *repository.KioskSessionRepository,
) *KioskUXService {
	return &KioskUXService{
		bookingSvc:  bookingSvc,
		kioskRepo:   kioskRepo,
		bookingRepo: bookingRepo,
		paymentRepo: paymentRepo,
		routeRepo:   routeRepo,
		cardRepo:    cardRepo,
		sessionRepo: sessionRepo,
	}
}

// GetSmartSuggestions returns AI-powered suggestions for the kiosk home screen.
// Considers time of day, airport, demand level, and popular destinations.
func (s *KioskUXService) GetSmartSuggestions(ctx context.Context, kioskID, tenantID, lang string) (*domain.KioskSuggestionsResponse, error) {
	if lang == "" {
		lang = "es"
	}

	kiosk, err := s.kioskRepo.GetByID(ctx, kioskID)
	if err != nil {
		return nil, fmt.Errorf("kiosk not found: %w", err)
	}

	hour := time.Now().Hour()
	demandLevel := classifyDemand(hour)

	// Build AI-driven suggestions
	suggestions := s.buildDestinationSuggestions(kiosk, hour, lang)

	// Get popular routes for this tenant
	routes, _ := s.routeRepo.ListByTenant(ctx, tenantID)
	popularRoutes := make([]domain.PopularRoute, 0, len(routes))
	for _, r := range routes {
		if r.Active {
			popularRoutes = append(popularRoutes, domain.PopularRoute{
				Name:        r.Name,
				Code:        r.Code,
				Origin:      r.Origin,
				Destination: r.Destination,
				PriceCents:  r.PriceCents,
				Currency:    r.Currency,
			})
		}
		if len(popularRoutes) >= 6 {
			break
		}
	}

	welcome := kioskWelcome(lang, hour)

	return &domain.KioskSuggestionsResponse{
		Suggestions:    suggestions,
		PopularRoutes:  popularRoutes,
		DemandLevel:    demandLevel,
		WelcomeMessage: welcome,
		Lang:           lang,
		GeneratedAt:    time.Now(),
	}, nil
}

// LookupFlight returns flight info + recommended transport options.
func (s *KioskUXService) LookupFlight(ctx context.Context, flightNumber, kioskID, tenantID, lang string) (*domain.FlightLookupResponse, error) {
	if lang == "" {
		lang = "es"
	}

	flightNumber = strings.ToUpper(strings.TrimSpace(flightNumber))
	if flightNumber == "" {
		return nil, fmt.Errorf("flight number is required")
	}

	kiosk, err := s.kioskRepo.GetByID(ctx, kioskID)
	if err != nil {
		return nil, fmt.Errorf("kiosk not found: %w", err)
	}

	// Simulated flight data (in production: call FlightAware/AeroAPI)
	flight := simulateFlightInfo(flightNumber)

	// Generate transport options based on flight
	options := s.generateTransportOptions(kiosk, flight, lang)

	return &domain.FlightLookupResponse{
		FlightNumber:     flightNumber,
		Airline:          flight.airline,
		Origin:           flight.origin,
		Status:           flight.status,
		ArrivalTime:      flight.arrivalTime,
		Terminal:         kiosk.TerminalID,
		Gate:             flight.gate,
		Passengers:       flight.passengers,
		TransportOptions: options,
	}, nil
}

// RecommendService uses AI to suggest the best service type based on context.
func (s *KioskUXService) RecommendService(ctx context.Context, passengers int, dropoffLat, dropoffLng float64, kioskID string) ([]domain.ServiceRecommendation, error) {
	kiosk, err := s.kioskRepo.GetByID(ctx, kioskID)
	if err != nil {
		return nil, fmt.Errorf("kiosk not found: %w", err)
	}

	// Estimate distance from airport (kiosk location is the pickup)
	pickupLat, pickupLng := airportCoords(kiosk.AirportID)
	dist := geo.Haversine(pickupLat, pickupLng, dropoffLat, dropoffLng)

	recommendations := make([]domain.ServiceRecommendation, 0, 4)

	// Score each service type
	types := []struct {
		svc       domain.ServiceType
		name      string
		baseCents int64
		maxPax    int
	}{
		{domain.ServiceTaxi, "Taxi", 5000, 4},
		{domain.ServiceShuttle, "Shuttle", 3500, 12},
		{domain.ServiceVan, "Van", 8000, 8},
		{domain.ServiceBus, "Bus", 2500, 40},
	}

	for _, t := range types {
		price := (t.baseCents + int64(dist*300)) * int64(passengers)
		eta := int(dist/0.8) + 5
		conf := 0.0
		reason := ""

		switch {
		case passengers <= 2 && dist < 30:
			if t.svc == domain.ServiceTaxi {
				conf = 0.95
				reason = "Best for 1-2 passengers, short distance"
			}
		case passengers <= 4 && dist < 30:
			if t.svc == domain.ServiceTaxi {
				conf = 0.85
				reason = "Good for small groups"
			}
		case passengers > 4 && passengers <= 8:
			if t.svc == domain.ServiceVan {
				conf = 0.92
				reason = "Ideal for medium groups with luggage"
			}
		case passengers > 8:
			if t.svc == domain.ServiceShuttle {
				conf = 0.88
				reason = "Best value for large groups"
			}
			if t.svc == domain.ServiceBus {
				conf = 0.90
				reason = "Most economical for large groups"
			}
		}

		if conf == 0 {
			// Default scoring
			if passengers <= t.maxPax {
				conf = 0.5
				reason = "Available option"
			} else {
				continue
			}
		}

		// Distance adjustments
		if dist > 50 && t.svc == domain.ServiceShuttle {
			conf += 0.1
			reason = "Comfortable for long distance"
		}
		if dist < 10 && t.svc == domain.ServiceTaxi {
			conf += 0.05
			reason = "Fastest for short trips"
		}
		if conf > 1.0 {
			conf = 0.99
		}

		recommendations = append(recommendations, domain.ServiceRecommendation{
			ServiceType: t.svc,
			Confidence:  conf,
			Reason:      reason,
			PriceCents:  price,
			Currency:    "MXN",
			ETAMinutes:  eta,
		})
	}

	// Sort by confidence descending (simple bubble sort for small slice)
	for i := 0; i < len(recommendations); i++ {
		for j := i + 1; j < len(recommendations); j++ {
			if recommendations[j].Confidence > recommendations[i].Confidence {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}

	return recommendations, nil
}

// QuickBook performs a streamlined one-step booking from the kiosk.
func (s *KioskUXService) QuickBook(ctx context.Context, kioskID, tenantID, sellerID string, req domain.QuickBookRequest) (*domain.QuickBookResponse, error) {
	kiosk, err := s.kioskRepo.GetByID(ctx, kioskID)
	if err != nil {
		return nil, fmt.Errorf("kiosk not found: %w", err)
	}

	if req.PassengerCount < 1 {
		req.PassengerCount = 1
	}
	if req.PassengerCount > 50 {
		return nil, fmt.Errorf("maximum 50 passengers per booking")
	}
	if !domain.ValidServiceType(string(req.ServiceType)) {
		return nil, fmt.Errorf("invalid service type: %s", req.ServiceType)
	}

	// Use kiosk/airport as pickup
	pickupLat, pickupLng := airportCoords(kiosk.AirportID)
	pickupAddr := kiosk.Location
	if pickupAddr == "" {
		pickupAddr = "Airport Terminal " + kiosk.TerminalID
	}

	dropoffAddr := req.DropoffAddress
	if dropoffAddr == "" {
		dropoffAddr = fmt.Sprintf("%.4f, %.4f", req.DropoffLat, req.DropoffLng)
	}

	// Create booking via booking service
	bookingReq := domain.CreateBookingRequest{
		ServiceType:    req.ServiceType,
		PickupAddress:  pickupAddr,
		DropoffAddress: dropoffAddr,
		PickupLat:      pickupLat,
		PickupLng:      pickupLng,
		DropoffLat:     req.DropoffLat,
		DropoffLng:     req.DropoffLng,
		PassengerCount: req.PassengerCount,
		FlightNumber:   req.FlightNumber,
		PaymentMethod:  req.PaymentMethod,
	}

	booking, err := s.bookingSvc.Create(ctx, tenantID, sellerID, kioskID, bookingReq)
	if err != nil {
		return nil, fmt.Errorf("creating booking: %w", err)
	}

	// Process payment
	var paymentMethod domain.PaymentMethod
	switch req.PaymentMethod {
	case "card":
		paymentMethod = domain.PaymentCard
	case "qr":
		paymentMethod = domain.PaymentQR
	case "transport_card":
		// Deduct from transport card
		if req.CardNumber == "" {
			return nil, fmt.Errorf("card_number required for transport_card payment")
		}
		if err := s.deductFromCard(ctx, tenantID, req.CardNumber, booking.PriceCents); err != nil {
			return nil, fmt.Errorf("card payment failed: %w", err)
		}
		paymentMethod = domain.PaymentQR // Map to QR for now
	default:
		paymentMethod = domain.PaymentCash
	}

	paymentRef, _ := generateReceiptNumber()
	payment := &domain.Payment{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		BookingID:   booking.ID,
		KioskID:     kioskID,
		Method:      paymentMethod,
		Status:      domain.PaymentCompleted,
		AmountCents: booking.PriceCents,
		Currency:    booking.Currency,
		Reference:   paymentRef,
	}
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("creating payment: %w", err)
	}

	// Update booking with payment
	_ = s.bookingRepo.SetPayment(ctx, booking.ID, payment.ID)
	booking.PaymentID = payment.ID

	// Generate receipt
	lang := req.Lang
	if lang == "" {
		lang = "es"
	}
	receipt := s.generateReceipt(kiosk, booking, payment, lang)

	// Calculate ETA
	dist := geo.Haversine(booking.PickupLat, booking.PickupLng, booking.DropoffLat, booking.DropoffLng)
	eta := int(dist/0.8) + 5

	msg := quickBookMessage(lang, booking.BookingNumber)

	return &domain.QuickBookResponse{
		Booking:    *booking,
		Payment:    *payment,
		Receipt:    receipt,
		QRCode:     booking.BookingNumber,
		Message:    msg,
		ETAMinutes: eta,
	}, nil
}

// GetReceipt generates a receipt for an existing booking.
func (s *KioskUXService) GetReceipt(ctx context.Context, bookingID, kioskID, lang string) (*domain.KioskReceipt, error) {
	booking, err := s.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("booking not found: %w", err)
	}

	kiosk, err := s.kioskRepo.GetByID(ctx, kioskID)
	if err != nil {
		return nil, fmt.Errorf("kiosk not found: %w", err)
	}

	var payment *domain.Payment
	if booking.PaymentID != "" {
		payment, _ = s.paymentRepo.GetByID(ctx, booking.PaymentID)
	}
	if payment == nil {
		payment = &domain.Payment{Currency: "MXN"}
	}

	if lang == "" {
		lang = "es"
	}

	receipt := s.generateReceipt(kiosk, booking, payment, lang)
	return &receipt, nil
}

// StartSession begins tracking a kiosk user session.
func (s *KioskUXService) StartSession(ctx context.Context, kioskID, tenantID, lang string) (*domain.KioskSession, error) {
	session := &domain.KioskSession{
		ID:        uuid.New().String(),
		KioskID:   kioskID,
		TenantID:  tenantID,
		Lang:      lang,
		StartedAt: time.Now(),
		Outcome:   "in_progress",
		StepCount: 0,
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("starting session: %w", err)
	}

	return session, nil
}

// EndSession marks a session as completed or abandoned.
func (s *KioskUXService) EndSession(ctx context.Context, sessionID, outcome, bookingID string, steps int) error {
	if outcome == "" {
		outcome = "completed"
	}
	now := time.Now()
	return s.sessionRepo.End(ctx, sessionID, outcome, bookingID, steps, &now)
}

// --- Transport Card Operations ---

// GetCardBalance looks up a transport card by number.
func (s *KioskUXService) GetCardBalance(ctx context.Context, tenantID, cardNumber string) (*domain.TransportCard, error) {
	return s.cardRepo.GetByNumber(ctx, tenantID, cardNumber)
}

// RechargeCard adds balance to a transport card.
func (s *KioskUXService) RechargeCard(ctx context.Context, tenantID, kioskID string, req domain.RechargeCardRequest) (*domain.RechargeCardResponse, error) {
	if req.AmountCents < 1000 {
		return nil, fmt.Errorf("minimum recharge is $10.00 MXN")
	}
	if req.AmountCents > 500000 {
		return nil, fmt.Errorf("maximum recharge is $5,000.00 MXN")
	}

	card, err := s.cardRepo.GetByNumber(ctx, tenantID, req.CardNumber)
	if err != nil {
		return nil, fmt.Errorf("card not found: %w", err)
	}
	if !card.Active {
		return nil, fmt.Errorf("card is inactive")
	}

	// Create payment record
	var method domain.PaymentMethod
	switch req.PaymentMethod {
	case "card":
		method = domain.PaymentCard
	case "qr":
		method = domain.PaymentQR
	default:
		method = domain.PaymentCash
	}

	ref, _ := generateReceiptNumber()
	payment := &domain.Payment{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		KioskID:     kioskID,
		Method:      method,
		Status:      domain.PaymentCompleted,
		AmountCents: req.AmountCents,
		Currency:    "MXN",
		Reference:   ref,
	}
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("creating payment: %w", err)
	}

	// Add balance
	if err := s.cardRepo.AddBalance(ctx, card.ID, req.AmountCents); err != nil {
		return nil, fmt.Errorf("adding balance: %w", err)
	}

	card.BalanceCents += req.AmountCents

	return &domain.RechargeCardResponse{
		Card:    *card,
		Payment: *payment,
	}, nil
}

// IssueCard creates a new transport card.
func (s *KioskUXService) IssueCard(ctx context.Context, tenantID, kioskID string, initialBalance int64) (*domain.TransportCard, error) {
	cardNum, err := generateCardNumber()
	if err != nil {
		return nil, fmt.Errorf("generating card number: %w", err)
	}

	card := &domain.TransportCard{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		CardNumber:   cardNum,
		BalanceCents: initialBalance,
		Currency:     "MXN",
		Active:       true,
	}

	if err := s.cardRepo.Create(ctx, card); err != nil {
		return nil, fmt.Errorf("issuing card: %w", err)
	}

	return card, nil
}

// --- Private helpers ---

func (s *KioskUXService) deductFromCard(ctx context.Context, tenantID, cardNumber string, amount int64) error {
	card, err := s.cardRepo.GetByNumber(ctx, tenantID, cardNumber)
	if err != nil {
		return fmt.Errorf("card not found")
	}
	if !card.Active {
		return fmt.Errorf("card is inactive")
	}
	if card.BalanceCents < amount {
		return fmt.Errorf("insufficient balance: have $%.2f, need $%.2f",
			float64(card.BalanceCents)/100, float64(amount)/100)
	}
	return s.cardRepo.AddBalance(ctx, card.ID, -amount)
}

func (s *KioskUXService) generateReceipt(kiosk *domain.Kiosk, booking *domain.Booking, payment *domain.Payment, lang string) domain.KioskReceipt {
	receiptNum, _ := generateReceiptNumber()

	footer := receiptFooter(lang)

	return domain.KioskReceipt{
		ReceiptNumber: receiptNum,
		KioskID:       kiosk.ID,
		KioskName:     kiosk.Name,
		TenantID:      kiosk.TenantID,
		BookingNumber: booking.BookingNumber,
		ServiceType:   string(booking.ServiceType),
		Pickup:        booking.PickupAddress,
		Dropoff:       booking.DropoffAddress,
		Passengers:    booking.PassengerCount,
		PriceCents:    booking.PriceCents,
		Currency:      booking.Currency,
		PaymentMethod: string(payment.Method),
		PaymentRef:    payment.Reference,
		FlightNumber:  booking.FlightNumber,
		QRCode:        booking.BookingNumber,
		IssuedAt:      time.Now(),
		Footer:        footer,
	}
}

func (s *KioskUXService) buildDestinationSuggestions(kiosk *domain.Kiosk, hour int, lang string) []domain.KioskSuggestion {
	// AI-driven suggestions based on airport, time of day, and common patterns
	suggestions := []domain.KioskSuggestion{}
	priority := 1

	// Popular hotel zones per airport (simulated — in production from ML model)
	destinations := airportDestinations(kiosk.AirportID)

	for _, d := range destinations {
		// Adjust pricing by time of day
		priceMult := 1.0
		if hour >= 22 || hour < 6 {
			priceMult = 1.3 // Night surcharge
		} else if (hour >= 7 && hour <= 9) || (hour >= 17 && hour <= 19) {
			priceMult = 1.2 // Peak hours
		}

		pickupLat, pickupLng := airportCoords(kiosk.AirportID)
		dist := geo.Haversine(pickupLat, pickupLng, d.lat, d.lng)
		baseCents := int64(5000 + dist*300)
		finalPrice := int64(float64(baseCents) * priceMult)
		eta := int(dist/0.8) + 5

		sug := domain.KioskSuggestion{
			Type:        "destination",
			Title:       localizeDestTitle(lang, d.name),
			Subtitle:    localizeDestSubtitle(lang, eta, finalPrice),
			ServiceType: "taxi",
			DropoffLat:  d.lat,
			DropoffLng:  d.lng,
			DropoffName: d.name,
			PriceCents:  finalPrice,
			Currency:    "MXN",
			ETAMinutes:  eta,
			Priority:    priority,
		}
		suggestions = append(suggestions, sug)
		priority++
	}

	// Add service type suggestions
	if hour >= 6 && hour <= 22 {
		suggestions = append(suggestions, domain.KioskSuggestion{
			Type:        "service",
			Title:       localizeServiceTitle(lang, "shuttle"),
			Subtitle:    localizeServiceSub(lang, "shuttle"),
			ServiceType: "shuttle",
			Priority:    priority + 10,
		})
	}

	return suggestions
}

func (s *KioskUXService) generateTransportOptions(kiosk *domain.Kiosk, flight flightInfo, lang string) []domain.TransportOption {
	pickupLat, pickupLng := airportCoords(kiosk.AirportID)

	// Common destinations from this airport
	dests := airportDestinations(kiosk.AirportID)
	avgDist := 25.0
	if len(dests) > 0 {
		avgDist = geo.Haversine(pickupLat, pickupLng, dests[0].lat, dests[0].lng)
	}

	options := []domain.TransportOption{
		{
			ServiceType:   domain.ServiceTaxi,
			Name:          localizeTransport(lang, "taxi"),
			Description:   localizeTransportDesc(lang, "taxi"),
			PriceCents:    int64(5000 + avgDist*300),
			Currency:      "MXN",
			ETAMinutes:    int(avgDist/0.8) + 5,
			MaxPassengers: 4,
			Recommended:   flight.passengers <= 4,
			ReasonTag:     "fastest",
		},
		{
			ServiceType:   domain.ServiceVan,
			Name:          localizeTransport(lang, "van"),
			Description:   localizeTransportDesc(lang, "van"),
			PriceCents:    int64(8000 + avgDist*300),
			Currency:      "MXN",
			ETAMinutes:    int(avgDist/0.8) + 8,
			MaxPassengers: 8,
			Recommended:   flight.passengers > 4 && flight.passengers <= 8,
			ReasonTag:     "group",
		},
		{
			ServiceType:   domain.ServiceShuttle,
			Name:          localizeTransport(lang, "shuttle"),
			Description:   localizeTransportDesc(lang, "shuttle"),
			PriceCents:    int64(3500 + avgDist*200),
			Currency:      "MXN",
			ETAMinutes:    int(avgDist/0.6) + 10,
			MaxPassengers: 12,
			Recommended:   false,
			ReasonTag:     "best_value",
		},
		{
			ServiceType:   domain.ServiceBus,
			Name:          localizeTransport(lang, "bus"),
			Description:   localizeTransportDesc(lang, "bus"),
			PriceCents:    int64(2500 + avgDist*150),
			Currency:      "MXN",
			ETAMinutes:    int(avgDist/0.5) + 15,
			MaxPassengers: 40,
			Recommended:   flight.passengers > 12,
			ReasonTag:     "economy",
		},
	}

	return options
}

// --- Data/localization helpers ---

type destInfo struct {
	name string
	lat  float64
	lng  float64
}

type flightInfo struct {
	airline     string
	origin      string
	status      string
	arrivalTime *time.Time
	gate        string
	passengers  int
}

func simulateFlightInfo(flightNumber string) flightInfo {
	// Simulated — in production: call FlightAware/Cirium API
	now := time.Now().Add(30 * time.Minute)
	prefix := strings.ToUpper(flightNumber[:2])

	airlines := map[string]string{
		"AM": "Aeromexico", "VB": "VivaAerobus", "4O": "Volaris",
		"Y4": "Volaris", "AA": "American Airlines", "UA": "United Airlines",
		"DL": "Delta Airlines", "LA": "LATAM Airlines",
	}
	airline := airlines[prefix]
	if airline == "" {
		airline = "Airline " + prefix
	}

	origins := []string{"CDMX", "GDL", "MTY", "CUN", "LAX", "MIA", "BOG", "SCL", "GRU"}
	origin := origins[len(flightNumber)%len(origins)]

	return flightInfo{
		airline:     airline,
		origin:      origin,
		status:      "on_time",
		arrivalTime: &now,
		gate:        fmt.Sprintf("%c%d", 'A'+rune(len(flightNumber)%4), 1+len(flightNumber)%20),
		passengers:  80 + len(flightNumber)*7%120,
	}
}

func airportCoords(airportID string) (float64, float64) {
	coords := map[string][2]float64{
		"CUN": {21.0365, -86.8771}, // Cancun
		"MEX": {19.4363, -99.0721}, // CDMX
		"GDL": {20.5218, -103.3113}, // Guadalajara
		"MTY": {25.7785, -100.1070}, // Monterrey
		"TIJ": {32.5411, -116.9700}, // Tijuana
		"SJD": {23.1518, -109.7215}, // Los Cabos
		"PVR": {20.6801, -105.2543}, // Puerto Vallarta
		"MID": {20.9370, -89.6577},  // Merida
	}
	if c, ok := coords[airportID]; ok {
		return c[0], c[1]
	}
	return 19.4363, -99.0721 // Default: CDMX
}

func airportDestinations(airportID string) []destInfo {
	destinations := map[string][]destInfo{
		"CUN": {
			{"Zona Hotelera Cancún", 21.1300, -86.7522},
			{"Playa del Carmen", 20.6296, -87.0739},
			{"Tulum", 20.2114, -87.4654},
			{"Puerto Morelos", 20.8403, -86.8753},
			{"Riviera Maya", 20.5000, -87.2200},
		},
		"MEX": {
			{"Centro Histórico CDMX", 19.4326, -99.1332},
			{"Polanco", 19.4352, -99.1944},
			{"Santa Fe", 19.3573, -99.2742},
			{"Reforma", 19.4270, -99.1677},
			{"Coyoacán", 19.3500, -99.1620},
		},
		"GDL": {
			{"Centro Histórico GDL", 20.6668, -103.3918},
			{"Zona Chapultepec", 20.6747, -103.3744},
			{"Tlaquepaque", 20.6410, -103.3438},
			{"Zapopan", 20.7214, -103.3882},
		},
		"SJD": {
			{"Cabo San Lucas", 22.8905, -109.9167},
			{"San José del Cabo Centro", 23.0573, -109.6981},
			{"Corredor Turístico", 22.9700, -109.8100},
		},
		"PVR": {
			{"Zona Romántica PV", 20.6600, -105.2340},
			{"Marina Vallarta", 20.6720, -105.2510},
			{"Nuevo Vallarta", 20.7000, -105.2940},
		},
	}

	if d, ok := destinations[airportID]; ok {
		return d
	}
	// Default destinations
	return []destInfo{
		{"Centro Ciudad", 19.4326, -99.1332},
		{"Zona Hotelera", 19.4200, -99.1700},
	}
}

func classifyDemand(hour int) string {
	switch {
	case hour >= 7 && hour <= 9:
		return "high"
	case hour >= 17 && hour <= 19:
		return "surge"
	case hour >= 22 || hour < 5:
		return "low"
	default:
		return "normal"
	}
}

func kioskWelcome(lang string, hour int) string {
	greetings := map[string][]string{
		"es": {"Buenos días", "Buenas tardes", "Buenas noches"},
		"en": {"Good morning", "Good afternoon", "Good evening"},
		"pt": {"Bom dia", "Boa tarde", "Boa noite"},
	}

	msgs, ok := greetings[lang]
	if !ok {
		msgs = greetings["es"]
	}

	idx := 0
	if hour >= 12 && hour < 18 {
		idx = 1
	} else if hour >= 18 || hour < 6 {
		idx = 2
	}

	suffixes := map[string]string{
		"es": " — ¡Bienvenido a GoDestino! Reserve su transporte en segundos.",
		"en": " — Welcome to GoDestino! Book your transport in seconds.",
		"pt": " — Bem-vindo ao GoDestino! Reserve seu transporte em segundos.",
	}

	suffix := suffixes[lang]
	if suffix == "" {
		suffix = suffixes["es"]
	}

	return msgs[idx] + suffix
}

func quickBookMessage(lang, bookingNumber string) string {
	msgs := map[string]string{
		"es": fmt.Sprintf("¡Reserva %s confirmada! Su conductor será asignado en breve. Presente el código QR al conductor.", bookingNumber),
		"en": fmt.Sprintf("Booking %s confirmed! Your driver will be assigned shortly. Show the QR code to your driver.", bookingNumber),
		"pt": fmt.Sprintf("Reserva %s confirmada! Seu motorista será designado em breve. Mostre o código QR ao motorista.", bookingNumber),
	}
	if m, ok := msgs[lang]; ok {
		return m
	}
	return msgs["es"]
}

func receiptFooter(lang string) string {
	footers := map[string]string{
		"es": "GoDestino — Transporte confiable en aeropuertos de LATAM. Conserve este recibo. Para soporte: soporte@godestino.com",
		"en": "GoDestino — Reliable airport transport in LATAM. Keep this receipt. Support: support@godestino.com",
		"pt": "GoDestino — Transporte confiável em aeroportos da LATAM. Guarde este recibo. Suporte: suporte@godestino.com",
	}
	if f, ok := footers[lang]; ok {
		return f
	}
	return footers["es"]
}

func localizeDestTitle(lang, name string) string {
	return name
}

func localizeDestSubtitle(lang string, eta int, price int64) string {
	switch lang {
	case "en":
		return fmt.Sprintf("~%d min • $%.2f MXN", eta, float64(price)/100)
	case "pt":
		return fmt.Sprintf("~%d min • $%.2f MXN", eta, float64(price)/100)
	default:
		return fmt.Sprintf("~%d min • $%.2f MXN", eta, float64(price)/100)
	}
}

func localizeServiceTitle(lang, svc string) string {
	titles := map[string]map[string]string{
		"shuttle": {"es": "Shuttle Compartido", "en": "Shared Shuttle", "pt": "Shuttle Compartilhado"},
	}
	if t, ok := titles[svc]; ok {
		if v, ok := t[lang]; ok {
			return v
		}
		return t["es"]
	}
	return svc
}

func localizeServiceSub(lang, svc string) string {
	subs := map[string]map[string]string{
		"shuttle": {
			"es": "La opción más económica — salidas cada 30 min",
			"en": "Most affordable option — departures every 30 min",
			"pt": "Opção mais econômica — partidas a cada 30 min",
		},
	}
	if s, ok := subs[svc]; ok {
		if v, ok := s[lang]; ok {
			return v
		}
		return s["es"]
	}
	return ""
}

func localizeTransport(lang, svc string) string {
	names := map[string]map[string]string{
		"taxi":    {"es": "Taxi Privado", "en": "Private Taxi", "pt": "Táxi Privado"},
		"van":     {"es": "Van Ejecutiva", "en": "Executive Van", "pt": "Van Executiva"},
		"shuttle": {"es": "Shuttle Compartido", "en": "Shared Shuttle", "pt": "Shuttle Compartilhado"},
		"bus":     {"es": "Autobús", "en": "Bus", "pt": "Ônibus"},
	}
	if n, ok := names[svc]; ok {
		if v, ok := n[lang]; ok {
			return v
		}
		return n["es"]
	}
	return svc
}

func localizeTransportDesc(lang, svc string) string {
	descs := map[string]map[string]string{
		"taxi": {
			"es": "Servicio puerta a puerta, directo a su destino",
			"en": "Door-to-door service, direct to your destination",
			"pt": "Serviço porta a porta, direto ao seu destino",
		},
		"van": {
			"es": "Ideal para grupos con equipaje, asientos cómodos",
			"en": "Ideal for groups with luggage, comfortable seats",
			"pt": "Ideal para grupos com bagagem, assentos confortáveis",
		},
		"shuttle": {
			"es": "Compartido con otros pasajeros, la opción más económica",
			"en": "Shared with other passengers, the most affordable option",
			"pt": "Compartilhado com outros passageiros, a opção mais econômica",
		},
		"bus": {
			"es": "Servicio regular, múltiples paradas, precio económico",
			"en": "Regular service, multiple stops, budget-friendly",
			"pt": "Serviço regular, múltiplas paradas, preço econômico",
		},
	}
	if d, ok := descs[svc]; ok {
		if v, ok := d[lang]; ok {
			return v
		}
		return d["es"]
	}
	return ""
}

func generateReceiptNumber() (string, error) {
	const charset = "0123456789"
	result := make([]byte, 12)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return "REC-" + string(result), nil
}

func generateCardNumber() (string, error) {
	const charset = "0123456789"
	result := make([]byte, 16)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	// Format: GD-XXXX-XXXX-XXXX-XXXX
	return fmt.Sprintf("GD-%s-%s-%s-%s", string(result[0:4]), string(result[4:8]), string(result[8:12]), string(result[12:16])), nil
}
