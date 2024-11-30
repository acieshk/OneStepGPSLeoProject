/****
MainLayout.vue
Handle Drawer logic
*/
<template>
	<q-layout view="hHh Lpr lFf" class="main-layout">
		<q-header elevated>
			<q-toolbar>
				<q-btn flat dense round icon="menu" aria-label="Menu" @click="toggleLeftDrawer" />

				<q-toolbar-title>
					Device Dashboard
				</q-toolbar-title>

				<div>Right side</div>
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
</template>

<script setup lang="ts">
import DeviceList from 'components/deviceList.vue';
import { storeToRefs } from 'pinia';
import { useDeviceStore } from 'src/stores/deviceStore';
import { onMounted, ref, watch } from 'vue';


defineOptions({
	name: 'MainLayout'
});

const deviceStore = useDeviceStore();
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
	background-color: rgba(25, 118, 210, 0.2); ;
}
</style>