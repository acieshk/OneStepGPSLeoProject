export interface Device {
	_id: string;
	activated_at: string | null;
	active_state: string;
	bcc_id: string;
	color?: string;
	iconUrl?: string;
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
	latest_device_point: DevicePoint | null;
	name: string;
	online: boolean;
	parent_device_id: string;
	serial_number: string | null;
	updated_at: string;
	vin: string | null;
	year: number | null;
	markerId?: number;
	// Additional properties for the custom object
	[key: string]: string | number | boolean | Date | object | undefined | null;
}
export interface DeviceSettings {
	_id: string;
	device_id: string;
	iconUrl?: string;
	version: number;
	updated_at?: string;
	begin_moving_speed: {
		value: number;
		unit: string;
		display: string;
	};
	begin_stopped_speed: {
		value: number;
		unit: string;
		display: string;
	};
	max_drift_distance: {
		value: number;
		unit: string;
		display: string;
	};
	min_num_satellites: number;
	ignore_unset_min_num_sats: boolean;
	max_hdop: number;
	drive_timeout: {
		value: number;
		unit: string;
		display: string;
	};
	stop_timeout: {
		value: number;
		unit: string;
		display: string;
	};
	offline_timeout: {
		value: number;
		unit: string;
		display: string;
	};
	history_calc_duration: {
		value: number;
		unit: string;
		display: string;
	};
	fuel_consumption: {
		calculation_method: string;
		measurement: string;
		fuel_type: string;
		fuel_cost: number;
		fuel_economy: number;
	};
	initial_device_point_delete_cutoff_time: string;
	engine_hours_counter_config: string;
	use_v3_engine_hours: boolean;
	history_retention_days: number;
	harsh_event_min_speed: {
		value: number;
		unit: string;
		display: string;
	};

}

export interface DevicePoint {
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
}

export interface DevicePointDetail {
	speed?: { value: number; unit: string; display: string };
}

export interface UserPreferences {
	version: number;
	userId: string;
	DeviceListWidth: number;
	unit: 'original' | 'metric' | 'imperial';
}