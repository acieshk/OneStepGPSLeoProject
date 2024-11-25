<template>
	<div class="icon-upload-container">
		<el-upload :action="uploadUrl" :on-success="handleSuccess" :on-error="handleError" :show-file-list="false"
			:before-upload="handleBeforeUpload" :headers="{ 'Content-Type': 'multipart/form-data' }">
			<el-button type="primary" size="small" :disabled="!uploadUrl">
				<el-icon v-if="!currentIconUrl">
					<Plus />
				</el-icon> </el-button>
			<span v-if="!currentIconUrl">Upload Icon</span>
			<div v-else class="uploaded-icon"> </div>
			<img :src="currentIconUrl" alt="Device Icon"> </img>
			<span>Replace Icon</span>
		</el-upload>
	</div>

</template>

<script setup lang="ts">
import { ref, defineEmits, defineProps } from 'vue';
import { ElMessage } from 'element-plus';
import { Plus } from '@element-plus/icons-vue'; // Import the Plus icon

const props = defineProps<{
	uploadUrl: string | null;
	currentIconUrl: string | null;
}>();

const emit = defineEmits(['icon-uploaded']);

const handleSuccess = (response) => {
	console.log('Icon uploaded successfully:', response);
	emit('icon-uploaded', response.iconURL); // Emit the new icon URL
	ElMessage({
		message: 'Icon uploaded successfully.',
		type: 'success',
	});
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
</script>
<style scoped>
.icon-upload-container {
  display: flex;
  justify-content: flex-start; /* Align items to the left */
  align-items: center;        /* Vertically center items */
  margin-bottom: 10px;       /* Add spacing below the container */
  padding: 10px;             /* Add some padding around content */
  border: 1px solid #ebeef5; /* Subtle border for visual separation */
  border-radius: 4px;        /* Rounded corners */
}

.icon-upload-container:deep(.el-upload) { /* Style the upload button container */
  display: inline-flex; /* Use inline-flex to prevent button from stretching full width */
  align-items: center; /* Vertically center button content */
}


.uploaded-icon {
  margin-left: 10px;
  display: inline-flex;  /* Use inline-flex for proper alignment with text*/
  align-items: center; /* Align items vertically */
}

.uploaded-icon img {
  width: 32px;
  height: 32px;
  margin-right: 5px;  /* Add a little space between the icon and the text */

}


</style>