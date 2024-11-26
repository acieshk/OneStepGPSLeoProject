import config from '@/config/config';
import { configService } from './config.service';
import type { UserPreferences } from '@/types/userPreferences'; // Correct path - no '.ts' extension needed


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

    async getDevices() {
		const endpoint = configService.getDevicesEndpoint();
        try {
            const response = await fetch(`${this.baseUrl}/devices`);
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        } catch (error) {
            console.error('Error fetching devices:', error);
            throw error;
        }
    }

    async refreshDatabase() {
		const endpoint = `${configService.getApiUrl()}/fetch-devices`;
        try {
            const response = await fetch(`${this.baseUrl}/fetch-devices`);
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        } catch (error) {
            console.error('Error refreshing database:', error);
            throw error;
        }
    }

	async updateDevice(deviceId: string, updatedDevice: Device): Promise<Device> {  // New function
		const endpoint = `${configService.getApiUrl()}/fetch-devices`;
		try {
            const response = await fetch(`${this.baseUrl}/devices/${deviceId}`, {  // Correct URL
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(updatedDevice), // Correctly stringify data
            });

            if (!response.ok) {
                const errorData = await response.json(); // Get error details from response
                throw new Error(`Failed to update device: ${errorData.error || response.statusText}`); // More informative error message
            }
            return response.json();
        } catch (error) {
            console.error('Error updating device:', error);
            throw error; // Re-throw the error for the component to handle
        }
    }



    async uploadDeviceIcon(deviceId: string, iconFile: File): Promise<any> {  // New function
        const endpoint = `${configService.getApiUrl()}/fetch-devices`;
		try {
            const formData = new FormData();
            formData.append('file', iconFile);

            const response = await fetch(`${this.baseUrl}/devices/${deviceId}/icon`, { // Correct URL for icon upload
                method: 'POST',
                body: formData, // Correctly sending FormData
            });

            if (!response.ok) {
                const errorData = await response.json(); // Try to get JSON error; fallback to statusText
                throw new Error(`Failed to upload icon: ${errorData.error || response.statusText}`);
            }


            return await response.json();
        } catch (error) {

            console.error('Error uploading device icon:', error);
            throw error;
        }
    }

	async getUserPreferences(userId: string): Promise<UserPreferences> {
        const endpoint = `${configService.getApiUrl()}`;
		try {
            const response = await fetch(`${endpoint}/user-preferences/${userId}`);
            if (!response.ok) {
                if (response.status === 404) {  // Handle "Not Found" specifically
                    // Return default preferences if not found
                    return {
                        userId: userId,
                        distanceUnit: 'km',  // Your default values
                        layout: 'horizontal', // ...
                        // ... other default preferences
                    };


                }

                // Other error, throw it
                const errorData = await response.json();
                throw new Error(`Failed to fetch preferences: ${errorData.error || response.statusText}`);

            }


            return await response.json();  // Return preferences
        } catch (error) {
            console.error("Error getting user preferences:", error);
            throw error; // Re-throw to be handled by the component
        }
    }

	async saveUserPreferences(preferences: UserPreferences): Promise<UserPreferences> {
		const endpoint = `${configService.getApiUrl()}`;
		try {

            const response = await fetch(`${endpoint}/user-preferences`, {
                method: 'POST',  // Use POST to create or update
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(preferences),
            });

            if (!response.ok) {
                 const errorData = await response.json();
                 throw new Error(`Failed to save preferences: ${errorData.error || response.statusText}`); // Improved error message

            }

            return await response.json(); // Correctly return the result
        } catch (error) {

            console.error("Error saving user preferences:", error);
            throw error;  // Re-throw for component to handle
        }

    }
}

export const apiService = ApiService.getInstance();