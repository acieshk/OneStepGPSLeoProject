import { defineStore } from 'pinia';
import type { UserPreferences } from '@/types/userPreferences';
import { apiService } from '@/services/api.service';
import type { Device } from '@/types/device';

// Color palette
// const COLOR_PALETTE = [
// 	'#1F77B4', // Muted Blue
// 	'#FF7F0E', // Vivid Orange
// 	'#2CA02C', // Fresh Green
// 	'#D62728', // Brick Red
// 	'#9467BD', // Soft Purple
// 	'#8C564B', // Brown
// 	'#E377C2', // Pink
// 	'#7F7F7F', // Gray
// 	'#BCBD22', // Olive Green
// 	'#17BECF', // Teal
// 	'#3498DB', // Bright Blue
// 	'#2ECC71', // Emerald Green
// 	'#E74C3C', // Coral Red
// 	'#9B59B6', // Lavender
// 	'#34495E', // Dark Slate Blue
// 	'#16A085', // Sea Green
// 	'#F39C12', // Sunflower Yellow
// 	'#2980B9', // Deep Blue
// 	'#8E44AD', // Deep Purple
// 	'#2C3E50'  // Navy Blue
// ];

// using https://github.com/pointhi/leaflet-color-markers
const COLOR_PALETTE = [
	"blue",
	"gold",
	"red",
	"green",
	"orange",
	"yellow",
	"violet",
	"grey",
	"black"
];

export const useUserStore = defineStore({
	id: 'user',
	state: () => ({
		devices: [] as Device[], // Devices array managed in the store
		userPreferences: {} as UserPreferences,
		loading: false,     // Add loading state
		error: null as string | null, // Add error state
	  }),

 getters: {
    getLayout: (state) => state.userPreferences.layout,
    getDistanceUnit: (state) => state.userPreferences.distanceUnit,
 },

	actions: {
		assignDeviceColor(device: Device, index: number): string {
			return device.color || COLOR_PALETTE[index % COLOR_PALETTE.length];
		},
		//This logic is needed because for leaflet
		async getDeviceIcon(device: Device): Promise<string> { //getIcon will always return something
			if (device.iconURL) {
				return device.iconURL;
			} else {
				const color = device.color || 'blue';
				return `https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-${color}.png`; // Construct default icon URL
			}
		},
		async fetchDevices() {
			try {
				const fetchedDevices = await apiService.getDevices();
				this.devices = await Promise.all(fetchedDevices.result_list.map(async (device: Device, index: number) => {
					const color = this.assignDeviceColor(device, index);
					const iconURL = await this.getDeviceIcon({ ...device, color: color });
		
					const newDevice = {
						...device,
						color: color,
						iconURL: iconURL,
						visible: true
					};
					
					if (newDevice.latest_device_point) {
						newDevice.latest_device_point.dt_tracker = new Date(newDevice.latest_device_point.dt_tracker);
					}
					return newDevice;
				}));
			} catch (error) {
				this.error = 'Failed to fetch devices';
				console.error(error);
			} finally {
				this.loading = false;
			}
		},
		updateUserPreferences(preferences: UserPreferences) {
			this.userPreferences = preferences; // Update user preferences
		},
	},
});