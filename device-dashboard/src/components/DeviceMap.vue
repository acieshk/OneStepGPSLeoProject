<template>
	<div class="device-map-container">
		<l-map ref="map" :zoom="zoom" :center="center" @ready="handleMapReady">
			<l-tile-layer :url="url" :attribution="attribution"></l-tile-layer>
            <l-marker 
				v-for="device in visibleDevices" 
                :key="device.device_id" 
                :lat-lng="getLatLng(device)"
                :icon="createLeafletIcon(device)"
            />
		</l-map>
	</div>

</template>


<script setup lang="ts">
import { inject, nextTick, h, ref, onMounted, watch, onUnmounted, defineComponent, computed } from 'vue';
import 'leaflet/dist/leaflet.css';
import * as L from 'leaflet'; // Correct import
import { LMap, LTileLayer, LMarker, LIcon } from '@vue-leaflet/vue-leaflet'; // Import LMarker and LIcon
import type { Device } from '@/types/device';
import { useUserStore } from '@/stores/userStore';
import { storeToRefs } from 'pinia';
import _ from 'lodash';
import MapMarkerIcon from './MapMarkerIcon.vue';
import { renderToString } from 'vue/server-renderer';

const zoom = ref(6);

const url = ref('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png');  // Base tile layer URL
//Have to include attribution otherwise it is going to show blank
const attribution = ref('&copy; <a target="_blank" href="http://osm.org/copyright">OpenStreetMap</a> contributors');

const userStore = useUserStore();
const { devices, selectedDevice } = storeToRefs(userStore);
var center = computed(() => {
    if (selectedDevice.value?.latest_device_point) {
        return [
            selectedDevice.value.latest_device_point.lat,
            selectedDevice.value.latest_device_point.lng
        ] as L.LatLngTuple;
    }
    return [37.25, -119.75] as L.LatLngTuple; // Default center
}); // Initial center point (California)

const visibleDevices = computed(() => {
    return devices.value.filter(device => device.visible);
});

// Correctly set Leaflet icon paths. Just a workaround for this library
L.Icon.Default.mergeOptions({
	iconRetinaUrl: new URL('leaflet/dist/images/marker-icon-2x.png', import.meta.url).href,
	iconUrl: new URL('leaflet/dist/images/marker-icon.png', import.meta.url).href,
	shadowUrl: new URL('leaflet/dist/images/marker-shadow.png', import.meta.url).href,
});

const getLatLng = (device: Device): L.LatLngTuple => { // Explicitly return L.LatLngTuple
	if (device.latest_device_point && device.latest_device_point.lat && device.latest_device_point.lng) {
		return [device.latest_device_point.lat, device.latest_device_point.lng] as L.LatLngTuple; // Type assertion if needed
	}
	return center.value as L.LatLngTuple; // Use type assertion
};

// Create a ref to hold the injected layout preference
const injectedLayoutPreference = inject<string | undefined>('layoutPreference');
const layoutPreference = ref(injectedLayoutPreference);

const map = ref();
let mapReady = ref(false);

const handleMapReady = () => {
    mapReady.value = true;
    if (devices.value.length > 0) {
        nextTick(() => {
            map.value?.mapObject?.fitBounds(devices.value.map(device => getLatLng(device)));
        });
    }
};

watch(() => devices.value, () => {
    if (map.value?.mapObject && devices.value.length > 0) {
        map.value.mapObject.fitBounds(devices.value.map(device => getLatLng(device)));
    }
}, { deep: true });

// Add watch for selectedDevice changes
watch(() => selectedDevice.value, (newDevice) => {
    if (newDevice?.latest_device_point && map.value?.mapObject) {
        const newCenter = [
            newDevice.latest_device_point.lat,
            newDevice.latest_device_point.lng
        ] as L.LatLngTuple;
        
         map.value.mapObject.panTo(newCenter);
    }
}, { immediate: true });

const debouncedUpdateMarkers = _.debounce(() => {
	updateMarkers();
}, 300);


const { getDeviceIcon } = userStore; 

const createLeafletIcon = (device: Device) => {
    // Use the iconURL that should already be in the device object from userStore
    return L.icon({
        iconUrl: device.iconURL || `https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-${device.color}.png`,
        iconSize: [25, 41],
        iconAnchor: [12, 41],
        popupAnchor: [0, -41]
    });
};


const createMarker = async (device: Device) => {
	console.log("createMarker");
    if (!map.value?.mapObject || !device.latest_device_point) {
        return;
    }
    const latLng = getLatLng(device);
    try {
        const iconUrl = await getDeviceIcon(device);
        
        // Create a simple div icon with an img element
        const icon = L.divIcon({
            html: `<img src="${iconUrl}" style="width: 25px; height: 41px;" />`,
            className: "",
            iconSize: [25, 41],
            iconAnchor: [12, 41],
            popupAnchor: [0, -41]
        });

        const marker = L.marker(latLng, { icon }).addTo(map.value.mapObject);
        markers.value[device.device_id] = marker;
    } catch (error) {
        console.error('Error creating marker:', error);
    }
};




const updateMarkers = async() => {
	if (!map.value?.mapObject || !devices.value) {
		return;  // No devices yet
	}

	for (const deviceId in markers.value) { // Clear existing markers
		markers.value[deviceId].remove();
	}
	markers.value = {};


	for (const device of devices.value) {
		await createMarker(device); // Wait for createMarker which is now async
	}
};

onMounted(async () => {
    console.log("Component mounted");
    await userStore.fetchDevices(); // Make sure devices are loaded
    if (mapReady.value) {
        updateMarkers();
    }
});



// Watch for changes in devices (including visibility changes)
watch(
    () => devices.value,
    () => {
        if (map.value?.mapObject && visibleDevices.value.length > 0) {
            map.value.mapObject.fitBounds(visibleDevices.value.map(device => getLatLng(device)));
        }
    },
    { deep: true }
);

</script>

<style scoped>
.device-map-container {
	height: 100%;
	width: 100%;
}
</style>