import { defineStore } from 'pinia';
import { Device, DevicePoint, DeviceSettings } from 'src/model/model';
import { apiService } from 'src/api/apiService';
import { Ref, ref } from 'vue';
import { Notify } from 'quasar';

export type PrimitiveValue = string | number | boolean | null | undefined;
export type DeviceValue = PrimitiveValue | Record<string, unknown> | unknown[];

// Define types for clarity 

export const useDeviceStore = defineStore('device', () => { 
	const deviceLoaded = ref(false);
	const deviceLoading = ref(false);
	const devices = ref([] as Device[]);
	const selectedDeviceId = ref<string | null>(null);
	const hoveredDeviceId = ref<string | null>(null);
	const editingDevice = ref<Device | null>(null);
	const mapReady = ref(false);
	const mapIconVisibility = ref(new Map<string, boolean>());
	const lastUpdate = ref('1970-01-01T00:00:00Z'); // RFC3339
	const pollingInterval: Ref<number | null> = ref(10000); // 10 seconds
    const deviceSettings = ref<DeviceSettings | null>(null); 
    const deviceSettingsLoaded = ref(false);  
    const deviceSettingsLoading = ref(false);
	const pollingActive = ref(false);

    const mergeUpdatedDevices = (updatedDevices: Device[], iconMap: {[key: string]: string} | null) => {
        if (!updatedDevices || updatedDevices.length === 0) {
            // Handle empty or null updatedDevices appropriately. Log a message or return early.
			if (!updatedDevices){
				console.log('updatedDevices is null or undefined, skipping merge')
			} else if (updatedDevices.length === 0){
				console.log('No updated devices received, skipping merge');
			}

            if(iconMap) { // Update iconUrl from iconMap even if updatedDevices is empty.
                console.log(iconMap);
				devices.value.forEach(device => {
					device.iconUrl = iconMap[device.device_id];
                });
            }
			devices.value = [...devices.value]; // Trigger reactivity
			return; // Return early to avoid errors
        }

		// Create a map of updated devices for efficient lookup
        const updatedDeviceMap = updatedDevices.reduce((map, device) => {
            if (typeof device.device_id === 'string') { // Ensure device_id is a string
                map[device.device_id] = device;
				return map;
            } else {
				console.warn('Invalid device_id format, skipping device:', device) // Log error if invalid device_id
				return map;
			}
        }, {} as { [key: string]: Device });

		// Iterate through existing devices and merge updates
        devices.value.forEach((device) => {
			// Correctly access updated device from map
            const updatedDevice = updatedDeviceMap[device.device_id as string]; // Access updated device by ID

			// Merge if an updated device exists for this ID, and if updatedDevice is not null/undefined
            if (updatedDevice) {

				device.updated_at = updatedDevice.updated_at || device.updated_at; // Correctly merge updated_at
				device.online = updatedDevice.online ?? device.online;       // Correctly merge online, using nullish coalescing operator (??)
				device.latest_device_point = updatedDevice?.latest_device_point || device.latest_device_point;
				device.latest_accurate_device_point = updatedDevice?.latest_accurate_device_point || device.latest_accurate_device_point;
                // Correctly set iconURL from iconMap
				if (iconMap) {

                    const iconURL = iconMap[device.device_id as string];

                    if(iconURL){ //Only set iconURL if it exists in the iconMap
                        device.iconUrl = iconURL
                    }
				}
            }
        });
        devices.value = [...devices.value]; // Trigger reactivity
    };
	const poll = async () => {
		console.log('polling');
		try {
		  const response = await apiService.checkForUpdates(lastUpdate.value);
		  lastUpdate.value = response.lastUpdate;
		  mergeUpdatedDevices(response.updatedDevices, response.icon_map);
		} catch (error) {
		  console.error('Polling error', error);
		}
	  };
	
	  const startPolling = async () => {
		// If polling is already active, just trigger an immediate poll
		if (pollingActive.value) {
		  await poll();
		  return;
		}
	
		pollingActive.value = true;
		
		try {
		  await loadDevices();
		  // Initial poll
		  await poll();
		  
		  // Clear any existing interval
		  if (pollingInterval.value) {
			clearInterval(pollingInterval.value);
		  }
		  
		  // Set up regular polling interval (10 seconds)
		  pollingInterval.value = window.setInterval(() => {
			poll();
		  }, 10000);
		} catch (error) {
		  pollingActive.value = false;
		  return;
		}
	  };
	
	  const stopPolling = () => {
		if (pollingInterval.value) {
		  clearInterval(pollingInterval.value);
		  pollingInterval.value = null;
		}
		pollingActive.value = false;
	  };



	const loadDevices = async () => {
		console.log('loadDevices');
		if (deviceLoading.value) return;

		deviceLoading.value = true;

		try {
			const loadedDevices = await apiService.getDevices();
			console.log(loadedDevices);
			devices.value = loadedDevices.map((device) => {
				let latestDevicePoint: DevicePoint | null = null;
				let latestAccurateDevicePoint: DevicePoint | null = null;

				if (isDevicePoint(device.latest_device_point)) {
					latestDevicePoint = device.latest_device_point as DevicePoint;
				}

				if (isDevicePoint(device.latest_accurate_device_point)) {
					latestAccurateDevicePoint = device.latest_accurate_device_point as DevicePoint;
				}

				return {
					...device,
					visible: device.visible ?? true,
					latest_device_point: latestDevicePoint,
					latest_accurate_device_point: latestAccurateDevicePoint,
				};
			});

			deviceLoading.value = false;
			deviceLoaded.value = true;
			lastUpdate.value = new Date().toISOString();
		} catch (error) {
			deviceLoading.value = false;
			console.error('Error loading devices:', error);
			// Consider adding more specific error handling or user feedback here.
		}
	};
	function isDevicePoint(obj: unknown): obj is DevicePoint {
		return typeof obj === 'object' && obj !== null;
	}

	async function fetchDeviceSettings(deviceId: string) {
		console.log('Fetching device settings', deviceId);
		
		if (!deviceId) return; // Handle null deviceId

		// If settings are already loaded or loading, return early
		if (deviceSettingsLoaded.value || deviceSettingsLoading.value) return;

		deviceSettingsLoading.value = true;
		deviceSettingsLoaded.value = false;

		try {
			console.log('Fetching device settings');
			const settings = await apiService.getDeviceSettings(deviceId);
			console.log('fetched');
			deviceSettings.value = settings; // Directly set the settings object
			console.log(deviceSettings.value);
			deviceSettingsLoaded.value = true;
		} catch (error) {
			console.error(`Failed to load settings for ${deviceId}: `, error);
			deviceSettings.value = null; // Reset settings if there is an error to prevent stale settings showing in edit page.

		} finally {
			deviceSettingsLoading.value = false;
		}
	}

	async function saveDeviceSettings(settings: DeviceSettings) {
		if (!settings.device_id) return;
	
		deviceSettingsLoading.value = true;
	
		try {
			const updatedSettings = await apiService.saveDeviceSettings(settings);
			deviceSettings.value = updatedSettings;
	
			Notify.create({
				type: 'positive',
				message: 'Settings saved successfully',
				position: 'top'
			});
	
		} catch (error) {
			console.error(`Failed to update settings for ${settings?.device_id}: `, error);
	
			// Simplified error handling
			Notify.create({
				type: 'negative',
				message: error instanceof Error ? error.message : 'An unexpected error occurred',
				caption: 'Please try again later',
				position: 'top',
				timeout: 5000
			});
		
		} finally {
			deviceSettingsLoading.value = false;
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

			// Trigger reactivity by replacing the devices array.
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

	async function updateIcon(deviceId: string, iconFile: File | null, defaultIcon: string | null) {
		if (!deviceId || !deviceSettings.value) {
			console.warn('Cannot update icon:', !deviceId ? 'deviceId is null' : 'deviceSettings not found in store');
			return;
		}
		
		try {
			let response;
			if (iconFile) {
				response = await apiService.uploadDeviceIcon(deviceId, iconFile, null);
			} else if (defaultIcon) {
				response = await apiService.uploadDeviceIcon(deviceId, null, defaultIcon);
			} else {
				response = await apiService.uploadDeviceIcon(deviceId, null, null);
			}
			deviceSettingsLoaded.value = false; //refetch the device
			// Update the store with the new settings
			await fetchDeviceSettings(deviceId);
			deviceSettings.value.iconUrl = response.iconUrl; //Set new iconURL
			deviceSettings.value.version = response.version; //Update version of settings

			// After successful update and after updating store's deviceSettings, trigger polling.
			// Trigger polling by calling the poll function directly.
			startPolling(); // Trigger poll after icon update


			// Notify success
			Notify.create({
				type: 'positive',
				message: 'Icon updated successfully'
			});
		} catch (error) {
			console.error('Failed to update icon:', error);
			Notify.create({
				type: 'negative',
				message: 'Failed to update icon'
			});
		}
	}
	
	return {
		//states
		deviceLoaded, deviceLoading, devices, selectedDeviceId,
		hoveredDeviceId, editingDevice, mapReady, mapIconVisibility,
		pollingInterval, lastUpdate, deviceSettings, deviceSettingsLoaded, deviceSettingsLoading,

		//actions
		mergeUpdatedDevices, startPolling, loadDevices, selectDevice, 
		deselectDevice, setMapReady,
		setHoveredDevice, toggleDeviceVisibility, setMapIconVisibility,
		setEditingDevice, setEditingDeviceByID, updateDeviceProperty,
		updateIcon, fetchDeviceSettings, saveDeviceSettings, stopPolling
	};
});
