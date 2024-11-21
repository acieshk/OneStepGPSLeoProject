import { ref, reactive } from 'vue';
import { Config } from '@/config/config.interface';
import { config as devConfig } from '@/config/config.development';
import { config as prodConfig } from '@/config/config.production';

export class ConfigService {
  private static instance: ConfigService;
  private config: Config;

  private constructor() {
    this.config = process.env.NODE_ENV === 'production' ? prodConfig : devConfig;
  }

  static getInstance(): ConfigService {
    if (!ConfigService.instance) {
      ConfigService.instance = new ConfigService();
    }
    return ConfigService.instance;
  }

  getConfig(): Config {
    return this.config;
  }

  getApiUrl(): string {
    return `${this.config.api.baseUrl}:${this.config.api.port}`;
  }

  getDevicesEndpoint(): string {
    return this.getApiUrl() + this.config.api.endpoints.devices;
  }
}

export const configService = ConfigService.getInstance();