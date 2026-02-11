import { createApp } from 'vue'
import { createPinia } from 'pinia'

// 本地字体 - 替代 Google Fonts，消除国内网络阻塞
import '@fontsource/jetbrains-mono/400.css'
import '@fontsource/jetbrains-mono/500.css'
import '@fontsource/jetbrains-mono/700.css'
import '@fontsource/outfit/300.css'
import '@fontsource/outfit/400.css'
import '@fontsource/outfit/500.css'
import '@fontsource/outfit/600.css'
import '@fontsource/outfit/700.css'

import './style.css'
import App from './App.vue'
import router from './router'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)
app.mount('#app')

// 延迟加载 Phosphor Icons - 避免阻塞首屏渲染
// 在 App 挂载后 100ms 异步加载图标库
setTimeout(() => {
    import('@phosphor-icons/web/regular')
}, 100)

