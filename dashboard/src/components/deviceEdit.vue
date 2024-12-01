<template>
	<q-page padding>
		<div class="device-edit-header">
			<q-toolbar>
				<q-toolbar-title>Device Edit</q-toolbar-title>
				<div class="buttons">
					<q-btn flat label="Back" @click="goBack" />
					<q-btn color="primary" :disable="saveDisabled" :loading="isLoading" @click="saveDevice">
						Save
					</q-btn>
				</div>
			</q-toolbar>
		</div>

		<q-tree :nodes="deviceNodes" node-key="_id" default-expand-all :loading="isLoading">
			<template v-slot:default-header="prop">
				<div class="row items-center q-gutter-sm">
					<div>{{ prop.node.label }}:</div>
					<!-- Add plus button for array nodes -->
					<q-btn v-if="isArrayNode(prop.node)" flat round dense icon="add" size="sm" class="q-ml-sm"
						@click.stop="addArrayItem(prop.node)" />
					<!-- Add a minus button for array children -->
					<q-btn v-if="isArrayChild(prop.node)" flat round dense icon="remove" size="sm" class="q-ml-sm"
						color="negative" @click.stop="removeArrayItem(prop.node)" />
					<!-- Handle Array type -->
					<template v-if="Array.isArray(prop.node.value)">
						<div class="column">
							<div v-for="(item, index) in prop.node.value" :key="index"
								class="row items-center q-gutter-sm">
								<q-input v-model="prop.node.value[index]" dense outlined @change="updateNodeValue(prop.node, null, prop.node.value)" />
							</div>
							<q-btn label="Add Item" color="positive" flat dense @click="addArrayItem(prop.node)" />
						</div>
					</template>

					<!-- Handle Boolean type -->
					<q-select v-else-if="typeof prop.node.value === 'boolean'" :disable="isDisabled(prop.node)"
						v-model="prop.node.value" :options="[
							{ label: 'True', value: true },
							{ label: 'False', value: false }
						]" dense outlined options-dense emit-value map-options :display-value="prop.node.value ? 'true' : 'false'"
						style="min-width: 100px" @change="updateNodeValue(prop.node, null, prop.node.value)" />

					<!-- Handle DateTime type -->
					<q-input v-else-if="isDateValue(prop.node.value)" :disable="isDisabled(prop.node)"
						v-model="prop.node.value" outlined dense @change="updateNodeValue(prop.node, null, prop.node.value)" />
					<!-- Handle String/Number type -->
					<q-input v-else-if="prop.node.value !== undefined" :disable="isDisabled(prop.node)"
						v-model="prop.node.value" dense outlined @change="updateNodeValue(prop.node, null, prop.node.value)" />
				</div>
			</template>
		</q-tree>

	</q-page>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router';
import { onMounted, ref, watch } from 'vue';
import { storeToRefs } from 'pinia';
import { PrimitiveValue, type DeviceValue, useDeviceStore } from 'src/stores/deviceStore';
import { QInput } from 'quasar';

interface TreeNode {
	label: string;
	_id: string;
	value?: unknown | PrimitiveValue;
	children?: TreeNode[];
	parent?: TreeNode;
}

const router = useRouter();
const route = useRoute();
const deviceStore = useDeviceStore();
const { editingDevice, deviceLoaded } = storeToRefs(deviceStore);

const deviceNodes = ref<TreeNode[]>([]);
const isLoading = ref(false);
const isSaving = ref(false);

const saveDisabled = ref(false);

// const hasChanges = computed(() => {
// 	// Implement change detection logic
// 	return true; // Placeholder
// });

const goBack = () => router.push('/');

const disabledFields = ['_id', 'bcc_id', 'visible']; // Fields to disable

const isDisabled = (node: TreeNode) => {
	return disabledFields.includes(node.label); // Check if node label is in disabledFields
};

const saveDevice = async () => {
  try {
    isSaving.value = true;
    console.log('Attempting to save device...');
    if (editingDevice.value) {
      console.log('Editing device:', editingDevice.value);
      await deviceStore.updateDevice(editingDevice.value);
      console.log('Device saved successfully!');
      router.push('/');
    } else {
      console.error('Editing device is null or undefined.');
    }
  } catch (error) {
    console.error('Failed to save device:', error);
  } finally {
    isSaving.value = false;
  }
};

