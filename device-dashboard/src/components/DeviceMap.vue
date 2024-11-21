<template>
	<div class="map-container">
	  <l-map ref="mapRef" :zoom="13" :center="mapCenter" @ready="handleMapReady">
		<l-tile-layer url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
		<l-marker
		  v-for="device in visibleDevices"
		  :key="device.device_id"
		  :lat-lng="getDeviceLatLng(device)"
		  :icon="getDeviceIcon(device)"
		  @click="(event) => handleMarkerClick(device, event)"
		></l-marker> 
	  </l-map>
	</div>
  </template>
  
  <script setup lang="ts">
  import { ref, watch, computed, onMounted } from 'vue';
  import "leaflet/dist/leaflet.css";
  import * as L from 'leaflet';
  import { LMap, LTileLayer, LMarker } from "@vue-leaflet/vue-leaflet";
  import { Device } from '@/App.vue';
  
  const deviceColors = [
	'#E53935', '#1E88E5', '#43A047', '#FB8C00', '#8E24AA',
	'#00ACC1', '#FFB300', '#546E7A', '#D81B60', '#6D4C41'
  ];
  
  delete L.Icon.Default.prototype._getIconUrl;
  L.Icon.Default.mergeOptions({
	iconRetinaUrl: new URL('leaflet/dist/images/marker-icon-2x.png', import.meta.url).href,
	iconUrl: new URL('leaflet/dist/images/marker-icon.png', import.meta.url).href,
	shadowUrl: new URL('leaflet/dist/images/marker-shadow.png', import.meta.url).href,
  });
    
  const props = defineProps<{
	devices: Device[];
	selectedDevice: Device | null;
	deviceVisibility: { [key: string]: boolean };
  }>();
  
  const visibleDevices = computed(() => props.devices.filter(device => props.deviceVisibility[device.device_id]));
  
  const getDeviceColor = (deviceId: string) => {
  const device = props.devices.find(d => d.device_id === deviceId);
  return device?.color || '#1976D2'; // Default color if not specified
};

const getDeviceIcon = (device: Device) => {
  const color = getDeviceColor(device.device_id);
  const scale = isSelected(device) ? 1.2 : 1;
  return L.divIcon({
    className: 'custom-div-icon',
    html: `<svg width="${24 * scale}" height="${36 * scale}" viewBox="0 0 24 36" xmlns="http://www.w3.org/2000/svg"><path d="M12 0C5.4 0 0 5.4 0 12c0 7.2 12 24 12 24s12-16.8 12-24c0-6.6-5.4-12-12-12z" fill="${color}" stroke="white" stroke-width="2"/><circle cx="12" cy="12" r="4" fill="white"/></svg>`,
    iconSize: [24 * scale, 36 * scale],
    iconAnchor: [12 * scale, 36 * scale],
    popupAnchor: [0, -36 * scale]
  });
};
  
const emit = defineEmits<{
	(e: 'select-device', device: Device): void;
	}>();
  
  
  const mapRef = ref();
  const defaultCenter = [34.0522, -118.2437];
  const mapCenter = computed(() => (props.selectedDevice?.latest_device_point?.lat && props.selectedDevice?.latest_device_point?.lng) ? [props.selectedDevice.latest_device_point.lat, props.selectedDevice.latest_device_point.lng] : defaultCenter);
  
  
  const markers = ref<L.Marker[]>([]);
  
  
  onMounted(() => {
	  markers.value.forEach(marker => marker.remove());
	  markers.value = [];
  
	  const mapInstance = (mapRef.value as any).leafletObject;
  
	  visibleDevices.value.forEach(createMarker);
  });
  
  
  watch(visibleDevices, (newDevices) => {
	  const mapInstance = (mapRef.value as any).leafletObject;
	  markers.value.forEach(marker => marker.remove());
	  markers.value = [];
	  newDevices.forEach(createMarker);
  
  }, {deep: true});
  
  const createMarker = (device: Device) => {
	  const mapInstance = (mapRef.value as any).leafletObject;
	  const markerLatLng = getDeviceLatLng(device);
	  const marker = L.marker(markerLatLng, { icon: getDeviceIcon(device) }).addTo(mapInstance);
	  const popupContent = `
		  <div class="popup-content">
			  <h3>${device.display_name}</h3>
			  <p>Location: ${formatLocation(markerLatLng[0], markerLatLng[1])}</p>
		  </div>
	  `;
	  marker.bindPopup(popupContent);
  
	  marker.on('click', (event) => {
		  handleMarkerClick(device, event);
	  });
  
  
	  markers.value.push(marker);
  };
  
  
  
  const handleMapReady = () => {
	console.log('Map is ready');
  };
  
  const getDeviceLatLng = (device: Device): [number, number] => (device.latest_device_point?.lat && device.latest_device_point?.lng) ? [device.latest_device_point.lat, device.latest_device_point.lng] : defaultCenter;
  
  const formatLocation = (lat: number, lng: number) => `${lat.toFixed(6)}, ${lng.toFixed(6)}`;
  
  const handleMarkerClick = (device: Device, event: L.LeafletMouseEvent) => {
	console.log('Marker clicked:', device.display_name);
	emit('select-device', device);
	event.target.openPopup();
  };
  
  const isSelected = (device: Device) => props.selectedDevice?.device_id === device.device_id;
  
  watch(() => props.selectedDevice, (newDevice) => {
	  if (newDevice?.latest_device_point?.lat && newDevice?.latest_device_point?.lng) {
		  const mapInstance = mapRef.value;
		  if (mapInstance) {
			  const leafletMap = (mapInstance as any).leafletObject;
			  if (leafletMap) {
				  leafletMap.setView([newDevice.latest_device_point.lat, newDevice.latest_device_point.lng], 13, { animate: true, duration: 1 });
			  }
		  }
	  }
  }, { deep: true });
  
  </script>
  
  <style scoped>
  .map-container {
	height: 100%;
	width: 100%;
	border-radius: 8px;
	overflow: hidden;
  }
  
  .popup-content {
	padding: 10px;
	min-width: 200px;
  }
  
  .popup-content h3 {
	margin: 0 0 8px 0;
	font-size: 1.1rem;
	color: #2c3e50;
  }
  
  .popup-content p {
	margin: 4px 0;
	font-size: 0.9rem;
	color: #666;
  }
  
  .status-row {
	display: flex;
	align-items: center;
	gap: 8px;
	margin-bottom: 8px;
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
  
  :deep(.leaflet-popup-content-wrapper) {
	border-radius: 8px;
  }
  
  :deep(.leaflet-popup-content) {
	margin: 8px 12px;
  }
  
  :deep(.leaflet-container) {
	font-family: inherit;
  }
  </style>