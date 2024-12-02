<template>
	<div id="map" ref="map" class="map"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, shallowRef, onUnmounted } from 'vue';
import Map from 'ol/Map';
import View from 'ol/View';
import TileLayer from 'ol/layer/Tile';
import OSM from 'ol/source/OSM';
import { fromLonLat, toLonLat } from 'ol/proj.js'; // For coordinate transformation
import Feature from 'ol/Feature';
import Point from 'ol/geom/Point';
import VectorLayer from 'ol/layer/Vector';
import VectorSource from 'ol/source/Vector';
import { Icon, Style } from 'ol/style';  // Import Icon and Style
import { useDeviceStore } from 'src/stores/deviceStore';
import { storeToRefs } from 'pinia';
import type { Device } from 'src/model/model';
import Overlay from 'ol/Overlay';
import { Coordinate } from 'ol/coordinate';


const deviceStore = useDeviceStore();
const { devices, selectedDeviceId } = storeToRefs(deviceStore);
const map = shallowRef<Map | null>(null);


const zoom = ref(6);
const center = ref([-119.75, 37.25]); // Longitude, Latitude for OpenLayers


// onMounted(() => {
// 	// Initialize the map after the component is mounted
// 	map.value = new Map({
// 		target: 'map', // Reference to your map div
// 		layers: [
// 			new TileLayer({
// 				source: new OSM()
// 			})
// 		],
// 		view: new View({
// 			center: fromLonLat(center.value), // Transform to OpenLayers coordinates
// 			zoom: zoom.value
// 		}),
// 		controls: [], // This hides all default controls.  Remove or customize this for fine-grained control visibility.
// 	});

// 	// Initial marker creation
// 	createMarkers();



// 	// Watch for changes in devices (including visibility, locations, etc.)
// 	watch(devices, () => {
// 		// Remove existing markers before recreating so you don't have duplicates and also takes care of visibility since markers are recreated
// 		if (map.value) {
// 			map.value.getLayers().forEach(layer => {
// 				if (layer instanceof VectorLayer) {
// 					map.value?.removeLayer(layer);
// 				}
// 			});
// 		}
// 		createMarkers(); // Recreate markers based on updated data
// 	}, { deep: true });
// });

const defaultIcon = new Icon({
	src: 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-blue.png',
	width: 48,
	height: 48
});

