import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { VueQueryPlugin } from '@tanstack/vue-query'
import router from './router'
import './assets/main.css'
import App from './App.vue'

// Instantiate the app with its plugins: Pinia (session), router (screens),
// TanStack Query (server state). The auth store binds itself to the api
// client on first use — see stores/auth.ts.
createApp(App).use(createPinia()).use(VueQueryPlugin).use(router).mount('#app')
