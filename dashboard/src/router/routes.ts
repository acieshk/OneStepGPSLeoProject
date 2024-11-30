import { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('layouts/MainLayout.vue'),
    children: [{ path: '', component: () => import('pages/mapPage.vue') }],
  },
  {
    path: '/devices/:id', // Dynamic route for editing
    name: 'DeviceEdit',         // Route name
    component: () => import('../components/deviceEdit.vue') // Your DeviceEdit component
  },
  // Always leave this as last one,
  // but you can also remove it
  {
    path: '/:catchAll(.*)*',
    component: () => import('pages/ErrorNotFound.vue'),
  },
];

export default routes;
