package models

// Config represents the configuration structure for the application
type Config struct {
	ServerPort           string `json:"server_port"`
	MongoDBURL           string `json:"mongodb_url"`
	MongoDBPort          string `json:"mongodb_port"`
	MongoDBUsername      string `json:"mongodb_username"`
	MongoDBPassword      string `json:"mongodb_password"`
	DatabaseName         string `json:"database_name"`
	DeviceCollectionName string `json:"device_collection_name"`
	UserCollectionName   string `json:"user_collection_name"`
	APIKey               string `json:"api_key"`
	APIURL               string `json:"api_url"`
	UpdateInterval       int    `json:"update_interval_seconds"`
	MockServerPort       string `json:"mock_server_port"`
}

type UserPreferences struct {
	Version         int    `bson:"version" json:"version"`
	UserID          string `bson:"user_id" json:"userId"`
	DeviceListWidth int    `bson:"device_list_width" json:"DeviceListWidth"`
	Unit            string `bson:"unit" json:"unit"`
}
