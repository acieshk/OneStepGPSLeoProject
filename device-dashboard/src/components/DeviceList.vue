<template>
	<div class="device-list-container">
	  <div class="search-container">
		<el-input
		  v-model="searchTerm"
		  placeholder="Search devices..."
		  clearable
		  @input="filterDevices"
		>
		  <template #prefix>
			<el-icon><Search /></el-icon>
		  </template>
		</el-input>
	  </div>
	  <el-table
		:data="filteredDevices"
		style="width: 100%"
		height="100%"
		:highlight-current-row="true"
		@current-change="handleCurrentChange"
		@row-click="handleRowClick"  
	  >
		<el-table-column label="Visibility" width="100">
		  <template #header>
			<el-checkbox
			  v-model="allDevicesVisible"
			  :indeterminate="indeterminateState"
			  @change="toggleAllVisibility"
			/>
		  </template>
		  <template #default="scope">
			<el-checkbox
			  v-model="deviceVisibility[scope.row.device_id]"
			  @change="(val) => updateDeviceVisibility(scope.row, val)"
			/>
		  </template>
		</el-table-column>
		<el-table-column label="Map" width="60">
		  <template #default="scope">
			<div 
			  class="map-icon-cell" 
			  @click.stop="centerMapOnDevice(scope.row)"
			>
			<svg 
				class="map-icon" 
				width="48" 
				height="72" 
				viewBox="0 0 24 36" 
				xmlns="http://www.w3.org/2000/svg">
				<path 	
					d="M12 0C5.4 0 0 5.4 0 12c0 7.2 12 24 12 24s12-16.8 12-24c0-6.6-5.4-12-12-12z" 
					:fill="scope.row.color || getDeviceColor(scope.row.device_id)" 
					stroke="white" 
					stroke-width="2"/>
				<circle 
					cx="12" 
					cy="12" 
					r="4" 
					fill="white"/>
				</svg>
			</div>
		  </template>
		</el-table-column>
		<el-table-column label="Name" min-width="180">
        <template #default="scope">
          <div class="device-name">{{ scope.row.display_name }}</div>
          <div class="device-id">{{ scope.row.device_id }}</div>
        </template>
      </el-table-column>
		<el-table-column label="Status" width="120">
		  <template #default="scope">
			<div class="status-cell">
			  <span :class="['status-indicator', scope.row.online ? 'online' : 'offline']"></span>
			  {{ scope.row.online ? 'Online' : 'Offline' }}
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
  </template>
  
  <script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue';
import { Search } from '@element-plus/icons-vue';
import { Device } from '@/App.vue'; 
  
  
const props = defineProps<{
  devices: Device[];
  selectedDevice: Device | null;
}>();

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

// State for all devices visibility
const allDevicesVisible = ref(true);
const indeterminateState = ref(false);

// Computed property to determine the state of the header checkbox
const allVisible = computed(() => {
  return props.devices.every(device => deviceVisibility.value[device.device_id]);
});

const noneVisible = computed(() => {
  return props.devices.every(device => !deviceVisibility.value[device.device_id]);
});

// Device color logic
const getDeviceColor = (deviceId: string) => {
  return props.devices.find(d => d.device_id === deviceId)?.color || '#1976D2';
};

const centerMapOnDevice = (device: Device) => {
	emit('center-map-on-device', device);
};

const handleRowClick = (device: Device) => {
	// Add any additional logic here. For now it prevents row click to do anything more
}

// Watch for changes in individual device visibility to update header checkbox
watch([allVisible,noneVisible], () => {
  indeterminateState.value = !allVisible.value && !noneVisible.value;
  allDevicesVisible.value = allVisible.value;

});

// Method to toggle the visibility of all devices
const toggleAllVisibility = (val: boolean) => {
  props.devices.forEach(device => {
    updateDeviceVisibility(device, val);
  });
  //Update deviceMap
};

// Initialize visibility and colors when component mounts
onMounted(() => {
  props.devices.forEach(device => {
    // Retrieve visibility from localStorage, defaulting to true if not found
    const storedVisibility = localStorage.getItem(`deviceVisibility-${device.device_id}`);
    deviceVisibility.value[device.device_id] = storedVisibility ? JSON.parse(storedVisibility) : true;

    // Use device's color or default
    deviceColors.value[device.device_id] = device.color || '#1976D2';
  });
});


  const updateDeviceVisibility = (device: Device, visible: boolean) => {
	deviceVisibility.value[device.device_id] = visible;
	emit('update-device-visibility', device.device_id, visible);
  };
  // Methods
  const handleCurrentChange = (currentRow: Device) => {
	if (currentRow) {
	  emit('select-device', currentRow);
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
  
  const formatSpeed = (speed?: number): string => {
	if (speed === undefined) return 'N/A';
	return `${speed} km/h`;
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
	if (!odometer?.value) return 'N/A';
	return `${odometer.value.toFixed(1)} ${odometer.unit}`;
  };
  
  const formatDateTime = (dateString?: string): string => {
	if (!dateString) return 'N/A';
	return new Date(dateString).toLocaleString();
  };
  </script>
  
  <style scoped>
  .device-list-container {
	height: 100%;
	position: relative;
	background: #f5f5f5;
	border-radius: 8px;
	overflow: hidden;
  }
  
  .search-container {
	padding: 10px;
	background: white;
	border-bottom: 1px solid #e0e0e0;
  }
  
  .map-icon-cell {
	position: relative;
	width: 40px;
	text-align: center;
  }
  
  .map-icon {
	cursor: pointer;
	width: 24px;
	height: 24px;
  }
  
  .device-name {
	font-weight: 500;
	color: #2c3e50;
  }
  
  .device-id {
	font-size: 0.8rem;
	color: #666;
	margin-top: 4px;
  }
  
  .status-cell {
	display: flex;
	align-items: center;
	gap: 8px;
  }
  
  .status-indicator {
	width: 8px;
	height: 8px;
	border-radius: 50%;
  }
  
  .status-indicator.online {
	background: #4caf50;
  }
  
  .status-indicator.offline {
	background: #f44336;
  }
  
  :deep(.el-table__row.current-row) {
	background-color: #e3f2fd;
  }
  .map-icon-cell {
		cursor: pointer; /* Make the icon clickable */
		display: flex;
		justify-content: center; /* Center the icon horizontally */
		align-items: center;   /* Center the icon vertically */
		height: 100%;         /* Make the cell fill the row height */
	}


	.map-icon-cell .map-icon{ 
		display: inline-block; 
		border-radius: 50%;  
	}


</style>