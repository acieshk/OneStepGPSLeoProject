<template>
	<q-tree
	  :nodes="nodes"
	  node-key="label"  />
  </template>
  
  <script setup lang="ts">
  import { ref } from 'vue';
  import type { QTreeNode } from 'quasar';
  
  
  const nodes = ref([
	{
	  label: 'Node 1',
	  children: [
		{ label: 'Node 1.1' },
		{ label: 'Node 1.2' }
	  ]
	},
	{
	  label: 'Node 2',
	  children: [
		{ label: 'Node 2.1' },
		{
		  label: 'Node 2.2',
		  children: [
			{ label: 'Node 2.2.1' },
			{ label: 'Node 2.2.2' }
		  ]
		}
	  ]
	}
  ] as QTreeNode[]);
  
  
  
  
  </script>