const updateNodeValue = (node: TreeNode, arrayIndex: number | null, newValue: DeviceValue) => {
	
	console.log('node value is updated');
	console.log('Node: ', node.value);
	console.log('array index: ', arrayIndex);
	console.log('New Value: ', newValue);
	const pathParts = node._id.split('.');

	if (arrayIndex !== null) {
		pathParts.push(`[${arrayIndex}]`);
	}

	deviceStore.updateDeviceProperty(pathParts, newValue);
};

const objectToTree = (obj: Record<string, unknown>): TreeNode[] => {
	const convertToNodes = (data: unknown, path = ''): TreeNode[] => {
		if (!data || typeof data !== 'object') {
			return [];
		}

		return Object.entries(data as Record<string, unknown>).map(([key, value]) => {
			const currentPath = path ? `${path}.${key}` : key;

			if (Array.isArray(value)) {
				return {
					label: key,
					_id: currentPath,
					children: value.map((item, index) => ({
						label: `[${index}]`,
						_id: `${currentPath}[${index}]`,
						value: isPrimitive(item) ? item : undefined,
						children: !isPrimitive(item) ? convertToNodes(item, `${currentPath}[${index}]`) : undefined
					}))
				};
			}

			if (value && typeof value === 'object') {
				return {
					label: key,
					_id: currentPath,
					children: convertToNodes(value, currentPath)
				};
			}

			return {
				label: key,
				_id: currentPath,
				value: value as DeviceValue
			};
		});
	};
	// Helper function to check if value is primitive
	const isPrimitive = (value: unknown): value is PrimitiveValue => {
		return (
			typeof value === 'string' ||
			typeof value === 'number' ||
			typeof value === 'boolean' ||
			value === null ||
			value === undefined
		);
	};

	return convertToNodes(obj);
}

const isArrayNode = (node: TreeNode): boolean => {
	return node.children?.some(child => /^\[\d+\]$/.test(child.label)) ?? false;
};

const isArrayChild = (node: TreeNode): boolean => {
	return /^\[\d+\]$/.test(node.label);
};

const addArrayItem = (node: TreeNode) => {
	if (!node.children) node.children = [];

	const newIndex = node.children.length;
	const newNode: TreeNode = {
		_id: `${node._id}_${newIndex}`,
		label: `[${newIndex}]`,
		value: null
	};

	node.children.push(newNode);
	updateNodeValue(node, newIndex, null);
};

const removeArrayItem = (node: TreeNode) => {
	if (!node.parent?.children) return;

	const index = node.parent.children.findIndex(child => child._id === node._id);
	if (index !== -1) {
		node.parent.children.splice(index, 1);
		node.parent.children.forEach((child, idx) => {
			child.label = `[${idx}]`;
		});
	}
};

const isDateValue = (value: unknown): boolean => {
	if (typeof value === 'string') {
		try {
			const parsed = new Date(value);
			return !isNaN(parsed.getTime());
		} catch {
			return false;
		}
	}
	return value instanceof Date && !isNaN(value.getTime());
};

watch(deviceLoaded, (loaded) => {
	if (!loaded) return;
	if (editingDevice.value === null) {
		// Use route.params.id if editingDevice is null
		// Refreshing on the edit page, or enter url directly would enter this route
		const deviceIdFromRoute = route.params.id;
		const deviceId = Array.isArray(deviceIdFromRoute) ? deviceIdFromRoute[0] : deviceIdFromRoute; // Take the first element if it's an array
		deviceStore.setEditingDeviceByID(deviceId as string); // Ensure type safety with type assertion
	}
}, { immediate: true });

watch(editingDevice, (newDevice) => {
	if (newDevice !== null) {
		// Convert the device object to a tree structure for easy editing
		deviceNodes.value = objectToTree(newDevice);
	} else deviceNodes.value = [];
}, { immediate: true });


// When initializing the tree, add parent references
const addParentReferences = (nodes: TreeNode[], parent?: TreeNode) => {
	nodes.forEach(node => {
		if (parent) {
			node.parent = parent;
		}
		if (node.children) {
			addParentReferences(node.children, node);
		}
	});
};


onMounted(async () => {
	await deviceStore.loadDevices(); // Ensure devices are loaded when component mounts
	addParentReferences(deviceNodes.value);
});

</script>

<style scoped>
.device-edit-header {
	border-bottom: 1px solid #ddd;
}

.buttons {
	display: flex;
	gap: 8px;
}
</style>
