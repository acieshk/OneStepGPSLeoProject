interface Config {
	api: {
		baseUrl: string;
		port: number;
		endpoints: {
			devices: string;
			fetchDevices: string;
			userPreferences: string;
			updateDevice: string;
			uploadDeviceIcon: string;
		};
	};
	server: {
		baseUrl: string;
		port: number;
		mongodb: {
			url: string;
			port: number;
			database: string;
			devicesCollection: string;
			userPreferencesCollection: string;
		};
	};

	userId: string;

}


const config: Config = {
	userId: "default",
    api: {
        baseUrl: 'http://localhost',
        port: 8080,
        endpoints: {
            devices: '/devices',
            fetchDevices: '/fetch-devices', // Correct path for refreshDatabase
            userPreferences: '/user-preferences', // Path for user preferences
            updateDevice: '/devices', // Endpoint for updating devices
            uploadDeviceIcon: '/devices', // Base endpoint for icon uploads (deviceId will be appended)
        },

    },
    server: {
        baseUrl: 'http://localhost',
        port: 8081,        // Or your server port
        mongodb: {
            url: 'your_mongodb_url', //e.g., 'localhost' or your MongoDB connection string
            port: 27017,             // Your MongoDB port
            database: 'your_database_name',
            devicesCollection: 'devices',          //  Collection name for device data
            userPreferencesCollection: 'user_preferences', // Collection name for user preferences
        },
    },

};

export default config;