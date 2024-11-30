export interface Device {
	_id: string;
	activated_at: string | null;
	active_state: string;
	bcc_id: string;
	color?: string;  
	iconURL?: string;
	conn_data: {
		auth_token: string;
		device_id: string;
		host: string;
		is_on_ctc: boolean;
		port: number;
	};
	created_at: string;
	customer_id: string | null;
	deleted_at: null;
	device_category: string;
	device_id: string;
	device_make: string;
	device_model: string;
	device_type: string;
	display_name: string;
	external_id: string;
	firmware_version: string;
	fuel_tank_size_liters: number | null;
	groups: string[];
	hardware_version: string;
	imei: string;
	integration_meta: {
		onestepgps: string;
	};
	latest_device_point: {
		conn_data: {
			device_id: string;
		};
		device_point_detail: {
			external_volt: number;
		};
		device_state: {
			fuel_percent: number;
			odometer: {
				unit: string,
				value: number,
			},
		};
		dt_tracker: string | Date;
		lat: number;
		lng: number;
		speed: number;
	} | null;
	name: string;
	online: boolean;
	parent_device_id: string;
	serial_number: string | null;
	updated_at: string;
	vin: string | null;
	year: number | null;

	// Additional properties for the custom object
	[key: string]: string | number | boolean | Date | object | undefined | null;
}