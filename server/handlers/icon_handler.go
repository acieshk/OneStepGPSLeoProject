package handlers

import (
	"context"
	"fmt"
	"image"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"OneStepGPSLeo/database" // Correct import path
	"OneStepGPSLeo/models"   // Correct import path

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IconHandlers struct {
	DB     *database.MongoDB
	Config models.Config
}

func NewIconHandlers(cfg models.Config, db *database.MongoDB) *IconHandlers {
	return &IconHandlers{Config: cfg, DB: db}
}

// HandleIconUpload handles both uploading and removing device icons.
func (h *IconHandlers) HandleIconUpload(c *gin.Context) {
	const iconDirectory = "./icons"

	deviceIDStr := c.Param("id")
	if deviceIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Device ID is required"})
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := h.validateDevice(ctx, deviceIDStr); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		return
	}

	// Initialize device settings
	deviceSettings := models.DeviceSettings{
		DeviceID: deviceIDStr,
	}

	// Handle icon removal if requested
	if c.Query("remove") == "true" {
		if err := h.handleIconRemoval(deviceSettings, iconDirectory); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove icon"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Icon removed successfully"})
		return
	}

	// Handle default icon if provided
	if defaultIcon := c.PostForm("defaultIcon"); defaultIcon != "" {
		if err := h.handleDefaultIcon(deviceSettings, defaultIcon); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set default icon"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"iconUrl": defaultIcon, "message": "Default icon set successfully"})
		return
	}

	// Handle file upload
	updatedSettings, err := h.handleFileUpload(c, deviceSettings, iconDirectory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process icon upload"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"iconUrl": updatedSettings.IconURL,
		"message": "Icon uploaded successfully",
		"version": updatedSettings.Version,
	})
}

func (h *IconHandlers) validateDevice(ctx context.Context, deviceID string) error {
	collection := h.DB.Client.Database(h.DB.DatabaseName).Collection(h.DB.DeviceCollectionName)
	return collection.FindOne(ctx, bson.M{"device_id": deviceID}).Err()
}

func (h *IconHandlers) handleIconRemoval(settings models.DeviceSettings, iconDir string) error {
	filename := fmt.Sprintf("%s.png", settings.DeviceID)
	filepath := filepath.Join(iconDir, filename)

	log.Printf("Attempting to remove icon at filepath: %s", filepath)

	if err := os.Remove(filepath); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error removing icon file: %v", err)
			return fmt.Errorf("failed to remove icon file: %w", err)
		}

	}

	settings.IconURL = ""
	_, err := h.DB.SaveDeviceSettings(settings) // Use SaveDeviceSettings to update the iconURL in device settings
	if err != nil {
		log.Printf("Failed to update DeviceSettings after icon removal: %v", err) // Log error. Wrap error for more informative message if needed.

		return fmt.Errorf("failed to remove iconURL from database after icon removal: %w", err)
	}
	log.Printf("Successfully removed icon for device: %s", settings.DeviceID) // Log message after successful removal

	return nil //Return nil if no error

}

func (h *IconHandlers) handleDefaultIcon(settings models.DeviceSettings, defaultIcon string) error {
	settings.IconURL = defaultIcon
	_, err := h.DB.SaveDeviceSettings(settings)
	return err
}

func (h *IconHandlers) handleFileUpload(c *gin.Context, settings models.DeviceSettings, iconDir string) (models.DeviceSettings, error) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		return settings, fmt.Errorf("failed to get uploaded file: %v", err)
	}
	defer file.Close()

	// Ensure directory exists with proper permissions
	if err := os.MkdirAll(iconDir, 0755); err != nil {
		return settings, fmt.Errorf("failed to create directory: %v", err)
	}

	filename := fmt.Sprintf("%s.png", settings.DeviceID)
	filepath := filepath.Join(iconDir, filename)
	tempFile := filepath + ".tmp"

	// Create temporary file with less restrictive flags
	out, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return settings, fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer func() {
		out.Close()
		if err != nil {
			os.Remove(tempFile) // Clean up temp file on error
		}
	}()

	// Copy file contents
	if _, err = io.Copy(out, file); err != nil {
		return settings, fmt.Errorf("failed to write file: %v", err)
	}

	// Ensure all data is written to disk
	if err = out.Sync(); err != nil {
		return settings, fmt.Errorf("failed to sync file: %v", err)
	}

	// Close the file before rename
	out.Close()

	// Atomically rename temporary file to final filename
	if err = os.Rename(tempFile, filepath); err != nil {
		return settings, fmt.Errorf("failed to rename file: %v", err)
	}

	settings.IconURL = fmt.Sprintf("http://%s/icons/%s", c.Request.Host, filename)
	return h.DB.SaveDeviceSettings(settings)
}

// validateImageFile, saveIconFile, getIconHandler, etc. - move here
func (h *IconHandlers) ValidateImageFile(file multipart.File, header *multipart.FileHeader) error {
	// Decode to get image dimensions and validate it's an image
	contentType := header.Header.Get("Content-Type")
	if contentType != "image/png" && contentType != "image/jpeg" && contentType != "image/jpg" && contentType != "image/gif" {
		return fmt.Errorf("invalid image content type: %s", contentType)
	}

	_, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("invalid image file: %w", err)

	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("error seeking file: %w", err)
	}

	return nil
}

func (h *IconHandlers) GetIconHandler(c *gin.Context) { // Add receiver, Rename to conventional case

	deviceId := c.Param("id") // Use Gin's param method

	iconPath := filepath.Join("./icons", deviceId+".png")

	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		c.String(http.StatusOK, "") // Return empty string with 200 OK if not found (Gin's way)
		return
	}

	// For serving static files, Gin's File function is often cleaner:
	c.File(iconPath) // Gin automatically handles Content-Type
}

func (h *IconHandlers) saveIconFile(file multipart.File, filename string) error {
	if err := os.MkdirAll("./icons", os.ModePerm); err != nil {
		return fmt.Errorf("failed to create icons directory: %w", err)
	}

	out, err := os.Create(filepath.Join("./icons", filename))
	if err != nil {
		return fmt.Errorf("failed to create icon file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return fmt.Errorf("failed to copy icon file: %w", err)
	}

	return nil
}

func (h *IconHandlers) UpdateDeviceIconURL(deviceID primitive.ObjectID, iconURL string) error {
	filter := bson.M{"_id": deviceID}
	update := bson.M{"$set": bson.M{"iconUrl": iconURL}} // Make sure field name matches your database

	// Use the database client from h.DB
	collection := h.DB.Client.Database(h.DB.DatabaseName).Collection(h.DB.DeviceCollectionName)
	_, err := collection.UpdateOne(context.TODO(), filter, update)

	return err // Directly return the error from UpdateOne for simpler error handling

}
