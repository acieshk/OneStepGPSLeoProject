// import config from '@/config/config';
// import { configService } from './config.service';
// import type { UserPreferences } from '@/types/userPreferences'; // Correct path - no '.ts' extension needed
// import type { Device } from '@/types/device';

import { Device } from 'src/types/device';

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

export async function getDevices(): Promise<Device[]> {
	try {
		const response = await fetch(`${config.api.baseUrl}:${config.api.port}${config.api.endpoints.devices}`);
		if (!response.ok) {
			const errorText = await response.text();
			throw new Error(`Network response was not ok: ${response.status} - ${errorText}`);
		}
		const data = await response.json();
		const devices = data.result_list;

		console.log(devices);
		return devices as Device[]; // Type assertion after validation
	} catch (error) {
		console.error('Error fetching devices:', error);
		throw error;
	}
}

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




	// async refreshDatabase() {
	// 	const endpoint = `${configService.getApiUrl()}/fetch-devices`;
	//     try {
	//         const response = await fetch(`${this.baseUrl}/fetch-devices`);
	//         if (!response.ok) {
	//             throw new Error('Network response was not ok');
	//         }
	//         return response.json();
	//     } catch (error) {
	//         console.error('Error refreshing database:', error);
	//         throw error;
	//     }
	// }

	async updateDevice(_id: string, updatedDevice: Device): Promise<Device> {
		console.log('update Device API');
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



	// async uploadDeviceIcon(deviceId: string, iconFile: File): Promise<any> {  // New function
	//     const endpoint = `${configService.getApiUrl()}/fetch-devices`;
	// 	try {
	//         const formData = new FormData();
	//         formData.append('file', iconFile);

	//         const response = await fetch(`${this.baseUrl}/devices/${deviceId}/icon`, { // Correct URL for icon upload
	//             method: 'POST',
	//             body: formData, // Correctly sending FormData
	//         });

	//         if (!response.ok) {
	//             const errorData = await response.json(); // Try to get JSON error; fallback to statusText
	//             throw new Error(`Failed to upload icon: ${errorData.error || response.statusText}`);
	//         }


	//         return await response.json();
	//     } catch (error) {

	//         console.error('Error uploading device icon:', error);
	//         throw error;
	//     }
	// }

	// async getUserPreferences(userId: string): Promise<UserPreferences> {
	//     const endpoint = `${configService.getApiUrl()}`;
	// 	try {
	//         const response = await fetch(`${endpoint}/user-preferences/${userId}`);
	//         if (!response.ok) {
	//             if (response.status === 404) {  // Handle "Not Found" specifically
	//                 // Return default preferences if not found
	//                 return {
	//                     userId: userId,
	//                     distanceUnit: 'km',  // Your default values
	//                     layout: 'horizontal', // ...
	//                     // ... other default preferences
	//                 };


	//             }

	//             // Other error, throw it
	//             const errorData = await response.json();
	//             throw new Error(`Failed to fetch preferences: ${errorData.error || response.statusText}`);

	//         }


	//         return await response.json();  // Return preferences
	//     } catch (error) {
	//         console.error("Error getting user preferences:", error);
	//         throw error; // Re-throw to be handled by the component
	//     }
	// }

	// // return path of the icon, or null if icon is not found
	// async getIcon(deviceId: string): Promise<string | null> {
	// 	if (!deviceId) {
	// 		return null;
	// 	}
	// 	return `${configService.getApiUrl()}/icons/${deviceId}.png`;

	// 	const iconUrl = `${configService.getApiUrl()}/getIcon/${deviceId}`; // Correct URL
	// 	console.log("GET ICON:" + deviceId);
	// 	console.log(`${configService.getApiUrl()}`)
	// 	console.log("icon url" + iconUrl);
	// 	try {
	// 		const response = await fetch(iconUrl);
	// 		if (!response.ok) {  // Check for any server error (not just 404)
	// 			const errorText = await response.text()
	// 			console.error(`Error getting icon for device ${deviceId}:`, response.status,  errorText);
	// 			throw new Error (`Error getting icon for device ${deviceId}: ${response.status} ${errorText}`) // Throw error for the component to handle
	// 		}

	// 		const iconPathFromServer = await response.text(); // Get the icon path from the server

	// 		if (iconPathFromServer !== "") {  // Custom icon exists
	// 			return `${configService.getApiUrl()}${iconPathFromServer}`; // Construct and return the full URL
	// 		} else {
	// 			return null; // No custom icon, return null (for default icon logic)
	// 		}

	// 	} catch (error) {
	// 		console.error("Error fetching icon:", error);
	// 		return null;  // Handle the error as needed (e.g., return null for default icon)
	// 	}
	// }

	// async saveUserPreferences(preferences: UserPreferences): Promise<UserPreferences> {
	// 	const endpoint = `${configService.getApiUrl()}`;
	// 	try {

	//         const response = await fetch(`${endpoint}/user-preferences`, {
	//             method: 'POST',  // Use POST to create or update
	//             headers: { 'Content-Type': 'application/json' },
	//             body: JSON.stringify(preferences),
	//         });

	//         if (!response.ok) {
	//              const errorData = await response.json();
	//              throw new Error(`Failed to save preferences: ${errorData.error || response.statusText}`); // Improved error message

	//         }

	//         return await response.json(); // Correctly return the result
	//     } catch (error) {

	//         console.error("Error saving user preferences:", error);
	//         throw error;  // Re-throw for component to handle
	//     }

	// }
}

export const apiService = ApiService.getInstance();