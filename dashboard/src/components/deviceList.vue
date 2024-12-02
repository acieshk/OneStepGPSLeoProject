<template>
	<q-table title="Devices" :rows="devices" :columns="columns" v-model:pagination.sync="pagination"
		:rows-per-page-options="[10, 25, 50, 100, 0]" row-key="_id" @row-click="handleRowClick"
		@mouseover="handleRowMouseover" @mouseout="handleRowMouseout">
		<template v-slot:header-cell-visible="props">
			<q-th :props="props">
				<q-btn flat dense icon="visibility" :class="{ 'visibility-on': allVisible }"
					@click="toggleAllVisibility">
					{{ `Visible (${visibleDevices}/${totalDevices})` }}
				</q-btn>
			</q-th>
		</template>
		<template v-slot:body-cell-actions="props">
			<q-td :props="props" class="action-cell">
				<q-btn flat round dense icon="edit" @click.stop="goToEditDevice(props.row)" />
			</q-td>
		</template>
		<template v-slot:body-cell-visible="props">
			<q-td :props="props" class="visibility-cell">
				<q-btn flat round dense :icon="props.row.visible ? 'visibility' : 'visibility_off'"
					@click.stop="handleVisibilityToggle(props.row)" />
			</q-td>
		</template>
		<template v-slot:body-cell-iconUrl="props">
			<q-td :props="props">
				<q-img :src="formatURL(props.row.iconUrl)" style="max-width: 50px; max-height:50px" />
			</q-td>
		</template>
		<template v-slot:body-cell-device="props"> 
			<q-td :props="props">
				<div class="name-status">
					<span :class="['status-indicator', props.row.online ? 'online' : 'offline']"></span>
					<span class="device-name">{{ props.row.display_name }}</span>
				</div>
				<div class="device-id-container">
					<span class="device-id">{{ props.row.device_id }}</span>
				</div>
			</q-td>
		</template>
	</q-table>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useDeviceStore } from 'src/stores/deviceStore';
import { useUserStore } from 'src/stores/userStore';
import { storeToRefs } from 'pinia';
import { Device } from 'src/model/model';
import { QTd } from 'quasar';
import { useRouter } from 'vue-router';



const deviceStore = useDeviceStore();
const { deviceLoaded, devices, selectedDeviceId } = storeToRefs(deviceStore);
const userStore = useUserStore();
const { userPreferences } = storeToRefs(userStore);
const selectedRow = ref<Device | null>(null);
const pagination = ref({
	page: 1,
	rowsPerPage: userPreferences.value.rowPerPage, 
	rowPerPageOptions: [10, 25, 50, 100, 0]
});



const columns = computed(() => {
	if (!devices.value || devices.value.length === 0) {
		return []; 
	}

	const visibleDevices = deviceStore.devices.filter((device: Device) => device.visible).length;
	const totalDevices = deviceStore.devices.length;

	return [
		{
			name: 'actions',
			label: 'Actions',
			field: 'actions',
			align: 'center' as 'left' | 'right' | 'center'
		},
		{
			name: 'visible',
			label: `Visible (${visibleDevices}/${totalDevices})`,
			field: 'visible',
			align: 'center' as 'left' | 'right' | 'center',
			sortable: false,
		},
		{
			name: 'iconUrl',
			label: 'icon',
			field: (row: Device) => formatURL(row.iconUrl), 
			align: 'center' as 'left' | 'right' | 'center',
			sortable: false // No sorting needed for this icons
		},
		{
			name: 'device', 
			label: 'Device', 
			field: 'device',  // Pass the whole row so both name and id can be used for formatting
			align: 'left' as 'left' | 'right' | 'center',
			sortable: true // Sorting will be based on the formatted value
		},
		{
			name: 'fuel',
			label: 'Fuel (%)',
			field: (row: Device) => row.latest_device_point?.device_state.fuel_percent?.toFixed(2),
			align: 'right' as 'left' | 'right' | 'center',
			sortable: true,
		},
		{
			name: 'odometer', 
			label: 'odometer', 
			field: (row: Device) => row.latest_device_point?.device_state.odometer,
			format: (val: {
				value: number,
				unit: string
			}) => {
				return formatOdometer(val); // Format value as a string with 2 decimal places and the unit appended
			},
			align: 'right' as 'left' | 'right' | 'center',
			sortable: true,
			sort: (a: { value: number; unit: string } | undefined, b: { value: number; unit: string } | undefined) => {
				// Convert to metric before sorting
				const valueA = convertToMetric(a?.value ?? null, a?.unit ?? null) ?? 0; // Default to 0 if null/undefined after conversion
				const valueB = convertToMetric(b?.value ?? null, b?.unit ?? null) ?? 0;
				return valueA - valueB;
			},
		},
		{
			name: 'location',
			label: 'Location',
			field: (row: Device) => row.latest_device_point, // Access latest_device_point
			format: (val: {
				lat: string,
				lng: string,
			}) => {
				const lat = Number(val.lat) || 0; // Convert to number
				const lng = Number(val.lng) || 0; // Convert to number
				return isNaN(lat) || isNaN(lng) ? 'N/A' : `[${lat.toFixed(2)}, ${lng.toFixed(2)}]`; // Check result
			},
			align: 'left' as 'left' | 'right' | 'center',
			sortable: false, // Can't really sort by location
		}
	];
});

const allVisible = computed(() => devices.value.every(device => device.visible));
const visibleDevices = computed(() => devices.value.filter(device => device.visible).length);
const totalDevices = computed(() => devices.value.length);

