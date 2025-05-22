package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/oschwald/maxminddb-golang"
	"github.com/patrickmn/go-cache"
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

// initializes a cache with 1-hour expiration and 10-minute cleanup
var c = cache.New(1*time.Hour, 10*time.Minute)

func main() {
	db, err := maxminddb.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatal("Error opening MMDB database:", err)
	}
	defer func(db *maxminddb.Reader) {
		err := db.Close()
		if err != nil {
			log.Println("Error closing MMDB database:", err)
		}
	}(db)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ip", handleRoot(db))
	mux.HandleFunc("GET /ip/{ip}/", handleIP(db))

	log.Println("Server running at http://localhost:8888")
	if err := http.ListenAndServe(":8888", mux); err != nil {
		log.Fatal("Error starting server:", err)
	}
}

func handleRoot(db *maxminddb.Reader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := getIPFromRequest(r)
		if ip == "" {
			http.Error(w, `{"error":"Unable to determine IP"}`, http.StatusBadRequest)
			return
		}

		response, err := lookupIP(db, ip)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		sendJSONResponse(w, response)
	}
}

func handleIP(db *maxminddb.Reader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.PathValue("ip")

		response, err := lookupIP(db, ip)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		sendJSONResponse(w, response)
	}
}

// extracts the IP from the request, considering X-Forwarded-For
func getIPFromRequest(r *http.Request) string {
	// Check X-Forwarded-For header (for proxies)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		// May contain multiple IPs, take the first one
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0])
	}

	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// queries the MMDB database for IP information
func lookupIP(db *maxminddb.Reader, ipStr string) (GeoIPResponse, error) {
	if cached, found := c.Get(ipStr); found {
		return cached.(GeoIPResponse), nil
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return GeoIPResponse{}, fmt.Errorf("invalid IP")
	}

	var record GeoIPRecord
	if err := db.Lookup(ip, &record); err != nil {
		log.Printf("Error querying database: %v", err)
		return GeoIPResponse{}, fmt.Errorf("internal server error")
	}

	if record.Country.ISOCode == "" {
		return GeoIPResponse{}, fmt.Errorf("information not found for IP")
	}

	response := GeoIPResponse{
		IP:          ipStr,
		Country:     record.Country.Names.En,
		CountryCode: record.Country.ISOCode,
		City:        record.City.Names.En,
		PostalCode:  record.Postal.Code,
		Latitude:    record.Location.Latitude,
		Longitude:   record.Location.Longitude,
		Timezone:    record.Location.TimeZone,
	}

	c.Set(ipStr, response, cache.DefaultExpiration)
	return response, nil
}

func sendJSONResponse(w http.ResponseWriter, response GeoIPResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Add("status", "200")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
	}
}
