<template>
  <div class="h-6 bg-white border-t border-gray-200 flex items-center justify-between px-3 text-[11px] text-gray-500 font-mono">
    <!-- Left section -->
    <div class="flex items-center gap-4">
      <div class="flex items-center gap-1.5">
        <i class="ph ph-git-branch text-xs"></i>
        <span>main</span>
      </div>
      <div class="flex items-center gap-1.5">
        <i class="ph ph-warning-circle text-xs text-gray-400"></i>
        <span>0</span>
      </div>
    </div>
    
    <!-- Right section -->
    <div class="flex items-center gap-4">
      <span>Ln {{ line }}, Col {{ col }}</span>
      <span>UTF-8</span>
      <span>{{ language }}</span>
      <div class="flex items-center gap-1.5">
        <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
        <span class="text-emerald-600">Ready</span>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  line: {
    type: Number,
    default: 1
  },
  col: {
    type: Number,
    default: 1
  },
  language: {
    type: String,
    default: 'Go'
  }
})
</script>
