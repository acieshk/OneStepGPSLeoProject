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
		<!-- the icon uploader section -->
		<IconUploader />

		<div class="tree-controls">
			<q-btn flat icon="unfold_more" @click="expandAll">Expand All</q-btn>
			<q-btn flat icon="unfold_less" @click="collapseAll">Collapse All</q-btn>
		</div>

		<q-tree :nodes="deviceNodes" ref="treeRef" node-key="_id" :loading="isLoading">
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
								<q-input v-model="prop.node.value[index]" dense outlined
									@update:model-value="val => updateNodeValue(prop.node, null, val)" />
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
						style="min-width: 100px" @update:model-value="val => updateNodeValue(prop.node, null, val)" />

					<!-- Handle DateTime type -->
					<q-input v-else-if="isDateValue(prop.node.value)" :disable="isDisabled(prop.node)"
						v-model="prop.node.value" outlined dense
						@change="updateNodeValue(prop.node, null, prop.node.value)" />

					<!-- Handle String/Number type -->
					<q-input v-else-if="prop.node.value !== undefined" :disable="isDisabled(prop.node)"
						v-model="prop.node.value" dense outlined
						@change="updateNodeValue(prop.node, null, prop.node.value)" />
				</div>
			</template>
		</q-tree>

	</q-page>
</template>

<script setup lang="ts">
import IconUploader from 'components/iconUploader.vue';
import { useRoute, useRouter } from 'vue-router';
import { nextTick, onMounted, ref, watch } from 'vue';
import { storeToRefs } from 'pinia';
import { PrimitiveValue, type DeviceValue, useDeviceStore } from 'src/stores/deviceStore';
import { QInput, QTree } from 'quasar';

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

const treeRef = ref<InstanceType<typeof QTree> | null>(null);

const expandAll = () => {
	if (treeRef.value) {
		treeRef.value.expandAll();
	}
};

const collapseAll = () => {
	if (treeRef.value) {
		treeRef.value.collapseAll();
	}
};

const goBack = () => router.push('/');

// Fields to disable
const disabledFields = ['_id', 'bcc_id', 'visible', 'iconUrl', 'markerId'];

const isDisabled = (node: TreeNode) => {
	return disabledFields.includes(node.label); // Check if node label is in disabledFields
};

const saveDevice = async () => {
	try {
		isSaving.value = true;
		if (editingDevice.value) {
			await deviceStore.updateDevice(editingDevice.value);
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
	const pathParts = node._id.split('.');

	if (arrayIndex !== null && arrayIndex !== undefined) { // Handle array index for updates
		pathParts[pathParts.length - 1] = `${pathParts[pathParts.length - 1]}[${arrayIndex}]`;
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
	if (!editingDevice.value) return;

	const pathParts = node._id.split('.');
	const currentArray = pathParts.reduce((acc: Record<string, unknown>, key) => {
		// Handle array index notation
		const arrayMatch = key.match(/(\w+)\[(\d+)\]/);
		if (arrayMatch) {
			const [, arrayName] = arrayMatch;
			return acc[arrayName] as Record<string, unknown>;
		}
		return acc[key] as Record<string, unknown>;
	}, editingDevice.value as Record<string, unknown>);

	if (Array.isArray(currentArray)) {
		currentArray.push(null);

		// Update the device property to trigger reactivity
		deviceStore.updateDeviceProperty(pathParts, currentArray);
	}
};

const removeArrayItem = (node: TreeNode) => {

	// Find the parent array node manually
	const findParentArrayNode = (nodes: TreeNode[]): TreeNode | null => {
		for (const n of nodes) {
			if (n.children && n.children.includes(node)) {
				return n;
			}
			if (n.children) {
				const found = findParentArrayNode(n.children);
				if (found) return found;
			}
		}
		return null;
	};

	const parentNode = findParentArrayNode(deviceNodes.value);

	if (!editingDevice.value || !parentNode) {
		console.log('Cannot remove item - no parent found');
		return;
	}

	// Extract the array path
	const arrayPathParts = parentNode._id.split('.');

	// Retrieve the current array
	const currentArray = arrayPathParts.reduce((acc: Record<string, unknown>, key) => {
		const arrayMatch = key.match(/(\w+)\[(\d+)\]/);
		if (arrayMatch) {
			const [, arrayName] = arrayMatch;
			return acc[arrayName] as Record<string, unknown>;
		}
		return acc[key] as Record<string, unknown>;
	}, editingDevice.value as Record<string, unknown>);

	if (Array.isArray(currentArray)) {
		// Extract the actual index from the node's label
		const indexMatch = node.label.match(/\[(\d+)\]/);
		if (!indexMatch) {
			console.log('No index match found');
			return;
		}

		const indexToRemove = parseInt(indexMatch[1], 10);

		// Remove the item at the specified index
		currentArray.splice(indexToRemove, 1);

		// Update the device property to trigger reactivity
		deviceStore.updateDeviceProperty(arrayPathParts, currentArray);

		// Refresh the tree nodes
		if (parentNode.children) {
			parentNode.children = parentNode.children
				.filter(child => child._id !== node._id)
				.map((child, newIndex) => ({
					...child,
					label: `[${newIndex}]`,
					_id: `${parentNode._id}[${newIndex}]`
				}));
		}
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

const DEFAULT_ICON_URL = 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-blue.png';

// This watch handles updating the tree and setting up the default icon URL.
let updatingDevice = false;

watch(
  editingDevice,
  (newDevice) => {
    if (updatingDevice) return; // Prevent recursion
    updatingDevice = true;
    if (newDevice) {
      deviceNodes.value = objectToTree(newDevice);
      // Only set default icon if iconUrl is missing or is the default
      if (!newDevice.iconUrl || newDevice.iconUrl === 'https://raw.githubusercontent.com/pointhi/leaflet-color-markers/master/img/marker-icon-2x-blue.png') {
        deviceStore.updateDeviceProperty(['iconUrl'], DEFAULT_ICON_URL); // This triggers reactivity
      }
    } else {
      deviceNodes.value = []; // Clear nodes if editingDevice is null
    }
    nextTick(() => {
      updatingDevice = false; // Reset the flag after reactivity settles
    });
  },
  { immediate: true }
);

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

.tree-controls {
	display: flex;
	gap: 10px;
	margin-bottom: 10px;
}

.hide {
	display: none;
}
</style>
