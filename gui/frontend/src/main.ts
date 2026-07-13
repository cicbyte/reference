import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import 'ant-design-vue/dist/reset.css'
import './assets/styles/variables.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(Antd)
app.use(i18n)
app.mount('#app')

document.addEventListener('contextmenu', (e) => e.preventDefault())
