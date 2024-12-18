package models

// Config represents the configuration structure for the application
type Config struct {
	ServerPort             string `json:"server_port"`
	MongoDBURL             string `json:"mongodb_url"`
	MongoDBPort            string `json:"mongodb_port"`
	MongoDBUsername        string `json:"mongodb_username"`
	MongoDBPassword        string `json:"mongodb_password"`
	DatabaseName           string `json:"database_name"`
	DeviceCollectionName   string `json:"device_collection_name"`
	UserCollectionName     string `json:"user_collection_name"`
	SettingsCollectionName string `json:"device_setting_collection_name"`
	APIKey                 string `json:"api_key"`
	APIURL                 string `json:"api_url"`
	UpdateInterval         int    `json:"update_interval_seconds"`
	MockServerPort         string `json:"mock_server_port"`
}

type UserPreferences struct {
	Version         int    `bson:"version" json:"version"`
	UserID          string `bson:"user_id" json:"userId"`
	DeviceListWidth int    `bson:"device_list_width" json:"DeviceListWidth"`
	Unit            string `bson:"unit" json:"unit"`
}

type DeviceSettings struct {
	DeviceID                           string          `bson:"device_id" json:"device_id"`
	IconURL                            string          `bson:"iconUrl" json:"iconUrl,omitempty"`
	Version                            int             `bson:"version" json:"version"`
	UpdatedAt                          string          `bson:"updated_at" json:"updated_at,omitempty"`
	BeginMovingSpeed                   Speed           `bson:"begin_moving_speed" json:"begin_moving_speed"`
	BeginStoppedSpeed                  Speed           `bson:"begin_stopped_speed" json:"begin_stopped_speed"`
	MaxDriftDistance                   Speed           `bson:"max_drift_distance" json:"max_drift_distance"`
	MinNumSatellites                   int             `bson:"min_num_satellites" json:"min_num_satellites"`
	IgnoreUnsetMinNumSats              bool            `bson:"ignore_unset_min_num_sats" json:"ignore_unset_min_num_sats"`
	MaxHdop                            float64         `bson:"max_hdop" json:"max_hdop"`
	DriveTimeout                       Speed           `bson:"drive_timeout" json:"drive_timeout"`
	StopTimeout                        Speed           `bson:"stop_timeout" json:"stop_timeout"`
	OfflineTimeout                     Speed           `bson:"offline_timeout" json:"offline_timeout"`
	HistoryCalcDuration                Speed           `bson:"history_calc_duration" json:"history_calc_duration"`
	FuelConsumption                    FuelConsumption `bson:"fuel_consumption" json:"fuel_consumption"`
	InitialDevicePointDeleteCutoffTime string          `bson:"initial_device_point_delete_cutoff_time" json:"initial_device_point_delete_cutoff_time"`
	EngineHoursCounterConfig           string          `bson:"engine_hours_counter_config" json:"engine_hours_counter_config"`
	UseV3EngineHours                   bool            `bson:"use_v3_engine_hours" json:"use_v3_engine_hours"`
	HistoryRetentionDays               int             `bson:"history_retention_days" json:"history_retention_days"`
	HarshEventMinSpeed                 Speed           `bson:"harsh_event_min_speed" json:"harsh_event_min_speed"`
}

type Speed struct {
	Value   float64 `bson:"value" json:"value"`
	Unit    string  `bson:"unit" json:"unit"`
	Display string  `bson:"display" json:"display"`
}

type FuelConsumption struct {
	CalculationMethod string  `bson:"calculation_method" json:"calculation_method"`
	Measurement       string  `bson:"measurement" json:"measurement"`
	FuelType          string  `bson:"fuel_type" json:"fuel_type"`
	FuelCost          float64 `bson:"fuel_cost" json:"fuel_cost"`
	FuelEconomy       float64 `bson:"fuel_economy" json:"fuel_economy"`
}
