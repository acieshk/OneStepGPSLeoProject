<template>
	<el-dialog v-model="showDialog" title="Edit Device" width="90%" @close="closeDialog">
		<div class="upload-container">
			<IconUpload :deviceId="deviceToEdit?._id" :uploadUrl="uploadUrl" :iconColor="deviceToEdit?.color" />
		</div>
		<div class="edit-table-container">
			<el-table :data="filteredTableData" border style="width: 100%"
				:row-class-name="(row) => row.modified ? 'modified-row fixed-height-row' : 'fixed-height-row'">
				<el-table-column v-for="(header, level) in headers" :key="level" :label="header"
					:width="level === 0 ? '150px' : '125px'">
					<template #default="scope">
						<span v-if="scope.row.path.length >= level + 1">
							{{ scope.row.path[level] }}
						</span>
					</template>
				</el-table-column>
				<el-table-column label="Value" width="275">
					<template #default="scope">
						<span v-if="!scope.row.editing">
							{{ displayValue(scope.row.value) }} {{ scope.row.unit ? scope.row.unit : '' }}
						</span>
						<div v-else-if="!scope.row.nonEditable" class="value-edit-container">
							<el-select v-if="getInputType(scope.row.originalValue) === 'boolean'"
								v-model="scope.row.value" @change="() => scope.row.modified = true">
								<el-option :value="true" label="true"></el-option>
								<el-option :value="false" label="false"></el-option>
							</el-select>
							<el-input v-else-if="getInputType(scope.row.originalValue) === 'text'"
								v-model="scope.row.value" @input="() => scope.row.modified = true"></el-input>
							<el-input-number v-else-if="getInputType(scope.row.originalValue) === 'number'"
								v-model="scope.row.value" @change="() => scope.row.modified = true"></el-input-number>
							<el-date-picker v-else-if="getInputType(scope.row.originalValue) === 'date'"
								v-model="scope.row.value" type="datetime" @change="() => scope.row.modified = true" />
							<el-button type="text" size="small" class="null-button"
								@click="setNullAndSave(scope.row, scope.$index)">Set Null</el-button>
						</div>
						<span v-else> <!-- Correctly placed v-else for non-editable fields -->
							{{ displayValue(scope.row.value) }} {{ scope.row.unit ? scope.row.unit : '' }}
						</span>
					</template>
				</el-table-column>
				<el-table-column label="Edit" width="120">
					<template #default="scope">
						<el-button type="primary" size="small" :disabled="scope.row.nonEditable"
							@click="editRow(scope.row, scope.$index)" v-if="!scope.row.editing">Edit</el-button>
						<el-button type="primary" size="small" @click="saveRow(scope.row, scope.$index)"
							v-if="scope.row.editing">Save</el-button>
					</template>
				</el-table-column>
			</el-table>
		</div>
		<template #footer>
			<span class="dialog-footer">
				<el-button @click="closeDialog">Cancel</el-button>
				<el-button type="primary" @click="saveDevice">Save</el-button>
			</span>
		</template>
	</el-dialog>
</template>



<script setup lang="ts">
import { ref, reactive, computed, watch, toRefs, defineEmits, defineProps, nextTick } from 'vue';
import { ElMessage } from 'element-plus';
import IconUpload from './IconUpload.vue';
import _ from 'lodash';
import { apiService } from '@/services/api.service';
import type { Device } from '@/types/device';

const isSaving = ref(false);

const emit = defineEmits(['update:modelValue', 'device-updated']);

const props = defineProps<{
	modelValue: boolean;
	deviceToEdit: Device | null;
}>();

const displayValue = (value: any) => {
	return value === null ? '[null]' : value;  // Display [null] for actual null values
};

const saveRow = (row, index) => {
    row.modified = !_.isEqual(row.originalValue, row.value);
    row.editing = false;
    editingRowIndex.value = -1;
};

const setNullAndSave = (row: any, index: number) => {
	row.value = null;
	saveRow(row, index); // Call saveRow which will handle setting modified and forceUpdate
};

const validateDevice = (device: Device): string[] => {
	const errors: string[] = [];
	console.log(device);
	// Example validations (adapt based on your Device type and requirements):
	if (!device._id) {
		errors.push("Name is required.");
	}
	if (device.location && typeof device.location.latitude !== 'number') { // Check type and presence
		errors.push("Latitude must be a number.");
	}
	return errors;
}

const tableData = ref([] as any[]);

const getUnit = (path: string[]): string | null => {
	if (path.join('.') === 'location.altitude') return 'meters';
	if (path.join('.') === 'power.battery_voltage') return 'volts';

	return null;
};

