<template>
	<el-dialog v-model="showDialog" title="Edit Device" width="90%" @close="closeDialog">
	  <div class="upload-container">
	  <IconUpload :uploadUrl="uploadUrl" :currentIconUrl="deviceToEdit?.iconURL" @icon-uploaded="handleIconUploaded" />
	  </div>
	  <div class="edit-table-container">
		<el-table :data="filteredTableData" border style="width: 100%" :row-class-name="(row) => row.modified ? 'modified-row fixed-height-row' : 'fixed-height-row'">
		  <el-table-column v-for="(header, level) in headers" :key="level" :label="header" :width="level === 0 ? '150px' : '125px'">
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
				<el-select v-if="getInputType(scope.row.originalValue) === 'boolean'" v-model="scope.row.value" @change="() => { scope.row.modified = true; forceUpdate(); }">
				  <el-option :value="true" label="true"></el-option>
				  <el-option :value="false" label="false"></el-option>
				</el-select>
				<el-input v-else-if="getInputType(scope.row.originalValue) === 'text'" v-model="scope.row.value" @input="() => scope.row.modified = true"></el-input>
				<el-input-number v-else-if="getInputType(scope.row.originalValue) === 'number'" v-model="scope.row.value"  @change="() => scope.row.modified = true"></el-input-number>
				<el-date-picker v-else-if="getInputType(scope.row.originalValue) === 'date'" v-model="scope.row.value" type="datetime" @change="() => scope.row.modified = true" />
				<el-button type="text" size="small" class="null-button" @click="setNullAndSave(scope.row, scope.$index)">Set Null</el-button>
			  </div>
			  <span v-else>  <!-- Correctly placed v-else for non-editable fields -->
				{{ displayValue(scope.row.value) }} {{ scope.row.unit ? scope.row.unit : '' }}
			  </span>
			</template>
		  </el-table-column>
		  <el-table-column label="Edit" width="120">
			<template #default="scope">
			  <el-button type="primary" size="small" :disabled="scope.row.nonEditable" @click="editRow(scope.row, scope.$index)" v-if="!scope.row.editing">Edit</el-button>
			  <el-button type="primary" size="small" @click="saveRow(scope.row, scope.$index)" v-if="scope.row.editing">Save</el-button>
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
import type { Device } from '@/App.vue';
import { ElMessage } from 'element-plus';
import IconUpload from './IconUpload.vue';
import _ from 'lodash';

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

	// Trigger a reactivity update by modifying the tableData ref directly:
	tableData.value[index] = { ...row }; // Create a new object
};

const setNullAndSave = (row: any, index: number) => {
	row.value = null;
	saveRow(row, index); // Call saveRow which will handle setting modified and forceUpdate
};


const tableData = computed(() => {
	if (!deviceToEdit.value) return [];
	const rows: any[] = [];
	flattenObjectToTableData(deviceToEdit.value, rows);

	rows.forEach(row => {
		row.editing = false;
		row.unit = getUnit(row.path);
		row.originalValue = row.value; // Store original value
	});
	return rows;
});

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

const flattenObjectToTableData = (obj: Record<string, any>, rows: any[], currentPath: string[] = [], level: number = 1, parentKey: string | null = null) => {
	for (const key in obj) {
		if (obj.hasOwnProperty(key)) {

			if (key !== 'device_id') {
				const value = obj[key];
				const newPath = [...currentPath, key];

				const row = {
					level1: level >= 1 ? newPath[0] : '', // Start from level 1
					level2: level >= 2 ? newPath[1] : '',
					level3: level >= 3 ? newPath[2] : '',
					level4: level >= 4 ? newPath[3] : '', // Include level 4
					value: value,
					path: newPath,
					nonEditable: nonEditableFields.value.includes(newPath.join('.')), // Check for non-editable fields using parentKey
				};

				if (typeof value === 'object' && value !== null && !Array.isArray(value) && Object.keys(value).length > 0) {
					flattenObjectToTableData(value, rows, newPath, level + 1, key); // Pass key as parentKey in recursive call
				} else {
					rows.push(row);
				}


			}


		}
	}

};


const unflattenObject = (tableData: any[]): Device => {
	const unflattenedObj = {} as Device;

	for (const row of tableData) {

		_.set(unflattenedObj, row.path, row.value);

	}
	return unflattenedObj;

};

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

	// Deep copy original value for comparison
	row.originalValue = JSON.parse(JSON.stringify(row.value));  // Use JSON stringify for deep copy
};






const editingRowIndex = ref(-1); // Keep track of the row index being edited

const editForm = ref<Device>({} as Device);

const rules = reactive({}); // Add validation rules as needed



watch(deviceToEdit, (newDevice) => {
	if (newDevice) {
		editForm.value = JSON.parse(JSON.stringify(newDevice)); // Deep copy
	} else {
		editForm.value = {} as Device; // Clear form
	}
});


const editFormRef = ref();


const closeDialog = () => {

	showDialog.value = false;
	editFormRef.value.resetFields();
};

const saveDevice = async () => {

	const valid = await editFormRef.value.validate();
	if (!valid) return;


	emit('device-updated', editForm.value);

	const updatedDevice = unflattenObject(tableData.value);
	emit('device-updated', updatedDevice);
	closeDialog();
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
	height: 60px;
}

.el-table ::v-deep(.fixed-height-row .el-table__cell) {
	height: 60px;
	box-sizing: border-box;
}


.value-edit-container {
	/* Style the container for layout */
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