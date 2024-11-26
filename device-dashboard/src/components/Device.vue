<template>
	<div class="devices-view" :style="layoutStyle" v-if="userPreferences">
		<DeviceList v-if="devices" :devices="devices" @select-device="handleDeviceSelect"
			:selectedDevice="selectedDevice" :userPreferences="userPreferences"
			@update-device-color="handleDeviceColorUpdate" @update-device-visibility="handleDeviceVisibilityUpdate" />
		<DeviceMap v-if="devices" :devices="devices" :selectedDevice="selectedDevice" :deviceColors="deviceColors"
			:deviceVisibility="deviceVisibility" @select-device="handleDeviceSelect" />
	</div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { storeToRefs } from 'pinia';
import { useUserStore } from '@/stores/userStore';
import DeviceList from './DeviceList.vue';

import DeviceMap from './DeviceMap.vue';

const userStore = useUserStore();

const { devices, userPreferences } = storeToRefs(userStore);

const selectedDevice = ref(null);


const handleDeviceSelect = (device) => {
	selectedDevice.value = device;
};

const layoutStyle = computed(() => {  // Dynamic style object
	return {
		display: 'flex',
		flexDirection: userPreferences.value.layout === 'vertical' ? 'column' : 'row',
		height: '100%',  // Always 100% height
		width: '100%',    // Always 100% width
	};
});


onMounted(() => {
	userStore.fetchDevices();
});

</script>
<style scoped>
.devices-view {
	display: flex;
	/* Essential: Enable flexbox layout */
	flex-direction: row;
	/* Arrange children horizontally (or use column for vertical as needed) */
	width: 100%;
	/* Occupy full width */
	height: 100%;
	/* Occupy full height */
	padding: 0;
	/* Remove any padding */
	margin: 0;
	/* Remove any margin */
}


/* Styles for horizontal layout */
.devices-view.horizontal>* {
	/* When layout is horizontal, children take 50% width */
	width: 50%;
	height: 100%;
	overflow-y: auto;
	/* Vertical scroll within children */
	box-sizing: border-box;

}

/* Styles for vertical layout */
.devices-view.vertical>* {
	/* When layout is vertical, children take full width */
	width: 100%;
	height: 50%;
	overflow-y: auto;
	box-sizing: border-box;

}

/* Responsive layout adjustments - Example: Stack children vertically on smaller screens */
@media (max-width: 768px) {

	/* Media query example */
	.devices-view {
		flex-direction: column;
		/* Stack vertically */
	}

	.devices-view>* {
		/* Target direct children*/
		width: 100%;
		/* Occupy full width */
		height: 50%;
		/* Occupy 50% height */
	}
}
</style>
