# Device Dashboard

This project provides a dashboard to visualize and manage device data fetched from the OneStepGPS API. It features a map view, a list view, customizable device icons, and an interface to edit device properties.

---

## Features

- **Map View**: Displays devices on a map with customizable icons.  
- **List View**: Presents device data in a sortable and filterable list.  
- **Device Icon Upload**: Allows users to upload custom icons for each device.  
- **Device Details Editing**: Provides an interface to edit device properties, with validation and change tracking.  
- **Database Refresh**: Enables manual refresh of device data from the OneStepGPS API.  
- **User Preferences**: Lets users set distance units \(km/mi\) and layout preferences \(horizontal/vertical\).

---

## Requirements

- **Go \(1.16 or later\)**: Backend server.  
- **Node.js and npm \(or yarn\)**: Required for the frontend \(Vue.js\).  
- **MongoDB**: Database for storing device data.

---

## Installation and Setup

### 1. Clone the repository
\`\`\`bash
git clone https://github.com/your-username/device-dashboard.git
\`\`\`

### 2. Backend \(Go\)

#### Navigate to the server directory:
\`\`\`bash
cd device-dashboard/server
\`\`\`

#### Install dependencies:
\`\`\`bash
go mod download
\`\`\`

#### Create a \`config.json\` file:  
Create a \`config.json\` file in the server directory and configure the following settings:
\`\`\`json
{
  "server_port": "8080",
  "mongodb_url": "your_mongodb_url",
  "mongodb_port": "27017",
  "database_name": "your_database_name",
  "device_collection_name": "your_device_collection_name",
  "user_collection_name": "your_user_preferences_collection",
  "api_url": "onestepgps_api_url",
  "api_key": "your_onestepgps_api_key",
  "frontend_url": "your_frontend_url",
  "frontend_port": "your_frontend_port",
  "mongodb_username": "your_mongodb_username",
  "mongodb_password": "your_mongodb_password"
}
\`\`\`

**Note**:  
- Replace the placeholder values with your actual settings.  
- If you're running MongoDB locally without authentication, you can omit the \`mongodb_username\` and \`mongodb_password\` fields.

#### Run the server:
\`\`\`bash
go run server.go
\`\`\`

### 3. Frontend \(Vue.js\)

#### Navigate to the frontend directory:
\`\`\`bash
cd device-dashboard/frontend
\`\`\`

#### Install dependencies:
\`\`\`bash
npm install  # or yarn install
\`\`\`

#### Configure environment variables \(optional\):  
Create a \`.env\` file in the frontend directory for sensitive data or configuration settings. Use it to store API keys, backend URLs, etc.

#### Run the development server:
\`\`\`bash
npm run dev  # or yarn dev
\`\`\`

---

## Usage

1. **Access the Dashboard**:  
   Open your web browser and navigate to \`http://localhost:<your_frontend_port>\` \(replace \`<your_frontend_port>\` with your configured port\).

2. **Refresh Data**:  
   Click the refresh button to fetch the latest device data from the OneStepGPS API.

3. **View Devices**:  
   Use the map and list views to visualize and manage device data.

4. **Customize Icons**:  
   Upload custom icons for devices using the upload button in the device edit dialog.

5. **Edit Device Details**:  
   Click on a device in the list to open the edit dialog and modify device properties. Changes are highlighted, allowing you to revert or save edits.

6. **User Preferences**:  
   Use the settings icon to adjust distance units and dashboard layout preferences.

---
