<template>
	<div class="app-container">
		<header class="dashboard-header">
			<h1>Device Dashboard</h1>
			<div>
				<el-tooltip content="Refresh Database" placement="bottom">
					<el-button type="primary" :icon="Refresh" @click="showRefreshDialog" circle
						:loading="isRefreshing" />
				</el-tooltip>
				<el-popover
					ref="settingsPopover"
					placement="bottom"
					:width="300"
					trigger="click"
					v-model="settingsVisible"
				>
					<template #reference>
						<el-button type="primary" :icon="Setting" circle @click="settingsVisible = !settingsVisible">
							<el-tooltip content="Setting" placement="bottom" />
						</el-button>
					</template>
					<div class="settings-content"> <div><span>Distance Unit:</span> <el-radio-group
								v-model="userPreferences.distanceUnit">
								<el-radio label="km">Kilometers</el-radio>
								<el-radio label="mi">Miles</el-radio>
							</el-radio-group>
						</div><div class="layout-setting"><span>Layout:</span> <el-radio-group v-model="layout">
								<el-radio label="horizontal">Horizontal</el-radio>
								<el-radio label="vertical">Vertical</el-radio>
							</el-radio-group>
						</div></div>
				</el-popover>
			</div>
		</header>
		<div :class="['app-content', layout]">
			<device-list :devices="devices" :selectedDevice="selectedDevice" :userPreferences="userPreferences"
				@select-device="handleDeviceSelect" @update-device-color="handleDeviceColorUpdate"
				@update-device-visibility="handleDeviceVisibilityUpdate" />
			<device-map :devices="devices" :selectedDevice="selectedDevice" :deviceColors="deviceColors"
				:deviceVisibility="deviceVisibility" @select-device="handleDeviceSelect" />
		</div>

		<el-dialog v-model="showDialog" title="Refresh Database" width="30%">
			<span>Do you want to update the DB?</span>
			<template #footer>
				<span class="dialog-footer">
					<el-button @click="showDialog = false">Cancel</el-button>
					<el-button type="primary" @click="handleRefreshDB" :loading="isRefreshing">
						Confirm
					</el-button>
				</span>
			</template>
		</el-dialog>
	</div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, onMounted, computed } from 'vue';
import { apiService } from '@/services/api.service';
import DeviceList from './components/DeviceList.vue';
import DeviceMap from './components/DeviceMap.vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Refresh, Setting, Loading } from '@element-plus/icons-vue'; // Import icons

// Professional color palette
const COLOR_PALETTE = [
	'#1F77B4', // Muted Blue
	'#FF7F0E', // Vivid Orange
	'#2CA02C', // Fresh Green
	'#D62728', // Brick Red
	'#9467BD', // Soft Purple
	'#8C564B', // Brown
	'#E377C2', // Pink
	'#7F7F7F', // Gray
	'#BCBD22', // Olive Green
	'#17BECF', // Teal
	'#3498DB', // Bright Blue
	'#2ECC71', // Emerald Green
	'#E74C3C', // Coral Red
	'#9B59B6', // Lavender
	'#34495E', // Dark Slate Blue
	'#16A085', // Sea Green
	'#F39C12', // Sunflower Yellow
	'#2980B9', // Deep Blue
	'#8E44AD', // Deep Purple
	'#2C3E50'  // Navy Blue
];

export interface Device {
	device_id: string;
	display_name: string;
	online: boolean;
	color?: string;
	visible?: boolean;
	latest_device_point?: {
		lat: number;
		lng: number;
		speed: number;
		dt_tracker: string;
		device_point_detail?: {
			external_volt: number;
		};
		device_state?: {
			fuel_percent: number;
			odometer: {
				value: number;
				unit: string;
			};
		};
	};
	[key: string]: any;
}

const devices = ref<Device[]>([]);
const selectedDevice = ref<Device | null>(null);
const deviceColors = ref<{ [key: string]: string }>({});
const deviceVisibility = ref<{ [key: string]: boolean }>({});
const showDialog = ref(false);
const isRefreshing = ref(false);
const showRefreshSuccessDialog = ref(false);
const layout = ref(localStorage.getItem('layout') || 'horizontal');

// Watch layout and store in localStorage
// Watch layout and store in localStorage
watch(layout, (newVal) => {
	localStorage.setItem('layout', newVal);
});
// Corrected userPreferences (speedUnit is now a ref)
const userPreferences = reactive({
	distanceUnit: localStorage.getItem('distanceUnit') || 'km',
	speedUnit: localStorage.getItem('speedUnit') || 'km/h'
});

const showRefreshDialog = () => {
	showDialog.value = true;
};

