// src/types/device.ts
export interface Device {
	device_id: string;
	display_name: string;
	active_state: string;
	online: boolean;
	latest_device_point: {
	  lat: number;
	  lng: number;
	  speed: number;
	  device_state: {
		drive_status: string;
		fuel_percent?: number;
	  }
	}
}