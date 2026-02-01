<template>
  <div class="h-9 border-b border-gray-200 flex items-center bg-white">
    <!-- Tabs -->
    <div 
      v-for="tab in tabs" 
      :key="tab.name"
      class="group h-full flex items-center gap-2 px-3 cursor-pointer border-r border-gray-200 transition-all relative"
      :class="activeTab === tab.name ? 'bg-gray-50' : 'hover:bg-gray-100'"
    >
      <!-- Active indicator -->
      <div 
        v-if="activeTab === tab.name"
        class="absolute top-0 left-0 right-0 h-0.5 bg-gradient-to-r from-primary to-secondary"
      ></div>
      
      <i :class="getFileIcon(tab.type)" class="text-sm"></i>
      <span 
        class="text-xs"
        :class="activeTab === tab.name ? 'text-gray-900 font-medium' : 'text-gray-500'"
      >{{ tab.name }}</span>
      <span v-if="tab.unsaved" class="text-secondary text-lg leading-none">•</span>
      <button 
        class="ml-1 w-4 h-4 rounded hover:bg-gray-200 flex items-center justify-center opacity-60 hover:opacity-100 transition-opacity"
      >
        <i class="ph ph-x text-[10px] text-gray-500"></i>
      </button>
    </div>

    <!-- Spacer -->
    <div class="flex-1"></div>
  </div>
</template>

<script setup>
defineProps({
  tabs: {
    type: Array,
    default: () => []
  },
  activeTab: {
    type: String,
    default: ''
  }
})

const getFileIcon = (type) => {
  const icons = {
    go: 'ph ph-file-code text-cyan-500',
    config: 'ph ph-gear text-red-400',
    md: 'ph ph-info text-blue-400',
    js: 'ph ph-file-js text-yellow-500',
    py: 'ph ph-file-code text-green-500',
    java: 'ph ph-file-code text-orange-500',
    cpp: 'ph ph-file-code text-purple-500',
    c: 'ph ph-file-code text-blue-600',
  }
  return icons[type] || 'ph ph-file-text text-gray-400'
}
</script>
