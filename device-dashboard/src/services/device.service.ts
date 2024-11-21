import config from '@/config/config';

interface LatLng {
  lat: number;
  lng: number;
}

interface DevicePoint {
  lat: number;
  lng: number;
  speed: number;
  dt_tracker: string;
}

interface Device {
  device_id: string;
  display_name: string;
  online: boolean;
  latest_device_point: {
    lat: number;
    lng: number;
    speed: number;
    dt_tracker: string;
  };
}

interface ApiResponse {
  result_list: Device[];
}

class DeviceService {
  private static instance: DeviceService;
  private baseUrl: string;

  private constructor() {
    this.baseUrl = `${config.api.baseUrl}:${config.api.port}`;
  }

  static getInstance(): DeviceService {
    if (!DeviceService.instance) {
      DeviceService.instance = new DeviceService();
    }
    return DeviceService.instance;
  }

  async getDevices(): Promise<any> {
    try {
      const response = await fetch(`${this.baseUrl}/devices`);
      if (!response.ok) {
        throw new Error('Failed to fetch devices');
      }
      const data = await response.json();
      console.log('Device service response:', data); // Debug log
      return data;
    } catch (error) {
      console.error('Error in device service:', error);
      throw error;
    }
  }

  getDeviceLocation(device: Device): LatLng {
    return {
      lat: device.latest_device_point.lat,
      lng: device.latest_device_point.lng
    };
  }

  formatDate(dateString: string): string {
    return new Date(dateString).toLocaleString();
  }
}

export const deviceService = DeviceService.getInstance();