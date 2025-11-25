package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GinWrap wraps a standard http.HandlerFunc to work with gin
// It preserves path parameters by copying them to the request
func GinWrap(fn http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Copy gin path parameters to request context via SetPathValue (Go 1.22+)
		for _, param := range c.Params {
			c.Request.SetPathValue(param.Key, param.Value)
		}

		// Copy gin headers to request (tenant/user context)
		if tenantID, exists := c.Get("tenant_id"); exists {
			c.Request.Header.Set("X-Tenant-ID", tenantID.(string))
		}
		if userID, exists := c.Get("user_id"); exists {
			c.Request.Header.Set("X-User-ID", userID.(string))
		}

		fn(c.Writer, c.Request)
	}
}
