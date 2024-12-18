<template>
	<q-page class="form-container">
		<div class="form-wrapper">
			<div class="device-edit-header header-section">
				<q-toolbar>
					<q-toolbar-title>Edit Device Settings</q-toolbar-title>
					<q-btn flat label="Back" @click="goBack" />
					<q-btn color="primary" :disable="!deviceSettings" :loading="isSaving" @click="saveSettings">
						Save
					</q-btn>
				</q-toolbar>
			</div>
			<div v-if="deviceSettingsLoading">Loading settings...</div>
			<div v-else-if="!deviceSettingsLoaded">Failed to load settings.</div>

			<q-form ref="formRef" @submit.prevent="saveSettings" class="settings-form" v-else-if="deviceSettings">

				<!-- Icon Upload Section -->
				<div class="row q-col-gutter-md">
					<div class="col-12">
						<q-input v-model="deviceSettings.iconUrl" label="Icon URL" disable />
						<IconUploader :device-settings="deviceSettings" :current-icon-url="currentIconUrl"
							:is-uploading="isUploading" @upload="handleIconUpload" @remove="handleRemoveIcon"
							@default-icon="handleDefaultIcon" @image-error="handleImageError" />

					</div>
				</div>
				<!-- System Fields (Disabled) -->
				<div class="text-h6">System Information</div>
				<div class="row q-col-gutter-md">
					<q-input class="col-12 col-md-6" v-model="deviceSettings.device_id" label="Device ID" disable />
					<q-input class="col-12 col-md-6" v-model="deviceSettings.version" label="Version" type="number"
						disable />
					<q-input class="col-12 col-md-6" v-model="deviceSettings.updated_at" label="Last Updated" disable />
				</div>

				<!-- Speed Settings -->
				<div class="text-h6">Speed Settings</div>
				<div class="row q-col-gutter-md">
					<!-- Begin Moving Speed -->
					<div class="col-12 col-md-6">
						<q-input v-model.number="deviceSettings.begin_moving_speed.value" label="Begin Moving Speed"
							type="number" :rules="[val => val > 0 || 'Speed must be greater than 0']" />
						<q-input v-model="deviceSettings.begin_moving_speed.unit" label="Unit" type="text" />
					</div>

					<!-- Begin Stopped Speed -->
					<div class="col-12 col-md-6">
						<q-input v-model.number="deviceSettings.begin_stopped_speed.value" label="Begin Stopped Speed"
							type="number" :rules="[val => val >= 0 || 'Speed must be non-negative']" />
						<q-input v-model="deviceSettings.begin_stopped_speed.unit" label="Unit" type="text" />
					</div>
				</div>

				<!-- Distance and GPS Settings -->
				<div class="text-h6">Distance and GPS Settings</div>
				<div class="row q-col-gutter-md">
					<!-- Max Drift Distance -->
					<div class="col-12 col-md-6">
						<q-input v-model.number="deviceSettings.max_drift_distance.value" label="Max Drift Distance"
							type="number" :rules="[val => val > 0 || 'Distance must be greater than 0']" />
						<q-input v-model="deviceSettings.max_drift_distance.unit" label="Unit" type="text" />
					</div>

					<!-- GPS Settings -->
					<div class="col-12 col-md-6">
						<q-input v-model.number="deviceSettings.min_num_satellites" label="Minimum Number of Satellites"
							type="number" :rules="[val => val > 0 || 'Must be greater than 0']" />
						<q-toggle v-model="deviceSettings.ignore_unset_min_num_sats"
							label="Ignore Unset Minimum Satellites" />
						<q-input v-model.number="deviceSettings.max_hdop" label="Maximum HDOP" type="number" />
					</div>
				</div>

				<!-- Timeout Settings -->
				<div class="text-h6">Timeout Settings</div>
				<div class="row q-col-gutter-md">
					<!-- Drive Timeout -->
					<div class="col-12 col-md-4">
						<q-input v-model.number="deviceSettings.drive_timeout.value" label="Drive Timeout"
							type="number" />
						<q-input v-model="deviceSettings.drive_timeout.unit" label="Unit" type="text" />
					</div>

					<!-- Stop Timeout -->
					<div class="col-12 col-md-4">
						<q-input v-model.number="deviceSettings.stop_timeout.value" label="Stop Timeout"
							type="number" />
						<q-input v-model="deviceSettings.stop_timeout.unit" label="Unit" type="text" />
					</div>

					<!-- Offline Timeout -->
					<div class="col-12 col-md-4">
						<q-input v-model.number="deviceSettings.offline_timeout.value" label="Offline Timeout"
							type="number" />
						<q-input v-model="deviceSettings.offline_timeout.unit" label="Unit" type="text" />
					</div>
				</div>

				<!-- Fuel Settings -->
				<div class="text-h6">Fuel Settings</div>
				<div class="row q-col-gutter-md">
					<div class="col-12 col-md-6">
						<q-input v-model="deviceSettings.fuel_consumption.calculation_method" label="Calculation Method"
							type="text" />
						<q-input v-model="deviceSettings.fuel_consumption.measurement" label="Measurement"
							type="text" />
						<q-input v-model="deviceSettings.fuel_consumption.fuel_type" label="Fuel Type" type="text" />
						<q-input v-model.number="deviceSettings.fuel_consumption.fuel_cost" label="Fuel Cost"
							type="number" />
						<q-input v-model.number="deviceSettings.fuel_consumption.fuel_economy" label="Fuel Economy"
							type="number" />
					</div>
				</div>

				<!-- Additional Settings -->
				<div class="text-h6">Additional Settings</div>
				<div class="row q-col-gutter-md">
					<div class="col-12 col-md-6">
						<q-input v-model="deviceSettings.initial_device_point_delete_cutoff_time"
							label="Initial Device Point Delete Cutoff Time" type="text" />
						<q-input v-model="deviceSettings.engine_hours_counter_config"
							label="Engine Hours Counter Config" />
						<q-toggle v-model="deviceSettings.use_v3_engine_hours" label="Use V3 Engine Hours" />
						<q-input v-model.number="deviceSettings.history_retention_days" label="History Retention Days"
							type="number" />
					</div>
				</div>
			</q-form>
		</div>
	</q-page>
