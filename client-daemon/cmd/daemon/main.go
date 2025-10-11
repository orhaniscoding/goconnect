package main

import (
	"crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var (
	version = "dev"
	commit  = "none"
	date    = "2025-09-22"
	builtBy = "orhaniscoding"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("goconnect-daemon %s (commit %s, build %s) built by %s\n", version, commit, date, builtBy)
		return
	}
	// Choose a pseudo-random port in a small range using crypto/rand
	var b [2]byte
	if _, err := rand.Read(b[:]); err != nil {
		log.Fatalf("failed to read random: %v", err)
	}
	n := binary.BigEndian.Uint16(b[:]) % 1000
	port := int(12000 + n)
	http.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"running":true,"wg":{"active":false}}`)); err != nil {
			// best-effort logging only
			log.Printf("write error: %v", err)
		}
	})
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
