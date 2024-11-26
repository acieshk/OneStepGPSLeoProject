<template>
	<el-form :model="preferences" label-width="120px">
		<el-form-item label="User ID">
			<el-input v-model="preferences.userId" disabled />
		</el-form-item>
		<el-form-item label="Distance Unit">
			<el-select v-model="preferences.distanceUnit">
				<el-option label="Kilometers" value="km" />
				<el-option label="Miles" value="mi" />
			</el-select>
		</el-form-item>
		<el-form-item label="Layout">
			<el-select v-model="preferences.layout">
				<el-option label="Horizontal" value="horizontal" />
				<el-option label="Vertical" value="vertical" />
			</el-select>
		</el-form-item>
		<el-form-item>
			<el-button type="primary" @click="savePreferences">Save</el-button>
			<el-button @click="goBack">Back</el-button>
		</el-form-item>
	</el-form>
</template>


<script setup lang="ts">
import { reactive, onMounted, provide, watch, ref } from 'vue';
import { useUserStore } from '@/stores/userStore';
import { apiService } from '@/services/api.service';
import type { UserPreferences } from '@/types/userPreferences';
import { ElMessage } from 'element-plus';
import { configService } from '@/services/config.service';
import { useRouter } from 'vue-router';
const router = useRouter();
const goBack = () => {
  router.push('/devices');  // Navigate to the devices route
};

const userStore = useUserStore();

// Provide the layout preference (assuming it's a string 'horizontal' or 'vertical')
const layoutPreference = ref(userStore.layoutPreference); // Initialize with current preference

provide('layoutPreference', layoutPreference); // Provide to descendant components

// Watch for changes in the Pinia store (if that's where layout preference is)
watch(() => userStore.layoutPreference, (newLayout) => {
    layoutPreference.value = newLayout; // Update the provided value when store changes
});

// Function to fetch user preferences
const fetchPreferences = async () => {

	try {
		const fetchedPreferences = await apiService.getUserPreferences(userId.value);
		// Update local reactive object
		Object.assign(preferences, fetchedPreferences);

		// Update the Pinia store
		userStore.updateUserPreferences(fetchedPreferences);
	} catch (error) {
		// Handle the error, e.g., display a message to the user
		console.error('Failed to fetch user preferences:', error);
	}
};



const savePreferences = async () => { // update preferences in the pinia store and call save function
	try {
		await apiService.saveUserPreferences(preferences); // Save changes first
		ElMessage({
			message: 'Preferences saved successfully.',
			type: 'success',
		});


	}
	catch (error) {
		// Handle error and provide feedback to the user
		ElMessage.error('Failed to save preferences. Please try again.');
		console.error('Error saving user preferences:', error);
	}

};


const userId = ref(configService.getConfig().userId);  

const preferences = reactive<UserPreferences>({
  userId: userId.value,          
  distanceUnit: 'km',     // Default distance unit
  layout: 'horizontal',   // Default layout
});


onMounted(fetchPreferences); // Fetch preferences when component mounts



watch(() => preferences, (newVal) => {
	userStore.updateUserPreferences(newVal);

}, { deep: true });



</script>