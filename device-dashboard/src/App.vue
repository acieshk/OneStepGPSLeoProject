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

// Device interface moved here
export interface Device {
  device_id: string;
  display_name: string;
  online: boolean;
  color?: string; // Optional color attribute
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
}
  
const devices = ref<Device[]>([]);
const selectedDevice = ref<Device | null>(null);
const deviceColors = ref<{ [key: string]: string }>({});
const deviceVisibility = ref<{ [key: string]: boolean }>({});
const showDialog = ref(false);
const isRefreshing = ref(false);
  
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
      color: assignDeviceColor(device, index)
    }));
    
    // Initialize visibility and colors for new devices
    devices.value.forEach(device => {
      if (deviceVisibility.value[device.device_id] === undefined) {
        deviceVisibility.value[device.device_id] = true;
      }
      if (!deviceColors.value[device.device_id]) {
        deviceColors.value[device.device_id] = device.color || '#1976D2';
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