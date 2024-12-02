<template>
	<q-form v-if="!isLoading" class="user-form" @submit="saveUserPreferences">
		<q-toolbar>
			<q-toolbar-title>User Preferences</q-toolbar-title>
		</q-toolbar>
		<div class="q-gutter-md">
			<q-select style="display:none" v-model="userPreferences.rowPerPage" :options="[20, 50, 100]" label="Rows per page" emit-value
				map-options />

			<q-input v-model="userPreferences.DeviceListWidth" label="Device List Width (px)" disable type="number" />

			<div>Unit:</div>
			<q-radio v-model="userPreferences.unit" val="original" label="Original" />
			<q-radio v-model="userPreferences.unit" val="metric" label="Metric" />
			<q-radio v-model="userPreferences.unit" val="imperial" label="Imperial" />

			<q-btn label="Save" type="submit" color="primary" />
		</div>
	</q-form>
	<q-spinner v-else />
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { QForm, QSelect, QInput, QRadio, QBtn, useQuasar } from 'quasar';
import { useUserStore } from 'src/stores/userStore';
import { storeToRefs } from 'pinia';

const userStore = useUserStore();
const router = useRouter();
const $q = useQuasar();
const isLoading = ref(false);

// Create a reactive object for the form data, linked to the store
const { userPreferences } = storeToRefs(userStore);

const saveUserPreferences = async () => {
  try {
    await userStore.saveUserPreferences();
    $q.notify({
      message: 'User preferences updated successfully!',
      type: 'positive',
      position: 'top',
    });
    router.push('/');
  } catch (error) {
	console.error(error);
  }
};

// Watch only the loading state
watch(
    () => userStore.userLoaded, // Watch userLoaded directly
    (loaded) => {
        isLoading.value = !loaded; // Set isLoading to the opposite of loaded
    },
    { immediate: true } // Run immediately on mount
);
onMounted(async () => {
  await userStore.ensurePreferencesLoaded();
});

</script>
<style scoped>
.user-form {
	max-width: 400px;
	margin: 20px auto;
	padding: 20px;
	border: 1px solid #ddd;
	border-radius: 5px;
}

.user-form .q-field {
	margin-bottom: 16px;
}

.user-form .q-btn {
	width: 100%;
}

.user-form .q-radio {
	margin-bottom: 8px;
}

@media (max-width: 500px) {
	.user-form {
		max-width: 90%;
	}
}
</style>