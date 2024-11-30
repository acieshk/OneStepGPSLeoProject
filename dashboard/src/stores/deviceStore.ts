import { defineStore } from 'pinia';
import { Device } from 'src/types/device';
import { getDevices } from 'src/api/apiService';
import { ref } from 'vue';

export const useDeviceStore = defineStore('device', () => { // No object, just the store ID
	const deviceLoaded = ref(false);
	const deviceLoading = ref(false);
	const devices = ref([] as Device[]);
	const selectedDeviceId  = ref<string | null>(null);
	const hoveredDeviceId = ref<string | null>(null);
	const editingDevice = ref(null as Device | null);
	const mapReady = ref(false);
	const mapIconVisibility = ref(new Map<string, boolean>());

	async function loadDevices() {
		deviceLoading.value = true;
		try {
			const loadedDevices = await getDevices(); // Call the service function

			// Ensure 'visible' property exists and defaults to true:
			devices.value = loadedDevices.map(device => ({
				...device,  // Spread existing properties
				visible: device.visible ?? true, // Add or update 'visible'
			}));
			deviceLoading.value = false;
			deviceLoaded.value = true;
		} catch (error) {
			if (error instanceof Error) {
				console.error('Error loading devices:', error.message); // Access properties for logging
			} else {
				console.error('An unexpected error occurred.');  // Log the unknown error
			}
			deviceLoading.value = false;
		}
	}

	function selectDevice(_id: string) {
		selectedDeviceId.value = _id;
	}

	function deselectDevice() {
		selectedDeviceId.value = null;
    }

	function setHoveredDevice(_id: string | null) {
        hoveredDeviceId.value = _id;
    }

	function setMapReady(isReady: boolean) {
		mapReady.value = isReady;
	}

	function toggleDeviceVisibility(_id: string) {
		const device = devices.value.find(d => d._id === _id);
		if (device) {
			device.visible = !device.visible; // Toggle visibility
			// Update the map icon visibility map:
			setMapIconVisibility(_id, device.visible);
		}
	}

	function setMapIconVisibility(_id: string, visibility: boolean) {
		mapIconVisibility.value.set(_id, visibility);
		mapIconVisibility.value = new Map(mapIconVisibility.value);
	}


	return {
		//states
		deviceLoaded, deviceLoading, devices, selectedDeviceId, 
		hoveredDeviceId, editingDevice, mapReady, mapIconVisibility,
		//actions
		loadDevices, selectDevice, deselectDevice, setMapReady, 
		setHoveredDevice, toggleDeviceVisibility, setMapIconVisibility
	};
});
