// mockserver/server.go
package mockserver

import (
	"OneStepGPSLeo/models"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
)

type MockAPIResponse struct {
	ResultList []map[string]interface{} `json:"result_list"` // Use map directly
}

// Datastore to store and retrieve mock devices
type Datastore struct {
	Devices []map[string]interface{}
	Mutex   sync.RWMutex //Use read write mutex
}

// NewDatastore creates and initializes new datastore
func NewDatastore() *Datastore {
	return &Datastore{
		Devices: make([]map[string]interface{}, 0),
		Mutex:   sync.RWMutex{},
	}
}

// AddDevice adds a device to the Datastore.
func (ds *Datastore) AddDevice(device map[string]interface{}) {
	ds.Mutex.Lock()
	defer ds.Mutex.Unlock()
	ds.Devices = append(ds.Devices, device)
}

// UpdateDeviceAtIndex updates a device at the given index.
func (ds *Datastore) UpdateDeviceAtIndex(index int, device map[string]interface{}) {
	ds.Mutex.Lock()
	defer ds.Mutex.Unlock()
	if index >= 0 && index < len(ds.Devices) {
		ds.Devices[index] = device
	}
}

// GetDevices returns a copy of the devices in the Datastore.
func (ds *Datastore) GetDevices() []map[string]interface{} {
	ds.Mutex.RLock()
	defer ds.Mutex.RUnlock()
	devicesCopy := make([]map[string]interface{}, len(ds.Devices))
	for i, device := range ds.Devices {
		// Create a new map for each device
		devicesCopy[i] = make(map[string]interface{})
		for k, v := range device {
			devicesCopy[i][k] = v
		}
	}
	return devicesCopy
}

// StartMockServer starts the mock server.
func StartMockServer(config models.Config, port string, updateInterval time.Duration, mutateChance float64, mutateDeviceCount int) {
	datastore := NewDatastore()

	err := initializeMockDevicesFromAPI(datastore, config)
	if err != nil {
		log.Fatalf("Failed to initialize mock devices from API: %v", err)
	}

	go updateMockDevices(datastore, updateInterval, mutateChance, mutateDeviceCount)

	router := gin.Default() // Create a Gin router

	router.GET("/api/v1/devices", func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		devices := datastore.GetDevices()
		resp := MockAPIResponse{
			ResultList: devices,
		}
		c.JSON(http.StatusOK, resp)
	})

	log.Printf("Mock server started on :%s\n", port)

	if err := router.Run(":" + port); err != nil { // Correct error handling for router.Run
		log.Fatalf("Failed to start mock server: %v", err)
	}
}

// initializeMockDevicesFromAPI initializes the mock devices from the API.
func initializeMockDevicesFromAPI(datastore *Datastore, config models.Config) error {
	apiURL := fmt.Sprintf("%s%s", config.APIURL, config.APIKey)
	resp, err := http.Get(apiURL)

	if err != nil {
		return fmt.Errorf("error fetching from API: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	var apiResponse struct {
		ResultList []map[string]interface{} `json:"result_list"`
	}

	body = bytes.TrimSpace(body)
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return fmt.Errorf("error parsing JSON: %w", err)
	}

	for _, deviceMap := range apiResponse.ResultList {
		datastore.AddDevice(deviceMap)
	}

	if len(datastore.Devices) == 0 {
		return fmt.Errorf("no devices found in API response")
	}

	log.Printf("Initialized %d mock devices from API.\n", len(datastore.GetDevices()))

	return nil

}

// updateMockDevices (Updated)
func updateMockDevices(datastore *Datastore, updateInterval time.Duration, mutateChance float64, mutateDeviceCount int) {
	for {
		time.Sleep(updateInterval)

		devices := datastore.GetDevices() //Get copy
		randomNumber := rand.Float64()

		//fmt.Printf("Random number: %f\n", randomNumber)
		//fmt.Printf("Mutating %.2f chance\n", mutateChance) // Print with 2 decimal places

		// Apply mutations with a certain chance
		if randomNumber < mutateChance {
			mutateDevices(&devices, mutateDeviceCount) // Pass devices as a pointer
			fmt.Println("Mock devices mutated")

			//Update the device in datastore
			datastore.Mutex.Lock()

			datastore.Devices = devices //Update the devices

			datastore.Mutex.Unlock()

		}

	}
}

func mutateDevices(devices *[]map[string]interface{}, mutateCount int) {
	if len(*devices) == 0 {
		return // Nothing to mutate
	}

	count := min(mutateCount, len(*devices))

	for i := 0; i < count; i++ {
		index := rand.Intn(len(*devices))
		device := (*devices)[index]

		deviceID, ok := device["device_id"].(string) // Get device_id
		if !ok {
			log.Println("device_id not found or not a string, skipping mutation for this device")
			continue // Skip to next device if device_id isn't there
		}

		// Mutate online status
		if _, ok := device["online"]; ok {
			device["online"] = rand.Intn(2) == 0

		}

		if _, ok := device["latest_device_point"]; ok {

			latestDevicePoint, ok := device["latest_device_point"].(map[string]interface{})
			if !ok {
				log.Printf("latest_device_point is not a map. Skipping device %v", device)

				continue

			}

			if _, ok := latestDevicePoint["lat"]; ok {

				if lat, ok := latestDevicePoint["lat"].(float64); ok {
					latestDevicePoint["lat"] = mutateLat(lat)

				} else {
					log.Printf("Latitude is not float64 for device %v", device)

				}

			}

			if _, ok := latestDevicePoint["lng"]; ok {

				if lng, ok := latestDevicePoint["lng"].(float64); ok {

					latestDevicePoint["lng"] = mutateLng(lng)

				} else {
					log.Printf("Longitude is not float64 for device %v", device)

				}

			}

			// Correctly mutate nested speed value and display:
			if devicePointDetail, ok := latestDevicePoint["device_point_detail"].(map[string]interface{}); ok {
				if speed, ok := devicePointDetail["speed"].(map[string]interface{}); ok {
					speed["value"] = rand.Intn(51)
					speed["display"] = fmt.Sprintf("%d km/h", speed["value"])
				} else {

					log.Printf("Speed not found or not a map %v", device)

				}

			}

		}

		if _, ok := device["updated_at"]; ok {
			device["updated_at"] = time.Now().Format(time.RFC3339) // Update timestamp
		}
		color.Green("Mutated device: %s\n", deviceID)
		(*devices)[index] = device

	}
}

// Helper functions to mutate latitude, longitude. Add or subtract a random number between 0.01 degree to 0.05 degree
func mutateLat(lat float64) float64 {
	change := (rand.Float64() * 0.04) + 0.01 // Random change between 0.01 and 0.05
	if rand.Intn(2) == 0 {
		lat += change
	} else {
		lat -= change
	}
	return lat
}

func mutateLng(lng float64) float64 {

	change := (rand.Float64() * 0.04) + 0.01 // Random change between 0.01 and 0.05
	if rand.Intn(2) == 0 {
		lng += change
	} else {
		lng -= change
	}
	return lng

}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
