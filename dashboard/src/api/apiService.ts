import { Device, UserPreferences } from 'src/model/model';

const config = {
	userId: 'default',
	api: {
		baseUrl: 'http://localhost',
		port: 8080,
		endpoints: {
			devices: '/devices',
			refreshDatabase: '/fetch-devices',
			userPreferences: '/user-preferences',
			updateDevice: '/devices',
			uploadDeviceIcon: '/devices',
		},

	}
};

class ApiService {
	private static instance: ApiService;
	private baseUrl: string;

	private constructor() {
		this.baseUrl = `${config.api.baseUrl}:${config.api.port}`;
	}


	static getInstance(): ApiService {
		if (!ApiService.instance) {
			ApiService.instance = new ApiService();
		}
		return ApiService.instance;
	}

	async getDevices(): Promise<Device[]> {
		try {
			const response = await fetch(`${config.api.baseUrl}:${config.api.port}${config.api.endpoints.devices}`);
			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(`Network response was not ok: ${response.status} - ${errorText}`);
			}
			const data = await response.json();
			const devices = data.result_list;
			return devices as Device[];
		} catch (error) {
			console.error('Error fetching devices:', error);
			throw error;
		}
	}

	async refreshDatabase(): Promise<string> {
		try {
			const response = await fetch(`${config.api.baseUrl}:${config.api.port}${config.api.endpoints.refreshDatabase}`, { // Use the correct config value for endpoint
				method: 'POST', // POST is usually appropriate for triggering an action
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(`Network response was not ok: ${response.status} - ${errorText}`); // Detailed error message
			}

			return await response.json(); // Return the response data (if any)
		} catch (error) {
			console.error('Error refreshing database:', error);
			throw error; // Re-throw for error handling in the component
		}
	}

	async updateDevice(_id: string, updatedDevice: Device): Promise<Device> {
		try {
			const response = await fetch(`${config.api.baseUrl}:${config.api.port}${config.api.endpoints.updateDevice}/${_id}`, { // Use correct endpoint and deviceId
				method: 'PUT', // Or PATCH if appropriate for your API
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify(updatedDevice),
			});

			if (!response.ok) {
				const errorText = await response.text(); // Get error text for debugging
				throw new Error(`Network response was not ok: ${response.status} - ${errorText}`); // More informative error message
			}

			const updatedDeviceFromServer = await response.json() as Device;

			return updatedDeviceFromServer
		} catch (error) {
			console.error('Error updating device:', error);
			throw error; // Re-throw for error handling in the component
		}
	}

	async getUserPreferences(userId: string): Promise<UserPreferences> {
		try {
			const response = await fetch(`${config.api.baseUrl}:${config.api.port}${config.api.endpoints.userPreferences}/${userId}`);
			if (!response.ok) {
				if (response.status === 404) {
					// Create default preferences for new users
					const defaultPrefs: UserPreferences = {
						rowPerPage: 20,
						DeviceListWidth: 200,
						unit: 'original',
					};
					const createResponse = await fetch(`${config.api.baseUrl}:${config.api.port}${config.api.endpoints.userPreferences}`, {  // POST to create new prefs
						method: 'POST',
						headers: { 'Content-Type': 'application/json' },
						body: JSON.stringify(defaultPrefs),
					});
					if (!createResponse.ok) {  // Check if POST was successful
						const createError = await createResponse.text();  // Get the error from the POST request
						throw new Error(`Failed to create default preferences: ${createResponse.status} - ${createError}`);  // Detailed error message
					}
					// Get the created preferences (if returned by the API, or use default values if not)
					return createResponse.json().catch(() => defaultPrefs) as Promise<UserPreferences>;
				} else {
					// Handle other errors (not 404)
					const errorText = await response.text();
					throw new Error(`Failed to fetch preferences: ${response.status} - ${errorText}`);
				}
			}
			return await response.json() as UserPreferences;
		} catch (error) {
			console.error('Error getting user preferences:', error);
			throw error;  // Re-throw for handling in component
		}
	}

	async saveUserPreferences(preferences: UserPreferences): Promise<UserPreferences> {
		try {
			const response = await fetch(`${config.api.baseUrl}:${config.api.port}${config.api.endpoints.userPreferences}`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(preferences),
			});

			if (!response.ok) {
				const errorText = await response.text(); // Get error content for debugging
				throw new Error(`Failed to save preferences: ${response.status} - ${errorText}`);
			}

			return await response.json() as UserPreferences;  // Return updated preferences after successful save, assuming server sends data back
		} catch (error) {
			console.error('Error saving user preferences:', error);
			throw error; // Re-throw for component to handle
		}
	}

	async removeDeviceIcon(deviceId: string): Promise<unknown> {
		const response = await fetch(`${this.baseUrl}/devices/${deviceId}/icon?remove=true`, {
			method: 'POST',
		});
		if (!response.ok) {
			throw new Error('Failed to remove icon');
		}
		return await response.json();
	}

	async uploadDeviceIcon(deviceId: string, iconFile: File): Promise<{ iconUrl: string }> {
		try {
			const formData = new FormData();
			formData.append('file', iconFile);
			const response = await fetch(`${this.baseUrl}/devices/${deviceId}/icon`, {
				method: 'POST',
				body: formData
			});
			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(JSON.stringify(errorData));
			}
			return { iconUrl: `${this.baseUrl}/icons/${deviceId}.png` };
		} catch (error) {
			console.error('Error uploading device icon:', error);
			throw error;
		}
	}

	// return path of the icon, or null if icon is not found
	async getIcon(deviceId: string): Promise<string | null> {
		try {
			const response = await fetch(`${config.api.baseUrl}:${config.api.port}/icons/${deviceId}`); // Correct URL structure.  Adjust as needed.
			if (!response.ok) {
				if (response.status === 404) { // If no icon exists for this device ID, this is a valid case.
					return null;  // Return null to indicate no icon, don't throw
				}
				// For other errors, throw
				const errorText = await response.text();
				throw new Error(`Error getting icon: ${response.status} - ${errorText}`); // Detailed error
			}
			// Assuming the API returns the icon URL directly:
			return await response.json() as string; // Type assertion since server should return path as string
		} catch (error) {
			console.error('Error fetching icon:', error);  // Log the error for debugging
			throw error; // then rethrow it
			// Or return a default icon URL if there is one, or return null and have UI use a placeholder icon.
		}
	}
}

export const apiService = ApiService.getInstance();