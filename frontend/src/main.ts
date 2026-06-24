import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import Antd from 'ant-design-vue'
import 'ant-design-vue/dist/reset.css'
import * as AntdIconsVue from '@ant-design/icons-vue'
import App from './App.vue'
import router from './router'
import './styles/global.css'
import './styles/ellipsis.css'
import vEllipsis from './directives/ellipsis'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(ElementPlus, { locale: zhCn })
app.use(Antd)
app.directive('ellipsis', vEllipsis)

for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  if (!app._context.components[key]) {
    app.component(key, component)
  }
}

for (const [key, component] of Object.entries(AntdIconsVue)) {
  if (!app._context.components[key]) {
    app.component('Antd' + key, component)
  }
}

app.mount('#app')