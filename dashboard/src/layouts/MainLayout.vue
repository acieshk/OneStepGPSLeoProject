<!--
	MainLayout.vue
	Handle Drawer logic
-->
<template>
	<q-layout view="hHh Lpr lFf" class="main-layout">
		<q-header elevated>
			<q-toolbar>
				<q-btn flat dense round icon="menu" aria-label="Menu" @click="toggleLeftDrawer" />

				<q-toolbar-title>
					Device Dashboard
				</q-toolbar-title>

				<div>Right side</div>
				<div class="right-controls">
					<q-btn flat label="Refresh Database" @click="refreshDatabase" />
					<q-btn round dense icon="person" />
				</div>
			</q-toolbar>
		</q-header>

		<q-drawer v-model="leftDrawerOpen" :width="drawerWidth" class="resizable-drawer" side="left" bordered>
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
import { useQuasar } from 'quasar';
import { apiService } from 'src/api/apiService';
import { useDeviceStore } from 'src/stores/deviceStore';
import { onMounted, ref, watch } from 'vue';

const $q = useQuasar();  // Keep only one instance of useQuasar
const deviceStore = useDeviceStore();
const isLoading = ref(false);
const showRefreshDialog = ref(false);

const refreshDatabase = async () => {
	showRefreshDialog.value = true;
};

const onConfirmRefreshDatabase = async () => {
	try {
		isLoading.value = true;
		await apiService.refreshDatabase();
		console.log('Database refreshed successfully');
		await deviceStore.loadDevices();

		showRefreshDialog.value = false;
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

defineOptions({
	name: 'MainLayout'
});

const { mapReady } = storeToRefs(deviceStore);
const leftDrawerOpen = ref(false);
const drawerWidth = ref(300);
let isResizing = false;
const minDrawerWidth = 200;

function toggleLeftDrawer() {
	leftDrawerOpen.value = !leftDrawerOpen.value;
}

function startResizing(event: MouseEvent) {
	isResizing = true;
	const startX = event.clientX;
	const startWidth = drawerWidth.value;

	const onMouseMove = (moveEvent: MouseEvent) => {
		if (!isResizing) return;
		const deltaX = moveEvent.clientX - startX;
		drawerWidth.value = Math.max(minDrawerWidth, startWidth + deltaX);
	};

	const onMouseUp = () => {
		isResizing = false;
		document.removeEventListener('mousemove', onMouseMove);
		document.removeEventListener('mouseup', onMouseUp);
	};

	document.addEventListener('mousemove', onMouseMove);
	document.addEventListener('mouseup', onMouseUp);
}

//When the map is ready, show the left drawer
watch(mapReady, (mapIsReady) => {
	if (mapIsReady) {
		leftDrawerOpen.value = true;
	}
});
onMounted(() => {
	deviceStore.loadDevices();
});
</script>
<style scoped>
.page-container {
	height: 100vh;
}

.resizable-drawer {
	position: relative;
	overflow: hidden;
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