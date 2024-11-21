import config from '@/config/config';

class ApiService {
    private static instance: ApiService;
    private baseUrl: string;

    private constructor() {
        this.baseUrl = `${config.api.baseUrl}:${config.api.port}`;
    }

    static getInstance(): ApiService {
        if (!ApiService.instance) {
            ApiService.instance = new ApiService();
        }
        return ApiService.instance;
    }

    async getDevices() {
        try {
            const response = await fetch(`${this.baseUrl}/devices`);
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        } catch (error) {
            console.error('Error fetching devices:', error);
            throw error;
        }
    }

    async refreshDatabase() {
        try {
            const response = await fetch(`${this.baseUrl}/fetch-devices`);
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        } catch (error) {
            console.error('Error refreshing database:', error);
            throw error;
        }
    }
}

export const apiService = ApiService.getInstance();