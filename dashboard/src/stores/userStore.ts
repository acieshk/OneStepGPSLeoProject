// src/stores/userStore.ts
import { defineStore } from 'pinia';
import { useQuasar } from 'quasar';
import { apiService } from 'src/api/apiService';
import { UserPreferences } from 'src/model/model';
import { ref } from 'vue';



export const useUserStore = defineStore('user', () => {
	const $q = useQuasar();
	const userLoaded = ref(false);
	const userLoading = ref(false);
	const userId = ref('default_user');
	const userPreferences = ref<UserPreferences>({
		rowPerPage: 20,
		DeviceListWidth: 200, // Default width in pixels
		unit: 'original',
	});

	async function loadUser() {  // Takes userId as argument
		if (userLoading.value) return;
		userLoading.value = true;
        try {
            const preferences = await apiService.getUserPreferences('default_user'); // the user is default_user right now
            userPreferences.value = preferences; // Update store state
			userLoaded.value = true; 
			userLoading.value = false; 
			console.log('User loaded');
			console.log(preferences);
        } catch (error) {
            // Use Quasar notification for error display
            $q.notify({
                type: 'negative', // Use a negative type for errors
                message: 'Error loading user preferences. Please try again later.', // User-friendly message
                position: 'top', // Or other preferred position
            });
        }
    }

	async function saveUserPreferences() {
        try {
			console.log('Saving preferences');
			console.log(userPreferences.value);
            const updatedPreferences = await apiService.saveUserPreferences(userPreferences.value);
            // Update the store with the saved preferences (if the API returns them)
			userPreferences.value = updatedPreferences; // Assuming server sends back updated preferences

        } catch (error) {
            $q.notify({
                type: 'negative', // Use a negative type for errors
                message: 'Error saving preferences. Please try again later.', // User-friendly message
                position: 'top', // Or other preferred position
            });
        }
    }

	function setRowPerPage(rowPerPage: number) {
		userPreferences.value.rowPerPage = rowPerPage
		saveUserPreferences()
	}

	function setDeviceListWidth(DeviceListWidth: number) {
		userPreferences.value.DeviceListWidth = DeviceListWidth
		saveUserPreferences()
	}

	function setUnit(unit: 'original' | 'metric' | 'imperial') {
		userPreferences.value.unit = unit
		saveUserPreferences()
	}

	const ensurePreferencesLoaded = async () => {
		if (!userLoaded.value) {
		  await loadUser();
		}
		return userPreferences.value;
	};

	return {
		// states
		userId,
		userPreferences,
		userLoading,
		userLoaded,
		// actions
		loadUser,
		saveUserPreferences,
		setRowPerPage,
		setDeviceListWidth,
		setUnit,
		ensurePreferencesLoaded,
	};
});
