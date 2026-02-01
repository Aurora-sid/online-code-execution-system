<template>
  <div class="relative" ref="selectorRef">
    <!-- Trigger Button -->
    <button 
      @click="toggleDropdown($event)"
      class="flex items-center gap-2 px-3 py-1.5 rounded-lg bg-gradient-to-r from-indigo-50 via-purple-50 to-pink-50 hover:from-indigo-100 hover:via-purple-100 hover:to-pink-100 border border-indigo-100/50 transition-all active:scale-95 hover:scale-105"
    >
      <img 
        v-if="currentLang?.icon && !currentLang.icon.startsWith('ph')" 
        :src="getIconUrl(currentLang.icon)" 
        class="w-6 h-6 object-contain" 
        alt="icon"
      />
      <i v-else :class="currentLangIcon" class="text-lg"></i>
      <span class="text-sm font-medium text-gray-700">{{ currentLangLabel }}</span>
      <!-- <i :class="isOpen ? 'ph ph-caret-up' : 'ph ph-caret-down'" class="text-xs text-gray-500 ml-1"></i> -->
    </button>

    <!-- Dropdown -->
    <Transition name="dropdown">
      <div 
        v-if="isOpen" 
        class="absolute top-full left-0 mt-2 w-48 bg-white rounded-xl shadow-2xl overflow-hidden z-50 border border-gray-100"
      >
        <div class="max-h-64 overflow-y-auto py-1">
          <button
            v-for="lang in languages"
            :key="lang.value"
            @click="selectLanguage(lang.value)"
            class="w-full px-3 py-2 flex items-center gap-2.5 text-left hover:bg-gray-50 transition-colors"
            :class="modelValue === lang.value ? 'bg-primary/5 text-primary' : 'text-gray-600'"
          >
            <img 
              v-if="lang.icon && !lang.icon.startsWith('ph')" 
              :src="getIconUrl(lang.icon)" 
              class="w-5 h-5 object-contain"
              alt="icon"
            />
            <i v-else :class="lang.icon" class="text-lg"></i>
            <span class="text-sm font-medium">{{ lang.label }}</span>
            <i 
              v-if="modelValue === lang.value" 
              class="ph ph-check ml-auto text-primary"
            ></i>
          </button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import axios from 'axios'

const props = defineProps({
  modelValue: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['update:modelValue'])

const isOpen = ref(false)
const selectorRef = ref(null)
const languages = ref([])

const languageIcons = {
  go: 'ph ph-file-code text-cyan-500',
  python: 'ph ph-file-py text-blue-500',
  java: 'ph ph-file-code text-orange-500',
  cpp: 'ph ph-file-cpp text-purple-500',
  c: 'ph ph-file-c text-blue-600',
  javascript: 'ph ph-file-js text-yellow-500',
  typescript: 'ph ph-file-ts text-blue-500',
  rust: 'ph ph-gear text-orange-600',
  ruby: 'ph ph-diamond text-red-500',
  php: 'ph ph-file-code text-indigo-500',
}

const defaultLanguages = [
  { value: "go", label: "Go", icon: "ph ph-file-code text-cyan-500" },
  { value: "python", label: "Python", icon: "ph ph-file-py text-blue-500" },
  { value: "java", label: "Java", icon: "ph ph-file-code text-orange-500" },
  { value: "cpp", label: "C++", icon: "ph ph-file-cpp text-purple-500" },
  { value: "c", label: "C", icon: "ph ph-file-c text-blue-600" },
]

const currentLang = computed(() => {
  return languages.value.find(l => l.value === props.modelValue) || defaultLanguages[0]
})

const currentLangLabel = computed(() => currentLang.value?.label || 'Go')
const currentLangIcon = computed(() => currentLang.value?.icon || 'ph ph-file-code text-cyan-500')

const toggleDropdown = (event) => {
  event?.stopPropagation()
  isOpen.value = !isOpen.value
}

const selectLanguage = (value) => {
  emit('update:modelValue', value)
  isOpen.value = false
}

const getIconUrl = (iconName) => {
  if (!iconName) return ''
  try {
    return new URL(`../assets/icons/${iconName}`, import.meta.url).href
  } catch (e) {
    return ''
  }
}

const fetchLanguages = async () => {
  try {
    const response = await axios.get('/api/languages')
    // API returns { languages: [...] }
    const langs = response.data?.languages || response.data
    
    if (langs && Array.isArray(langs) && langs.length > 0) {
      languages.value = langs.map(lang => ({
        value: lang.value,
        label: lang.label,
        icon: lang.icon // API returns filename like "go.png"
      }))
    } else {
      console.warn('API returned empty languages, using defaults')
      languages.value = defaultLanguages
    }
  } catch (error) {
    console.warn('Failed to fetch languages, using defaults:', error)
    languages.value = defaultLanguages
  }
}

// Click outside to close
const handleClickOutside = (event) => {
  if (selectorRef.value && !selectorRef.value.contains(event.target)) {
    isOpen.value = false
  }
}

onMounted(() => {
  fetchLanguages()
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
  transform: translateY(-8px);
}

/* Custom scrollbar for light theme */
.max-h-64::-webkit-scrollbar {
  width: 6px;
}
.max-h-64::-webkit-scrollbar-track {
  background: #f1f5f9;
}
.max-h-64::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 3px;
}
.max-h-64::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}
</style>