</template>
<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router';
import { ref, watch, onMounted, computed } from 'vue';
import { useDeviceStore } from 'src/stores/deviceStore';
import IconUploader from 'components/iconUploader.vue'; // Import the IconUploader
import { storeToRefs } from 'pinia';
import { Notify } from 'quasar';
import { QForm } from 'quasar';

const router = useRouter();
const route = useRoute();
const deviceStore = useDeviceStore();

const deviceId = ref<string | null>(null);
const { deviceSettings, deviceSettingsLoaded, deviceSettingsLoading } = storeToRefs(deviceStore);
const isSaving = ref(false);
const formRef = ref<QForm | null>(null);
const isUploading = ref(false);

const DEFAULT_ICON_URL = 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-blue.png';

const currentIconUrl = computed(() => {
	if (!deviceSettings.value?.iconUrl) {
		return DEFAULT_ICON_URL;
	}
	return deviceSettings.value.iconUrl.startsWith('http')
		? deviceSettings.value.iconUrl
		: `https://${deviceSettings.value.iconUrl}`;
});

onMounted(() => {
	//Set device ID for data fetching
	deviceId.value = route.params.id as string;
	loadDeviceSettings();
});

watch(() => route.params.id, () => {
	deviceId.value = route.params.id as string
	loadDeviceSettings();
})