const toggleAllVisibility = () => {
	const newValue = !allVisible.value;
	devices.value.forEach(device => {
		deviceStore.setMapIconVisibility(device._id, newValue);
	});
};
/*
	Row Functionality
*/
const handleRowClick = (_evt: Event, row: Device) => {
	const _id = row._id
	if (deviceStore.selectedDeviceId == null) {
		deviceStore.selectDevice(_id);
	}
	else if (deviceStore.selectedDeviceId !== _id) {
		deviceStore.selectDevice(_id);
	}
	else if (deviceStore.selectedDeviceId == _id) {
		deviceStore.deselectDevice();
	}
};
const handleRowMouseover = (evt: Event, row: Device | undefined) => { // Indicate that row could be undefined
	if (row) {  // Check if row is defined
		deviceStore.setHoveredDevice(row._id);
	}
};
const handleRowMouseout = () => {
	deviceStore.setHoveredDevice(null);
}

const formatOdometer = (odometerData: { value: number, unit: string }) => {
	if (!odometerData) return 'N/A'; // Handle missing data

	let { value, unit } = odometerData;
	if (typeof value !== 'number') {
		value = parseFloat(value) // Convert to number if needed. If invalid number, handle as needed
		if (isNaN(value)) 
			return 'N/A'
	}

	switch (userPreferences.value.unit) { // Access user preference reactively
		case 'imperial':
			if (unit === 'km' || unit === 'm') {
				value = convertToMiles(value, unit); // Convert to miles
				unit = 'mi';
			}
			break; // No conversion needed if already miles
		case 'metric':
			if (unit === 'mi') {
				value = convertToKilometers(value);
				unit = 'km';
			} else if (unit === 'm') {
				value = value / 1000; // Convert meters to kilometers
				unit = 'km';
			}
			break; // No conversion needed if already kilometers
		case 'original': // No conversion needed
		    break;  
		default: // Handle invalid or unknown units from user preferences
			console.warn(`Unknown unit preference: ${userPreferences.value.unit}`);
	}

	return `${value.toFixed(2)} ${unit}`; // Format with 2 decimal places
};

const convertToMiles = (value: number, unit: string) => { // Keep other conversion functions if needed
	if (unit === 'km') {
		return value * 0.621371;
	} else if (unit === 'm') {
		return value * 0.000621371;
	}
	return value; // Return original value if unit is not km or m
};

const convertToKilometers = (value: number) => { // Keep other conversion functions if needed
	return value * 1.60934;
};

const convertToMetric = (value: number | null | undefined, unit: string | null | undefined): number | null => {
	if (value == null || unit == null) return null; // Handle missing data

	if (typeof value !== 'number') {
		value = parseFloat(value as string);
		if (isNaN(value)) {
			return null; // Return null or handle invalid values
		}
	}

	if (typeof unit !== 'string') { // If unit is invalid, return null, or handle as needed
		return null;
	}

	switch (unit) {
		case 'mi':
			return convertToKilometers(value);
		case 'm':
			return value / 1000; // Convert meters to kilometers
		case 'km':
			return value; // Already in kilometers
		default:
			console.warn(`Unknown unit: ${unit}`);
			return null; // Or handle the unknown unit differently
	}
};
/* 
	  Edit Button functionality
 */
const router = useRouter();
const goToEditDevice = (device: Device) => {
	deviceStore.setEditingDevice(device);
	router.push(`/devices/${device._id}`); // Or your edit route
};

/*
	Visibility functionality
*/
const handleVisibilityToggle = (device: Device) => {
	// Update the device's visibility in the store and Map:
	console.log('Update Visibility');
	deviceStore.toggleDeviceVisibility(device._id);	//update visibility Map for the icons
};


watch(deviceLoaded, () => {
	if (deviceLoaded)
		console.log('Devices loaded:', devices.value);
}, { immediate: true });

// Watch selectedDeviceId from the store
watch(selectedDeviceId, (newSelectedDeviceId) => {
	if (newSelectedDeviceId) {
		selectedRow.value = devices.value.find((device: Device) => device._id === newSelectedDeviceId) || null;
	} else {
		selectedRow.value = null;
	}
});
/*
	icon
*/

const DEFAULT_ICON_URL = 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-blue.png';
const formatURL = (url: string | null | undefined) => { 
	if (!url) return DEFAULT_ICON_URL; // Return default if null, undefined, or empty

	if (typeof url !== 'string') {
		console.warn(`formatURL received a non-string value: ${typeof url}, ${url}`); // Log for debugging
		return DEFAULT_ICON_URL;
	}

	if (url.startsWith('http://') || url.startsWith('https://')) {
		return url;
	}
	return `http://${url}`; // Prepend http:// if no protocol
};
</script>
<style scoped>
.q-table__tr--selected {
	background-color: rgba(0, 0, 0, 0.1) !important;
}

/* Hovered row style */
.q-table__tr--highlighted {
	background-color: rgba(0, 0, 0, 0.05);
}

.action-cell {
	text-align: center;
	cursor: pointer;
}

visibility-cell {
	cursor: pointer;
}

.device-name {
	font-weight: 500;
	color: #2c3e50;
}

.device-id {
	font-size: 0.8rem;
	color: #999;
	margin-top: 4px;
}

.name-status-container {
	display: flex;
	align-items: center;
}

.device-id-container {
	margin-top: 4px;
}

.status-indicator {
	width: 10px;
	height: 10px;
	border-radius: 50%;
	display: inline-block;
	margin-right: 5px;
}

.status-indicator.online {
	background-color: #4caf50;
}

.status-indicator.offline {
	background-color: #f44336;
}

.visibility-on {
	color: green;
}
</style>