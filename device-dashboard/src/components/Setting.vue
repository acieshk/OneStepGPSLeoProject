// src/components/Settings.vue
<template> </template>
    <el-form :model="settingsForm" label-width="120px">
        <el-form-item label="Distance Unit">  </el-form-item>

        <el-form-item label="Layout">  </el-form-item>

        <el-form-item>  </el-form-item>
            <el-button type="primary" @click="savePreferences">Save</el-button>
    </el-form>



<script setup lang="ts">
import { reactive, ref, watch } from 'vue';
import { useUserStore } from '@/stores/userStore'; 
import { apiService } from '@/services/api.service';
import type { UserPreferences } from '@/types/userPreferences';
import { ElMessage } from 'element-plus';
import { configService } from '@/services/config.service'; // Import configService




const userStore = useUserStore();

const settingsForm = reactive<UserPreferences>({
	userId: configService.getConfig().userId, // Get userId from config
	distanceUnit: userStore.userPreferences.distanceUnit,
	layout: userStore.userPreferences.layout,
});



watch(() => userStore.userPreferences, (newVal) => { // watcher
    Object.assign(settingsForm, newVal)

}, {deep: true});



const savePreferences = async () => { // update preferences in the pinia store and call save function
    try {


        const updatedPreferences = await apiService.saveUserPreferences(settingsForm);

        userStore.updateUserPreferences(updatedPreferences);  // Updates the Pinia store

    ElMessage({
            message: 'Preferences saved successfully.',
            type: 'success',
          });

    } catch (error) {

        // Handle error and provide feedback to the user
        ElMessage.error('Failed to save preferences. Please try again.');
        console.error('Error saving user preferences:', error);
    }

};





</script>