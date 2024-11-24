<template>
	<el-dialog v-model="showDialog" title="Edit Device" width="90%" @close="closeDialog">
	  <div class="edit-table-container" style="max-height: 500px; overflow-y: auto;">
		<el-table :data="tableData" border style="width: 100%">
		  <el-table-column label="Level 1" prop="level1" width="180" />
		  <el-table-column label="Level 2" prop="level2" width="180" />
		  <el-table-column label="Level 3" prop="level3" width="180" />
		  <el-table-column label="Level 4" prop="level4" width="180" />
		  <el-table-column label="Value" width="180">
			<template #default="scope">
			  <div v-if="scope.row.editing">
				<el-input v-if="getInputType(scope.row.value) === 'text'" v-model="scope.row.value"></el-input>
				<el-input-number v-else-if="getInputType(scope.row.value) === 'number'" v-model="scope.row.value" />
				 <el-select v-else-if="getInputType(scope.row.value) === 'boolean'" v-model="scope.row.value">
				  <el-option :value="true" label="true"></el-option>
				  <el-option :value="false" label="false"></el-option>
				</el-select>
				<el-date-picker v-else-if="getInputType(scope.row.value) === 'date'" v-model="scope.row.value" type="date" />
  
			  </div>
				<span v-else-if="scope.row.unit">
				  {{ scope.row.value }} {{ scope.row.unit }}
				</span>
			  <span v-else>
				{{ scope.row.value }}
			   </span>
				<el-button type="primary" size="small" @click="editRow(scope.row)" v-if="!scope.row.editing">Edit</el-button>
  
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
import { ref, reactive, computed, watch, toRefs, defineEmits, defineProps } from 'vue';
import type { Device } from '@/App.vue';
import { ElMessage } from 'element-plus';
import _ from 'lodash';

const emit = defineEmits(['update:modelValue', 'device-updated']);

const props = defineProps<{
  modelValue: boolean;
  deviceToEdit: Device | null;
}>();

const tableData = computed(() => {
  if (!deviceToEdit.value) return [];
  const rows: any[] = [];
  flattenObjectToTableData(deviceToEdit.value, rows);
  // Add editing property and unit property
  rows.forEach(row => {
    row.editing = false;
    row.unit = getUnit(row.path); // Assuming you have a getUnit function
  });
  return rows;
});

const getUnit = (path: string[]): string | null => {
  if (path.join('.') === 'location.altitude') return 'meters';
  if (path.join('.') === 'power.battery_voltage') return 'volts';

  return null;
};

const getInputType = (value: unknown) => {
  if (typeof value === 'number') return 'number';
  if (typeof value === 'boolean') return 'boolean';
    if (value instanceof Date) return 'date';
  return 'text';
};

const editRow = (row: any) => {
    row.editing = true;

};

const flattenObjectToTableData = (obj: Record<string, any>, rows: any[], currentPath: string[] = [], level: number = 1) => {
  for (const key in obj) {
    if (obj.hasOwnProperty(key)) {

			if (key !== 'device_id') {
				const value = obj[key];
				const newPath = [...currentPath, key];

				const row = {
					level1: level >= 1 ? newPath[0] : '',
					level2: level >= 2 ? newPath[1] : '',
					level3: level >= 3 ? newPath[2] : '',
					level4: level >= 4 ? newPath[3] : '', // Add more levels
					value: value,
					path: newPath, // Store the full path to the value
				};

				if (typeof value === 'object' && value !== null && !Array.isArray(value) && Object.keys(value).length > 0) {
					flattenObjectToTableData(value, rows, newPath, level + 1); // Recursive call for nested objects

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


const editForm = ref<Device>({} as Device);

const flattenedFields = computed(() => {
	if (!deviceToEdit.value) return {};
	return _.transform(deviceToEdit.value, (result, value, key) => {

		if (key !== 'device_id') { // Exclude device_id from editable fields

			if (typeof value === 'object' && value !== null && !Array.isArray(value)) {
				// Recursively flatten nested objects
				_.forEach(value, (innerValue, innerKey) => {
					result[`${key}.${innerKey}`] = innerValue;
				});
			} else {
				result[key] = value;
			}

		}


	}, {});
});


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

</script>