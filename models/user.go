package models

import (
	"errors"

	"github.com/alejandrolopezj/events-booking-rest-api/db"
	"github.com/alejandrolopezj/events-booking-rest-api/utils"
)

type User struct {
	ID       int64  `json:"id"`
	Email    string `binding:"required,email" json:"email"`
	Password string `binding:"required" json:"password"`
}

func (u *User) Save() error {
	errorUserMessage := "Error saving the user"
	query := `
	INSERT INTO users (email, password)
	VALUES (?, ?)
	`
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		technicalMessage := "Could not prepare statement to save the user"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}
	defer stmt.Close()

	hashedPassword, err := utils.HashPassword(u.Password)
	if err != nil {
		technicalMessage := "Could not hash the user's password"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}

	result, err := stmt.Exec(u.Email, hashedPassword)
	if err != nil {
		technicalMessage := "Could not execute statement to save the user"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}

	id, err := result.LastInsertId()
	if err != nil {
		technicalMessage := "Could not retrieve last inserted user ID"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}

	u.ID = id

	return nil
}

func (u *User) ValidateCredentials(password string) error {
	errorUserMessage := "Invalid user credentials"
	query := `SELECT id, password FROM users WHERE email = ?`
	row := db.DB.QueryRow(query, u.Email)

	var hashedPassword string
	err := row.Scan(&u.ID, &hashedPassword)
	if err != nil {
		technicalMessage := "Could not retrieve user credentials"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}

	if !utils.CheckPasswordHash(password, hashedPassword) {
		technicalMessage := "User password does not match"
		return utils.LogAndReturnError(errors.New("password does not match"), technicalMessage, errorUserMessage)
	}

	return nil
}
