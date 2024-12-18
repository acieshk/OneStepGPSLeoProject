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
	const userID = ref('default_user');
	const userPreferences = ref<UserPreferences>({
		userId: 'default_user',
		version: 0,
		DeviceListWidth: 400, 
		unit: 'original',
	});

	async function loadUser() {  
		if (userLoading.value) return;
		userLoading.value = true;
        try {
            const preferences = await apiService.getUserPreferences(userID.value); // the user is default_user right now
            userPreferences.value = preferences; 
			console.log(preferences);
			userLoaded.value = true; 
			userLoading.value = false; 
        } catch (error) {
            $q.notify({
                type: 'negative', 
                message: 'Error loading user preferences. Please try again later.',
                position: 'top', 
            });
        }
    }

	async function saveUserPreferences() {
		try {
			console.log(userPreferences.value);
			const updatedPreferences = await apiService.saveUserPreferences(userPreferences.value);
			userPreferences.value = updatedPreferences;
	
		} catch (error) {
			$q.notify({
				type: 'warning',
				message: 'Your client is outdated. Please refresh your browser to get the latest changes.',
				position: 'top',
				timeout: 0,  // 0 means the notification will stay indefinitely
				actions: [
					{
						label: 'Refresh Now',
						color: 'white',
						handler: () => {
							window.location.reload();
						}
					}
				],
				closeBtn: true  // Adds an X button to manually dismiss if needed
			});
		}
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
		userID,
		userPreferences,
		userLoading,
		userLoaded,
		// actions
		loadUser,
		saveUserPreferences,
		setDeviceListWidth,
		setUnit,
		ensurePreferencesLoaded,
	};
});
