package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"OneStepGPSLeo/api"
	"OneStepGPSLeo/database"
	"OneStepGPSLeo/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DeviceHandlers struct {
	DB              *database.MongoDB
	Config          models.Config
	UpdateMutex     sync.RWMutex         // Add mutex field
	LastUpdateTimes map[string]time.Time // Add map field
	LastChecked     time.Time            // Add lastChecked field
}

func NewDeviceHandlers(cfg models.Config, db *database.MongoDB) *DeviceHandlers {
	return &DeviceHandlers{
		Config:          cfg,
		DB:              db,
		UpdateMutex:     sync.RWMutex{},
		LastUpdateTimes: make(map[string]time.Time),
		LastChecked:     time.Now(),
	}
}

func (h *DeviceHandlers) GetDevices(c *gin.Context) {
	devices, err := h.DB.GetDevices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"result_list": devices})
}

func (h *DeviceHandlers) UpdateDeviceHandler(c *gin.Context) {
	deviceIDStr := c.Param("id")
	deviceID, err := primitive.ObjectIDFromHex(deviceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	var updatedDevice map[string]interface{}
	if err := c.ShouldBindJSON(&updatedDevice); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Remove _id to prevent accidental replacement
	delete(updatedDevice, "_id")

	versionStr := c.Query("version")
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid version"}) //Handle invalid version parameter
		return
	}

	if err := h.DB.UpdateDevice(deviceID, updatedDevice, version); err != nil { // Pass version to UpdateDevice
		if err.Error() == "device not found" || err.Error() == "Outdated device version" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()}) // Return 409 Conflict and specific message
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating device"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Device updated successfully"})
}

func (h *DeviceHandlers) CheckForUpdates(c *gin.Context) {

	deviceID := c.Query("deviceId") //Get deviceID
	clientLastUpdateStr := c.Query("lastUpdate")

	if deviceID == "" || clientLastUpdateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing parameters"})
		return
	}

	clientLastUpdate, err := time.Parse(time.RFC3339, clientLastUpdateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timestamp format"})
		return
	}

	h.UpdateMutex.RLock()
	serverLastUpdate, deviceExists := h.LastUpdateTimes[deviceID]
	h.UpdateMutex.RUnlock()

	if !deviceExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	needsUpdate := clientLastUpdate.Before(serverLastUpdate)

	var responseData gin.H

	if needsUpdate {
		// Efficiently fetch the updated device data, including _id
		updatedDevice, err := h.fetchUpdatedDevice(deviceID)
		if err != nil {
			if err == mongo.ErrNoDocuments { // Handle device not found during fetch

				c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
			} else {

				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Error fetching Device %v", err)}) //Descriptive error message
			}

			return
		}

		responseData = gin.H{
			"needsUpdate": true,
			"lastUpdate":  serverLastUpdate.Format(time.RFC3339),
			"data":        updatedDevice, // Return the updated device data
		}
	} else {
		responseData = gin.H{
			"needsUpdate": false,
			"lastUpdate":  serverLastUpdate.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, responseData)
}

// Helper function to fetch a single updated device efficiently
func (h *DeviceHandlers) fetchUpdatedDevice(deviceID string) (map[string]interface{}, error) {

	objID, err := primitive.ObjectIDFromHex(deviceID) // Convert to ObjectID

	if err != nil {
		return nil, fmt.Errorf("invalid device ID: %w", err) //Wrap the error
	}

	collection := h.DB.Client.Database(h.Config.DatabaseName).Collection(h.Config.DeviceCollectionName)

	// Project only necessary fields as you did in earlier fetch functions. Add more fields if needed.
	projection := bson.D{
		{"online", 1},
		{"latest_device_point", 1},
		{"latest_accurate_device_point", 1},
		{"updated_at", 1},
		{"device_id", 1},
		{"active_state", 1},
		{"_id", 1},
	}

	var updatedDevice map[string]interface{} //Correctly use updatedDevice here

	err = collection.FindOne(context.TODO(), bson.M{"_id": objID}, options.FindOne().SetProjection(projection)).Decode(&updatedDevice) //Use correct filter, options

	if err != nil {
		return nil, err //Return the error from database query, for specific error handling
	}

	return updatedDevice, nil
}

// RefreshDatabaseHandler clears the device and user preferences collections and then re-fetches device data from the external API.
func (h *DeviceHandlers) RefreshDatabaseHandler(c *gin.Context) {
	if err := h.DB.ClearCollections(); err != nil { // Clear both collections
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear database collections"})
		return
	}

	//Refetch devices from API.
	api.FetchAndStoreDevices(h.DB, h.Config, &h.UpdateMutex, h.LastUpdateTimes, &h.LastChecked)

	c.JSON(http.StatusOK, gin.H{"message": "Database refreshed successfully"})
}

// Add new handler functions for settings.
func (h *DeviceHandlers) GetDeviceSettingsHandler(c *gin.Context) {
	deviceID := c.Param("id")

	settings, err := h.DB.GetDeviceSettings(deviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (h *DeviceHandlers) SaveDeviceSettingsHandler(c *gin.Context) {
	var settings models.DeviceSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedSettings, err := h.DB.SaveDeviceSettings(settings) // Updated to match changes
	if err != nil {
		if err.Error() == "Outdated device settings version" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "currentPrefs": updatedSettings})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return

	}

	c.JSON(http.StatusOK, updatedSettings) //Return updated settings
}
