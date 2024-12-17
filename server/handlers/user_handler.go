/*
Package handlers provides HTTP request handlers for user-related operations.

This file contains handlers for retrieving, saving, and updating user preferences.
It interacts with the database to manage user specific settings.
*/
package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"OneStepGPSLeo/database"
	"OneStepGPSLeo/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
			if saveErr := h.DB.SaveUserPreferences(prefs); saveErr != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create and save default user preferences"})
				return
			}
			c.JSON(http.StatusCreated, prefs)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get user preferences: %v", err)}) //Return more specific error for easier debuging
		return

	}

	c.JSON(http.StatusOK, prefs)
}

// SaveUserPreferencesHandler saves or updates user preferences in the database.

func (h *UserHandlers) SaveUserPreferencesHandler(c *gin.Context) {
	var prefs models.UserPreferences
	if err := c.ShouldBindJSON(&prefs); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	filter := bson.M{"user_id": prefs.UserID} // Filter only by UserID for upsert

	// Initialize version if it's not already set (for new documents)
	// Because Go initializes value as 0
	if prefs.Version == 0 {
		// Initial save (insert) - set version to 1
		prefs.Version = 1
		update := bson.M{"$setOnInsert": prefs} // Use $setOnInsert

		opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After) // Use FindOneAndUpdate
		result := h.DB.Client.Database(h.DB.DatabaseName).Collection(h.DB.UserCollectionName).FindOneAndUpdate(context.TODO(), filter, update, opts)

		if result.Err() != nil { //Check error from FindOneAndUpdate
			if result.Err() == mongo.ErrNoDocuments { // No update/insert occurred. Should not happen for upsert
				c.JSON(http.StatusNotFound, gin.H{"error": "No matching document found for update"}) //Should not occur
			} else {

				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save preferences"})
			}
			return
		}
		//If no error occurs during find and update, then decode the new document
		var updatedPrefs models.UserPreferences
		if err := result.Decode(&updatedPrefs); err != nil { // Decode the updated document

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode updated preferences"}) //Return the decoded error
			return
		}
		c.JSON(http.StatusOK, updatedPrefs)
	} else {
		// Update existing document - increment version
		update := bson.M{"$set": prefs, "$inc": bson.M{"version": 1}}
		opts := options.FindOneAndUpdate().SetReturnDocument(options.After) // Use FindOneAndUpdate for update and retrieve
		result := h.DB.Client.Database(h.DB.DatabaseName).Collection(h.DB.UserCollectionName).FindOneAndUpdate(context.TODO(), filter, update, opts)

		if result.Err() != nil { //Check error from FindOneAndUpdate
			if result.Err() == mongo.ErrNoDocuments {
				// No update/insert occurred
				c.JSON(http.StatusNotFound, gin.H{"error": "No matching document found for update"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save preferences"})
			}
			return
		}
		var updatedPrefs models.UserPreferences
		if err := result.Decode(&updatedPrefs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode updated preferences"})
			return
		}
		c.JSON(http.StatusOK, updatedPrefs)
	}
}
