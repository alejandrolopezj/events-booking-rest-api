package routes

import (
	"net/http"

	"github.com/alejandrolopezj/events-booking-rest-api/models"
	"github.com/alejandrolopezj/events-booking-rest-api/utils"
	"github.com/gin-gonic/gin"
)

// signup creates a user account from the JSON request body.
func signup(context *gin.Context) {
	var user models.User

	err := context.ShouldBindJSON(&user)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = user.Save()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	context.JSON(http.StatusCreated, gin.H{"message": "User signed up successfully"})
}

// login validates the submitted credentials and returns an authentication JWT.
func login(context *gin.Context) {
	var user models.User

	err := context.ShouldBindJSON(&user)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = user.ValidateCredentials(user.Password)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": utils.LogAndReturnError(
				err,
				"Could not generate the authentication token",
				"Error generating the authentication token",
			).Error(),
		})
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "Login successful", "token": token})
}