const getInputType = (originalValue: unknown) => {
	if (typeof originalValue === 'boolean') return 'boolean';
	if (typeof originalValue === 'number') return 'number';
	if (originalValue instanceof Date) return 'date';
	// Check if the original value is a date STRING:
	if (typeof originalValue === 'string' && !isNaN(new Date(originalValue).getTime())) {
		return 'date'
	}
	return 'text';
};

const headers = ['Properties', 'Level 2', 'Level 3', 'Level 4'];
const nonEditableFields = ref([  // Array of non-editable fields
	'_id',
]);
const filteredTableData = computed(() => {
	return tableData.value.filter(row => !hiddenFields.value.includes(row.path.join('.')));
});

const hiddenFields = ref([
	// Add fields you want to hide here, using dot notation for nested fields.  For example:
	'location.latitude',
	'power.battery_level'
]);

function flattenObjectToTableData(obj: Record<string, any>, path: string[] = [], level: number = 1): any[] {
    const rows: any[] = [];
    for (const key in obj) {
        const newPath = [...path, key];
        const value = obj[key];

        if (typeof value === 'object' && value !== null && !Array.isArray(value) && Object.keys(value).length > 0) {
            rows.push(...flattenObjectToTableData(value, newPath, level + 1));
        } else {
            const row = reactive({
                level1: level >= 1 ? newPath[0] : '',
                level2: level >= 2 ? newPath[1] : '',
                level3: level >= 3 ? newPath[2] : '',
                level4: level >= 4 ? newPath[3] : '',
                value: value,                     // Current value (editable)
                path: newPath,                     // Path to the property (e.g., ['location', 'latitude'])
                originalValue: JSON.parse(JSON.stringify(value)), // Original value (for change detection)
                modified: false,                 // Whether the value has been modified
                editing: false,                  // Whether the row is currently being edited
                unit: getUnit(newPath),           // Unit of the value (if applicable)
                nonEditable: nonEditableFields.value.includes(newPath.join('.')) // Whether the field is editable
            });
            rows.push(row);
        }
    }
    return rows;
}


function unflattenObject(tableData: any[]): Device {
    const unflattenedObj = {} as Device;
    for (const row of tableData) {
        _.set(unflattenedObj, row.path, row.value);
    }
    return unflattenedObj;
}

const { modelValue, deviceToEdit } = toRefs(props);


const showDialog = computed({
	get() {
		return modelValue.value;
	},
	set(value) {
		emit('update:modelValue', value);
	}
});

const editRow = (row: any, index: number) => {

	if (editingRowIndex.value !== -1 && editingRowIndex.value !== index) {
		tableData.value[editingRowIndex.value].editing = false;
	}
	row.editing = true;
	editingRowIndex.value = index;
};





const editingRowIndex = ref(-1); // Keep track of the row index being edited

const editForm = ref<Device>({} as Device);

const rules = reactive({}); // Add validation rules as needed



// Watch deviceToEdit to update tableData reactively:
watch(deviceToEdit, (newDevice) => {
	if (newDevice) {
		tableData.value = flattenObjectToTableData(newDevice);
	} else {
		tableData.value = [];
	}
	console.log("deviceToEdit changed:", newDevice);
    console.log("deviceId changed:", newDevice?._id);
}, { deep: true });

const closeDialog = () => {
	showDialog.value = false;
};


const saveDevice = async () => {
    isSaving.value = true;
    try {

        const updatedDevice = unflattenObject(tableData.value);

        // Call the API service to update the device:
        const savedDevice = await apiService.updateDevice(updatedDevice._id, updatedDevice);  // Ensure _id is correct

        // Handle success:
        ElMessage.success('Device Updated Successfully!');
        emit('device-updated', savedDevice);
        closeDialog();

    } catch (error) {
        console.error("Error saving device:", error);
        ElMessage.error('Failed to update Device.');
    } finally {
        isSaving.value = false;
    }
};


const formatLabel = (key: unknown) => {

	if (typeof key === 'string') {
		return _.startCase(key);
	} else {
		return String(key);
	}
};

const uploadUrl = computed(() => {
	return deviceToEdit.value?.device_id ? `/devices/${deviceToEdit.value.device_id}/icon` : null // or just ''
});

</script>
<style scoped>
/* This is how to set the row to fixed height since Element is overridding the height */
.el-table ::v-deep(.fixed-height-row) {
	height: 65px;
}

.el-table ::v-deep(.fixed-height-row .el-table__cell) {
	height: 65px;
	box-sizing: border-box;
}


.value-edit-container {
	display: flex;
	align-items: center;
	width: 100%;
	padding: 5px 0;
}

.edit-table-container {
	display: flex;
	justify-content: center;
	width: 100%;
}

.null-button {
	margin-left: 8px;
}

.el-table ::v-deep(.modified-row) {
	background-color: lightgreen;
}
</style>