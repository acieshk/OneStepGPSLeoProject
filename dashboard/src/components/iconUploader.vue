<template>
	<div class="icon-uploader q-mb-md">
		<div>
			<p> Icon Uploader </p>
		</div>
		<div v-if="editingDevice" class="row items-center q-gutter-md">
			<div class="icon-preview">
				<q-img :src="currentIconUrl" style="width: 48px; height: 48px;" fit="contain"
					@error="handleImageError" />
			</div>

			<div class="icon-controls">
				<q-select v-if="showDefaultIconSelector" v-model="selectedDefaultColor" :options="defaultColorOptions"
					label="Select default marker" outlined dense style="min-width: 200px"
					@update:model-value="updateDefaultIcon">
					<template v-slot:option="{ itemProps, opt }">
						<q-item v-bind="itemProps">
							<q-item-section avatar>
								<q-img :src="'http://' + getMarkerUrl(opt.value)" style="width: 24px" />
							</q-item-section>
							<q-item-section>
								<q-item-label>{{ opt.label }}</q-item-label>
							</q-item-section>
						</q-item>
					</template>
				</q-select>
			</div>

			<div class="icon-controls custom-icon-controls">
				<q-file v-model="iconFile" label="Upload custom icon" outlined dense accept=".png, .jpg, .jpeg, .gif"
					@update:model-value="handleIconUpload" style="max-width: 250px">
					<template v-slot:prepend>
						<q-icon name="upload" />
					</template>
				</q-file>
				<q-btn v-if="currentIconUrl && currentIconUrl !== 'https://' + DEFAULT_ICON_URL" flat dense
					color="negative" icon="delete" @click="removeIcon" class="q-ml-sm" />
			</div>
		</div>
		<div v-else class="loading-container">
			<q-spinner-dots color="primary" size="md" />
		</div>
	</div>
</template>


<script setup lang="ts">
import { storeToRefs } from 'pinia';
import { useDeviceStore } from 'src/stores/deviceStore';
import { computed, ref, watch } from 'vue';
import { useQuasar } from 'quasar'; 

const $q = useQuasar();
const deviceStore = useDeviceStore();
const { editingDevice } = storeToRefs(deviceStore);

const iconFile = ref<File | null>(null);
const selectedDefaultColor = ref('Blue');
const defaultColorOptions = [
	{ label: 'Blue(Default)', value: 'Blue' },
	{ label: 'Gold', value: 'Gold' },
	{ label: 'Red', value: 'Red' },
	{ label: 'Green', value: 'Green' },
	{ label: 'Orange', value: 'Orange' },
	{ label: 'Yellow', value: 'Yellow' },
	{ label: 'Violet', value: 'Violet' },
	{ label: 'Grey', value: 'Grey' },
	{ label: 'Black', value: 'Black' }
];

const DEFAULT_ICON_COLOR = 'blue'; // Store default color name
const DEFAULT_ICON_PREFIX = 'raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-';


const getMarkerUrl = (color: string) => {
	return `${DEFAULT_ICON_PREFIX}${color.toLowerCase()}.png`;
};

// To make sure the url is correct
const formatURL = (url: string | null | undefined) => {
	if (!url) return '';

	if (typeof url !== 'string') {
		console.warn(`formatURL received a non-string value: ${typeof url}, ${url}`);
		return '';
	}
	console.log(url);
	if (url.startsWith(DEFAULT_ICON_PREFIX)) { // Check if it's a default icon
		return `https://${url}`; // Add https if default icon
	} else if (url.startsWith('http://') || url.startsWith('https://')) {
		return url; // Already has protocol
	} else if (url !== '') { // Check if url is not an empty string.  Only prepend if not empty.
		return `http://${url}`;  // For custom icons, prepend the protocol
	} else {
		return ''; // Return empty for no URL
	}
};

const DEFAULT_ICON_URL = getMarkerUrl(DEFAULT_ICON_COLOR); // Use function with default color

