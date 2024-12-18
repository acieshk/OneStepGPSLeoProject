/*
Package api provides functions for fetching, storing, and retrieving device data.

This package handles interactions with the external API and the database,
including fetching device data, updating the database, and checking for updates.
*/
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"OneStepGPSLeo/database"
	"OneStepGPSLeo/models"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CheckForUpdatesResponse struct for returning response to checkForUpdates
type CheckForUpdatesResponse struct {
	NeedsUpdate    bool                     `json:"needsUpdate"`
	LastUpdate     string                   `json:"lastUpdate"`
	UpdatedDevices []map[string]interface{} `json:"updatedDevices,omitempty"` // Include updated devices
	IconMap        map[string]string        `json:"icon_map"`
}

// FetchAndStoreDevices fetches device data from the external API and updates the database.
// It handles both inserting new devices and updating existing ones, including their settings.
func FetchAndStoreDevices(db *database.MongoDB, config models.Config, updateMutex *sync.RWMutex, lastUpdateTimes map[string]time.Time, lastChecked *time.Time) {
	// API fetching and response handling
	log.Printf("Fetching device data from the external API")
	apiURL := fmt.Sprintf("%s%s", config.APIURL, config.APIKey)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("Error fetching from API: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}

	var response struct {
		ResultList []map[string]interface{} `json:"result_list"`
	}
	body = bytes.TrimSpace(body)
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return
	}

	collection := db.Client.Database(config.DatabaseName).Collection(config.DeviceCollectionName)

	currentDevices, err := getCurrentDevicesMap(collection)
	if err != nil {
		log.Printf("Failed to get current devices: %v\n", err)
		return
	}

	for _, device := range response.ResultList {
		deviceID, ok := device["device_id"].(string)
		if !ok {
			log.Printf("Error: device_id not found or not a string in device: %+v", device)
			continue
		}

		updatedAtStr, ok := device["updated_at"].(string)
		if !ok {
			log.Printf("Error: updated_at not found or not a string for device %s: %+v", deviceID, device)
			continue
		}

		updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			log.Printf("Error parsing updated_at '%s' for device %s: %v", updatedAtStr, deviceID, err)
			continue
		}

		settingsMap, settingsOK := device["settings"].(map[string]interface{})
		delete(device, "settings")

		var settings models.DeviceSettings
		if settingsOK {
			bData, _ := bson.Marshal(settingsMap)
			bson.Unmarshal(bData, &settings)
			settings.DeviceID = deviceID
			settings.UpdatedAt = updatedAt.Format(time.RFC3339)
			if v, versionOK := settingsMap["version"].(int); versionOK {
				settings.Version = v
			}
		}

		_, deviceExists := currentDevices[deviceID]

		if !deviceExists {
			// Insert new device
			_, err := collection.InsertOne(context.TODO(), device)
			if err != nil {
				log.Printf("Error inserting new device data %s: %v\n", deviceID, err)
				continue
			}

			// For new devices, always insert the settings
			if settingsOK {
				_, err := db.SaveDeviceSettings(settings)
				if err != nil {
					log.Printf("Failed to insert new device settings for device %s: %v\n", deviceID, err)
					continue
				}
			}
			log.Printf("Inserted new device: %s, updated_at: %s\n", deviceID, updatedAt)
		} else {
			updateMutex.RLock()
			lastUpdatedAt, _ := lastUpdateTimes[deviceID]
			updateMutex.RUnlock()

			if updatedAt.After(lastUpdatedAt) {
				// Update device data
				_, err := collection.ReplaceOne(context.TODO(), bson.M{"device_id": deviceID}, device)
				if err != nil {
					log.Printf("Failed to replace device %s: %v\n", deviceID, err)
					continue
				}

				if settingsOK {
					// Check if settings already exist for this device
					existingSettings, err := db.GetDeviceSettings(deviceID)
					if err != nil || existingSettings == (models.DeviceSettings{}) {
						// Only save settings if they don't exist
						settings, err = db.SaveDeviceSettings(settings)
						if err != nil {
							log.Printf("Failed to update device settings for device %s: %v\n", deviceID, err)
						} else {
							log.Printf("Initialized settings for existing device: %s\n", deviceID)
						}
					} else {
						log.Printf("Preserving existing user settings for device %s\n", deviceID)
					}
				}

				updateMutex.Lock()
				lastUpdateTimes[deviceID] = updatedAt
				updateMutex.Unlock()

				color.Green("Updated device: %s, last update was %s ago, updated_at: %s\n",
					deviceID, time.Since(lastUpdatedAt).Round(time.Second), updatedAt)
			} else {
				log.Printf("Device %s not updated. Current updated_at: %s is before or equal to last updated_at: %s\n",
					deviceID, updatedAt, lastUpdatedAt)
			}
		}
	}

	now := time.Now()
	updateMutex.Lock()
	*lastChecked = now
	updateMutex.Unlock()
}

