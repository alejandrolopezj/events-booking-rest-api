// Package middlewares contains reusable HTTP middleware for the API.
package middlewares

import (
	"net/http"

	"github.com/alejandrolopezj/events-booking-rest-api/utils"
	"github.com/gin-gonic/gin"
)

const errorMessage = "Not authorized"

// Authenticate validates the Authorization header and ensures the request carries
// a valid JWT. If the token is missing or invalid, the request is aborted with a
// 401 Unauthorized response.
func Authenticate(context *gin.Context) {
	token := context.Request.Header.Get("Authorization")
	if token == "" {
		utils.Logger.Error("Missing Authorization header")
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
		return
	}

	claims, err := utils.ValidateJWT(token)
	if err != nil {
		utils.Logger.Error("Invalid Authorization token", "error", err)
		context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": errorMessage})
		return
	}

	context.Set("claims", claims)
	context.Next()
}
