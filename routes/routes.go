// Package routes registers the HTTP endpoints for the API.
package routes

import (
	"github.com/alejandrolopezj/events-booking-rest-api/middlewares"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers public and authenticated API endpoints on server.
func RegisterRoutes(server *gin.Engine) {
	server.GET("/events", getEvents)
	server.GET("/events/:id", getEvent)

	authenticated := server.Group("/")
	authenticated.Use(middlewares.Authenticate)
	authenticated.POST("/events", createEvent)
	authenticated.PUT("/events/:id", updateEvent)
	authenticated.DELETE("/events/:id", deleteEvent)
	authenticated.POST("/events/:id/registration", registerForEvent)
	authenticated.DELETE("/events/:id/registration", unregisterFromEvent)

	server.POST("/signup", signup)
	server.POST("/login", login)
}
