package main

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/orhaniscoding/goconnect/client-daemon/internal/api"
	"github.com/orhaniscoding/goconnect/client-daemon/internal/identity"
)

var (
	version = "dev"
	commit  = "none"
	date    = "2025-09-22"
	builtBy = "orhaniscoding"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	serverURL := flag.String("server", "http://localhost:8080", "GoConnect server URL")
	flag.Parse()

	if *showVersion {
		fmt.Printf("goconnect-daemon %s (commit %s, build %s) built by %s\n", version, commit, date, builtBy)
		return
	}

	// Initialize Identity Manager
	idMgr, err := identity.NewManager()
	if err != nil {
		log.Fatalf("failed to init identity manager: %v", err)
	}

	id, err := idMgr.LoadOrGenerate()
	if err != nil {
		log.Fatalf("failed to load/generate identity: %v", err)
	}
	log.Printf("Device Identity: %s (Registered: %v)", id.PublicKey, id.DeviceID != "")

	// Initialize API Client
	apiClient := api.NewClient(*serverURL)

	// Choose a pseudo-random port in a small range using crypto/rand
	var b [2]byte
	if _, err := rand.Read(b[:]); err != nil {
		log.Fatalf("failed to read random: %v", err)
	}
	n := binary.BigEndian.Uint16(b[:]) % 1000
	port := int(12000 + n)

	// Setup Handlers
	setupHandlers(idMgr, apiClient)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	log.Printf("daemon bridge at http://%s", addr)
	srv := &http.Server{
		Addr:              addr,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
		Handler:           nil,
	}
	log.Fatal(srv.ListenAndServe())
}

func setupHandlers(idMgr *identity.Manager, apiClient *api.Client) {
	// CORS middleware
	cors := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*") // Dev only, restrict in prod
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next(w, r)
		}
	}

	http.HandleFunc("/status", cors(func(w http.ResponseWriter, _ *http.Request) {
		id := idMgr.Get()
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"running": true,
			"version": version,
			"device": map[string]interface{}{
				"registered": id.DeviceID != "",
				"public_key": id.PublicKey,
				"device_id":  id.DeviceID,
			},
			"wg": map[string]interface{}{
				"active": false,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))

	http.HandleFunc("/register", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if req.Token == "" {
			http.Error(w, "Token required", http.StatusBadRequest)
			return
		}

		hostname, _ := os.Hostname()

		// Register with server
		regReq := api.RegisterDeviceRequest{
			Name:      hostname,
			Platform:  runtime.GOOS,
			PubKey:    idMgr.Get().PublicKey,
			HostName:  hostname,
			OSVersion: runtime.GOOS, // TODO: Get real OS version
			DaemonVer: version,
		}

		resp, err := apiClient.Register(r.Context(), req.Token, regReq)
		if err != nil {
			log.Printf("Registration failed: %v", err)
			http.Error(w, fmt.Sprintf("Registration failed: %v", err), http.StatusInternalServerError)
			return
		}

		// Save device ID
		if err := idMgr.Update(resp.ID, req.Token); err != nil {
			log.Printf("Failed to save identity: %v", err)
			http.Error(w, "Failed to save identity", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":    "success",
			"device_id": resp.ID,
		})
	}))
}
