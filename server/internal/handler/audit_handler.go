package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/orhaniscoding/goconnect/server/internal/audit"
	"github.com/orhaniscoding/goconnect/server/internal/domain"
	"github.com/orhaniscoding/goconnect/server/internal/metrics"
	"net/http"
	"strconv"
	"time"
)

type AuditExportProvider interface {
	ExportIntegrity(ctx gin.Context, limit int) (audit.IntegrityExport, error)
}

// AuditListHandler returns a gin handler for listing audit logs
func AuditListHandler(aud audit.Auditor) gin.HandlerFunc {
	// Attempt to unwrap to *audit.SqliteAuditor
	sa, _ := aud.(*audit.SqliteAuditor)
	return func(c *gin.Context) {
		if sa == nil {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "audit log listing not supported for this auditor"})
			return
		}

		// Get tenant from context (set by auth middleware)
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userClaims := claims.(*domain.TokenClaims)

		// Pagination
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 20
		}
		offset := (page - 1) * limit

		logs, total, err := sa.QueryLogs(c.Request.Context(), userClaims.TenantID, limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query logs"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": logs,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
			},
		})
	}
}

// AuditIntegrityHandler returns a gin handler exposing chain head + anchors.
func AuditIntegrityHandler(aud audit.Auditor) gin.HandlerFunc {
	// Attempt to unwrap to *audit.SqliteAuditor for export functionality.
	sa, _ := aud.(*audit.SqliteAuditor) // if not sqlite returns empty 501
	return func(c *gin.Context) {
		start := time.Now()
		anchorsParam := c.Query("anchors")
		limit := 20
		if anchorsParam != "" {
			if v, err := strconv.Atoi(anchorsParam); err == nil && v > 0 && v <= 500 { // cap to 500
				limit = v
			}
		}
		if sa == nil {
			c.JSON(http.StatusNotImplemented, gin.H{"error": "integrity export not supported for this auditor"})
			return
		}
		exp, err := sa.ExportIntegrity(c.Request.Context(), limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export integrity"})
			return
		}
		c.JSON(http.StatusOK, exp)
		metrics.ObserveIntegrityExport(time.Since(start).Seconds())
	}
}
