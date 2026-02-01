<template>
  <div>
    <!-- 当前节点 -->
    <div 
      @click="handleClick"
      @dblclick="handleDoubleClick"
      class="group flex items-center gap-1.5 py-1 px-2 rounded cursor-pointer transition-all text-sm select-none"
      :class="[
        isActive ? 'bg-primary/20 text-primary font-medium' : 'text-gray-600 hover:text-gray-900 hover:bg-white/40',
      ]"
      :style="{ paddingLeft: `${depth * 12 + 8}px` }"
    >
      <!-- 展开/折叠图标 (仅文件夹) -->
      <i 
        v-if="node.type === 'directory'"
        :class="node.expanded ? 'ph ph-caret-down' : 'ph ph-caret-right'"
        class="text-xs text-gray-400 w-3"
      ></i>
      <span v-else class="w-3"></span>

      <!-- 文件/文件夹图标 -->
      <i :class="getIcon()" class="text-base"></i>

      <!-- 名称 -->
      <span class="truncate flex-1">{{ node.name }}</span>
    </div>

    <!-- 子节点 (仅展开的文件夹) -->
    <div v-if="node.type === 'directory' && node.expanded && node.children">
      <FileTreeNode
        v-for="child in node.children"
        :key="child.path"
        :node="child"
        :activeFile="activeFile"
        :depth="depth + 1"
        @select-file="$emit('select-file', $event)"
        @toggle-folder="$emit('toggle-folder', $event)"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  node: {
    type: Object,
    required: true
  },
  activeFile: {
    type: String,
    default: ''
  },
  depth: {
    type: Number,
    default: 0
  }
})

const emit = defineEmits(['select-file', 'toggle-folder'])

// 是否为当前激活的文件
const isActive = computed(() => props.activeFile === props.node.path)

// 单击处理
const handleClick = () => {
  if (props.node.type === 'directory') {
    emit('toggle-folder', props.node)
  }
}

// 双击处理
const handleDoubleClick = () => {
  if (props.node.type === 'file') {
    emit('select-file', props.node)
  }
}

// 获取图标
const getIcon = () => {
  if (props.node.type === 'directory') {
    return props.node.expanded 
      ? 'ph ph-folder-open text-yellow-500' 
      : 'ph ph-folder text-yellow-500'
  }

  // 根据扩展名返回不同图标
  const iconMap = {
    // 编程语言
    go: 'ph ph-file-code text-cyan-500',
    py: 'ph ph-file-code text-green-500',
    js: 'ph ph-file-js text-yellow-500',
    ts: 'ph ph-file-ts text-blue-500',
    jsx: 'ph ph-file-jsx text-cyan-400',
    tsx: 'ph ph-file-tsx text-blue-400',
    java: 'ph ph-file-code text-orange-500',
    cpp: 'ph ph-file-code text-purple-500',
    c: 'ph ph-file-code text-blue-600',
    h: 'ph ph-file-code text-purple-400',
    rs: 'ph ph-file-code text-orange-600',
    
    // Web
    html: 'ph ph-file-html text-orange-500',
    css: 'ph ph-file-css text-blue-500',
    scss: 'ph ph-file-css text-pink-500',
    vue: 'ph ph-file-vue text-green-500',
    
    // 数据/配置
    json: 'ph ph-brackets-curly text-yellow-600',
    yaml: 'ph ph-file-code text-red-400',
    yml: 'ph ph-file-code text-red-400',
    xml: 'ph ph-file-code text-orange-400',
    toml: 'ph ph-file-code text-gray-500',
    
    // 文档
    md: 'ph ph-file-text text-blue-400',
    txt: 'ph ph-file-text text-gray-400',
    pdf: 'ph ph-file-pdf text-red-500',
    
    // 其他
    svg: 'ph ph-image text-yellow-500',
    png: 'ph ph-image text-green-500',
    jpg: 'ph ph-image text-green-500',
    jpeg: 'ph ph-image text-green-500',
    gif: 'ph ph-image text-purple-500',
    
    // Go 特定
    mod: 'ph ph-gear text-cyan-600',
    sum: 'ph ph-list text-cyan-400',
  }

  return iconMap[props.node.extension] || 'ph ph-file text-gray-400'
}
</script>
