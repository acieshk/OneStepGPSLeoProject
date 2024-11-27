<template>
	<div class="device-list-container">
		<div class="search-container">
			<el-input v-model="searchTerm" placeholder="Search devices Name/ID" clearable @input="filterDevices">
				<template #prefix>
					<el-icon>
						<Search />
					</el-icon>
				</template>
			</el-input>
		</div>
		<div class="table-wrapper">
			<el-table :data="filteredDevices" style="width: 100%" height="100%" :highlight-current-row="true"
				@current-change="handleCurrentChange" @row-click="handleRowClick">
				<el-table-column label="Visibility" width="100">
					<template #header>
						<div class="visibility-icon-container" @click="toggleAllVisibility">
							<i v-if="anyVisible" class="fas fa-eye"></i>
							<i v-else class="fas fa-eye-slash"></i>
							<div style="margin-left: 5px; vertical-align: middle;">{{ visibleDevicesCount }} / {{
								totalDevicesCount }}</div>
						</div>
					</template>
					<template #default="scope">
						<div class="visibility-icon-container" @click="toggleDeviceVisibility(scope.row)">
							<i v-if="deviceVisibility[scope.row.device_id]" class="fas fa-eye"></i>
							<i v-else class="fas fa-eye-slash"></i>
						</div>
					</template>
				</el-table-column>
				<el-table-column label="Actions" width="120"> <!-- New Column -->
					<template #default="scope">
						<el-button size="small" @click="openEditDialog(scope.row)">Edit</el-button>
					</template>
				</el-table-column>
				<el-table-column label="Name" min-width="180">
					<template #default="scope">
						<div class="name-status-container">
							<div class="map-icon-cell" @click.stop="centerMapOnDevice(scope.row)">
								<MapMarkerIcon :deviceId="scope.row.device_id" :color="scope.row.color" />
							</div>
							<div>
								<div class="name-status">
									<span :class="['status-indicator', scope.row.online ? 'online' : 'offline']"></span>
									<span class="device-name">{{ scope.row.display_name }}</span>
								</div>
								<div class="device-id-container">
									<span class="device-id">{{ scope.row.device_id }}</span>
								</div>
							</div>
						</div>
					</template>
				</el-table-column>
				<el-table-column label="Location" min-width="180">
					<template #default="scope">
						{{ formatLocation(
							scope.row.latest_device_point?.lat,
							scope.row.latest_device_point?.lng
						) }}
					</template>
				</el-table-column>

				<el-table-column label="Speed" width="120">
					<template #default="scope">
						{{ formatSpeed(scope.row.latest_device_point?.speed) }}
					</template>
				</el-table-column>

				<el-table-column label="Battery" width="120">
					<template #default="scope">
						{{ formatBattery(scope.row.latest_device_point?.device_point_detail?.external_volt) }}
					</template>
				</el-table-column>

				<el-table-column label="Fuel" width="120">
					<template #default="scope">
						{{ formatFuel(scope.row.latest_device_point?.device_state?.fuel_percent) }}
					</template>
				</el-table-column>

				<el-table-column label="Odometer" min-width="150">
					<template #default="scope">
						{{ formatOdometer(scope.row.latest_device_point?.device_state?.odometer) }}
					</template>
				</el-table-column>

				<el-table-column label="Last Updated" min-width="180">
					<template #default="scope">
						{{ formatDateTime(scope.row.latest_device_point?.dt_tracker) }}
					</template>
				</el-table-column>
			</el-table>
		</div>
		<device-edit-dialog v-if="deviceToEdit" v-model="editDialogVisible" :device-to-edit="deviceToEdit" @device-updated="handleDeviceUpdated"></device-edit-dialog>
	</div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch, toRefs } from 'vue';
import { Search } from '@element-plus/icons-vue';

import MapMarkerIcon from '@/components/MapMarkerIcon.vue';
import { convertUnits } from '@/utils/UnitConverter';
import DeviceEditDialog from './DeviceEditDialog.vue';
import { ElMessage } from 'element-plus';
import { useUserStore } from '@/stores/userStore'; // Import your user store
import { storeToRefs } from 'pinia';
import type { Device } from '@/types/device';



