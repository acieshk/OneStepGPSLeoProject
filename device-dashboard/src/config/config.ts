interface Config {
    api: {
        baseUrl: string;
        port: number;
    };
    googleMaps: {
        apiKey: string;
    };
}

const config: Config = {
    api: {
        baseUrl: 'http://localhost',
        port: 8080
    },
    googleMaps: {
        apiKey: 'AIzaSyCv-LNnOxVRA2IKpQO-trqiyo79eUD38Kk'
    }
};

export default config;