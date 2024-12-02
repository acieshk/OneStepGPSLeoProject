<template>
	<div id="map" ref="map" class="map"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, shallowRef } from 'vue';
import Map from 'ol/Map';
import View from 'ol/View';
import TileLayer from 'ol/layer/Tile';
import OSM from 'ol/source/OSM';
import { fromLonLat } from 'ol/proj.js'; // For coordinate transformation
import Feature from 'ol/Feature';
import Point from 'ol/geom/Point';
import VectorLayer from 'ol/layer/Vector';
import VectorSource from 'ol/source/Vector';
import { Icon, Style } from 'ol/style';  // Import Icon and Style
import { useDeviceStore } from 'src/stores/deviceStore';
import { storeToRefs } from 'pinia';
import type { Device } from 'src/model/model';

const deviceStore = useDeviceStore();
const { devices, selectedDeviceId } = storeToRefs(deviceStore);
const map = shallowRef<Map | null>(null);


const zoom = ref(6);
const center = ref([-119.75, 37.25]); // Longitude, Latitude for OpenLayers


onMounted(() => {
	// Initialize the map after the component is mounted
	map.value = new Map({
		target: 'map', // Reference to your map div
		layers: [
			new TileLayer({
				source: new OSM()
			})
		],
		view: new View({
			center: fromLonLat(center.value), // Transform to OpenLayers coordinates
			zoom: zoom.value
		}),
		controls: [], // This hides all default controls.  Remove or customize this for fine-grained control visibility.
	});

	// Initial marker creation
	createMarkers();

	// Watch for changes in devices (including visibility, locations, etc.)
	watch(devices, () => {
		// Remove existing markers before recreating so you don't have duplicates and also takes care of visibility since markers are recreated
		if (map.value)
		{
			map.value.getLayers().forEach(layer => {
				if (layer instanceof VectorLayer) {
					map.value?.removeLayer(layer);
				}
			});
		}
		createMarkers(); // Recreate markers based on updated data
	}, { deep: true });
});

const defaultIcon = new Icon({
  src: 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-blue.png',
  width: 48,  
  height: 48
});




const createMarkers = () => {
	const features = devices.value.filter(device => device.visible).map(device => { // Filters out non-visible devices.
		return new Feature({
			geometry: new Point(fromLonLat(deviceLatLng(device))), // Convert coordinates
			device: device // Store the device data directly in the feature for easy access later.
		});
	});

	const vectorSource = new VectorSource({
		features: features
	});

	const vectorLayer = new VectorLayer({
    source: vectorSource,
    style: feature => {
      const device: Device = feature.get('device');
      const icon = device.iconUrl ? new Icon({ 
        src: device.iconUrl, 
        width: 48,  
        height: 48
      }) : defaultIcon; 

      return new Style({
        image: icon
		
      });
    }
  });


	map.value?.addLayer(vectorLayer);
};



const deviceLatLng = (device: Device): [number, number] => { // Returns [lon, lat]
	if (device.latest_device_point) {
		return [device.latest_device_point.lng, device.latest_device_point.lat];
	}
	return [0, 0]; // Default if no location data is available
};

onMounted(() => {
    // ... (map initialization)

    // Watch for changes in selectedDeviceId
    watch(selectedDeviceId, (newDeviceId) => {
        if (newDeviceId && map.value) {
            const selectedDevice = devices.value.find(device => device._id === newDeviceId);
            if (selectedDevice && selectedDevice.latest_device_point) {
                const coordinates = fromLonLat([selectedDevice.latest_device_point.lng, selectedDevice.latest_device_point.lat]);

				// Option 1: Instant pan
                // map.value.getView().setCenter(coordinates);

				// Option 2: Animated pan (smoother)
                map.value.getView().animate({
                    center: coordinates,
                    duration: 500 // Animation duration in milliseconds
                });


            }
        }
    });
});
</script>

<style scoped>
.map {
	height: 100%;
	width: 100%;
}
</style>