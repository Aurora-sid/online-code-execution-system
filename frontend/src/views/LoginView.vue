<template>
  <div class="min-h-screen w-full bg-[#0F172A] text-slate-200 flex items-center justify-center font-sans selection:bg-cyan-500/30 relative overflow-hidden pb-15">
    <!-- 背景特效 -->
    <div class="absolute inset-0 z-0">
      <img src="../assets/login.webp" alt="Background" class="w-full h-full object-cover opacity-60" />
      <!-- 网格覆盖层（科技感装饰） -->
      <div class="absolute inset-0 bg-[linear-gradient(rgba(255,255,255,0.03)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.03)_1px,transparent_1px)] bg-[size:50px_50px] opacity-20"></div>
    </div>

    <!-- 卡片容器 -->
    <div class="relative z-10 w-full max-w-md mx-4">

      <!-- Logo 区域 -->
      <div class="flex flex-col items-center justify-center mb-8 gap-4 group">
        <div class="w-20 h-20 bg-gradient-to-br from-cyan-400 to-blue-600 rounded-2xl flex items-center justify-center shadow-lg shadow-cyan-500/30 group-hover:scale-105 transition-transform duration-500 ring-1 ring-white/20 overflow-hidden">
          <img src="../assets/Dog.webp" alt="Logo" class="w-full h-full object-cover">
        </div>
        <!-- <h1 class="text-6xl text-white drop-shadow-lg mt-2 font-normal" style="font-family: 'Zhi Mang Xing', cursive;"> -->
          <h1 class="text-6xl text-white drop-shadow-lg mt-2 font-bold" style="font-family: 'Ma Shan Zheng', cursive;">
          <span class="text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 to-blue-500 pl-[0.1em]">在线编程系统</span>
        </h1>
      </div>

      <!-- 卡片容器 -->
      <div class="bg-slate-900/50 backdrop-blur-xl rounded-3xl border border-white/20 shadow-2xl overflow-hidden">
        
        <!-- 标签页（分段控制风格） -->
        <div class="p-6 pb-0">
          <div class="flex bg-black/20 p-1.5 rounded-2xl gap-2">
            <button 
              @click="activeTab = 'login'"
              :class="[
                'flex-1 py-2.5 text-sm font-bold rounded-xl transition-all duration-300 relative overflow-hidden',
                activeTab === 'login' 
                  ? 'text-white bg-white/10 shadow-sm ring-1 ring-white/10' 
                  : 'text-slate-400 hover:text-slate-200 hover:bg-white/5'
              ]"
            >
              登录
            </button>
            <button 
              @click="activeTab = 'register'"
              :class="[
                'flex-1 py-2.5 text-sm font-bold rounded-xl transition-all duration-300 relative overflow-hidden',
                activeTab === 'register' 
                  ? 'text-white bg-white/10 shadow-sm ring-1 ring-white/10' 
                  : 'text-slate-400 hover:text-slate-200 hover:bg-white/5'
              ]"
            >
              注册
            </button>
          </div>
        </div>

        <!-- 表单 -->
        <form @submit.prevent="handleSubmit" class="min-h-[360px] p-6 pt-6 space-y-6">
          <!-- 用户名 -->
          <div class="space-y-2">
            <label class="text-sm font-semibold text-slate-300 flex items-center gap-2 ml-1">
              <svg class="w-4 h-4 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
              </svg>
              用户名
            </label>
            <div class="relative group">
              <input 
                v-model="form.username"
                type="text" 
                placeholder="请输入用户名"
                autocomplete="username"
                class="w-full bg-slate-900/50 border border-slate-600/50 rounded-xl px-4 py-3.5 text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500/50 focus:ring-4 focus:ring-cyan-500/10 transition-all duration-300 hover:border-slate-500 hover:bg-slate-900/70"
                required
              />
            </div>
          </div>

          <!-- 密码 -->
          <div class="space-y-2">
            <label class="text-sm font-semibold text-slate-300 flex items-center gap-2 ml-1">
              <svg class="w-4 h-4 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
              </svg>
              密码
            </label>
            <div class="relative group">
              <input 
                v-model="form.password"
                :type="showPassword ? 'text' : 'password'" 
                placeholder="请输入密码"
                autocomplete="current-password"
                class="w-full bg-slate-900/50 border border-slate-600/50 rounded-xl px-4 py-3.5 pr-12 text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500/50 focus:ring-4 focus:ring-cyan-500/10 transition-all duration-300 hover:border-slate-500 hover:bg-slate-900/70"
                required
              />
              <button 
                type="button"
                @click="showPassword = !showPassword"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-cyan-400 transition-colors p-1"
              >
                <svg v-if="!showPassword" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
                </svg>
                <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- 确认密码（仅注册） -->
          <div v-if="activeTab === 'register'" class="space-y-2">
            <label class="text-sm font-semibold text-slate-300 flex items-center gap-2 ml-1">
              <svg class="w-4 h-4 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"/>
              </svg>
              确认密码
            </label>
            <div class="relative group">
              <input 
                v-model="form.confirmPassword"
                type="password" 
                placeholder="请再次输入密码"
                autocomplete="new-password"
                class="w-full bg-slate-900/50 border border-slate-600/50 rounded-xl px-4 py-3.5 text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500/50 focus:ring-4 focus:ring-cyan-500/10 transition-all duration-300 hover:border-slate-500 hover:bg-slate-900/70"
                required
              />
            </div>
          </div>

          <!-- 错误消息 -->
          <div v-if="error" class="bg-red-500/10 border border-red-500/20 rounded-xl p-4 flex items-center gap-3 animate-in fade-in slide-in-from-top-1">
            <svg class="w-5 h-5 text-red-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <span class="text-red-200 text-sm font-medium">{{ error }}</span>
          </div>

          <!-- 成功消息 -->
          <div v-if="success" class="bg-emerald-500/10 border border-emerald-500/20 rounded-xl p-4 flex items-center gap-3 animate-in fade-in slide-in-from-top-1">
            <svg class="w-5 h-5 text-emerald-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <span class="text-emerald-200 text-sm font-medium">{{ success }}</span>
          </div>

          <!-- 提交按钮 -->
          <button 
            type="submit"
            :disabled="loading"
            class="w-full py-3.5 rounded-xl font-bold text-white shadow-lg shadow-blue-500/20 transition-all duration-300 transform hover:-translate-y-0.5 active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none flex items-center justify-center gap-2 bg-gradient-to-r from-cyan-500 to-blue-600 hover:from-cyan-400 hover:to-blue-500 border border-white/10"
          >
            <svg v-if="loading" class="animate-spin h-5 w-5" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            <span class="tracking-wide">{{ loading ? '处理中...' : (activeTab === 'login' ? '登 录' : '注 册') }}</span>
          </button>
        </form>

        <!-- 底部 -->
        <div class="px-8 pb-8 text-center bg-black/10 pt-4 border-t border-white/5">
          <p class="text-slate-400 text-sm">
            {{ activeTab === 'login' ? '还没有账号？' : '已有账号？' }}
            <button 
              @click="activeTab = activeTab === 'login' ? 'register' : 'login'"
              class="text-cyan-400 hover:text-cyan-300 transition-colors font-semibold ml-1"
            >
              {{ activeTab === 'login' ? '立即注册' : '去登录' }}
            </button>
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
// 引入API模块
import api from '../api/index'
// 引入认证状态管理
import { useAuth } from '../stores/auth'
import loginBg from '../assets/login.webp'

