<template>
    <div v-if="isLoading">
        <div class="placeholder-icon"></div> 
    </div>
    <img v-else-if="isCustomIcon" :src="iconUrl" alt="Device Icon" class="map-icon" @error="handleImageError">
    <svg v-else class="map-icon" width="24" height="36" viewBox="0 0 24 36" xmlns="http://www.w3.org/2000/svg">
        <path d="M12 0C5.4 0 0 5.4 0 12c0 7.2 12 24 12 24s12-16.8 12-24c0-6.6-5.4-12-12-12z" :fill="color" stroke="white" stroke-width="2" />
        <circle cx="12" cy="12" r="4" fill="white" />
    </svg>
</template>

<script setup lang="ts">
import { defineProps, ref, onMounted, watch, toRefs } from 'vue';
import { apiService } from '@/services/api.service';
import type { Device } from '@/types/device';

const props = defineProps<{
    deviceId: string;
    color: string;
}>();

const { color } = toRefs(props);
const iconUrl = ref('');
const isLoading = ref(true);
const isCustomIcon = ref(false);

console.log("Props received in MapMarkerIcon:", props);

const handleImageError = () => {
	console.error("Error loading custom icon. Displaying default.");
	isCustomIcon.value = false; // Fallback to default SVG if image fails to load
};


const fetchIcon = async () => {
	console.log("fetchIcon called with deviceId:", props.deviceId); // Log when fetchIcon starts

    isLoading.value = true;
    if (props.deviceId) {
        try {
            // console.log("Calling apiService.getIcon with deviceId:", props.deviceId); // Log API call
            console.log("apiService.getIcon with deviceId:", props.deviceId);

			const fetchedIconUrl = await apiService.getIcon(props.deviceId); 

            if (fetchedIconUrl) { // Check if fetchedIconUrl is not null. 
                iconUrl.value = fetchedIconUrl;
                isCustomIcon.value = true;
            }
            else{
                isCustomIcon.value = false;
            }
           

        } catch(error) {
            console.error("Error fetching icon:", error);
            isCustomIcon.value = false; // Fallback to default in case of error
        } finally {
            isLoading.value = false;
        }
    } else {
        isLoading.value = false; // If no deviceId, no need to load, show default
        isCustomIcon.value = false; // Make sure to reset this if deviceId is absent
    }
	
	console.log("Device: " + props.deviceId);
	console.log("isCustomIcon:", isCustomIcon.value);
    console.log("iconUrl:", iconUrl.value);
    console.log("isLoading:", isLoading.value);
	console.log("color:", color.value);
};


onMounted(fetchIcon);

watch(() => props.deviceId, fetchIcon, { immediate: true });
</script>

<style scoped>
.map-icon {
	width: 24px;
	height: 36px;
	display: inline-block;
}

/* Styles for the default icon */
.default-icon {
	width: 24px;
	height: 36px;
	border-radius: 2px;
	display: inline-block;
	background-color: v-bind(color);  
}

.placeholder-icon { /* Style for placeholder while icon is loading */
  width: 24px;
  height: 36px;
  background-color: #eee; /* Light gray placeholder */
  display: inline-block; 
}
</style>