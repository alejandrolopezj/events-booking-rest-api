// Package models contains the data models and persistence logic for the API.
package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/alejandrolopezj/events-booking-rest-api/db"
	"github.com/alejandrolopezj/events-booking-rest-api/utils"
)

// Event represents an event created by a user and stored in the database.
type Event struct {
	ID          int64     `json:"id"`
	Title       string    `binding:"required" json:"title"`
	Description string    `binding:"required" json:"description"`
	Location    string    `binding:"required" json:"location"`
	DateTime    time.Time `binding:"required" json:"date_time"`
	UserID      int64     `json:"user_id"`
}

// Save inserts the event into the events table and stores the generated ID.
func (e *Event) Save() error {
	errorUserMessage := "Error saving the event"
	query := `
	INSERT INTO events (title, description, location, date_time, user_id)
	VALUES (?, ?, ?, ?, ?)
	`
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		technicalMessage := "Could not prepare statement to save the event"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}
	defer stmt.Close()

	result, err := stmt.Exec(e.Title, e.Description, e.Location, e.DateTime, e.UserID)
	if err != nil {
		technicalMessage := "Could not execute statement to save the event"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}

	id, err := result.LastInsertId()
	if err != nil {
		technicalMessage := "Could not retrieve last inserted ID"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}

	e.ID = id

	return nil
}

// Update modifies the event data in the database using the current event ID.
func (e *Event) Update() error {
	errorUserMessage := "Error updating the event"
	query := `
	UPDATE events
	SET title = ?, description = ?, location = ?, date_time = ?
	WHERE id = ?
	`
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		technicalMessage := "Could not prepare statement to update the event"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}
	defer stmt.Close()

	_, err = stmt.Exec(e.Title, e.Description, e.Location, e.DateTime, e.ID)
	if err != nil {
		technicalMessage := "Could not execute statement to update the event"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}

	return nil
}

// Delete removes the event from the database by its ID.
func (e *Event) Delete() error {
	errorUserMessage := "Error deleting the event"
	query := `
	DELETE FROM events
	WHERE id = ?
	`
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		technicalMessage := "Could not prepare statement to delete the event"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}
	defer stmt.Close()

	_, err = stmt.Exec(e.ID)
	if err != nil {
		technicalMessage := "Could not execute statement to delete the event"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}

	return nil
}

// GetAllEvents returns all events stored in the database.
func GetAllEvents() ([]Event, error) {
	errorUserMessage := "Error fetching events"
	query := `
	SELECT
		id,
		title,
		description,
		location,
		date_time,
		user_id
	FROM events
	`
	rows, err := db.DB.Query(query)
	if err != nil {
		technicalMessage := "Could not prepare statement to fetch events"
		return nil, utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}
	defer rows.Close()

	var events []Event = []Event{}

	for rows.Next() {
		var event Event

		err := rows.Scan(&event.ID, &event.Title, &event.Description, &event.Location, &event.DateTime, &event.UserID)
		if err != nil {
			technicalMessage := "Could not scan event row"
			return nil, utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		technicalMessage := "Could not iterate over events"
		return nil, utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}

	return events, nil
}

// GetEventByID returns a single event matching the provided ID.
func GetEventByID(id int64) (*Event, error) {
	errorUserMessage := "Error fetching the event"
	query := `
	SELECT
		id,
		title,
		description,
		location,
		date_time,
		user_id
	FROM events
	WHERE id = ?
	`
	row := db.DB.QueryRow(query, id)

	var event Event

	err := row.Scan(&event.ID, &event.Title, &event.Description, &event.Location, &event.DateTime, &event.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		technicalMessage := "Could not scan event row"
		return nil, utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}

	return &event, nil
}

// Register adds a user registration for the current event.
func (e *Event) Register(userID int64) error {
	errorUserMessage := "Error registering for the event"
	query := `
	INSERT INTO registrations (event_id, user_id)
	VALUES (?, ?)
	`
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		technicalMessage := "Could not prepare statement to register for the event"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}
	defer stmt.Close()

	_, err = stmt.Exec(e.ID, userID)
	if err != nil {
		technicalMessage := "Could not execute statement to register for the event"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}

	return nil
}

// Unregister removes a user registration for the current event.
func (e *Event) Unregister(userID int64) error {
	errorUserMessage := "Error unregistering from the event"
	query := `
	DELETE FROM registrations
	WHERE event_id = ? AND user_id = ?
	`
	stmt, err := db.DB.Prepare(query)
	if err != nil {
		technicalMessage := "Could not prepare statement to unregister from the event"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}
	defer stmt.Close()

	_, err = stmt.Exec(e.ID, userID)
	if err != nil {
		technicalMessage := "Could not execute statement to unregister from the event"
		return utils.LogAndReturnError(err, technicalMessage, errorUserMessage)
	}

	return nil
}
