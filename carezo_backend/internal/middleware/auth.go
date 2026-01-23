package middleware

import (
	"net/http"
	"strings"

	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/utils"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *configs.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get token from header in format "Bearer <toke>"
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return 
		}

		// Extract token (remove "Bearer ")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "Invalid authorization header format")
			c.Abort()
			return 
		}
		token := parts[1]

		// Validate token
		claims, err := utils.ValidateToken(token, cfg)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return 
		}

		// Add user info to conxtext so handlers can access it
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)


		// continue to next handler
		c.Next()
	}
}


// function to check if user has required role(admin, user etc)
func RequireRole(role string) gin.HandlerFunc {
	return func (c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		// Check if user has a required role
		if userRole != role {
			response.Error(
				c,
				http.StatusForbidden, "Permission denied")
				return
		}
		c.Next()
	}
}