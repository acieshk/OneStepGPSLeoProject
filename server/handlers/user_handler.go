/*
Package handlers provides HTTP request handlers for user-related operations.

This file contains handlers for retrieving, saving, and updating user preferences.
It interacts with the database to manage user specific settings.
*/
package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"OneStepGPSLeo/database"
	"OneStepGPSLeo/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// UserHandlers struct to manage dependencies for user-related operations. Contains a database client and configuration for the handlers.
type UserHandlers struct {
	DB     *database.MongoDB
	Config models.Config
}

// NewUserHandlers creates a new instance of UserHandlers with the provided dependencies.
func NewUserHandlers(cfg models.Config, db *database.MongoDB) *UserHandlers {
	return &UserHandlers{Config: cfg, DB: db}
}

// GetUserPreferencesHandler retrieves user preferences from the database.
// It handles cases where preferences are not found by returning default values.
func (h *UserHandlers) GetUserPreferencesHandler(c *gin.Context) {
	userID := c.Param("userId")

	prefs, err := h.DB.GetUserPreferences(userID)
	if err != nil {
		// Check if it's a "not found" error to return the default
		if errors.Is(err, mongo.ErrNoDocuments) {
			fmt.Println(err)
			prefs = models.UserPreferences{
				UserID:          userID,
				DeviceListWidth: 400,
				Unit:            "imperial",
				Version:         1,
			}
			savedPrefs, saveErr := h.DB.SaveUserPreferences(prefs)
			if saveErr != nil { // Check saveErr
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create and save default user preferences"})
				return
			}
			c.JSON(http.StatusCreated, savedPrefs)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get user preferences: %v", err)}) //Return more specific error for easier debuging
		return

	}

	c.JSON(http.StatusOK, prefs)
}

// SaveUserPreferencesHandler saves or updates user preferences in the database.

func (h *UserHandlers) SaveUserPreferencesHandler(c *gin.Context) {
	userId := c.Param("userId") // Get userId from URL parameter

	log.Printf("SaveUserPreferencesHandler called for userId: %s", userId)

	var prefs models.UserPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil { // existing code ...
		return

	}
	log.Printf("Attempting to save preferences: %+v", prefs)
	// Set UserID from the URL parameter.
	prefs.UserID = userId // Ensure UserID is set correctly

	// Attempt to save preferences.  Use returned updatedPrefs to update frontend store if needed.

	updatedPrefs, err := h.DB.SaveUserPreferences(prefs)
	if err != nil {
		if errors.Is(err, database.ErrPreferencesNotFound) { // Return 404 Not Found if preferences don't exist yet
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		} else if errors.Is(err, database.ErrOutdatedVersion) { // Return 409 Conflict for version mismatch
			c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "currentPrefs": updatedPrefs}) // existing code, include returned currentPrefs in response
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save user preferences: %v", err)}) //Return descriptive error message
		return
	}

	c.JSON(http.StatusOK, updatedPrefs) // Return 200 OK and the updated preferences
}
