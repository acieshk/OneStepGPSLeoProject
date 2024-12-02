<template>
    <div class="icon-upload-container">
        <div class="current-icon" v-if="displayedIconUrl || (!displayedIconUrl && props.deviceId)">
            <span>Customized Icon:</span>
            <img v-if="isCustomIcon" :src="displayedIconUrl" alt="Device Icon" class="icon-display" @error="handleImageError">
            <MapMarkerIcon v-else :deviceId="props.deviceId" :color="iconColor" />
        </div>

        <el-upload
            :show-file-list="false"
            :before-upload="handleBeforeUpload"
            :auto-upload="false"
            :on-change="handleFileChange"
            ref="uploadRef"
        >
            <el-button type="primary" size="small" :disabled="uploading" @click="handleUploadClick">
                <el-icon v-if="!isCustomIcon"><Plus /></el-icon> <span v-if="!isCustomIcon">Upload Icon</span> <span
                    v-else>Replace Icon</span>
            </el-button>
        </el-upload>

    </div>
</template>

<script setup lang="ts">
import { ref, defineEmits, defineProps, nextTick, watch, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { Plus } from '@element-plus/icons-vue'; // Import the Plus icon
import { apiService } from '@/services/api.service';
import { configService } from '@/services/config.service';
import MapMarkerIcon from './MapMarkerIcon.vue'; // Import MapMarkerIcon
import type { Device } from '@/types/device';

const uploadRef = ref();
const uploading = ref(false);
const defaultIconUrl = ref(new URL('/src/assets/default.png', import.meta.url).href); // Provide default path
const displayedIconUrl = ref(defaultIconUrl.value); // Use a separate ref for display
const isCustomIcon = ref(false);

const props = defineProps<{
    deviceId: string;
	device: Device | null;
    currentIconUrl: string | null;
    iconColor: string; 
}>();

// Call getIcon and set displayedIconUrl on mount and when deviceId changes:
onMounted(async () => {
    await loadIcon();
	console.log(props.iconColor);
});

watch(() => props.deviceId, async (newDeviceId, oldDeviceId) => {
    // Only load icon if deviceId changes AND there's a new deviceId:
    if (newDeviceId && newDeviceId !== oldDeviceId) {
        console.log("deviceId changed:", props.deviceId);
        await loadIcon();
    }
});




const loadIcon = async () => {
    if (props.deviceId) {
        const iconUrl = await apiService.getIcon(props._id);
        if (iconUrl) {
            displayedIconUrl.value = iconUrl;
            isCustomIcon.value = true;
        } else {
            displayedIconUrl.value = null;  // Or '' if you prefer, but null is generally clearer
            isCustomIcon.value = false;
        }
    }
};



const currentIconUrl = ref(props.currentIconUrl);

watch(() => props.currentIconUrl, (newUrl) => {
	currentIconUrl.value = newUrl;
});

const emit = defineEmits(['icon-uploaded']);

const handleSuccess = (response) => {
	console.log('Icon uploaded successfully:', response);
	emit('icon-uploaded', response.iconUrl); // Emit the new icon URL
	ElMessage({
		message: 'Icon uploaded successfully.',
		type: 'success',
	});
};

const handleImageError = (event) => { // Function to display default icon
    console.log("Failed to load custom icon. Displaying default icon.");
	displayedIconUrl.value = defaultIconUrl.value;
};

const handleError = (err) => {
	console.error('Icon upload failed:', err);
	ElMessage.error('Icon upload failed. Please try again.');
};

const handleBeforeUpload = (file) => {
	const isJPG = file.type === 'image/jpeg';
	const isPNG = file.type === 'image/png';
	const isLt2M = file.size / 1024 / 1024 < 2;

	if (!isJPG && !isPNG) {
		ElMessage.error('Upload image can only be JPG/PNG format!');
	}
	if (!isLt2M) {
		ElMessage.error('Upload image size can not exceed 2MB!');
	}
	return (isJPG || isPNG) && isLt2M; // Must return a value!
};


const handleFileChange = async (uploadFile, fileList) => {
	if (uploadFile && props.deviceId) {
		try {
			uploading.value = true;
			const response = await apiService.uploadDeviceIcon(props.deviceId, uploadFile.raw);

			currentIconUrl.value = `${configService.getApiUrl()}/icons/${props.deviceId}.png`; // Update icon URL

			emit('icon-uploaded', currentIconUrl.value); // Emit the updated URL

			ElMessage.success('Icon uploaded successfully.');

		} catch (error) {
			console.error('Icon upload failed:', error);
			ElMessage.error('Icon upload failed. Please try again.');
		} finally {
			uploading.value = false;
			(uploadRef.value as any).clearFiles(); // Clear the selected files after upload
			await loadIcon(); 
		}
	}
};



const handleUploadClick = () => {  //  Only triggers the file picker
	(uploadRef.value as any).handleRemove((uploadRef.value as any).uploadFiles[0]); // Clear the upload component's file list
	(uploadRef.value as any).handleClick();

};

</script>
<style scoped>
..icon-upload-container {
	display: flex;
	align-items: center;
	/* Align items vertically */
	gap: 10px;
	/* Add some space between icon and uploader */
}


.current-icon {
	display: flex;
	align-items: center;
	/* Vertically align icon and label */
}

.icon-display {
	width: 32px;
	/* Set desired width */
	height: 32px;
	/* Set desired height */
	margin-left: 5px;
	/* Small margin between label and icon */
}

.icon-upload-container:deep(.el-upload) {
	/* Style the upload button container */
	display: inline-flex;
	/* Use inline-flex to prevent button from stretching full width */
	align-items: center;
	/* Vertically center button content */
}


.uploaded-icon {
	margin-left: 10px;
	display: inline-flex;
	/* Use inline-flex for proper alignment with text*/
	align-items: center;
	/* Align items vertically */
}

.uploaded-icon img {
	width: 32px;
	height: 32px;
	margin-right: 5px;
	/* Add a little space between the icon and the text */

}
</style>