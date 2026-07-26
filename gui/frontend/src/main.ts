import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Antd from 'ant-design-vue'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import { useThemeStore } from './stores/theme'
import { vReveal, vRevealGroup } from './composables/useMotion'
import 'ant-design-vue/dist/reset.css'
import './assets/styles/variables.css'
import './themes/derivation.css'

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.use(Antd)
app.use(i18n)

// 全局动画指令（v-reveal / v-reveal-group）；import 即触发 useMotion 加载偏好 + gsap.defaults
app.directive('reveal', vReveal)
app.directive('reveal-group', vRevealGroup)

// 主题初始化（mount 前同步，配合 index.html FOUC 脚本首帧即正确主题）
useThemeStore().initialize()

app.mount('#app')

document.addEventListener('contextmenu', (e) => e.preventDefault())