const handleRefreshDB = async () => {
	isRefreshing.value = true;
	showDialog.value = false;

	try {
		const response = await apiService.refreshDatabase();

		// Show success dialog
		const confirmUpdate = await ElMessageBox.confirm(
			'Database refreshed successfully. Do you want to update the device list?',
			'Refresh Successful',
			{
				confirmButtonText: 'Yes',
				cancelButtonText: 'No',
				type: 'success'
			}
		);

		if (confirmUpdate) {
			// Clear existing devices before fetching
			devices.value = [];
			await fetchDevices();

			ElMessage({
				type: 'success',
				message: 'Device list updated successfully!'
			});
		}
	} catch (error) {
		// Show error dialog
		ElMessageBox.alert(
			'Failed to refresh database. Please try again.',
			'Refresh Error',
			{
				type: 'error'
			}
		);
		console.error('Error refreshing database:', error);
	} finally {
		isRefreshing.value = false;
	}
};

// Function to assign a color to a device
const assignDeviceColor = (device: Device, index: number): string => {
	// If device already has a color, return it
	if (device.color) return device.color;

	// Assign color from the predefined palette
	return COLOR_PALETTE[index % COLOR_PALETTE.length];
};

const fetchDevices = async () => {
	try {
		const response = await apiService.getDevices();
		devices.value = response.result_list.map((device: Device, index: number) => ({
			...device,
			color: assignDeviceColor(device, index),
			visible: true, // Default visibility to true
		}));

		// Update device colors
		devices.value.forEach(device => {
			deviceColors.value[device.device_id] = device.color || '#1976D2';
		});
	} catch (error) {
		console.error('Error fetching devices:', error);
	}
};

const handleDeviceVisibilityUpdate = (deviceId: string, visible: boolean) => {
	// Find the device and update its visibility
	const deviceIndex = devices.value.findIndex(d => d.device_id === deviceId);
	if (deviceIndex !== -1) {
		devices.value[deviceIndex].visible = visible;
	}
};

const handleDeviceSelect = (device: Device) => {
	selectedDevice.value = device;
};

const handleDeviceColorUpdate = (deviceId: string, color: string) => {
	deviceColors.value[deviceId] = color;

	// Update the color in the devices array
	const deviceIndex = devices.value.findIndex(d => d.device_id === deviceId);
	if (deviceIndex !== -1) {
		devices.value[deviceIndex].color = color;
	}
};
const settingsVisible = ref(false); // Control settings popover visibility

// Watch distanceUnit to update speedUnit 
watch(() => userPreferences.distanceUnit, (newVal) => {
	userPreferences.speedUnit = newVal === 'km' ? 'km/h' : 'mph';
	localStorage.setItem('distanceUnit', newVal);
}, { immediate: true });

// Watch speedUnit 
watch(() => userPreferences.speedUnit, (newVal) => {
	localStorage.setItem('speedUnit', newVal);
}, { immediate: true });

onMounted(() => {
	fetchDevices();
	// Retrieve from localStorage (corrected)
	const storedDistanceUnit = localStorage.getItem('distanceUnit');
	if (storedDistanceUnit) {
		userPreferences.distanceUnit = storedDistanceUnit; // No JSON.parse needed
	}
	const storedSpeedUnit = localStorage.getItem('speedUnit');
	if (storedSpeedUnit) {

		userPreferences.speedUnit = storedSpeedUnit;  // No JSON.parse needed

	}
	const storedLayout = localStorage.getItem('layout');
	if (storedLayout) {
		layout.value = storedLayout;
	}
});
</script>

<style>
* {
	margin: 0;
	padding: 0;
	box-sizing: border-box;
}

body {
	margin: 0;
	padding: 0;
}
</style>
<style scoped>
.app-container {
	display: flex;
	flex-direction: column;
	height: 100vh;
	overflow: hidden;
}

.dashboard-header {
	display: flex;
	justify-content: space-between;
	align-items: center;
	background-color: #2c3e50;
	color: white;
	padding: 1rem 2rem;
	box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.dashboard-header h1 {
	margin: 0;
	font-size: 1.8rem;
	font-weight: 600;
}

.app-content {
	display: grid;
	grid-template-columns: 1fr 1fr;
	gap: 20px;
	padding: 20px;
	flex: 1;
	min-height: 0;
	/* Important for proper scrolling */
}

.layout-setting {
	margin-top: 10px;
}

/* Horizontal layout */
.app-content.horizontal {
	/* Added horizontal class */
	display: grid;
	grid-template-columns: 1fr 1fr;
}

/* Vertical layout */
.app-content.vertical {
	display: flex;
	flex-direction: column;
}

.el-button.is-circle .el-icon {  /* Target circular buttons specifically */
  display: flex;   /* Use flexbox for centering */
  justify-content: center; /* Center horizontally */
  align-items: center; /* Center vertically */
  width: 100%;      /* Icon takes full width of button */
  height: 100%;     /* Icon takes full height of button */
}

@media (max-width: 768px) {
	.app-content {
		grid-template-columns: 1fr;
	}
}
</style>