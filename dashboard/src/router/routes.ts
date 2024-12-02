import { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: () => import('layouts/MainLayout.vue'),
    children: [
      { path: '', component: () => import('pages/mapPage.vue') }, // Map page route
      { path: 'devices/:id', name: 'DeviceEdit', component: () => import('components/deviceEdit.vue') }, // DeviceEdit route *nested* inside MainLayout
	  { path: 'user', name: 'User', component: () => import('components/userComponent.vue') } // DeviceEdit route *nested* inside MainLayout
    ],
  },  

  // Always leave this as last one,
  {
    path: '/:catchAll(.*)*',
    component: () => import('pages/ErrorNotFound.vue'),
  },
];

export default routes;