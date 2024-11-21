<template>
	<div class="app-container">
	  <header class="dashboard-header">
		<h1>Device Dashboard</h1>
		<el-button 
		  type="primary"
		  @click="showRefreshDialog"
		  :loading="isRefreshing"
		>
		  Refresh Database
		</el-button>
	  </header>
	  <div class="app-content">
		<device-list 
		  :devices="devices" 
		  :selectedDevice="selectedDevice"
		  @select-device="handleDeviceSelect"
		  @update-device-color="handleDeviceColorUpdate"
		  @update-device-visibility="handleDeviceVisibilityUpdate"
		/>
		<device-map 
		  :devices="devices"
		  :selectedDevice="selectedDevice"
		  :deviceColors="deviceColors"
		  :deviceVisibility="deviceVisibility"
		  @select-device="handleDeviceSelect"
		/>
	  </div>
  
	  <el-dialog
		v-model="showDialog"
		title="Refresh Database"
		width="30%"
	  >
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
  import { ref, onMounted } from 'vue';
  import { apiService } from '@/services/api.service';
  import DeviceList from './components/DeviceList.vue';
  import DeviceMap from './components/DeviceMap.vue';
  
  const devices = ref([]);
  const selectedDevice = ref(null);
  const deviceColors = ref({});
  const deviceVisibility = ref({});
  const showDialog = ref(false);
  const isRefreshing = ref(false);
  
  const fetchDevices = async () => {
	try {
	  const response = await apiService.getDevices();
	  devices.value = response.result_list;
	  
	  // Initialize visibility and colors for new devices
	  devices.value.forEach(device => {
		if (deviceVisibility.value[device.device_id] === undefined) {
		  deviceVisibility.value[device.device_id] = true;
		}
		if (!deviceColors.value[device.device_id]) {
		  deviceColors.value[device.device_id] = '#1976D2';
		}
	  });
	} catch (error) {
	  console.error('Error fetching devices:', error);
	}
  };
  
  const showRefreshDialog = () => {
	showDialog.value = true;
  };
  
  const handleRefreshDB = async () => {
	isRefreshing.value = true;
	try {
	  await apiService.refreshDatabase();
	  await fetchDevices();
	  showDialog.value = false;
	} catch (error) {
	  console.error('Error refreshing database:', error);
	} finally {
	  isRefreshing.value = false;
	}
  };
  
  const handleDeviceSelect = (device) => {
	selectedDevice.value = device;
  };
  
  const handleDeviceColorUpdate = (deviceId: string, color: string) => {
	deviceColors.value[deviceId] = color;
  };
  
  const handleDeviceVisibilityUpdate = (deviceId: string, visible: boolean) => {
	deviceVisibility.value[deviceId] = visible;
  };
  
  onMounted(() => {
	fetchDevices();
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
	min-height: 0; /* Important for proper scrolling */
  }
  
  @media (max-width: 768px) {
	.app-content {
	  grid-template-columns: 1fr;
	}
  }
  </style>