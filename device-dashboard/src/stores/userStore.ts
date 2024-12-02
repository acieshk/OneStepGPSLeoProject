import { defineStore } from 'pinia';
import type { UserPreferences } from '@/types/userPreferences';
import { apiService } from '@/services/api.service';
import type { Device } from '@/types/device';
import {ref } from 'vue';

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
		selectedDevice: null as Device | null,
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
			if (device.iconUrl) {
				return device.iconUrl;
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
					const iconUrl = device.iconUrl || await this.getDeviceIcon({ ...device, color: color });

					const newDevice = {
						...device,
						color: color,
						iconUrl: iconUrl,
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