const router = useRouter()
const route = useRoute()
const { login } = useAuth()

const activeTab = ref('login')
watch(activeTab, () => {
  // 切换标签页时始终清空表单，尤其是确保注册表单为空
  form.username = ''
  form.password = ''
  form.confirmPassword = ''
  error.value = ''
  success.value = ''
})
// 使用ref和reactive来管理表单状态、加载状态、错误和成功消息，以及密码可见性
const loading = ref(false)
const error = ref('')
const success = ref('')
const showPassword = ref(false)

const form = reactive({
  username: '',
  password: '',
  confirmPassword: ''
})

onMounted(() => {
  // 清除持久状态并处理登出后的上一个用户
  form.username = route.query.lastUser || ''
  form.password = ''
  form.confirmPassword = ''
})

// 密码最小长度检查仅使用本地验证
// 登录时不需要确认密码，注册时需要，并且两者必须匹配
const handleSubmit = async () => {
  error.value = ''
  success.value = ''
  
  // 注册时检查密码和确认密码是否匹配
  if (activeTab.value === 'register' && form.password !== form.confirmPassword) {
    error.value = '两次输入的密码不一致'
    return
  }
  
  if (form.password.length < 6) {
    error.value = '密码长度至少为6位'
    return
  }
  
  loading.value = true
  
  try {
    // 登录和注册的API调用逻辑
    if (activeTab.value === 'login') {
      const response = await api.post('/login', {
        username: form.username,
        password: form.password
      })
      // 存储token和用户信息到全局状态管理，并跳转到主页
      login(response.data.token, response.data.user)
      router.push('/')
    } else {
      await api.post('/register', {
        username: form.username,
        password: form.password
      })
      
      success.value = '注册成功！正在跳转到登录...'
      setTimeout(() => {
        activeTab.value = 'login'
        success.value = ''
      }, 1500)
    }
  } catch (err) {
    error.value = err.response?.data?.error || '操作失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>
