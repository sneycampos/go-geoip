package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLookupEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		lookup     func(string) (GeoIPResponse, error)
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			lookup: func(ip string) (GeoIPResponse, error) {
				return GeoIPResponse{
					IP:          ip,
					Country:     "Portugal",
					CountryCode: "PT",
					City:        "Lisbon",
					PostalCode:  "1000-005",
					Latitude:    38.7219,
					Longitude:   -9.1398,
					Timezone:    "Europe/Lisbon",
				}, nil
			},
			wantStatus: http.StatusOK,
			wantBody: `{"ip":"172.68.100.190","country":"Portugal","country_code":"PT",` +
				`"city":"Lisbon","postal_code":"1000-005","latitude":38.7219,` +
				`"longitude":-9.1398,"timezone":"Europe/Lisbon"}` + "\n",
		},
		{
			name: "invalid IP",
			lookup: func(string) (GeoIPResponse, error) {
				return GeoIPResponse{}, errInvalidIP
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "{\"error\":\"invalid IP\"}\n",
		},
		{
			name: "IP not found",
			lookup: func(string) (GeoIPResponse, error) {
				return GeoIPResponse{}, errIPNotFound
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "{\"error\":\"information not found for IP\"}\n",
		},
		{
			name: "internal error",
			lookup: func(string) (GeoIPResponse, error) {
				return GeoIPResponse{}, errors.New("database unavailable")
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "{\"error\":\"internal server error\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/172.68.100.190", nil)
			response := httptest.NewRecorder()

			newHandler(test.lookup).ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d; want %d", response.Code, test.wantStatus)
			}

			if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Content-Type = %q; want %q", contentType, "application/json")
			}

			if statusHeader := response.Header().Get("status"); statusHeader != "" {
				t.Errorf("unexpected status header %q", statusHeader)
			}

			if body := response.Body.String(); body != test.wantBody {
				t.Errorf("body = %q; want %q", body, test.wantBody)
			}
		})
	}
}

func TestNoContentRoutes(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "preflight", method: http.MethodOptions, path: "/172.68.100.190"},
		{name: "health check", method: http.MethodGet, path: "/healthz"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, test.path, nil)
			response := httptest.NewRecorder()
			lookup := func(string) (GeoIPResponse, error) {
				t.Fatal("lookup should not be called")
				return GeoIPResponse{}, nil
			}

			newHandler(lookup).ServeHTTP(response, request)

			if response.Code != http.StatusNoContent {
				t.Fatalf("status = %d; want %d", response.Code, http.StatusNoContent)
			}

			assertCORSHeaders(t, response)

			if response.Body.Len() != 0 {
				t.Errorf("body = %q; want empty body", response.Body.String())
			}
		})
	}
}

func TestLookupIPRejectsInvalidAddress(t *testing.T) {
	_, err := lookupIP(nil, "not-an-ip")
	if !errors.Is(err, errInvalidIP) {
		t.Fatalf("error = %v; want %v", err, errInvalidIP)
	}
}

func assertCORSHeaders(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()

	wantHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type",
	}

	for header, want := range wantHeaders {
		if got := response.Header().Get(header); got != want {
			t.Errorf("%s = %q; want %q", header, got, want)
		}
	}
}
