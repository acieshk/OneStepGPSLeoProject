<template>
	{{ selectedDeviceId }}
	<q-table title="Devices" :rows="devices" :columns="columns" v-model:pagination.sync="pagination"
		:rows-per-page-options="[10, 25, 50, 100, 0]" row-key="_id" @row-click="handleRowClick"
		@mouseover="handleRowMouseover" @mouseout="handleRowMouseout">
		<template v-slot:body-cell-actions="props">
			<q-td :props="props" class="action-cell">
				<q-btn flat round dense icon="edit" @click.stop="goToEditDevice(props.row)" />
			</q-td>
		</template>
		<template v-slot:body-cell-visible="props">
			<q-td :props="props" class="visibility-cell">
				<q-icon :name="props.row.visible ? 'visibility' : 'visibility_off'"
					@click.stop="handleVisibilityToggle(props.row)" />
			</q-td>
		</template>
		<template v-slot:body-cell-device="props"> <!-- Named slot for the "device" column -->
			<q-td :props="props">
				<div class="name-status">
					<span :class="['status-indicator', props.row.online ? 'online' : 'offline']"></span>
					<span class="device-name">{{ props.row.display_name }}</span>
				</div>
				<div class="device-id-container">
					<span class="device-id">{{ props.row.device_id }}</span>
				</div>
			</q-td>
		</template>
	</q-table>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useDeviceStore } from 'src/stores/deviceStore';
import { useUserStore } from 'src/stores/userStore';
import { storeToRefs } from 'pinia';
import { Device } from 'src/types/device';
import { QIcon, QTd } from 'quasar';
import { useRouter } from 'vue-router';


const deviceStore = useDeviceStore();
const { deviceLoaded, devices, selectedDeviceId } = storeToRefs(deviceStore);
const userStore = useUserStore();
const { rowPerPage } = storeToRefs(userStore);
const selectedRow = ref<Device | null>(null);
const pagination = ref({
	page: 1,
	rowsPerPage: rowPerPage, // Default rows per page
	rowPerPageOptions: [10, 20]
});


const columns = computed(() => {
	if (!devices.value || devices.value.length === 0) {
		return []; // Or a default set of columns if needed
	}

	const visibleDevices = deviceStore.devices.filter((device: Device) => device.visible).length;
	const totalDevices = deviceStore.devices.length;

	return [
		{
			name: 'actions',
			label: 'Actions',
			field: 'actions',
			align: 'center' as 'left' | 'right' | 'center'
		},
		{
			name: 'visible',
			label: `Visible (${visibleDevices}/${totalDevices})`,
			field: 'visible',
			align: 'center' as 'left' | 'right' | 'center',
		},
		{
			name: 'device', // Unique name for the column
			label: 'Device', // Display label
			field: 'device',  // Pass the whole row so both name and id can be used for formatting
			align: 'left' as 'left' | 'right' | 'center',
			sortable: true // Sorting will be based on the formatted value
		},
		{
			name: 'location',
			label: 'Location',
			field: (row: Device) => row.latest_device_point, // Access latest_device_point
			format: (val: {
				lat: string,
                lng: string,
			}) => {
				const lat = Number(val.lat) || 0; // Convert to number
				const lng = Number(val.lng) || 0; // Convert to number
				return isNaN(lat) || isNaN(lng) ? 'N/A' : `[${lat.toFixed(2)}, ${lng.toFixed(2)}]`; // Check result
			},
			align: 'left' as 'left' | 'right' | 'center',
			sortable: false, // Can't really sort by location
		}
	];
});
/*
	Row Functionality
*/
const handleRowClick = (_evt: Event, row: Device) => {
	const _id = row._id
	if (deviceStore.selectedDeviceId == null) {
		deviceStore.selectDevice(_id);
	}
	else if (deviceStore.selectedDeviceId !== _id) {
		deviceStore.selectDevice(_id);
	}
	else if (deviceStore.selectedDeviceId == _id) {
		deviceStore.deselectDevice();
	}
};
const handleRowMouseover = (evt: Event, row: Device | undefined) => { // Indicate that row could be undefined
	if (row) {  // Check if row is defined
		deviceStore.setHoveredDevice(row._id);
	}
};
const handleRowMouseout = () => {
	deviceStore.setHoveredDevice(null);
}
/* 
	  Edit Button functionality
 */
const router = useRouter();
const goToEditDevice = (device: Device) => {
	router.push(`/devices/${device._id}`); // Or your edit route
};

/*
	Visibility functionality
*/
const handleVisibilityToggle = (device: Device) => {
	// Update the device's visibility in the store and Map:
	console.log('Update Visibility');
	deviceStore.toggleDeviceVisibility(device._id);	//update visibility Map for the icons
};


watch(deviceLoaded, () => {
	if (deviceLoaded)
		console.log('Devices loaded:', devices.value);
}, { immediate: true });

// Watch selectedDeviceId from the store
watch(selectedDeviceId, (newSelectedDeviceId) => {
	if (newSelectedDeviceId) {
		selectedRow.value = devices.value.find((device: Device) => device._id === newSelectedDeviceId) || null;
	} else {
		selectedRow.value = null;
	}
});

</script>
<style scoped>
.q-table__tr--selected {
	background-color: rgba(0, 0, 0, 0.1) !important;
}

/* Hovered row style */
.q-table__tr--highlighted {
	background-color: rgba(0, 0, 0, 0.05);
}

.action-cell {
	text-align: center;
	cursor: pointer;
}

visibility-cell {
	cursor: pointer;
}

.device-name {
	font-weight: 500;
	color: #2c3e50;
}

.device-id {
	font-size: 0.8rem;
	color: #999;
	margin-top: 4px;
}

.name-status-container {
	display: flex;
	align-items: center;
}

.device-id-container {
	margin-top: 4px;
}

.status-indicator {
	width: 10px;
	height: 10px;
	border-radius: 50%;
	display: inline-block;
	margin-right: 5px;
}

.status-indicator.online {
	background-color: #4caf50;
}

.status-indicator.offline {
	background-color: #f44336;
}
</style>