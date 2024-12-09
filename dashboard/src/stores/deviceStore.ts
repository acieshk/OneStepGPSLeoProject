import { defineStore } from 'pinia';
import { Device } from 'src/model/model';
import { apiService } from 'src/api/apiService';
import { ref } from 'vue';

export type PrimitiveValue = string | number | boolean | null | undefined;
export type DeviceValue = PrimitiveValue | Record<string, unknown> | unknown[];

export const useDeviceStore = defineStore('device', () => { 
	const deviceLoaded = ref(false);
	const deviceLoading = ref(false);
	const devices = ref([] as Device[]);
	const selectedDeviceId = ref<string | null>(null);
	const hoveredDeviceId = ref<string | null>(null);
	const editingDevice = ref<Device | null>(null);
	const mapReady = ref(false);
	const mapIconVisibility = ref(new Map<string, boolean>());


	async function loadDevices() {
		//prevent race condition
		if (deviceLoading.value == true) return;
		deviceLoading.value = true;
		try {
			const loadedDevices = await apiService.getDevices(); // Call the service function

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

	async function updateDevice(updatedDevice: Device) {
		console.log(updatedDevice);
		if (!editingDevice.value) {
			console.error('No device to save. editingDevice is null.');
			return; // Or throw an error if appropriate
		}
		try {
			const updatedDevice = editingDevice.value;


			// Find the device in the array by _id
			const deviceIndex = devices.value.findIndex((d) => d._id === updatedDevice._id);

			if (deviceIndex !== -1) {
				// Update the device in the array using splice to maintain reactivity:
				devices.value.splice(deviceIndex, 1, updatedDevice);
			}
			await apiService.updateDevice(updatedDevice._id, updatedDevice);
		}
		catch (error) {
			console.error('Error updating device in store:', error);
			// Add more robust error handling here (e.g., display error to user)
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
        const deviceIndex = devices.value.findIndex(d => d._id === _id);

        if (deviceIndex !== -1) {
            devices.value[deviceIndex].visible = !devices.value[deviceIndex].visible;

			// Trigger reactivity by replacing the devices array.  Deep copy isn't always needed, but better to be safe for arrays
			devices.value = [...devices.value];
        }
    }

	function setMapIconVisibility(_id: string, visibility: boolean) {
        const deviceIndex = devices.value.findIndex(d => d._id === _id);

        if (deviceIndex !== -1) {
            devices.value[deviceIndex].visible = visibility;

			// Trigger reactivity by replacing the devices array.  Deep copy isn't always needed, but better to be safe for arrays
			devices.value = [...devices.value];
        }
	}

	function setEditingDevice(device: Device | null) {
		editingDevice.value = device ? { ...device } : null;
	}

	function setEditingDeviceByID(_id: string) {
		const device = devices.value.find(d => d._id === _id);
		if (device) {
			setEditingDevice(device);
		}
	}

	function updateDeviceProperty(path: string[], newValue: DeviceValue) {
		if (!editingDevice.value) return;

		const updateNestedProperty = (
			obj: Record<string, unknown>,
			pathSegments: string[],
			value: DeviceValue
		): void => {
			const currentKey = pathSegments[0];
			const remainingPath = pathSegments.slice(1);

			// Handle array index notation
			const arrayMatch = currentKey.match(/(\w+)\[(\d+)\]/);
			if (arrayMatch) {
				const [, arrayName, indexStr] = arrayMatch;
				const index = parseInt(indexStr, 10);

				if (Array.isArray(obj[arrayName])) {
					const arr = obj[arrayName] as unknown[];
					if (remainingPath.length === 0) {
						arr[index] = value;
					} else {
						updateNestedProperty(arr[index] as Record<string, unknown>, remainingPath, value);
					}
				}
				return;
			}

			// For non-array properties
			if (remainingPath.length === 0) {
				obj[currentKey] = value;
				return;
			}

			// Recursive descent
			if (typeof obj[currentKey] === 'object' && obj[currentKey] !== null) {
				updateNestedProperty(
					obj[currentKey] as Record<string, unknown>,
					remainingPath,
					value
				);
			}
		};

		// Create a deep copy to trigger reactivity
		const updatedDevice = JSON.parse(JSON.stringify(editingDevice.value));
		updateNestedProperty(updatedDevice, path, newValue);

		editingDevice.value = updatedDevice;
	}

    async function updateIcon(deviceId: string, iconFile: File | null, iconPath: string | null) {
        try {
            deviceLoading.value = true;

            if (iconFile) {  // Handle file upload
                const response = await apiService.uploadDeviceIcon(deviceId, iconFile);
				const newIconURL = response.iconUrl; // or however your API returns the URL

				if (!newIconURL) {
					throw new Error('API did not return an icon URL after upload.');
				}

                // Update iconUrl in editingDevice and devices array
                editingDevice.value!.iconUrl = newIconURL;
				const deviceIndex = devices.value.findIndex(d => d._id === deviceId);
				if (deviceIndex !== -1) {
					devices.value[deviceIndex].iconUrl = newIconURL;
                    devices.value = [...devices.value]; // Trigger reactivity for devices array
				}

            } else if (iconPath) { // Update existing Device
				const deviceIndex = devices.value.findIndex(d => d._id === deviceId);
				if (deviceIndex !== -1) {
                    devices.value[deviceIndex].iconUrl = iconPath;
                    devices.value = [...devices.value]; // Trigger reactivity
                }
            } else {
                editingDevice.value!.iconUrl = '';
				const deviceIndex = devices.value.findIndex(d => d._id === deviceId);
				if (deviceIndex !== -1) {
					devices.value[deviceIndex].iconUrl = '';
                    devices.value = [...devices.value]; // Trigger reactivity for devices array
				}
				await apiService.removeDeviceIcon(deviceId);
			}

        } catch (error) {
			console.error('Error updating icon:', error);
        } finally {
            deviceLoading.value = false;
        }
    }


	return {
		//states
		deviceLoaded, deviceLoading, devices, selectedDeviceId,
		hoveredDeviceId, editingDevice, mapReady, mapIconVisibility,
		//actions
		loadDevices, selectDevice, deselectDevice, setMapReady,
		setHoveredDevice, toggleDeviceVisibility, setMapIconVisibility,
		setEditingDevice, setEditingDeviceByID, updateDeviceProperty,
		updateDevice, updateIcon
	};
});
