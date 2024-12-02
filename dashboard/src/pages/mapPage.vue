<template>
  <l-map ref="map" :zoom="zoom" :center="center" :use-global-leaflet="false" @ready="onMapReady">
    <l-tile-layer :url="url" :attribution="attribution"></l-tile-layer>
    <l-marker v-for="device in devices" :key="device._id" :lat-lng="deviceLatLng(device)" :icon="deviceIcon(device)">
    </l-marker>
  </l-map>
</template>
<script setup lang="ts">
import 'leaflet/dist/leaflet.css';  // Import Leaflet CSS
import { LMap, LTileLayer, LMarker } from '@vue-leaflet/vue-leaflet';
import { ref, onMounted, nextTick, watch } from 'vue';
import L, { Icon } from 'leaflet';
import type { PointTuple } from 'leaflet';
import { useDeviceStore } from 'src/stores/deviceStore';
import { storeToRefs } from 'pinia';
import { Device } from 'src/model/model';

const markers = ref(new Map<string, L.Marker>());

const deviceStore = useDeviceStore();
const { devices } = storeToRefs(deviceStore);
const center = ref<PointTuple>([37.25, -119.75]); // Default center coordinates (New York City)
const zoom = ref(6);
let map = ref<L.Map>();
const url = ref('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png');  // Base tile layer URL
const attribution = ref('');

// Fix for default icon paths
// See https://vue2-leaflet.netlify.app/quickstart/#marker-icons-are-missing
// @ts-expect-error Ignore the error because it is exactly what documention tells me to do
delete Icon.Default.prototype._getIconUrl;
Icon.Default.mergeOptions({
	iconRetinaUrl: new URL('leaflet/dist/images/marker-icon-2x.png', import.meta.url).href,
	iconUrl: new URL('leaflet/dist/images/marker-icon.png', import.meta.url).href,
	shadowUrl: new URL('leaflet/dist/images/marker-shadow.png', import.meta.url).href,
});

const DEFAULT_ICON = L.icon({ // Create a default icon for consistent default icon implementation
    iconUrl: 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-blue.png', // Use HTTPS URL
    iconSize: [25, 41], // Adjust as needed
    iconAnchor: [12, 41], // Adjust as needed
    popupAnchor: [1, -34], // Adjust as needed
    tooltipAnchor: [16, -28], // Adjust as needed
	shadowSize: [41, 41] // Adjust as needed
});

const deviceLatLng = (device: Device): L.LatLngTuple  => {
    if (device.latest_device_point) {
        return [device.latest_device_point.lat, device.latest_device_point.lng];
    }
    return [0,0]; 
};

const onMapReady = (leafletMap: L.Map) => {
	map.value = leafletMap;
	deviceStore.setMapReady(true);
};


onMounted(async () => {
    await nextTick();

    if (map.value) {
        if (map.value) {
            deviceStore.setMapReady(true);
            createMapMarkers();
        }
    }
});

const deviceIcon = (device: Device) => {
    if (!device.iconUrl) return DEFAULT_ICON;
    
    try {
        return L.icon({ 
            iconUrl: String(device.iconUrl), // Convert to string explicitly
            iconSize: [25, 41],
            iconAnchor: [12, 41],
            popupAnchor: [1, -34]
        });
    } catch (error) {
        console.error(`Error creating icon for device: ${device._id}`, error);
        return DEFAULT_ICON;
    }
};

interface LeafletMarker extends L.Marker {
    _leaflet_id?: number;
}

// Then update the marker creation:
const createMapMarkers = () => {
    if (!map.value) return;
    
    // First, remove all existing markers from the map
    map.value.eachLayer((layer) => {
        if (layer instanceof L.Marker) {
            map.value?.removeLayer(layer);
        }
    });
    
    // Clear the markers Map
    markers.value.clear();

    // Create new markers
    devices.value.forEach(device => {
        if(!map.value || !deviceLatLng(device)) return;
        const marker = L.marker(deviceLatLng(device)!, { icon: deviceIcon(device) }) as LeafletMarker;
        marker.addTo(map.value);
        markers.value.set(device._id, marker);
        if (marker._leaflet_id) {
            device.markerId = marker._leaflet_id;
        }
    });
};

watch(devices, (newDevices) => {
    if (!map.value) return;
    newDevices.forEach(device => {
        if (!device.markerId) return;
        
        const marker = markers.value.get(device._id);
        if (!marker) return;

        marker.setIcon(deviceIcon(device));
        const newLatLng = deviceLatLng(device);
        if (newLatLng) {
            marker.setLatLng(newLatLng);
        }
    });
}, { deep: true });
</script>

<style lang="scss" scoped>
.map-container {
	width: 100%;
	height: 100%;
}
.l-map {
    height: 400px; /* or whatever height you want */
    width: 100%;
}
</style>