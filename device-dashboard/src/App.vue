<template>
	<div class="app-container">
		<header class="dashboard-header">
			<h1>Device Dashboard</h1>
			<div class="header-buttons">
				<router-link to="/some-route">
					<el-button type="primary" @click="showRefreshDialog" :disabled="isRefreshing">
						<span v-if="isRefreshing">Refreshing...</span>
						<span v-else>Refresh Data</span>
					</el-button>
				</router-link>
				<router-link to="/user-preferences">
					<el-button type="primary">Settings</el-button>
				</router-link>
			</div>
		</header>

		<router-view /> 

		<el-dialog v-model="showDialog" title="Refresh Database" width="30%">
		</el-dialog>
	</div>
</template>


<script setup lang="ts">
import { ref, reactive, watch, onMounted, computed } from 'vue';
import { useUserStore } from '@/stores/userStore'; // Import the user store
import { apiService } from '@/services/api.service';
import DeviceList from './components/DeviceList.vue';
import DeviceMap from './components/DeviceMap.vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Refresh, Setting, Loading } from '@element-plus/icons-vue';
import { useRouter } from 'vue-router';
import { RouterLink, RouterView } from 'vue-router';

// Color palette
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

const devices = ref<Device[]>([]);
const selectedDevice = ref<Device | null>(null);
const deviceColors = ref<{ [key: string]: string }>({});
const deviceVisibility = ref<{ [key: string]: boolean }>({});
const showDialog = ref(false);
const isRefreshing = ref(false);
const showRefreshSuccessDialog = ref(false);
const router = useRouter();
const userStore = useUserStore(); // Initialize the Pinia store

const layout = ref(userStore.userPreferences.layout); // Now get layout from userStore

// Watch layout changes in Pinia and update locally
watch(() => userStore.userPreferences.layout, (newLayout) => {
	layout.value = newLayout;
});
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

const goToSettings = () => {
	router.push('/user-preferences'); // Navigates to the settings page
};

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

		devices.value = response.result_list.map((device: Device, index: number) => {
			const newDevice = {
				...device,
				color: assignDeviceColor(device, index),
				visible: true,
			};

			// Convert date strings to Date objects
			if (newDevice.latest_device_point?.dt_tracker) {
				newDevice.latest_device_point.dt_tracker = new Date(newDevice.latest_device_point.dt_tracker);
			}


			// Iterate through all keys to find and convert any date strings.
			for (const key in device) {
				const value = device[key];
				if (typeof value === 'string') {
					const parsedDate = new Date(value);
					if (!isNaN(parsedDate.getTime())) {  // Check if it's a valid date
						newDevice[key] = parsedDate;  // Update with Date object if possible
					}
				}

			}

			return newDevice;
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

onMounted(async () => {
	try {
		await fetchDevices();
		const storedPreferences = await apiService.getUserPreferences("default"); // Fetch preferences first
		userStore.updateUserPreferences(storedPreferences); // Initialize Pinia store with fetched preferences
		router.push('/devices') 
	} catch (error) {
		// Display error
		ElMessage.error('Failed to fetch preferences. Please try again.');
		console.error('Error fetching preferences:', error);
	}
});

</script>

<style>
* {
	margin: 0;
	padding: 0;
	box-sizing: border-box;
}

html,
body {
	height: 100%;
	margin: 0;
	padding: 0;

}

.app-container {
	display: flex;
	flex-direction: column;
	height: 100%;
	overflow: auto;
}
</style>
<style scoped>
.app-container {
	display: flex;
	flex-direction: column;
	height: 100vh;
}

.header-buttons {
	display: flex;
	align-items: center;
	gap: 10px;
}


.dashboard-header button {

	height: fit-content;


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

.el-button.is-circle .el-icon {
	/* Target circular buttons specifically */
	display: flex;
	/* Use flexbox for centering */
	justify-content: center;
	/* Center horizontally */
	align-items: center;
	/* Center vertically */
	width: 100%;
	/* Icon takes full width of button */
	height: 100%;
	/* Icon takes full height of button */
}

@media (max-width: 768px) {
	.app-content {
		flex-direction: column;


	}
}
</style>