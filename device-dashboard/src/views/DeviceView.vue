<template>
	<el-container class="devices-view">
	  <el-drawer v-model="drawerVisible" title="Device List" direction="ltr" size="50%" append-to-body>
		  <DeviceList v-if="devices" :devices="devices" @select-device="handleDeviceSelect" :selectedDevice="selectedDevice" :userPreferences="userPreferences" @update-device="handleDeviceUpdate"/> 
	  </el-drawer>
	  <el-button @click="drawerVisible = !drawerVisible">Toggle Device List</el-button>
  
	  <DeviceMap v-if="devices" :devices="devices" :selectedDevice="selectedDevice" @select-device="handleDeviceSelect" /> 
	</el-container>
  </template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { storeToRefs } from 'pinia';
import { useUserStore } from '@/stores/userStore';
import DeviceList from '@/components/DeviceList.vue';
import DeviceMap from '@/components//DeviceMap.vue';

import type { Device } from '@/types/device';

const userStore = useUserStore();
const { devices, userPreferences } = storeToRefs(userStore);
const emit = defineEmits<{
	(e: 'refresh-list'): void;
}>();

const drawerVisible = ref(false); // Control drawer visibility
const selectedDevice = ref(null);




const handleDeviceUpdate = (updatedDevice: Device) => {
  // Create a new array with the updated device
  const updatedDevices = devices.value.map(device => 
    device.device_id === updatedDevice.device_id ? updatedDevice : device
  );
  
  // Update the store directly
  userStore.devices = updatedDevices;
  emit('refresh-list');
};


const handleDeviceSelect = (device) => {
	selectedDevice.value = device;
};

onMounted(() => {
	
});

</script>
<style scoped>
.devices-view {
  height: 100%; /* Ensure the container takes full height */
  width: 100%;   /* Ensure the container takes full width */
}

.device-map-container {
    height: 100%;
    width: 100%;
    z-index: 1; /* Or a higher value if needed */
}

/* If needed, also adjust the drawer's z-index (in a global stylesheet or less specifically) */
.el-drawer__wrapper {
    z-index: 2; /* Make sure this is higher than the map's z-index */
}

</style>
