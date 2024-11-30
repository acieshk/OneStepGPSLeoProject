// src/stores/deviceStore.ts
import { defineStore } from 'pinia';
import type { Device } from '@/types/device';


export const useDeviceStore = defineStore({
  id: 'device',
  state: () => ({
	loaded: false,						// if device is already loaded
	loading: false,                    	// Add loading state. Initialized but not loaded
    devices: [] as Device[],			// Devices array managed in the store	
	selectedDevice: null as Device | null,
  }),
  getters: {
	
  },
  actions: {
	selectDevice(_id: string) {  // Use regular function
		this.selectedDevice = this.devices.find((device) => device._id === _id) ?? null;
	},
	deselectDevice: () => {
		this.selectedDevice = null;
	}
  }
});