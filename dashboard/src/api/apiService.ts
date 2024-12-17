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
		const url = `${this.baseUrl}/api/devices`; // Use baseUrl and correct endpoint
		try {
			const response = await fetch(url);
			if (!response.ok) {
				throw await this.handleError(response);
			}
			const data = await response.json();
			return data.result_list as Device[]; 
		} catch (error) {
			console.error('Error fetching devices:', error);
			throw error; // Re-throw to be handled by components
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

    async updateDevice(deviceId: string, updatedDevice: Device): Promise<Device> { // Add deviceId parameter
        const url = `${this.baseUrl}${config.api.endpoints.updateDevice}/${deviceId}`; // Use deviceId in URL

        try {
            const response = await fetch(url, {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(updatedDevice),
            });

            if (!response.ok) {
                throw await this.handleError(response); 
            }

			return await response.json() as Device;
        } catch (error) {
            console.error('Error updating device:', error);
            throw error; // Re-throw error
        }

    }


	async getUserPreferences(userId: string): Promise<UserPreferences> {
		const url = `${this.baseUrl}/api/users/${userId}/preferences`;
		try {
			const response = await fetch(url);
			if (!response.ok) {
				//backend should always return an default user
				throw await this.handleError(response);
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
                throw await this.handleError(response); 
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

	async checkForUpdates(lastUpdate: string | null): Promise<{ needsUpdate: boolean; lastUpdate: string; updatedDevices: Device[] }> {
        const url = `${this.baseUrl}/api/devices/check-updates?lastUpdate=${lastUpdate || '1970-01-01T00:00:00Z'}`; // Provide default timestamp

        try {
            const response = await fetch(url);
            if (!response.ok) {
                throw await this.handleError(response); 
            }
            return await response.json() as { needsUpdate: boolean; lastUpdate: string; updatedDevices: Device[] };
        } catch (error) {
            console.error('Error checking for updates:', error);
            throw error; 
        }
    }

	private async handleError(response: Response): Promise<Error> {
        try {
            const errorData = await response.json(); 
            if (errorData.error) { 
				return new Error(errorData.error as string);  //Return the error message from server. Type guard to prevent error
            } else {
                return new Error(JSON.stringify(errorData)); //Generic error message if correct format is not received from server.
            }


        } catch (error) {
            // If parsing JSON fails, return a generic error message and status
            console.error('Failed to parse error response as JSON', error)
			return new Error(`Network response was not ok ${response.status}`);
        }
    }
}

export const apiService = ApiService.getInstance();