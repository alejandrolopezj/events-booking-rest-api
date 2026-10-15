package routes

import (
	"net/http"
	"strconv"

	"github.com/alejandrolopezj/events-booking-rest-api/models"
	"github.com/alejandrolopezj/events-booking-rest-api/utils"
	"github.com/gin-gonic/gin"
)

// registerForEvent registers the authenticated user for the event.
func registerForEvent(context *gin.Context) {
	claims, _ := context.Get("claims")

	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := models.GetEventByID(id)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if event.UserID != claims.(*utils.JWTClaims).UserID {
		context.JSON(http.StatusBadRequest, gin.H{"error": "You cannot register for this event"})
		return
	}

	err = event.Register(claims.(*utils.JWTClaims).UserID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "Successfully registered for event"})
}

// unregisterFromEvent removes the authenticated user's registration from the event.
func unregisterFromEvent(context *gin.Context) {
	claims, _ := context.Get("claims")

	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := models.GetEventByID(id)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if event.UserID != claims.(*utils.JWTClaims).UserID {
		context.JSON(http.StatusBadRequest, gin.H{"error": "You cannot unregister from this event"})
		return
	}

	err = event.Unregister(claims.(*utils.JWTClaims).UserID)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Successfully unregistered from event"})
}
