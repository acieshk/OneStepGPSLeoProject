<template>
	<div class="device-map-container">
		<l-map ref="map" :zoom="zoom" :center="center" @ready="handleMapReady">
			<l-tile-layer :url="url" :attribution="attribution"></l-tile-layer>
			<l-marker v-for="device in devices" :key="device.device_id" :lat-lng="getLatLng(device)"
				:icon="getIcon(device)" @click="() => handleMarkerClick(device)"></l-marker>
			<!-- Iterate over devices -->
		</l-map>
	</div>

</template>


<script setup lang="ts">
import { inject, nextTick, ref, onMounted, watch, onUnmounted  } from 'vue';
import 'leaflet/dist/leaflet.css';
import * as L from 'leaflet'; // Correct import
import { LMap, LTileLayer, LMarker, LIcon } from '@vue-leaflet/vue-leaflet'; // Import LMarker and LIcon
import type { Device } from '@/types/device';
import { useUserStore } from '@/stores/userStore';
import { storeToRefs } from 'pinia';
import _ from 'lodash';

const zoom = ref(6);
const center = ref([37.25, -119.75]);  // Initial center point (California)
const url = ref('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png');  // Base tile layer URL
//Have to include attribution otherwise it is going to show blank
const attribution = ref('&copy; <a target="_blank" href="http://osm.org/copyright">OpenStreetMap</a> contributors');

const userStore = useUserStore();
const { devices } = storeToRefs(userStore);

// Correctly set Leaflet icon paths. Just a workaround for this library
L.Icon.Default.mergeOptions({
	iconRetinaUrl: new URL('leaflet/dist/images/marker-icon-2x.png', import.meta.url).href,
	iconUrl: new URL('leaflet/dist/images/marker-icon.png', import.meta.url).href,
	shadowUrl: new URL('leaflet/dist/images/marker-shadow.png', import.meta.url).href,
});

const getLatLng = (device: Device) => { // Corrected return type
	if (device.latest_device_point && device.latest_device_point.lat && device.latest_device_point.lng) {
		return [device.latest_device_point.lat, device.latest_device_point.lng];
	}
	return center.value; // Or some default location if lat/lng is not available
};

// Create a ref to hold the injected layout preference
const injectedLayoutPreference = inject<string | undefined>('layoutPreference');
const layoutPreference = ref(injectedLayoutPreference);

const map = ref();
let mapReady = ref(false);

const handleMapReady = () => {
	mapReady.value = true;
	updateMarkers();
};

const getIcon = (device) => {
	// Existing getIcon logic - construct icon based on iconURL or color
	if (device.iconURL) {
		return new L.Icon({

			iconUrl: device.iconURL,
			iconSize: [32, 32],
			iconAnchor: [16, 32],
		});
	}
	const color = device.color || 'blue';

	return new L.Icon({
		iconUrl: `https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-${color}.png`,
		iconSize: [25, 41],
		iconAnchor: [12, 41],
	});

};

const isSelected = (device) => {
	return props.selectedDevice?.device_id === device.device_id;
};

const debouncedUpdateMarkers = _.debounce(() => {
	updateMarkers();
}, 300);


const markers = ref({});  // Stores markers by device_id



const createMarker = (device) => {  // Creates a single marker

	if (!map.value || !map.value.mapObject || !device.latest_device_point) {
		return; // Return early to prevent errors

	}

	const latLng = getLatLng(device);  // Get device coordinates
	// Creates a leaflet marker
	const marker = L.marker(latLng, { icon: getIcon(device) }).addTo(map.value.mapObject); // Add marker to map

	markers.value[device.device_id] = marker;

};




const updateMarkers = () => {
	if (!map.value?.mapObject || !devices.value) {
		return;  // No devices yet
	}

	for (const deviceId in markers.value) { // Clear existing markers
		markers.value[deviceId].remove();
	}
	markers.value = {};


	devices.value.forEach(device => { // Create markers for all current devices
		createMarker(device);
	});

};

onMounted(() => {

});

// Remove existing watch on layoutKey (it's not needed)
// Watch for changes to the injected layoutPreference
watch(layoutPreference, () => {  // layoutPreference is now a ref
    nextTick(() => {
        if (map.value?.mapObject) {
            // Redraw each layer:
            map.value.mapObject.eachLayer((layer) => {  // Force redraw of each layer
                layer.redraw();
            });

            // Optionally call invalidateSize afterwards (might help in some cases):
            map.value.mapObject.invalidateSize(false);
        }	
    });
}, { immediate: true }); // Trigger initially for proper sizing on mount

// Watch for changes to the devices array and re-render markers if needed
watch(() => devices.value, () => { // Correctly call fitBounds in the watcher

	console.log(devices);
	if (map.value?.mapObject && devices.value.length > 0) { // Check if map and devices are available
		map.value.mapObject.fitBounds(devices.value.map(device => getLatLng(device))); // Call fitBounds here
	}
}, { deep: true });

</script>

<style scoped>
.device-map-container {
	height: 100%;
	width: 100%;
}
</style>