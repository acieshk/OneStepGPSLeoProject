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

	"OneStepGPSLeo/database" // Correct import path
	"OneStepGPSLeo/models"   // Correct import path

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type IconHandlers struct {
	DB     *database.MongoDB
	Config models.Config
}

func NewIconHandlers(cfg models.Config, db *database.MongoDB) *IconHandlers {
	return &IconHandlers{Config: cfg, DB: db}
}

func (h *IconHandlers) HandleIconUpload(c *gin.Context) {
	log.Printf("Starting icon upload/remove for device ID: %s", c.Param("id"))

	deviceIDStr := c.Param("id")
	deviceID, err := primitive.ObjectIDFromHex(deviceIDStr)
	if err != nil {
		log.Printf("Invalid device ID: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid device ID"})
		return
	}

	// Check if the device exists
	collection := h.DB.Client.Database(h.DB.DatabaseName).Collection(h.DB.DeviceCollectionName) // Access from DB instance
	filter := bson.M{"_id": deviceID}

	if err := collection.FindOne(context.TODO(), filter).Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Device not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"}) // Don't leak internal details
		}
		return
	}

	// Check for the "remove" query parameter
	removeIcon := c.Query("remove") == "true"
	filename := fmt.Sprintf("%s.png", deviceID.Hex()) // Define filename here
	filepath := filepath.Join("icons", filename)      // Define filepath here

	if removeIcon {
		log.Println("Removing icon...")

		// Remove icon file (ignore errors if the file doesn't exist)
		_ = os.Remove(filepath)

		// Update database to remove iconURL (clear it)
		if err := h.UpdateDeviceIconURL(deviceID, ""); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove device icon URL"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Icon removed successfully"})
		return
	} else {
		// Handle upload only if not removing

		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}
		defer file.Close()

		if err := h.ValidateImageFile(file, header); err != nil { // Call h.ValidateImageFile
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if _, err := file.Seek(0, io.SeekStart); err != nil { // Rewind!
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset file reader"})
			return
		}

		if err := h.saveIconFile(file, filename); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save icon file"})
			return
		}

		iconURL := fmt.Sprintf("http://localhost:%s/icons/%s", h.Config.ServerPort, filename)

		if err := h.UpdateDeviceIconURL(deviceID, iconURL); err != nil {
			os.Remove(filepath) // Cleanup on error
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update icon URL"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"iconUrl": iconURL})
	}

}

// validateImageFile, saveIconFile, getIconHandler, etc. - move here
func (h *IconHandlers) ValidateImageFile(file multipart.File, header *multipart.FileHeader) error {
	// Decode to get image dimensions and validate it's an image
	_, _, err := image.DecodeConfig(file)
	if err != nil {
		return fmt.Errorf("invalid image file: %v", err)
	}

	// ... other validation if needed (e.g., file size)

	return nil
}

func (h *IconHandlers) GetIconHandler(c *gin.Context) { // Add receiver, Rename to conventional case

	deviceId := c.Param("id") // Use Gin's param method

	iconPath := filepath.Join("icons", deviceId+".png")

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
