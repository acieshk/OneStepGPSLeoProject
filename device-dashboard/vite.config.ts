import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import { quasar, transformAssetUrls } from '@quasar/vite-plugin'

// https://vite.dev/config/
export default defineConfig({
	plugins: [
		vue({
			template: { transformAssetUrls }, // This fixes the asset URLs transformation
		}),
		quasar({
			sassVariables: 'src/quasar-variables.sass', // Path to your custom Quasar variables (optional)
		}),
	],
	resolve: {
		alias: {
			'@': fileURLToPath(new URL('./src', import.meta.url))
		},
	},
})