const currentIconUrl = computed(() => {
	if (!editingDevice.value) return `https://${DEFAULT_ICON_URL}`; // Default URL with https

	const iconUrl = editingDevice.value.iconUrl;
	if (!iconUrl) return `https://${DEFAULT_ICON_URL}`; // Use default if no URL

	if (iconUrl.startsWith(DEFAULT_ICON_PREFIX)) {
		return `https://${iconUrl}`; // Add https for default icons
	} else if (iconUrl.startsWith('http://') || iconUrl.startsWith('https://') || iconUrl === '') {
		return iconUrl; // Already a full URL or empty
	} else {
		return formatURL(iconUrl); // Custom icon, use formatURL
	}
});

const updateDefaultIcon = (colorOption: { label: string; value: string }) => {
	if (editingDevice.value) {
		deviceStore.updateDeviceProperty(['iconUrl'], getMarkerUrl(colorOption.value));
	}
};

const handleImageError = () => {
	if (editingDevice.value) {
		deviceStore.updateDeviceProperty(['iconUrl'], '');
		selectedDefaultColor.value = 'Blue';
	}
};

const handleIconUpload = async (file: File | null) => {
	if (!file || !editingDevice.value) return;

	if (editingDevice.value?._id) {
		deviceStore.updateIcon(editingDevice.value._id, file, null); // Call updateIcon with the File object
	}
};

const removeIcon = () => {
	if (!editingDevice.value || !editingDevice.value._id) return;
	try {
		deviceStore.updateIcon(editingDevice.value._id, null, '').then(() => {
			iconFile.value = null; 
			selectedDefaultColor.value = 'blue';
			$q.notify({ 
				type: 'positive',
				message: 'Icon removed successfully!'
			});
		});

	} catch (error) {
		console.error('Failed to remove icon:', error);
	}
};

const showDefaultIconSelector = computed(() => {
	if (!editingDevice.value) return false; // Hide if no device is being edited
	console.log(editingDevice.value.iconUrl);
	return !editingDevice.value.iconUrl || isDefaultIcon(editingDevice.value.iconUrl);
});

const isDefaultIcon = (iconUrl:string) => {
	console.log(iconUrl);
	console.log((iconUrl.startsWith(DEFAULT_ICON_PREFIX)));
	return (iconUrl.startsWith(DEFAULT_ICON_PREFIX)) || iconUrl.startsWith('https://' + DEFAULT_ICON_PREFIX)
	|| iconUrl == '';
};

// Watch editingDevice to reset icon controls when device changes
watch(editingDevice, (newDevice) => {
	if (!newDevice) {
		iconFile.value = null; // Clear file upload
		selectedDefaultColor.value = 'Blue'; // Reset selected color
	}
});

</script>
<style scoped>
.icon-uploader {
	display: flex;
	flex-direction: column;
	align-items: flex-start;
	padding: 16px;
	border: 1px solid #ddd; 
	border-top: 0px;
	border-top-left-radius: 0px;
	border-top-right-radius: 0px;;
	border-radius: 0 0 8px 8px;
	background: #f9f9f9; 
}

.icon-preview {
	margin-right: 20px;
	border: 1px solid #ddd;
	border-radius: 4px;
	overflow: hidden; /* Prevents image overflow */
}

.icon-controls,
.custom-icon-controls {
	display: flex;
	align-items: center;
	gap: 12px; /* Spacing between controls */
}


.icon-uploader .q-file {
	max-width: 200px;
}


/* Style the remove button */
.q-btn.q-ml-sm {
	border: 1px solid #ddd;
	border-radius: 4px;
	padding: 4px 8px;
	background: #FFEBEE; /* negative color */
}


.q-btn.q-ml-sm:hover {
	background-color: #FFCDD2; /* Slightly darker on hover */
}


/* Responsive adjustments */
@media (max-width: 600px) {
	.icon-uploader {
		flex-direction: column;
		align-items: stretch; /* Fill width on smaller screens */
	}

	.icon-controls,
	.custom-icon-controls {
		flex-direction: column;
		align-items: stretch;
		margin-top: 12px;
		gap: 8px;
	}
}
</style>