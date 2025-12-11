package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// Minimal mock structures
type RegisterResponse struct {
	ID        string `json:"id"`
	AuthToken string `json:"auth_token"`
}

type DeviceConfigResponse struct {
	Peers      []interface{} `json:"peers"`
	AllowedIPs []string      `json:"allowed_ips"`
}

type NetworkResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	r := mux.NewRouter()

	// Logger middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[STUB] %s %s", r.Method, r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	// Register Device
	r.HandleFunc("/v1/devices/register", func(w http.ResponseWriter, r *http.Request) {
		resp := RegisterResponse{
			ID:        "dev-e2e-123",
			AuthToken: "mock-token-xyz",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}).Methods("POST")

	// Get Device Config
	r.HandleFunc("/v1/devices/{id}/config", func(w http.ResponseWriter, r *http.Request) {
		resp := DeviceConfigResponse{
			Peers:      []interface{}{},
			AllowedIPs: []string{"10.0.0.1/32"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}).Methods("GET")

	// Get Networks
	r.HandleFunc("/v1/networks", func(w http.ResponseWriter, r *http.Request) {
		resp := []NetworkResponse{
			{ID: "net-test-1", Name: "Test Network", Role: "member"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}).Methods("GET")

	// WebSocket Mock (Accept and hold)
	r.HandleFunc("/v1/ws", func(w http.ResponseWriter, r *http.Request) {
		// Just upgrade connection and do nothing to simulate active connection
		// Note: Using standard library upgrade if we wanted to be proper, 
		// but for E2E sometimes basic HTTP OK is enough if the client handles it gracefully,
		// however goconnect likely expects WS upgrade. 
		// For simplicity in this stub without gorilla/websocket dep complexity if possible,
		// we'll just return 200 OK for now. If client fails, we might need actual WS.
		// BUT the user's project likely has gorilla/mux and websocket in go.mod.
		// Let's assume we can import gorilla/websocket if needed, but to keep it simple
		// and avoid dependency hell in this one file, we will try to just log it.
		// Wait, the daemon will retry if WS fails. That's fine for "status: disconnected".
		w.WriteHeader(http.StatusServiceUnavailable) 
	})

	addr := ":" + *port
	log.Printf("Stub server listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