const createMarkers = () => {
    const features = devices.value.filter(device => device.visible).map(device => {
        return new Feature({
            geometry: new Point(fromLonLat(deviceLatLng(device))),
            device: device
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

    // Add click handler
    map.value!.on('click', (event) => {
		console.log('marker clicked');
        map.value!.forEachFeatureAtPixel(event.pixel, (feature, layer) => {
            if (layer === vectorLayer) {
                const device: Device = feature.get('device');
                showPopup(device, event.coordinate);
            }
        });

        if (!map.value!.hasFeatureAtPixel(event.pixel)) {
            popup.setPosition(undefined);
        }
    });

    // Add layer only ONCE
    map.value?.addLayer(vectorLayer);
};

const popupElement = document.createElement('div');
popupElement.className = 'ol-popup'; // Add styling class

const popup = new Overlay({
    element: popupElement,
    positioning: 'bottom-center',
    stopEvent: false,
    offset: [0, -50]
});

const showPopup = (device: Device, coordinate: Coordinate) => {
    const lonLat = toLonLat(coordinate);
    const onlineStatus = device.online ? 
        '<span style="color: #22c55e; font-size: 12px;">●</span> Online' : 
        '<span style="color: #ef4444; font-size: 12px;">●</span> Offline';
    
    const fuelPercent = device.latest_device_point?.device_state.fuel_percent || 0;
    const fuelBarWidth = Math.min(Math.max(fuelPercent, 0), 100);
    
    popup.getElement()!.innerHTML = `
        <div style="
            background: white;
            padding: 16px;
            border-radius: 12px;
            box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
            min-width: 300px;
            font-family: system-ui, -apple-system, sans-serif;
        ">
            <div style="
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 12px;
            ">
                <div style="
                    font-size: 16px;
                    font-weight: 600;
                    color: #1f2937;
                ">
                    ${device.display_name}
                </div>
                <div style="
                    font-size: 14px;
                    padding: 4px 8px;
                    border-radius: 9999px;
                    background: ${device.online ? '#dcfce7' : '#fee2e2'};
                    color: ${device.online ? '#166534' : '#991b1b'};
                ">
                    ${onlineStatus}
                </div>
            </div>
            
            <div style="
                font-size: 13px;
                color: #6b7280;
                margin-bottom: 8px;
            ">
                ID: ${device.device_id}
            </div>
            
            <div style="
                background: #f3f4f6;
                padding: 12px;
                border-radius: 8px;
                margin: 12px 0;
            ">
                <div style="
                    font-size: 14px;
                    color: #374151;
                    margin-bottom: 8px;
                ">
                    Location: ${lonLat.map(val => val.toFixed(2)).join(', ')}
                </div>
                
                <div style="margin-top: 12px;">
                    <div style="
                        font-size: 14px;
                        color: #374151;
                        margin-bottom: 4px;
                    ">
                        Fuel Level: ${fuelPercent}%
                    </div>
                    <div style="
                        background: #e5e7eb;
                        border-radius: 9999px;
                        height: 8px;
                    ">
                        <div style="
                            width: ${fuelBarWidth}%;
                            background: #3b82f6;
                            height: 100%;
                            border-radius: 9999px;
                            transition: width 0.3s ease;
                        "></div>
                    </div>
                </div>
            </div>
        </div>
    `;
    
    popup.setPosition(coordinate);
};

const deviceLatLng = (device: Device): [number, number] => { // Returns [lon, lat]
	if (device.latest_device_point) {
		return [device.latest_device_point.lng, device.latest_device_point.lat];
	}
	return [0, 0]; // Default if no location data is available
};

onMounted(() => {
    // Map initialization
    map.value = new Map({
        target: 'map',
        layers: [
            new TileLayer({
                source: new OSM()
            })
        ],
        view: new View({
            center: fromLonLat(center.value),
            zoom: zoom.value
        }),
        controls: []
    });
    
    map.value.addOverlay(popup); // Ensure popup is added to the map


    createMarkers(); // Call createMarkers after map initialization

    // Watch selectedDeviceId
    watch(
        selectedDeviceId,
        (newDeviceId) => {
            if (newDeviceId && map.value) {
                const selectedDevice = devices.value.find(device => device._id === newDeviceId);
                if (selectedDevice && selectedDevice.latest_device_point) {
                    const coordinates = fromLonLat([selectedDevice.latest_device_point.lng, selectedDevice.latest_device_point.lat]);
                    map.value.getView().animate({
                        center: coordinates,
                        duration: 500
                    });
                }
            }
        }
    );

    // Watch devices for changes in visibility or other properties
    watch(devices, () => {
        if(map.value) {  // Safeguard, only execute if map exists
           map.value.getLayers().getArray().forEach((layer) => {
               if (layer instanceof VectorLayer) {
                   map.value!.removeLayer(layer); // Remove old vector layers
               }
           });
           createMarkers(); // Update markers
        }
    }, {deep:true});

});

onUnmounted(() => {
    if (map.value) {
        map.value.setTarget(undefined);
        map.value = null;
    }
});
</script>

<style scoped>
.map {
	height: 100%;
	width: 100%;
}
.ol-popup {
    position: absolute;
    background-color: white;
    box-shadow: 0 1px 4px rgba(0,0,0,0.2);
    padding: 15px;
    border-radius: 10px;
    border: 1px solid #cccccc;
    min-width: 280px;
}

.popup-content { 
	font-size: 12px;
	padding: 15px;
	border-radius: 5px;
	background-color: white;
}

</style>