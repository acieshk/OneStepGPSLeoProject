# Device Dashboard

This project provides a dashboard to visualize and manage device data fetched from the OneStepGPS API. 

---

## Features

- **Map View**: Displays devices on a map with customizable icons.
- **List View**: Presents device data in a sortable and filterable list.  
- **Device Icon Upload**: Allows users to upload custom icons for each device.  
- **Device Details Editing**: Provides an treeview to edit device properties. 
- **Database Refresh**: Enables manual refresh of device data from the OneStepGPS API. 
- **Mock Server**: For testing purposes. It speeds up development when the device is not updating. 
- **User Preferences**: Lets users set distance units (km/mi) and save width of the left side drawer.

---

## Requirements

- **Go (1.16 or later)**: Backend server.  
- **Node.js and npm**: Required for the frontend (Vue.js).  
- [**MongoDB**](https://www.mongodb.com/products/tools/compass): Database for storing device data.

---

## Installation and Setup

### 1. Clone the repository
```bash
git clone https://github.com/acieshk/OneStepGPSLeoProject.git
```

### 2. Backend (Go)

#### Navigate to the server directory:
```bash
cd OneStepGPSLeoProject\server
```

#### Install dependencies:
```bash
go mod download
```

#### Edit the config.json file:  
Edit the config.json file in the server directory and configure the following settings:
```json
{
    "server_port": "8080",
    "mongodb_url": "localhost",
    "mongodb_port": "27017",
    "mongodb_username": "",
    "mongodb_password": "",
    "database_name": "onestepgps",  
    "device_collection_name": "device_data",
	"device_setting_collection_name": "device_setting",
    "user_collection_name": "user_preferences", 
	"icon_dir": "icons",
    "api_url": "https://track.onestepgps.com/v3/api/public/device?latest_point=true&api-key=",
    "api_key": "",
	"update_interval_seconds": 10
}


```
Failed to load config: APIKey is missing in config.json
exit status 1
> [!CAUTION]  
> If the API key is missing, the following error will occur:
> ``` Failed to load config: APIKey is missing in config.json exit status 1 ```

**Note**:  
- You have to enter api_key as it is removed for security purpose.
- Replace the placeholder values with your actual settings.  
- If you're running MongoDB locally without authentication, you can omit the `mongodb_username` and `mongodb_password` fields.
- The server will automatically create a database with collections. You may change to a different database name.

#### Run the server:
```bash
go run server.go
```
## Running the Mock Server (Optional)

A mock server is provided for development and testing purposes. This allows you to run the dashboard and test its functionality 

The mock server can randomly mutate device data, allowing you to test how the dashboard handles dynamic updates and changes.

The data that will be mutated are: latitude, longitude, online

To run the mock server, use the following command instead:

```bash
go run main.go -mock -mutateDevice 4 -mutateChance 0.5
```

-mock: Enables mock server mode. Required to use mock server.
-mutateChance=0.3 (optional): Sets the probability of data mutation to 30%. The value should be between 0.0 and 1.0. Defaults to 0.3.
-mutateDevice=2 (optional): Sets the number of devices to mutate if a mutation occurs. Defaults to 2.
The mock server will start on the port specified in your config.json file, or port 8081 if not configured. 

### 3. Frontend (Vue.js)

#### Open a new console. Navigate to the frontend directory:
```bash
cd OneStepGPSLeoProject\dashboard
```

#### Install dependencies:
```bash
npm install  # or yarn install
```

#### Run the development server:
```bash
npm run dev  # or yarn dev
```

---

## Usage

1. **Access the Dashboard**:  
   Open your web browser and navigate to `http://localhost:<your_frontend_port>` (replace `<your_frontend_port>` with your configured port).

2. **Refresh Data**:  
   Real time polling is supported.

3. **View Devices**:  
   Use the map and list views to visualize and manage device data.

4. **Customize Icons**:  
   Upload custom icons for devices using the upload button in the device edit dialog.

5. **Edit Device Details**:  
   Click on a device in the list to open the edit dialog and modify device properties. Changes are highlighted, allowing you to revert or save edits.

6. **User Preferences**:  
   Use the settings icon to adjust distance units and dashboard layout preferences.

7. **Concurrency**:
  Handle situation where multiple clients are updating. 
---
## Lessons Learnt
Transitioning from Angular to Vue for this project has been a significant learning experience for me. As a former Angular developer, I encountered several challenges while adapting to Vue's ecosystem. One of the major hurdles was navigating the various UI tools available. I found that some libraries lacked sufficient documentation and functionality, which hindered my progress. For instance, I initially used Element Plus for form components but had to switch to Quasar to better meet my needs. Similarly, I started with Leaflet for mapping but ultimately transitioned to OpenLayers for its enhanced capabilities.

Despite these challenges, I genuinely enjoyed working within the Vue ecosystem. The flexibility and simplicity of Vue have greatly facilitated my development process, allowing me to focus more on building features rather than getting bogged down by complex configurations. This project has not only improved my technical skills but also deepened my appreciation for Vue as a powerful framework for building user interfaces.

Overall, this experience has taught me the importance of choosing the right tools and being adaptable in the face of challenges, which I believe will be invaluable in my future projects.
---
## Future Improvements

* **Cloud Storage for Icons:** Currently, device icons are stored on the server's local file system. For production deployments, migrating to a cloud storage service like Google Cloud Storage, Amazon S3, or Azure Blob Storage is recommended for improved scalability, security, and maintainability.
* **Form Validation:** Currently, basic client-side form validation is implemented using Vue.js
* **User Authentication and Authorization:** Could implement a user system with Implement a secure authentication and authorization mechanism to control access to the dashboard and its features.  Consider using industry-standard authentication protocols like OAuth 2.0 or OpenID Connect.
* **API Rate Limiting:** Implement API rate limiting to prevent abuse of the OneStepGPS API/ and ensure the application's stability.
* **Improved Error Handling:** Enhance error handling throughout the application to provide more informative and user-friendly error messages.
* **Unit and Integration Tests:** Write comprehensive unit and integration tests to improve code quality and ensure the application's reliability.
* **Deployment Automation:** Automate the deployment process using tools like Docker and Kubernetes for easier and more reliable deployments.
* **Real-time Updates:** Implement real-time updates for device locations using WebSockets or Server-Sent Events.  This would provide a more dynamic and responsive user experience.
* **Performance Optimization:** Optimize database queries and frontend rendering to improve the application's performance and responsiveness, especially when handling a large number of devices.
