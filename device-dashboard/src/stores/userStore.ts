import { defineStore } from 'pinia';
import type { UserPreferences } from '@/types/userPreferences';
import { apiService } from '@/services/api.service';
import Device from '@/components/Device.vue';

export const useUserStore = defineStore({
	id: 'user',
	state: () => ({
		devices: [] as Device[], // Devices array managed in the store
		userPreferences: {
		  userId: 'default',
		  distanceUnit: 'km',
		  layout: 'horizontal',
		} as UserPreferences,
		loading: false,     // Add loading state
		error: null as string | null, // Add error state
	  }),

 getters: {
    getLayout: (state) => state.userPreferences.layout,
    getDistanceUnit: (state) => state.userPreferences.distanceUnit,
 },

	actions: {
		async fetchDevices() {
			this.loading = true; // Set loading state
			this.error = null;    // Clear any previous errors
			try {
			  const fetchedDevices = await apiService.getDevices(); // Use apiService
			  this.devices = fetchedDevices.result_list.map((device: any) => ({   // Map over results
				...device,
				visible: true // Set default visibility to true
			  }));
	  
	  
			  // Convert date/time string properties to Date objects
			  this.devices.forEach(device => {
				for (const key in device) {
	  
					const value = device[key];
					if (typeof value === 'string') {
	  
						const parsedDate = new Date(value)
						if (!isNaN(parsedDate)) {
	  
							device[key] = parsedDate
						}
					}
				}
	  
	  
			  });
	  
			} catch (error) {
			  this.error = 'Failed to fetch devices'; // Set error state
			  console.error(error);
			} finally {
			  this.loading = false; // Clear loading state
			}
		  },
		updateUserPreferences(preferences: UserPreferences) {
			this.userPreferences = preferences; // Update user preferences
		},
	},
});