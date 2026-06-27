package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"syscall"

	"github.com/oschwald/maxminddb-golang/v2"
)

var (
	errInvalidIP  = errors.New("invalid IP")
	errIPNotFound = errors.New("information not found for IP")
)

type GeoIPResponse struct {
	IP          string  `json:"ip"`
	Country     string  `json:"country,omitempty"`
	CountryCode string  `json:"country_code,omitempty"`
	City        string  `json:"city,omitempty"`
	PostalCode  string  `json:"postal_code,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	Timezone    string  `json:"timezone,omitempty"`
}

type GeoIPRecord struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code"`
		Names   struct {
			En string `maxminddb:"en"`
		} `maxminddb:"names"`
	} `maxminddb:"country"`
	City struct {
		Names struct {
			En string `maxminddb:"en"`
		} `maxminddb:"names"`
	} `maxminddb:"city"`
	Location struct {
		Latitude  float64 `maxminddb:"latitude"`
		Longitude float64 `maxminddb:"longitude"`
		TimeZone  string  `maxminddb:"time_zone"`
	} `maxminddb:"location"`
	Postal struct {
		Code string `maxminddb:"code"`
	} `maxminddb:"postal"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func main() {
	if err := run(); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	db, err := maxminddb.Open("GeoLite2-City.mmdb")
	if err != nil {
		return fmt.Errorf("opening MMDB database: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Println("Error closing MMDB database:", err)
		}
	}()

	server := &http.Server{
		Addr: ":8888",
		Handler: newHandler(func(ip string) (GeoIPResponse, error) {
			return lookupIP(db, ip)
		}),
	}

	signalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	log.Println("Server running at http://localhost:8888")

	select {
	case err := <-serverErrors:
		return fmt.Errorf("serving HTTP: %w", err)
	case <-signalContext.Done():
		log.Println("Shutting down server")
	}

	if err := server.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("shutting down HTTP server: %w", err)
	}

	if err := <-serverErrors; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serving HTTP: %w", err)
	}

	log.Println("Server stopped")

	return nil
}

func newHandler(lookup func(string) (GeoIPResponse, error)) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /{ip}", handleIP(lookup))
	mux.HandleFunc("OPTIONS /{ip}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	return corsMiddleware(mux)
}

func handleIP(lookup func(string) (GeoIPResponse, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.PathValue("ip")

		response, err := lookup(ip)
		if err != nil {
			handleLookupError(w, ip, err)
			return
		}

		sendJSONResponse(w, http.StatusOK, response)
	}
}

func handleLookupError(w http.ResponseWriter, ip string, err error) {
	status := http.StatusInternalServerError
	message := "internal server error"

	switch {
	case errors.Is(err, errInvalidIP):
		status = http.StatusBadRequest
		message = errInvalidIP.Error()
	case errors.Is(err, errIPNotFound):
		status = http.StatusNotFound
		message = errIPNotFound.Error()
	default:
		log.Printf("Error looking up IP %q: %v", ip, err)
	}

	sendJSONResponse(w, status, errorResponse{Error: message})
}

// lookupIP queries the MMDB database for IP information.
func lookupIP(db *maxminddb.Reader, ipStr string) (GeoIPResponse, error) {
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return GeoIPResponse{}, errInvalidIP
	}

	result := db.Lookup(ip)
	if err := result.Err(); err != nil {
		return GeoIPResponse{}, fmt.Errorf("looking up IP %q: %w", ipStr, err)
	}

	if !result.Found() {
		return GeoIPResponse{}, errIPNotFound
	}

	var record GeoIPRecord
	if err := result.Decode(&record); err != nil {
		return GeoIPResponse{}, fmt.Errorf("decoding result for IP %q: %w", ipStr, err)
	}

	if record.Country.ISOCode == "" {
		return GeoIPResponse{}, errIPNotFound
	}

	return GeoIPResponse{
		IP:          ipStr,
		Country:     record.Country.Names.En,
		CountryCode: record.Country.ISOCode,
		City:        record.City.Names.En,
		PostalCode:  record.Postal.Code,
		Latitude:    record.Location.Latitude,
		Longitude:   record.Location.Longitude,
		Timezone:    record.Location.TimeZone,
	}, nil
}

func sendJSONResponse(w http.ResponseWriter, status int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		next.ServeHTTP(w, r)
	})
}
