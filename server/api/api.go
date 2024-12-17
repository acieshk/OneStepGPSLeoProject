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
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CheckForUpdatesResponse struct for returning response to checkForUpdates
type CheckForUpdatesResponse struct {
	NeedsUpdate    bool                     `json:"needsUpdate"`
	LastUpdate     string                   `json:"lastUpdate"`
	UpdatedDevices []map[string]interface{} `json:"updatedDevices,omitempty"` // Include updated devices
}

// FetchAndStoreDevices fetches device data from the external API and updates the database.
// It handles both inserting new devices and updating existing ones based on their "updated_at" timestamps.
func FetchAndStoreDevices(db *database.MongoDB, config models.Config, updateMutex *sync.RWMutex, lastUpdateTimes map[string]time.Time, lastChecked *time.Time) {
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

	// Get current devices from the database for efficient checking
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

		var lastUpdatedAt time.Time
		updateMutex.RLock()
		lastUpdatedAt, ok = lastUpdateTimes[deviceID]
		updateMutex.RUnlock()

		if _, exists := currentDevices[deviceID]; !exists {
			// Device doesn't exist, insert it
			_, err := collection.InsertOne(context.TODO(), device)
			if err != nil {
				log.Printf("Error inserting device %s: %v\n", deviceID, err)
				continue
			}
			log.Printf("Inserted new device: %s, updated_at: %s\n", deviceID, updatedAt)

		} else {
			if !ok {
				log.Printf("Warning: No last updated time found for existing device %s. Updating anyway...", deviceID)
			} else if updatedAt.After(lastUpdatedAt) {
				update := bson.M{
					"latest_device_point":          device["latest_device_point"],
					"active_state":                 device["active_state"],
					"updated_at":                   device["updated_at"],
					"latest_accurate_device_point": device["latest_accurate_device_point"],
					"online":                       device["online"],
				}

				_, err := collection.UpdateOne(context.TODO(), bson.M{"device_id": deviceID}, bson.M{"$set": update})
				if err != nil {
					log.Printf("Failed to update device %s: %v\n", deviceID, err)
					continue
				}

				timeDiff := time.Since(lastUpdatedAt)
				color.Green("Updated device: %s, last update was %s ago, updated_at: %s\n", deviceID, timeDiff.Round(time.Second), updatedAt)

			}
			// else {
			// 	log.Printf("Device %s not updated. \n Current updated_at: %s is not after last updated_at: %s\n", deviceID, updatedAt, lastUpdatedAt)
			// }
		}

		updateMutex.Lock()
		lastUpdateTimes[deviceID] = updatedAt
		updateMutex.Unlock()
	}

	now := time.Now()
	updateMutex.Lock()
	*lastChecked = now
	updateMutex.Unlock()
}

// CheckForUpdates checks if any device has been updated since the client's last check and returns updated devices.
func CheckForUpdates(c *gin.Context, db *database.MongoDB, config models.Config, lastChecked *time.Time, lastUpdateTimes map[string]time.Time) { // Add db, config, lastUpdateTimes as parameter

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

	//Fetch devices only if needs update and there is no error.
	if needsUpdate && err == nil {

		updatedDevices, err = fetchUpdatedDevicesSince(db, config, clientLastUpdate, lastUpdateTimes)
		if needsUpdate {
			if len(updatedDevices) == 0 && err == nil {
				log.Println("needsUpdate is true, but no updated devices found. Check database/timestamps.")

			}

		}

		if err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch updated devices %v", err)}) //Return error message if fetching fails
			return

		}

	}

	response := CheckForUpdatesResponse{
		NeedsUpdate:    needsUpdate,
		LastUpdate:     lastChecked.Format(time.RFC3339),
		UpdatedDevices: updatedDevices,
	}

	c.JSON(http.StatusOK, response)
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
