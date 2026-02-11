<template>
  <!-- Overlay -->
  <Transition name="overlay">
    <div 
      v-if="show" 
      class="fixed inset-0 bg-black/50 z-40"
      @click="$emit('update:show', false)"
    ></div>
  </Transition>

  <!-- Panel -->
  <Transition name="slide">
    <div 
      v-if="show"
      class="fixed top-0 right-0 h-full w-full max-w-md bg-surface/95 border-l border-white/10 shadow-2xl z-50 flex flex-col"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-white/10 bg-surface/50">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-lg bg-gradient-to-br from-accent-purple to-accent-pink flex items-center justify-center">
            <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
          </div>
          <div>
            <h2 class="text-lg font-bold text-white">历史记录</h2>
            <p class="text-xs text-gray-400">查看您之前提交的代码</p>
          </div>
        </div>
        <button 
          @click="$emit('update:show', false)"
          class="w-8 h-8 rounded-lg bg-white/5 hover:bg-white/10 flex items-center justify-center text-gray-400 hover:text-white transition-all"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-4 space-y-3">
        <!-- Loading State -->
        <div v-if="loading" class="flex items-center justify-center h-40">
          <svg class="animate-spin h-8 w-8 text-accent-cyan" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </div>

        <!-- Empty State -->
        <div v-else-if="records.length === 0" class="flex flex-col items-center justify-center h-40 text-gray-500">
          <svg class="w-16 h-16 mb-4 opacity-30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
          </svg>
          <p class="text-sm">暂无历史记录</p>
          <p class="text-xs text-gray-600 mt-1">运行代码后会自动保存</p>
        </div>

        <!-- Record List -->
        <div 
          v-else
          v-for="record in records" 
          :key="record.id"
          class="bg-background/50 rounded-xl border border-white/5 overflow-hidden hover:border-accent-purple/30 transition-all duration-300 group cursor-pointer"
          @click="$emit('load-code', record)"
        >
          <!-- Record Header -->
          <div class="px-4 py-3 bg-surface/30 border-b border-white/5 flex items-center justify-between">
            <div class="flex items-center gap-3">
              <!-- Language Icon -->
              <div :class="[
                'w-8 h-8 rounded-lg flex items-center justify-center overflow-hidden',
                getLanguageColor(record.language)
              ]">
                <img :src="getLanguageIcon(record.language)" class="w-5 h-5 object-contain" />
              </div>
              <div>
                <p class="text-sm font-medium text-white">{{ getLanguageName(record.language) }}</p>
                <p class="text-xs text-gray-500">{{ formatTime(record.createdAt) }}</p>
              </div>
            </div>
            <!-- Status Badge -->
            <span :class="[
              'px-2 py-1 rounded-full text-xs font-medium',
              statusColors[record.status] || 'bg-gray-600/50 text-gray-300'
            ]">
              {{ statusLabels[record.status] || record.status }}
            </span>
          </div>

          <!-- Code Preview -->
          <div class="px-4 py-3">
            <pre class="text-xs text-gray-400 font-mono line-clamp-3 overflow-hidden">{{ record.code }}</pre>
          </div>

          <!-- Action Hint -->
          <div class="px-4 py-2 bg-accent-purple/5 border-t border-white/5 opacity-0 group-hover:opacity-100 transition-opacity">
            <p class="text-xs text-accent-purple flex items-center gap-1">
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
              </svg>
              点击加载此代码
            </p>
          </div>
        </div>
      </div>

      <!-- Refresh Button -->
      <div class="px-4 py-4 border-t border-white/10 bg-surface/50">
        <button 
          @click="fetchHistory"
          :disabled="loading"
          class="w-full py-3 rounded-lg bg-gradient-to-r from-accent-purple/20 to-accent-cyan/20 border border-white/10 text-white font-medium hover:from-accent-purple/30 hover:to-accent-cyan/30 transition-all flex items-center justify-center gap-2 disabled:opacity-50"
        >
          <svg :class="['w-4 h-4', loading ? 'animate-spin' : '']" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
          </svg>
          刷新记录
        </button>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, watch, onMounted } from 'vue'
import { fetchLanguages, fetchSubmissions } from '../api'

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:show', 'load-code'])

const loading = ref(false)
const records = ref([])
const languagesData = ref([])

// 图标映射 - 用于加载本地图标
const iconModules = import.meta.glob('../assets/icons/*.webp', { eager: true })

// 根据图标文件名获取图标路径
const getIconUrl = (iconName) => {
  const path = `../assets/icons/${iconName.replace('.png', '.webp')}`
  return iconModules[path]?.default || ''
}

// 从 API 加载语言配置
const loadLanguages = async () => {
  try {
    languagesData.value = await fetchLanguages()
  } catch (error) {
    console.error('Failed to load languages:', error)
  }
}

// 获取语言显示名称
const getLanguageName = (value) => {
  return languagesData.value.find(l => l.value === value)?.label || value
}

// 获取语言图标
const getLanguageIcon = (value) => {
  const lang = languagesData.value.find(l => l.value === value)
  return lang ? getIconUrl(lang.icon) : ''
}

// 语言颜色映射
const languageColors = {
  python: 'bg-yellow-500/20',
  java: 'bg-orange-500/20',
  cpp: 'bg-blue-500/20',
  c: 'bg-indigo-500/20',
  go: 'bg-cyan-500/20',

  javascript: 'bg-yellow-400/20',
  rust: 'bg-orange-600/20',
  csharp: 'bg-indigo-600/20',
  typescript: 'bg-blue-600/20'
}

const getLanguageColor = (value) => {
  return languageColors[value] || 'bg-gray-600/20'
}

const statusColors = {
  Success: 'bg-green-500/20 text-green-400',
  Failed: 'bg-red-500/20 text-red-400',
  Timeout: 'bg-yellow-500/20 text-yellow-400',
  Running: 'bg-blue-500/20 text-blue-400',
  Pending: 'bg-gray-500/20 text-gray-400'
}

const statusLabels = {
  Success: '成功',
  Failed: '失败',
  Timeout: '超时',
  Running: '运行中',
  Pending: '等待中'
}

const formatTime = (dateStr) => {
  const date = new Date(dateStr)
  const now = new Date()
  const diff = now - date
  
  // Less than 1 minute
  if (diff < 60000) return '刚刚'
  
  // Less than 1 hour
  if (diff < 3600000) {
    const mins = Math.floor(diff / 60000)
    return `${mins} 分钟前`
  }
  
  // Less than 1 day
  if (diff < 86400000) {
    const hours = Math.floor(diff / 3600000)
    return `${hours} 小时前`
  }
  
  // Format as date
  return date.toLocaleDateString('zh-CN', { 
    month: 'short', 
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const fetchHistory = async () => {
  loading.value = true
  try {
    records.value = await fetchSubmissions()
  } catch (error) {
    console.error('Failed to fetch history:', error)
    records.value = []
  } finally {
    loading.value = false
  }
}

// Fetch when panel opens
watch(() => props.show, (newVal) => {
  if (newVal) {
    fetchHistory()
  }
})

onMounted(() => {
  loadLanguages()
})
</script>

<style scoped>
.overlay-enter-active,
.overlay-leave-active {
  transition: opacity 0.3s ease;
}
.overlay-enter-from,
.overlay-leave-to {
  opacity: 0;
}

.slide-enter-active,
.slide-leave-active {
  transition: transform 0.3s ease;
}
.slide-enter-from,
.slide-leave-to {
  transform: translateX(100%);
}

.line-clamp-3 {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
