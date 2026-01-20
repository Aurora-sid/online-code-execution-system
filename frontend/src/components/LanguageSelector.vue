<template>
  <div class="relative" ref="dropdownRef">
    <!-- 触发按钮 -->
    <button
      @click="toggleDropdown"
      type="button"
      class="flex items-center gap-3 bg-zinc-100 border border-zinc-200 rounded-lg px-4 py-2 text-zinc-700 hover:bg-zinc-200 hover:border-zinc-300 focus:outline-none focus:ring-2 focus:ring-accent-purple/50 cursor-pointer transition-all duration-200 min-w-[140px]"
      :class="{ 'ring-2 ring-accent-purple/50 border-accent-purple/50': isOpen }"
    >
      <!-- 当前语言图标 -->
      <img :src="currentLangIcon" class="w-5 h-5 flex-shrink-0 object-contain" />
      <span class="font-medium flex-1 text-left">{{ currentLangLabel }}</span>
      <!-- 箭头图标 -->
      <svg 
        class="w-4 h-4 text-zinc-500 transition-transform duration-200" 
        :class="{ 'rotate-180': isOpen }"
        fill="none" stroke="currentColor" viewBox="0 0 24 24"
      >
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
      </svg>
    </button>

    <!-- 下拉菜单 -->
    <Transition
      enter-active-class="transition ease-out duration-150"
      enter-from-class="opacity-0 scale-95 -translate-y-1"
      enter-to-class="opacity-100 scale-100 translate-y-0"
      leave-active-class="transition ease-in duration-100"
      leave-from-class="opacity-100 scale-100 translate-y-0"
      leave-to-class="opacity-0 scale-95 -translate-y-1"
    >
      <div 
        v-if="isOpen"
        class="absolute top-full left-0 mt-2 w-full bg-white border border-zinc-200 rounded-lg shadow-xl shadow-black/10 overflow-hidden z-50"
      >
        <!-- 加载中状态 -->
        <div v-if="loading" class="flex items-center justify-center py-4">
          <svg class="animate-spin h-5 w-5 text-accent-purple" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </div>
        <!-- 语言列表 -->
        <button
          v-else
          v-for="lang in languages"
          :key="lang.value"
          @click="selectLanguage(lang.value)"
          @mouseenter="hoveredValue = lang.value"
          @mouseleave="hoveredValue = null"
          class="flex items-center gap-3 w-full px-4 py-2.5 text-left cursor-pointer transition-colors"
          :class="[
            modelValue === lang.value ? 'bg-accent-purple/10 text-accent-purple' : 'text-zinc-700',
            hoveredValue === lang.value && modelValue !== lang.value ? 'bg-zinc-100 text-zinc-900' : ''
          ]"
        >
          <img :src="getIconUrl(lang.icon)" class="w-5 h-5 flex-shrink-0 object-contain" />
          <span class="font-medium">{{ lang.label }}</span>
        </button>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { fetchLanguages } from '../api'

const props = defineProps({
  modelValue: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['update:modelValue'])

const isOpen = ref(false)
const hoveredValue = ref(null)
const dropdownRef = ref(null)
const loading = ref(false)
const languages = ref([])

// 图标映射 - 用于加载本地图标
const iconModules = import.meta.glob('../assets/icons/*.png', { eager: true })

// 根据图标文件名获取图标路径
const getIconUrl = (iconName) => {
  const path = `../assets/icons/${iconName}`
  return iconModules[path]?.default || ''
}

// 从 API 加载语言列表
const loadLanguages = async () => {
  loading.value = true
  try {
    const data = await fetchLanguages()
    languages.value = data
  } catch (error) {
    console.error('Failed to load languages:', error)
  } finally {
    loading.value = false
  }
}

const currentLangLabel = computed(() => {
  return languages.value.find(l => l.value === props.modelValue)?.label || 'Select'
})

const currentLangIcon = computed(() => {
  const lang = languages.value.find(l => l.value === props.modelValue)
  return lang ? getIconUrl(lang.icon) : ''
})

const toggleDropdown = () => {
  isOpen.value = !isOpen.value
}

const selectLanguage = (value) => {
  emit('update:modelValue', value)
  isOpen.value = false
}

// 点击外部关闭下拉菜单
const handleClickOutside = (event) => {
  if (dropdownRef.value && !dropdownRef.value.contains(event.target)) {
    isOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  loadLanguages()
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>
