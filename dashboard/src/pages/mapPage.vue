<template>
	<div class="map-container">
		<l-map ref="map" :zoom="zoom" :center="center" :use-global-leaflet="false" @Ready="onMapReady">
			<l-tile-layer :url="url" :attribution="attribution"></l-tile-layer>
			<l-marker v-for="device in devices" :key="device._id" :lat-lng="deviceLatLng(device)" ref="marker"></l-marker>
		</l-map>
	</div>
</template>

<script setup lang="ts">
import 'leaflet/dist/leaflet.css';  // Import Leaflet CSS
import { LMap, LTileLayer, LMarker } from '@vue-leaflet/vue-leaflet';
import { ref, onMounted, nextTick } from 'vue';
import L, { Icon } from 'leaflet';
import type { LatLngTuple, PointTuple } from 'leaflet';
import { useDeviceStore } from 'src/stores/deviceStore';
import { storeToRefs } from 'pinia';
import { Device } from 'src/types/device';


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

const deviceLatLng = (device: Device): LatLngTuple => {
	if (device.latest_device_point) {
		return [device.latest_device_point.lat, device.latest_device_point.lng];
	}
	return center.value;
}

const onMapReady = (leafletMap: L.Map) => {
	map.value = leafletMap;
	deviceStore.setMapReady(true);
};


onMounted(async () => {
	await nextTick();  // No other logic needed in onMounted since vue-leaflet creates markers
});

</script>

<style lang="scss" scoped>
.map-container {
	width: 100%;
	height: 100%;
}
</style>