const props = defineProps<{
	devices: Device[];
	selectedDevice: Device | null;
	userPreferences: { distanceUnit: string, speedUnit: string };
}>();


const userStore = useUserStore();
const { devices, selectedDevice, userPreferences } = toRefs(props);

const emit = defineEmits<{
	(e: 'select-device', device: Device): void;
	(e: 'update-device-color', deviceId: string, color: string): void;
	(e: 'update-device-visibility', deviceId: string, visible: boolean): void;
}>();

// Search functionality
const searchTerm = ref('');
const filteredDevices = computed(() => {
	if (!searchTerm.value) return props.devices;

	const searchLower = searchTerm.value.toLowerCase();
	return props.devices.filter(device =>
		device.display_name.toLowerCase().includes(searchLower) ||
		device.device_id.toLowerCase().includes(searchLower)
	);
});

const filterDevices = () => {
	// Additional filtering logic if needed
};

// State for color picker and visibility
const deviceColors = ref<{ [key: string]: string }>({});
const deviceVisibility = ref<{ [key: string]: boolean }>({});

const toggleDeviceVisibility = (device: Device) => {
	updateDeviceVisibility(device, !deviceVisibility.value[device.device_id]);
};

const updateDeviceVisibility = (device: Device, visible: boolean) => {
	deviceVisibility.value[device.device_id] = visible;
	emit('update-device-visibility', device.device_id, visible);
};

const toggleAllVisibility = () => {
	const newVisibility = !allVisible.value;
	props.devices.forEach(device => {
		updateDeviceVisibility(device, newVisibility);
	});
};

const allVisible = computed(() => {
	return props.devices.every(device => deviceVisibility.value[device.device_id]);
});

const anyVisible = computed(() => {
	return props.devices.some(device => deviceVisibility.value[device.device_id]);
});

const noneVisible = computed(() => {
	return props.devices.every(device => !deviceVisibility.value[device.device_id]);
});

const totalDevicesCount = computed(() => {
	return props.devices.length;
});

const visibleDevicesCount = computed(() => {
	return props.devices.filter(device => deviceVisibility.value[device.device_id]).length;
});

const indeterminateState = computed(() =>
	!allVisible.value && !noneVisible.value
);


// Device color logic
const getDeviceColor = (deviceId: string) => {
	return props.devices.find(d => d.device_id === deviceId)?.color || '#1976D2';
};

const centerMapOnDevice = (device: Device) => {
	emit('center-map-on-device', device);
};

// Initialize visibility when devices prop changes
watch(
	() => props.devices,
	(newDevices) => {
		const initialVisibility: { [key: string]: boolean } = {};
		newDevices.forEach((device) => {
			initialVisibility[device.device_id] = true; // Default to visible
		});
		deviceVisibility.value = initialVisibility;

	},
	{ immediate: true }
);

onMounted(() => {
	// props.devices.forEach(device => {
	// });
});

// Methods
const handleCurrentChange = (currentRow: Device | null) => {
	if (currentRow && deviceVisibility.value[currentRow.device_id]) { // Check visibility
		emit('select-device', currentRow);
	}
};

const handleRowClick = (device: Device) => {

	if (deviceVisibility.value[device.device_id]) {
		emit('select-device', device);
	}
};


const updateDeviceColor = (device: Device, color: string) => {
	deviceColors.value[device.device_id] = color;
	emit('update-device-color', device.device_id, color);
};

const formatLocation = (lat?: number, lng?: number): string => {
	if (!lat || !lng) return 'N/A';
	return `${lat.toFixed(6)}, ${lng.toFixed(6)}`;
};

const formatBattery = (voltage?: number): string => {
	if (!voltage) return 'N/A';
	return `${voltage.toFixed(2)}V`;
};

const formatFuel = (percent?: number): string => {
	if (percent === undefined || percent === null) return 'N/A';
	return `${percent.toFixed(1)}%`;
};