// CheckForUpdates checks if any device has been updated since the client's last check and returns updated devices.
func CheckForUpdates(c *gin.Context, db *database.MongoDB, config models.Config, lastChecked *time.Time, lastUpdateTimes map[string]time.Time) {
	clientLastUpdateStr := c.Query("lastUpdate")
	if clientLastUpdateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing lastUpdate parameter"})
		return
	}

	clientLastUpdate, err := time.Parse(time.RFC3339, clientLastUpdateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid lastUpdate timestamp format. Use RFC3339."})
		return
	}

	needsUpdate := clientLastUpdate.Before(*lastChecked)

	var updatedDevices []map[string]interface{}

	// Get the icon map - always fetch, regardless of needsUpdate
	iconMap, err := db.GetIconMap()
	if err != nil {
		log.Printf("Failed to get icon map: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get icon map"})
		return // Return early on error
	}

	if needsUpdate {
		updatedDevices, err = fetchUpdatedDevicesSince(db, config, clientLastUpdate, lastUpdateTimes)
		if err != nil {
			log.Printf("Failed to fetch updated devices: %v", err) // Log error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated devices"})
			return // Return early on error
		}

	}

	response := CheckForUpdatesResponse{
		NeedsUpdate:    needsUpdate,
		LastUpdate:     lastChecked.Format(time.RFC3339),
		UpdatedDevices: updatedDevices,
		IconMap:        iconMap, // Include iconMap in response
	}

	c.JSON(http.StatusOK, response)
}

// New helper function to efficiently build icon map based on updated devices.
func buildIconMap(db *database.MongoDB, currentDevices map[string]primitive.ObjectID) (map[string]string, error) {
	iconMap := make(map[string]string) // Initialize an empty iconMap

	for deviceID := range currentDevices { // Iterate through ALL devices

		deviceSettings, err := db.GetDeviceSettings(deviceID)
		if err != nil { //Handle error for fetching device settings, return early if failed and log the error.
			if !errors.Is(err, mongo.ErrNoDocuments) {

				return nil, fmt.Errorf("failed to get device settings: %w", err)
			}

			log.Printf("Settings not found for device: %s. Using default icon\n", deviceID)
			iconMap[deviceID] = "" // Use empty string if no settings found, frontend should handle null/empty string
			continue
		}

		iconMap[deviceID] = deviceSettings.IconURL // Add iconURL to iconMap. No timestamp check here
	}

	return iconMap, nil
}

// New helper function to fetch updated devices efficiently. Fetches only the fields you need.
func fetchUpdatedDevicesSince(db *database.MongoDB, config models.Config, since time.Time, lastUpdateTimes map[string]time.Time) ([]map[string]interface{}, error) {

	
	collection := db.Client.Database(config.DatabaseName).Collection(config.DeviceCollectionName)

	projection := bson.D{
		{"online", 1},
		{"latest_device_point", 1},
		{"latest_accurate_device_point", 1},
		{"updated_at", 1},
		{"device_id", 1},
		{"active_state", 1},
		{"_id", 1},
	}

	//Filter based on timestamp and query
	filter := bson.M{"updated_at": bson.M{"$gt": since.Format(time.RFC3339)}}
	opts := options.Find().SetProjection(projection)
	cursor, err := collection.Find(context.TODO(), filter, opts)

	if err != nil {
		return nil, fmt.Errorf("failed to find updated devices: %w", err) // Wrap and return the error

	}
	defer cursor.Close(context.TODO())

	var updatedDevices []map[string]interface{}

	if err := cursor.All(context.TODO(), &updatedDevices); err != nil {
		log.Printf("Found %d updated devices since %s", len(updatedDevices), since.Format(time.RFC3339))
		return nil, fmt.Errorf("failed to decode updated devices: %w", err) //Wrap the error
	}

	return updatedDevices, nil
}
