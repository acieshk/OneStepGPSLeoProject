// src/stores/deviceStore.ts
import { defineStore } from 'pinia';
import type { Device, UserPreferences } from '@/types/device';

export const useDeviceStore = defineStore('devices', {
  state: () => ({
    devices: [] as Device[],
    preferences: {
      sortOrder: 'name',
      hiddenDevices: [],
      deviceIcons: {}
    } as UserPreferences,
    loading: false,
    error: null as string | null
  }),

  actions: {
    async fetchDevices() {
      this.loading = true;
      try {
        // For now, we'll load the JSON directly
        const response = await fetch('/api/devices');
        const data = await response.json();
        this.devices = data.result_list;
      } catch (error) {
        this.error = 'Failed to fetch devices';
        console.error(error);
      } finally {
        this.loading = false;
      }
    },

    updatePreferences(prefs: Partial<UserPreferences>) {
      this.preferences = {
        ...this.preferences,
        ...prefs
      };
      // Save to localStorage
      localStorage.setItem('devicePreferences', JSON.stringify(this.preferences));
    }
  }
});