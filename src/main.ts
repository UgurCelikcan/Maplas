import './leaflet-setup'
import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import i18n, { setLanguage } from './i18n'

const app = createApp(App)
app.use(i18n)

// Load the saved language before showing the page (no flash of Turkish text).
setLanguage(localStorage.getItem('lang') || 'tr')
  .catch(() => {})
  .finally(() => app.mount('#app'))
