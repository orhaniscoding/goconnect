package handler

import (
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/orhaniscoding/goconnect/server/internal/service"
)

// IPFilterMiddleware filters requests based on IP rules
type IPFilterMiddleware struct {
	svc *service.IPRuleService
}

// NewIPFilterMiddleware creates a new IP filter middleware
func NewIPFilterMiddleware(svc *service.IPRuleService) *IPFilterMiddleware {
	return &IPFilterMiddleware{svc: svc}
}

// Middleware returns an HTTP middleware that filters requests based on IP rules
func (m *IPFilterMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			// No tenant context, skip IP filtering
			next.ServeHTTP(w, r)
			return
		}

		clientIP := getClientIP(r)
		if clientIP == "" {
			log.Printf("WARN: Could not determine client IP for request")
			next.ServeHTTP(w, r)
			return
		}

		allowed, matchedRule, err := m.svc.CheckIP(r.Context(), tenantID, clientIP)
		if err != nil {
			log.Printf("ERROR: IP check failed: %v", err)
			// Fail open on error to avoid blocking legitimate traffic
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			ruleInfo := "no matching allow rule"
			if matchedRule != nil {
				ruleInfo = "matched deny rule: " + matchedRule.CIDR
			}
			log.Printf("INFO: IP %s blocked for tenant %s: %s", clientIP, tenantID, ruleInfo)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"access_denied","message":"IP address is not allowed"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxied requests)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if parsedIP := net.ParseIP(ip); parsedIP != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		if parsedIP := net.ParseIP(xri); parsedIP != nil {
			return xri
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr might not have a port
		return r.RemoteAddr
	}

	return host
}
