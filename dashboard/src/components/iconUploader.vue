<template>
	<div class="icon-uploader q-mb-md">
		<div>
			<p> Icon Uploader </p>
		</div>
		<div v-if="deviceSettings" class="row items-center q-gutter-md">
			<div class="icon-preview">
				<q-img :src="currentIconUrl" style="width: 48px; height: 48px;" fit="contain"
					@error="handleImageError" />
			</div>

			<div class="icon-controls">
				<q-select v-if="deviceSettings" v-model="selectedDefaultColor" :options="defaultColorOptions"
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
					:disable="isUploading" @update:model-value="handleIconUpload" style="max-width: 250px">
					<template v-slot:prepend>
						<q-icon name="upload" />
					</template>
					<template v-slot:append>
						<q-spinner v-if="isUploading" color="primary" size="sm" />
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
import { ref, watch } from 'vue';
import { DeviceSettings } from 'src/model/model';


//Using props to ensure consisntency between device Edit and uploader
const props = defineProps<{
  deviceSettings: DeviceSettings | null;

}>();
const emit = defineEmits<{
    (e: 'update:deviceSettings', settings: DeviceSettings): void;
    (e: 'upload', file: File):void,
	(e: 'remove'): void,
	(e: 'defaultIcon', colorOption: {label: string, value: string}): void,
	(e: 'imageError'): void;
}>()

const selectedDefaultColor = ref<string>('Blue')
// const isSaving = ref(false);
const iconFile = ref<File | null>(null)
const isUploading = ref(false);
const DEFAULT_ICON_PREFIX = 'raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-';
const DEFAULT_ICON_COLOR = 'blue';
const defaultColorOptions = [
	{ label: 'Blue', value: 'blue' },
	{ label: 'Gold', value: 'gold' },
	{ label: 'Red', value: 'red' },
	{ label: 'Green', value: 'green' },
	{ label: 'Orange', value: 'orange' },
	{ label: 'Violet', value: 'violet' },
	{ label: 'Grey', value: 'grey' },
	{ label: 'Black', value: 'black' },
];


const getMarkerUrl = (color: string) => `${DEFAULT_ICON_PREFIX}${color}.png`;
const DEFAULT_ICON_URL = getMarkerUrl(DEFAULT_ICON_COLOR);


function formatURL(url: string | null | undefined): string {
	if (url && url !== '') {
		if (!url.startsWith('http://') && !url.startsWith('https://')) {
			return `https://${url}`; // This might be causing issues locally
		} else {
			return url;
		}
	}
	return `https://${DEFAULT_ICON_URL}`;
}

const currentIconUrl = ref(formatURL(null))
const getDefaultColorFromURL = (iconURL: string | undefined | null) => {
    try {
        if (!iconURL) {
            return DEFAULT_ICON_COLOR.charAt(0).toUpperCase() + DEFAULT_ICON_COLOR.slice(1);
        }

        const formattedURL = formatURL(iconURL);
        const url = new URL(formattedURL);
        const pathname = url.pathname;
        const filename = pathname.substring(pathname.lastIndexOf('/') + 1);
        
        if (filename.includes('marker-icon-')) {
            const color = filename.substring('marker-icon-'.length, filename.lastIndexOf('.'));
            return color.charAt(0).toUpperCase() + color.slice(1);
        }
        
        return DEFAULT_ICON_COLOR.charAt(0).toUpperCase() + DEFAULT_ICON_COLOR.slice(1);
    } catch (error) {
        console.error('Error parsing icon URL:', error);
        return DEFAULT_ICON_COLOR.charAt(0).toUpperCase() + DEFAULT_ICON_COLOR.slice(1);
    }
};

watch(() => props.deviceSettings?.iconUrl, (newIconUrl) => {
	console.log(props.deviceSettings?.iconUrl);
    try {
        if (newIconUrl !== undefined) {
            currentIconUrl.value = formatURL(newIconUrl);
            selectedDefaultColor.value = getDefaultColorFromURL(newIconUrl);
        } else {
            currentIconUrl.value = formatURL(null);
        }
    } catch (error) {
        console.error('Error in iconUrl watcher:', error);
        // Set to default values if there's an error
        currentIconUrl.value = formatURL(null);
        selectedDefaultColor.value = DEFAULT_ICON_COLOR.charAt(0).toUpperCase() + DEFAULT_ICON_COLOR.slice(1);
    }
}, { immediate: true });


const handleIconUpload = (file: File | null) => {
	if (!file) return;
    emit('upload', file);

};


const removeIcon = () => {
    emit('remove')
};


const updateDefaultIcon = (colorOption: { label: string; value: string }) => {
	emit('defaultIcon', colorOption);
}


const handleImageError = () => {
    emit('imageError');
};

//Set selectedDefaultColor based on deviceSettings iconUrl when component is mounted,
//or when deviceSettings changes. No need to fetch default color when image fails to load.
watch(() => props.deviceSettings, (newSettings) => {
    if(newSettings) {
        selectedDefaultColor.value = getDefaultColorFromURL(newSettings.iconUrl);
    }
}, {deep:true, immediate: true})

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
	border-top-right-radius: 0px;
	;
	border-radius: 0 0 8px 8px;
	background: #f9f9f9;
}

.icon-preview {
	margin-right: 20px;
	border: 1px solid #ddd;
	border-radius: 4px;
	overflow: hidden;
	/* Prevents image overflow */
}

.icon-controls,
.custom-icon-controls {
	display: flex;
	align-items: center;
	gap: 12px;
	/* Spacing between controls */
}


.icon-uploader .q-file {
	max-width: 200px;
}


/* Style the remove button */
.q-btn.q-ml-sm {
	border: 1px solid #ddd;
	border-radius: 4px;
	padding: 4px 8px;
	background: #FFEBEE;
	/* negative color */
}


.q-btn.q-ml-sm:hover {
	background-color: #FFCDD2;
	/* Slightly darker on hover */
}


/* Responsive adjustments */
@media (max-width: 600px) {
	.icon-uploader {
		flex-direction: column;
		align-items: stretch;
		/* Fill width on smaller screens */
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