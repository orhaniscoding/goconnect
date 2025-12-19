package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/metrics"
)

// MetricsHandler provides JSON API for metrics dashboard.
type MetricsHandler struct{}

// NewMetricsHandler creates a new metrics handler.
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// GetSummary returns a JSON summary of key metrics for the dashboard.
// GET /v1/metrics/summary
func (h *MetricsHandler) GetSummary(c *gin.Context) {
	summary := metrics.GetSummary()
	c.JSON(http.StatusOK, gin.H{
		"data": summary,
	})
}

// GetHealth returns health status with additional metrics.
// GET /v1/metrics/health
func (h *MetricsHandler) GetHealth(c *gin.Context) {
	summary := metrics.GetSummary()

	status := "healthy"
	if summary.WSConnections == 0 && summary.PeersOnline == 0 {
		status = "idle"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         status,
		"ws_connections": summary.WSConnections,
		"networks":       summary.NetworksActive,
		"peers_online":   summary.PeersOnline,
	})
}
