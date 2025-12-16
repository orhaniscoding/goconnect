package main

import (
	"encoding/json"
	"flag"
	"log/slog" // Changed from "log" to "log/slog"
	"net/http"
	"os" // Added for os.Exit

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
	ID         string `json:"id"`
	Name       string `json:"name"`
	InviteCode string `json:"invite_code,omitempty"`
	Role       string `json:"role"`
}

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()

	r := mux.NewRouter()

	// Logger middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Info("[STUB]", "method", r.Method, "path", r.URL.Path)
			next.ServeHTTP(w, r)
		})
	})

	// Register Device
	r.HandleFunc("/v1/devices/register", func(w http.ResponseWriter, _ *http.Request) {
		resp := RegisterResponse{
			ID:        "dev-e2e-123",
			AuthToken: "mock-token-xyz",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}).Methods("POST")

	// Get Device Config
	r.HandleFunc("/v1/devices/{id}/config", func(w http.ResponseWriter, _ *http.Request) {
		resp := DeviceConfigResponse{
			Peers:      []interface{}{},
			AllowedIPs: []string{"10.0.0.1/32"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}).Methods("GET")

	// Get Networks
	r.HandleFunc("/v1/networks", func(w http.ResponseWriter, _ *http.Request) {
		resp := []NetworkResponse{
			{ID: "net-test-1", Name: "Test Network", Role: "member"},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}).Methods("GET")

	// Create Network (for E2E)
	r.HandleFunc("/v1/networks", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name string `json:"name"`
			CIDR string `json:"cidr"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		slog.Info("[STUB] CreateNetwork", "name", req.Name, "cidr", req.CIDR)
		
		resp := NetworkResponse{
			ID:         "net-created-e2e",
			Name:       req.Name,
			InviteCode: "inv-e2e-code",
			Role:       "owner",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}).Methods("POST")

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
	slog.Info("Stub server listening", "addr", addr) // Changed from log.Printf to slog.Info
	if err := http.ListenAndServe(addr, r); err != nil { // Kept 'r' as the handler, not 'nil'
		slog.Error("Server error", "error", err) // Changed from log.Fatal to slog.Error
		os.Exit(1)                               // Added os.Exit(1)
	}
}
