<template>
  <header class="bg-white/90 backdrop-blur-md border-b border-zinc-200 sticky top-0 z-50">
    <div class="px-6 py-4">
      <div class="flex justify-between items-center relative">
        <!-- Logo 部分 -->
        <div class="h-10 w-10 flex items-center justify-center group cursor-default group-hover:scale-105 transition-transform duration-300">
          <img src="../assets/nano-banana.png" alt="Nano Banana" class="w-full h-full object-contain drop-shadow-md" />
        </div>
        
        <!-- 居中标题 - 小屏幕隐藏 -->
        <div class="absolute left-1/2 -translate-x-1/2 hidden lg:block">
          <h1 class="text-xl font-bold tracking-[0.5em] bg-clip-text text-transparent bg-gradient-to-r from-cyan-500 to-blue-500 font-sans whitespace-nowrap">
            在线编程系统
          </h1>
        </div>
        
        <!-- 操作部分 -->
        <div class="flex items-center gap-2 sm:gap-4">
          <LanguageSelector v-model="internalLanguage" />
          
          <div class="h-6 w-[1px] bg-zinc-300 mx-1 hidden sm:block"></div>

          <!-- 历史记录按钮 -->
          <button 
            @click="$emit('toggle-history')"
            class="p-2.5 rounded-lg bg-zinc-100 border border-zinc-200 hover:bg-zinc-200 hover:border-zinc-300 text-zinc-500 hover:text-zinc-800 transition-all duration-200 group relative"
            title="查看历史记录"
          >
            <svg class="w-5 h-5 group-hover:scale-110 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <span class="absolute -bottom-10 left-1/2 -translate-x-1/2 px-2 py-1 bg-zinc-900 border border-white/10 rounded text-xs text-white opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none">历史记录</span>
          </button>

          <!-- 保存按钮 -->
          <button 
            @click="$emit('save')"
            class="p-2.5 rounded-lg bg-zinc-100 border border-zinc-200 hover:bg-zinc-200 hover:border-zinc-300 text-zinc-500 hover:text-zinc-800 transition-all duration-200 group relative"
            title="保存到本地"
          >
            <svg class="w-5 h-5 group-hover:scale-110 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4"/>
            </svg>
            <span class="absolute -bottom-10 left-1/2 -translate-x-1/2 px-2 py-1 bg-zinc-900 border border-white/10 rounded text-xs text-white opacity-0 group-hover:opacity-100 transition-opacity whitespace-nowrap pointer-events-none">保存代码</span>
          </button>

          <!-- 运行按钮（高可见性） -->
          <button 
            @click="$emit('run')"
            :disabled="loading"
            class="bg-gradient-to-r from-emerald-600 to-teal-700 hover:from-emerald-500 hover:to-teal-600 text-white px-3 sm:px-6 py-2 sm:py-2.5 rounded-xl font-bold shadow-lg shadow-emerald-900/50 hover:shadow-emerald-900/70 transition-all duration-300 transform hover:-translate-y-0.5 active:translate-y-0 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none flex items-center gap-2 ring-1 ring-white/10"
          >
            <template v-if="loading">
              <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              <span class="hidden sm:inline">运行中...</span>
            </template>
            <template v-else>
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
              <span class="hidden sm:inline">运行代码</span>
            </template>
          </button>

          <!-- User Menu -->
          <div class="relative ml-2" ref="userMenuRef">
            <button 
              @click="showUserMenu = !showUserMenu"
              class="flex items-center gap-2 px-3 py-2 rounded-lg bg-zinc-100 border border-zinc-200 hover:bg-zinc-200 hover:border-zinc-300 transition-all duration-200"
            >
              <!-- 头像已移除 -->
              <span class="text-sm font-medium text-zinc-700 hidden sm:inline">{{ user?.username || '用户' }}</span>
              <svg class="w-3.5 h-3.5 text-zinc-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
              </svg>
            </button>

            <!-- 下拉菜单 -->
            <Transition name="dropdown">
              <div 
                v-if="showUserMenu"
                class="absolute right-0 mt-3 w-56 bg-zinc-900/95 backdrop-blur-xl rounded-xl border border-white/10 shadow-2xl overflow-hidden origin-top-right z-50 ring-1 ring-black/5"
              >
                <div class="px-5 py-4 border-b border-white/10 bg-white/5">
                  <p class="text-sm font-bold text-white">{{ user?.username }}</p>
                  <p class="text-xs text-zinc-400 mt-0.5 flex items-center gap-1.5">
                    <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span> 在线
                  </p>
                </div>
                <div class="p-1">
                  <button 
                    @click="handleLogout"
                    class="w-full px-3 py-2.5 text-left text-sm text-red-400 hover:bg-red-500/10 hover:text-red-300 rounded-lg flex items-center gap-2.5 transition-colors"
                  >
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/>
                    </svg>
                    退出登录
                  </button>
                </div>
              </div>
            </Transition>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import LanguageSelector from './LanguageSelector.vue'
import { useAuth } from '../stores/auth'

const props = defineProps({
  currentLanguage: {
    type: String,
    required: true
  },
  loading: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:currentLanguage', 'run', 'save', 'toggle-history'])

const router = useRouter()
const { user, logout } = useAuth()

const showUserMenu = ref(false)
const userMenuRef = ref(null)

// 为 v-model 绑定创建计算属性
const internalLanguage = computed({
  get: () => props.currentLanguage,
  set: (value) => emit('update:currentLanguage', value)
})

const userInitial = computed(() => {
  return user.value?.username?.charAt(0).toUpperCase() || 'U'
})

const handleLogout = () => {
  const lastUser = user.value?.username
  logout()
  router.push({ path: '/login', query: { lastUser } })
}

// 点击外部关闭用户菜单
const handleClickOutside = (event) => {
  if (showUserMenu.value && userMenuRef.value && !userMenuRef.value.contains(event.target)) {
    showUserMenu.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.dropdown-enter-active,
.dropdown-leave-active {
  transition: all 0.2s ease;
}
.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
