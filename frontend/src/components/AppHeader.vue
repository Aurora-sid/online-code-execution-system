<template>
  <header class="h-14 border-b border-border flex items-center justify-between px-2 md:px-4 z-10 bg-gray-300/80 shadow-sm">
    <!-- 左侧：Logo -->
    <div class="flex items-center gap-2 md:gap-3 flex-shrink-0">
      <div class="w-10 h-10 md:w-12 md:h-12 rounded-lg bg-gradient-to-br from-primary to-secondary flex items-center justify-center shadow-glow-primary flex-shrink-0">
        <img src="../assets/TX.webp" class="w-full h-full object-contain">
      </div>
      <div class="hidden sm:block">
        <h1 class="font-bold text-sm md:text-lg tracking-tight">
          <span class="text-transparent bg-clip-text bg-gradient-to-r from-indigo-500 via-purple-500 to-pink-500 animate-gradient-x tracking-[0.1em] [word-spacing:0.3em] md:[word-spacing:0.6em]">Aurora Code</span>
        </h1>
        <div class="hidden md:block text-[9px] font-bold tracking-widest uppercase text-transparent bg-clip-text bg-gradient-to-r from-cyan-500 to-blue-600">Online Execution Environment</div>
      </div>
    </div>

    <!-- 右侧：操作区 -->
    <div class="flex items-center gap-1 md:gap-4">
      <!-- 语言选择器 -->
      <LanguageSelector v-model="internalLanguage" />

      <!-- 操作按钮 -->
      <div class="flex items-center gap-1 md:gap-2">
        <!-- 文件夹按钮（切换侧边栏） -->
        <button 
          @click="$emit('toggle-sidebar')"
          class="flex items-center justify-center gap-2 p-2 md:px-4 md:py-1.5 rounded-lg bg-gradient-to-r from-blue-400 to-cyan-500 text-white font-medium text-sm transition-all active:scale-95 hover:scale-105 shadow-md hover:shadow-lg hover:brightness-110 border-none"
          :title="sidebarOpen ? '收起文件夹' : '打开文件夹'"
        >
          <i :class="sidebarOpen ? 'ph ph-folder-minus' : 'ph ph-folder-open'" class="text-sm"></i>
          <span class="hidden lg:inline">{{ sidebarOpen ? '收起文件夹' : '打开文件夹' }}</span>
        </button>

        <!-- 历史记录按钮 -->
        <button 
          @click="$emit('toggle-history')"
          class="flex items-center justify-center gap-2 p-2 md:px-4 md:py-1.5 rounded-lg bg-gradient-to-r from-orange-400 to-pink-500 text-white font-medium text-sm transition-all active:scale-95 hover:scale-105 shadow-md hover:shadow-lg hover:brightness-110 border-none"
          title="历史记录"
        >
          <i class="ph ph-clock-counter-clockwise text-sm"></i>
          <span class="hidden lg:inline">历史记录</span>
        </button>

        <!-- 保存按钮 -->
        <button 
          @click="$emit('save')"
          class="flex items-center justify-center gap-2 p-2 md:px-4 md:py-1.5 rounded-lg bg-gradient-to-r from-emerald-400 to-teal-500 text-white font-medium text-sm transition-all active:scale-95 hover:scale-105 shadow-md hover:shadow-lg hover:brightness-110 border-none"
          title="保存至本地"
        >
          <i class="ph ph-download-simple text-sm"></i>
          <span class="hidden lg:inline">保存至本地</span>
        </button>

        <!-- AI 分析按钮 -->
        <button 
          @click="$emit('toggle-analysis')"
          class="flex items-center justify-center gap-2 p-2 md:px-4 md:py-1.5 rounded-lg bg-gradient-to-r from-purple-500 to-indigo-600 text-white font-medium text-sm transition-all active:scale-95 hover:scale-105 shadow-md hover:shadow-lg hover:brightness-110 border-none"
          title="智能分析"
        >
          <i class="ph ph-sparkle text-sm"></i>
          <span class="hidden lg:inline">智能分析</span>
        </button>

        <!-- 运行按钮 -->
        <button 
          @click="$emit('run')"
          :disabled="loading"
          class="flex items-center justify-center gap-2 p-2 md:px-5 md:py-1.5 rounded-lg bg-gradient-to-r from-[#6366f1] to-[#8b5cf6] text-white font-semibold text-sm shadow-lg shadow-indigo-500/30 hover:shadow-indigo-500/50 hover:scale-105 active:scale-95 transition-all disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:scale-100"
        >
          <img v-if="!loading" src="../assets/Vue-icons/run.webp" alt="运行" class="w-5 h-5 md:w-6 md:h-6">
          <i v-else class="ph ph-spinner animate-spin text-lg"></i>
          <span class="hidden md:inline tracking-[0.3em]">{{ loading ? '运行中...' : '运行代码' }}</span>
        </button>
      </div>

      <!-- 用户头像 -->
      <div class="relative flex-shrink-0" ref="userMenuRef">
        <button
          @click="showUserMenu = !showUserMenu"
          class="h-8 md:h-10 px-2 md:px-3 rounded-xl bg-gradient-to-r from-indigo-50 via-purple-50 to-pink-50 hover:from-indigo-100 hover:via-purple-100 hover:to-pink-100 text-gray-700 text-sm font-bold flex items-center gap-1 md:gap-2 shadow-sm hover:shadow-md transition-all active:scale-95 hover:scale-105 border border-indigo-100/50"
        >
          <img src="../assets/Vue-icons/user.webp" alt="用户头像" class="w-5 h-5 md:w-6 md:h-6 rounded-full">
          <span class="w-2 h-2 rounded-full bg-emerald-400 hidden md:block"></span>
          <span class="hidden sm:inline">{{ user?.username || 'Guest' }}</span>
        </button>

        <!-- 下拉菜单 -->
        <Transition name="dropdown">
          <div 
            v-if="showUserMenu"
            class="absolute right-0 mt-3 w-56 bg-white rounded-xl shadow-2xl overflow-hidden origin-top-right z-50 border border-gray-100"
          >
            <div class="px-4 py-3 border-b border-gray-100 bg-gray-50">
              <p class="text-sm font-semibold text-gray-900">{{ user?.username }}</p>
              <p class="text-xs text-gray-500 mt-0.5 flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span> Online
                <span v-if="user?.role === 'admin'" class="ml-2 px-1.5 py-0.5 bg-purple-100 text-purple-600 rounded text-[10px] font-medium">管理员</span>
              </p>
            </div>
            <div class="p-1">
              <button 
                v-if="user?.role === 'admin'"
                @click="goToAdmin"
                class="w-full px-3 py-2 text-left text-sm text-purple-600 hover:bg-purple-50 rounded-lg flex items-center gap-2.5 transition-colors"
              >
                <i class="ph ph-gauge text-lg"></i>
                管理后台
              </button>
              <button 
                @click="handleLogout"
                class="w-full px-3 py-2 text-left text-sm text-red-500 hover:bg-red-50 hover:text-red-600 rounded-lg flex items-center gap-2.5 transition-colors"
              >
                <i class="ph ph-sign-out text-lg"></i>
                退出登录
              </button>
            </div>
          </div>
        </Transition>
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
  },
  sidebarOpen: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:currentLanguage', 'run', 'save', 'toggle-history', 'toggle-sidebar', 'toggle-analysis'])

const router = useRouter()
const { user, logout } = useAuth()

const showUserMenu = ref(false)
const userMenuRef = ref(null)

// v-model 绑定的计算属性
const internalLanguage = computed({
  get: () => props.currentLanguage,
  set: (value) => emit('update:currentLanguage', value)
})

const handleLogout = () => {
  const lastUser = user.value?.username
  logout()
  router.push({ path: '/login', query: { lastUser } })
}

const goToAdmin = () => {
  showUserMenu.value = false
  router.push('/admin')
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