const formatOdometer = (odometer: any): string => {
    if (!odometer || typeof odometer.value !== 'number' || typeof odometer.unit !== 'string') return 'N/A';

    try {
        const convertedOdometer = convertUnits(odometer.value, odometer.unit, userPreferences.value.distanceUnit);
        const displayedUnit = userPreferences.value.distanceUnit === "km" ? "km" : "mi"; // Use store for unit
        return `${convertedOdometer.toFixed(1)} ${displayedUnit}`;
    } catch (error) {
        console.error("Odometer unit conversion error:", error);
        return "N/A";
    }
};
const formatSpeed = computed(() => (device: Device) => {
    if (!device || !device.latest_device_point?.speed) {
		return "0";
	}
    try {
		const speedKmh = device.latest_device_point.speed * 3.6;
        const unit = userPreferences.value.distanceUnit === 'mi' ? 'mph' : 'km/h';
		const convertedSpeed = convertUnits(speedKmh, 'km/h', unit); // Use convertUnits
		return parseFloat(convertedSpeed.toFixed(1)) + ' ' + unit;
    } catch (error) {
        console.error("Unit conversion error:", error);
        return "N/A"; // Or handle the error as needed
    }
});

const formatDateTime = (dateString?: string): string => {
	if (!dateString) return 'N/A';
	return new Date(dateString).toLocaleString();
};

const editDialogVisible = ref(false);
const deviceToEdit = ref<Device | null>(null); // Store device to edit


const openEditDialog = (device: Device) => {
	deviceToEdit.value = device;
	editDialogVisible.value = true;
};

const handleDeviceUpdated = (updatedDevice: Device) => {
	// Find the index of the updated device in the devices array
	const index = props.devices.findIndex(d => d.device_id === updatedDevice.device_id);

	if (index !== -1) {
		// Update the device in the array
		props.devices.splice(index, 1, updatedDevice);
	}
	ElMessage.success("Device information updated")
};


</script>

<style scoped>
.visibility-icon-container {
	cursor: pointer;
	display: flex;
	justify-content: center;
	align-items: center;
	height: 100%;
	/* Make the cell fill the row height */
}

.device-list-container {
	height: 100%;
	position: relative;
	background: #f5f5f5;
	border-radius: 8px;
	overflow-x: auto;
}

.table-wrapper {
  width: fit-content; 
  min-width: 100%;   
  margin: 0 auto; 
}

.search-container {
	padding: 10px;
	background: white;
	border-bottom: 1px solid #e0e0e0;
}

.name-status-container {
	display: flex;
	align-items: stretch;
	/* Stretch to occupy full height */
}

.map-icon-cell {
	cursor: pointer;
	margin-right: 10px;
	/* Add some space between the icon and name */
	height: 100%;
	/* Occupy full cell height */
	display: flex;
	/* Use flexbox for vertical centering */
	align-items: center;
	/* Vertically center the icon */
}



.map-icon {
	/* Styles for the map icon */
	width: 24px;
	height: 36px;
	/* Maintain aspect ratio */
	display: inline-block;
}

.name-status {
	/* Keep name and status in one line */
	display: flex;
	align-items: center;
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

.status-cell {
	display: flex;
	align-items: center;
	gap: 8px;
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
	/* Adjust size */
	height: 10px;
	/* Adjust size */
	border-radius: 50%;
	display: inline-block;
	margin-right: 5px;
	/* Adjust spacing */
}

.status-indicator.online {
	background-color: #4caf50;
	/* Green */
}

.status-indicator.offline {
	background-color: #f44336;
	/* Red */
}

:deep(.el-table__row.current-row) {
	background-color: #e3f2fd;
}

.map-icon-cell {
	cursor: pointer;
	/* Make the icon clickable */
	display: flex;
	justify-content: center;
	/* Center the icon horizontally */
	align-items: center;
	/* Center the icon vertically */
	height: 100%;
	/* Make the cell fill the row height */
}

.map-icon-cell .map-icon {
	display: inline-block;
	border-radius: 50%;
}
</style>