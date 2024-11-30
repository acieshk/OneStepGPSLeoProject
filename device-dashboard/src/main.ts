import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import App from './App.vue'
import router from './router';
import { Quasar } from 'quasar'
import '@quasar/extras/material-icons/material-icons.css' // Or the icon set you want
import 'quasar/src/css/index.sass'
import 'leaflet/dist/leaflet.css';

const app = createApp(App)
app.use(createPinia())
app.use(router);
app.use(ElementPlus)
app.use(Quasar, {
	plugins: {}, // import Quasar plugins and add here
})
app.mount('#app')
