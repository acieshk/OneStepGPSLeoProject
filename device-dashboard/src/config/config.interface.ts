export interface Config {
	api: {
	  baseUrl: string;
	  port: number;
	  endpoints: {
		devices: string;
	  }
	};
	server: {
	  port: number;
	  mongodb: {
		url: string;
		port: number;
		database: string;
		collection: string;
	  }
	};
}