<!--
	MainLayout.vue
	Handle Drawer logic
-->
<template>
	<q-layout view="hHh Lpr lFf" class="main-layout">
		<q-header elevated>
			<q-toolbar>
				<q-btn flat dense round icon="menu" aria-label="Menu" @click="toggleLeftDrawer" />

				<q-toolbar-title class="page-title" @click="backToMain">
					Device Dashboard
				</q-toolbar-title>

				<div class="right-controls">
					<q-btn flat label="Refresh Database" @click="refreshDatabase" />
					<q-btn round dense icon="person" @click="goToUserPage" />
				</div>
			</q-toolbar>
		</q-header>

		<q-drawer v-model="leftDrawerOpen" :width="drawerWidth" class="resizable-drawer" side="left" bordered
			:behavior="'desktop'" :breakpoint="500" :mini="false" :overlay="false">
			<q-list>
				<device-list />
			</q-list>
			<div class="resize-handle" @mousedown="startResizing">
				<div class="resize-knob"></div>
			</div>
		</q-drawer>

		<q-page-container class="page-container">
			<router-view />
		</q-page-container>
	</q-layout>
	<q-dialog v-model="showRefreshDialog" persistent>
		<q-card>
			<q-card-section>
				<div class="text-h6">Confirm Refresh</div>
				<div class="text-body2">Are you sure you want to refresh the database?</div>
			</q-card-section>

			<q-card-actions align="right">
				<q-btn flat label="Cancel" color="primary" v-close-popup />
				<q-btn flat label="OK" color="primary" :loading="isLoading" @click="onConfirmRefreshDatabase">
					<template v-slot:loading>
						<q-spinner-hourglass class="on-left" />
						Loading...
					</template>
				</q-btn>
			</q-card-actions>
		</q-card>
	</q-dialog>
</template>

<script setup lang="ts">
import DeviceList from 'components/deviceList.vue';
import { storeToRefs } from 'pinia';
import { debounce, useQuasar } from 'quasar';
import { apiService } from 'src/api/apiService';
import { useDeviceStore } from 'src/stores/deviceStore';
import { useUserStore } from 'src/stores/userStore';
import { onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

defineOptions({
	name: 'MainLayout'
});

/*
	Buttons control logic
*/
const $q = useQuasar();  // Keep only one instance of useQuasar
const deviceStore = useDeviceStore();
const userStore = useUserStore();
const { userPreferences } = storeToRefs(userStore);
const isLoading = ref(false);
const showRefreshDialog = ref(false);
const router = useRouter();
const refreshDatabase = async () => {
	showRefreshDialog.value = true;
};

const backToMain = () => {
	router.push('/');
};

const onConfirmRefreshDatabase = async () => {
	try {
		isLoading.value = true;
		await apiService.refreshDatabase();
		console.log('Database refreshed successfully');
		await deviceStore.loadDevices();

		showRefreshDialog.value = false;
		router.push('/');
		$q.notify({
			type: 'positive',
			message: 'Database refreshed successfully!',
		});
	} catch (error) {
		console.error('Error refreshing database:', error);
		$q.notify({
			type: 'negative',
			message: 'Failed to refresh database. Please try again later.'
		});
	} finally {
		isLoading.value = false;
	}
}

const goToUserPage = () => {
	router.push('/user'); // Navigate to the user page
};

/*
	Drawer logic 
*/
const leftDrawerOpen = ref(false);
const drawerWidth = ref(400);
let isResizing = false;
const minDrawerWidth = 200;

const getPreferredWidth = () => {
	return userPreferences.value?.DeviceListWidth || 400; // fallback width of 400
};

// Update your toggleLeftDrawer function
const toggleLeftDrawer = () => {
	if (!leftDrawerOpen.value) {
		drawerWidth.value = getPreferredWidth();
	}
	leftDrawerOpen.value = !leftDrawerOpen.value;

	// if (!leftDrawerOpen.value) {
	// 	debouncedUpdateWidth(drawerWidth.value);
	// }
};


const debouncedUpdateWidth = debounce((width: number) => {
	userStore.setDeviceListWidth(width);
}, 300);

let initialWidth = 0;
function startResizing(event: MouseEvent) {
	isResizing = true;
	const startX = event.clientX;
	initialWidth = drawerWidth.value;
	const maxWidth = window.innerWidth * 0.8;

	const onMouseMove = (moveEvent: MouseEvent) => {
		if (!isResizing) return;
		const deltaX = moveEvent.clientX - startX;
		drawerWidth.value = Math.min(
			Math.max(minDrawerWidth, initialWidth + deltaX),
			maxWidth
		);
	};

	const onMouseUp = () => {
		isResizing = false;
		debouncedUpdateWidth(drawerWidth.value);
		document.removeEventListener('mousemove', onMouseMove);
		document.removeEventListener('mouseup', onMouseUp);
	};

	document.addEventListener('mousemove', onMouseMove);
	document.addEventListener('mouseup', onMouseUp);
}

const route = useRoute();
// Add a watch for userPreferences to ensure drawer width stays in sync
watch(() => userPreferences.value?.DeviceListWidth, (newWidth) => {
    if (leftDrawerOpen.value && newWidth) {
        drawerWidth.value = newWidth;
    }
});

watch(
    [() => userStore.userLoaded, () => route.path],
    ([userIsLoaded, currentPath]) => {
        const isMainPage = currentPath === '/' || currentPath === '/map';
        if (userIsLoaded && isMainPage && !leftDrawerOpen.value) {
            try {
                drawerWidth.value = getPreferredWidth();
                leftDrawerOpen.value = true;
            } catch (error) {
                console.error('Error setting drawer width:', error);
                // Fallback to default width
                drawerWidth.value = 400;
                leftDrawerOpen.value = true;
            }
        }
    },
    { immediate: true }
);

onMounted(() => {
	deviceStore.loadDevices();
	userStore.loadUser();
});
</script>
<style scoped>
.page-title {
	cursor: pointer;
}

.page-container {
	height: 100vh;
}

.q-drawer {
    transition: transform 0.3s ease;
}

.resizable-drawer {
	position: relative;
	overflow: hidden;
	transition: width 0.3s ease;
}

.resize-handle {
	position: absolute;
	top: 0;
	right: 0;
	width: 5px;
	height: 100%;
	cursor: ew-resize;
	background-image: linear-gradient(to left, rgba(0, 0, 0, 0.1),
			rgba(0, 0, 0, 0.05) 50%,
			rgba(0, 0, 0, 0.0));

	.resize-knob {
		position: relative;
		top: calc(50% - 50px);
		width: 5px;
		height: 50px;
		border-radius: 2.5px;
		background-color: rgba(0, 0, 0, 0.3);
	}
}

.resize-handle:hover {
	background-color: rgba(25, 118, 210, 0.2);
	;
}

.right-controls {
	display: flex;
	align-items: center;
	gap: 10px;
	/* Adjust spacing as needed */
}
</style>