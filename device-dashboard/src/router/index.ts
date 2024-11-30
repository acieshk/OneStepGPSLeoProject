import { createRouter, createWebHistory } from 'vue-router';
// import Devices from '../components/Device.vue';
import DeviceView from '@/views/DeviceView.vue';
import UserPreferences from '../components/UserPreferences.vue';


const router = createRouter({ // createRouter returns a Router instance, no need for type declaration
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
        { path: '/', name: 'Devices', component: DeviceView },
        { path: '/devices', name: 'Devices', component: DeviceView },
        { path: '/user-preferences', name: 'UserPreferences', component: UserPreferences },
        // Wildcard route should always be last
		{ path: '/:pathMatch(.*)*', redirect: '/devices' },

    ],
});

export default router