// Fetch device settings when component is mounted or route param ID changes.
const loadDeviceSettings = async () => {
	if (!deviceId.value) return;
	deviceSettings.value = null;
	deviceSettingsLoaded.value = false;
	try {
		await deviceStore.fetchDeviceSettings(deviceId.value);

	} catch (error) {
		// Handle error appropriately
		console.error('Failed to load device settings:', error);
	}
};



const goBack = () => {
	router.push('/');
};

const saveSettings = async () => {
	if (!deviceSettings.value) return;

	try {
		isSaving.value = true;
		await deviceStore.saveDeviceSettings(deviceSettings.value);
		goBack();
	} catch (error) {
		console.error('Failed to save settings:', error);
		//error already caught in store, no need to display another one
	} finally {
		isSaving.value = false;
	}
};

// Event Handlers from iconUploader. 

const handleIconUpload = async (file: File) => {
	if (!deviceId.value || !file) return;

	isUploading.value = true;
	try {
		await deviceStore.updateIcon(deviceId.value, file, null);
	} catch (error) {
		console.error('Failed to upload icon:', error);
		Notify.create({
			type: 'negative',
			message: 'Failed to upload icon'
		});
	} finally {
		isUploading.value = false;
	}
};

const handleRemoveIcon = async () => {
	if (!deviceId.value) return;

	try {
		await deviceStore.updateIcon(deviceId.value, null, null);
	} catch (error) {
		console.error('Failed to remove icon:', error);
		Notify.create({
			type: 'negative',
			message: 'Failed to remove icon'
		});
	}
};

const handleDefaultIcon = async (colorOption: { label: string, value: string }) => {
	if (!deviceId.value) return;

	//   console.log('trying to update default icon:', deviceId.value);
	try {
		await deviceStore.updateIcon(deviceId.value, null, defaultURLBuilder(DEFAULT_ICON_URL, colorOption.value));
	} catch (error) {
		console.error('Failed to update default icon:', error);
		Notify.create({
			type: 'negative',
			message: 'Failed to update default icon'
		});
	}
};

function defaultURLBuilder(iconUrl: string, newColor: string): string {
	try {
		// Check if the URL matches the expected pattern
		if (!iconUrl.includes('marker-icon-')) {
			return iconUrl;
		}

		// Extract the base URL and file extension
		const baseUrl = iconUrl.substring(0, iconUrl.lastIndexOf('marker-icon-') + 'marker-icon-'.length);
		const extension = iconUrl.substring(iconUrl.lastIndexOf('.'));

		// Construct the new URL with the desired color
		return `${baseUrl}${newColor.toLowerCase()}${extension}`;
	} catch (error) {
		console.error('Error changing marker color:', error);
		return iconUrl; // Return original URL if there's an error
	}
}

const handleImageError = async () => {
	if (!deviceId.value) return;

	try {
		await deviceStore.updateIcon(deviceId.value, null, null);
	} catch (error) {
		console.error('Failed to handle image error:', error);
	}
};

</script>

<style lang="scss" scoped>
.form-container {
	display: flex;
	justify-content: center;
	padding: 24px;
	background-color: #f5f5f5;
}

.form-wrapper {
	width: 100%;
	max-width: 600px;
	background: white;
	border-radius: 8px;
	box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.header-section {
	padding: 16px;
	border-bottom: 1px solid #e0e0e0;
}

.settings-form {
	padding: 24px;
}

.form-section {
	margin-bottom: 32px;

	&:last-child {
		margin-bottom: 0;
	}
}

.section-title {
	font-size: 1.1rem;
	font-weight: 500;
	color: #333;
	margin-bottom: 16px;
	padding-bottom: 8px;
	border-bottom: 1px solid #eee;
}

.form-field {
	margin-bottom: 16px;
	max-width: 400px;

	&:last-child {
		margin-bottom: 0;
	}
}

// Responsive adjustments
@media (max-width: 600px) {
	.form-container {
		padding: 16px;
	}

	.form-wrapper {
		border-radius: 0;
	}

	.settings-form {
		padding: 16px;
	}
}
</style>