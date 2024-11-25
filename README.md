# Device Dashboard

This project provides a dashboard to visualize and manage device data fetched from the OneStepGPS API.  It features a map view, a list view, customizable device icons, and an interface to edit device properties.

## Features

* **Map View:** Displays devices on a map with customizable icons.
* **List View:** Presents device data in a sortable and filterable list.
* **Device Icon Upload:**  Allows users to upload custom icons for each device.
* **Device Details Editing:** Provides an interface to edit device properties, with validation and change tracking.
* **Database Refresh:** Enables manual refresh of device data from the OneStepGPS API.
* **User Preferences:**  Allows users to set distance units (km/mi) and layout preferences (horizontal/vertical).

## Requirements

* **Go (1.16 or later):** Used for the backend server.
* **Node.js and npm (or yarn):**  Required for the frontend (Vue.js).
* **MongoDB:** Used as the database to store device data.

## Installation and Setup

1. **Clone the repository:**

   ```bash
   git clone https://github.com/your-username/device-dashboard.git

   Navigate to the server directory:

cd device-dashboard/server
Install dependencies:

go mod download
Create a config.json file: Create a config.json file in the server directory and configure the following settings:

{
  "server_port": "8080",
  "mongodb_url": "your_mongodb_url",  // e.g., localhost or mongodb://your-mongodb-hostname
  "mongodb_port": "27017",
  "database_name": "your_database_name",
  "device_collection_name": "your_device_collection_name",
  "user_collection_name": "your_user_preferences_collection",
  "api_url": "onestepgps_api_url",  // OneStepGPS API endpoint
  "api_key": "your_onestepgps_api_key",
  "frontend_url": "your_frontend_url", // Frontend URL for CORS configuration (important for development)
  "frontend_port": "your_frontend_port",
  "mongodb_username": "your_mongodb_username",  // If MongoDB authentication is enabled
  "mongodb_password": "your_mongodb_password"   // If MongoDB authentication is enabled

}
Replace the placeholder values with your actual settings.
If you're running MongoDB locally without authentication, you can omit the mongodb_username and mongodb_password fields.
Run the server:

go run server.go
Frontend (Vue.js)
Navigate to the frontend directory:

cd device-dashboard/device-dashboard
Install dependencies:

npm install  // or yarn install
Configure environment variables (optional): Create a .env file in the frontend directory to store sensitive data or configuration settings. You can use environment variables for your API keys, backend URLs, etc.

Run the development server:

npm run dev   // or yarn dev
Usage
Access the dashboard: Open your web browser and navigate to http://localhost:your_frontend_port (replace with your configured frontend port).

Refresh Data: Click the refresh button to fetch the latest device data from the OneStepGPS API.

View Devices: Use the map and list views to visualize and manage device data.

Customize Icons: Upload custom icons for devices using the upload button in the device edit dialog.

Edit Device Details: Click on a device in the list to open the edit dialog and modify device properties. Changes are highlighted, and you can revert or save edits.

User Preferences: Use the settings icon to change the distance units and layout of the dashboard.