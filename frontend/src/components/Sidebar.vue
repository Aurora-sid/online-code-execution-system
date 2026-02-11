<template>
  <!-- 侧边栏容器 -->
  <aside 
    class="sidebar-responsive w-64 border-r border-gray-150/20 flex flex-col bg-white/40 z-10 transition-all duration-300 overflow-hidden"
    v-show="!collapsed"
  >
    <!-- 顶部栏 -->
    <div class="px-3 py-3 border-b border-gray-200/20">
      <div class="flex items-center justify-between mb-2">
        <div class="text-[10px] font-bold bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600 bg-clip-text text-transparent uppercase tracking-widest">资源管理器</div>
        <div class="flex items-center gap-1">
          <!-- <button 
            @click="openFolder"
            class="p-1 hover:bg-white/50 rounded transition-colors"
            title="打开文件夹"
          >
            <i class="ph ph-folder-plus text-sm bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600 bg-clip-text text-transparent"></i>
          </button> -->
          <img 
            :src="smallFolder" 
            alt="打开文件夹" 
            class="w-5 h-5 mr-4 drop-shadow-sm opacity-90 cursor-pointer hover:scale-105 transition-transform" 
            @click="openFolder"
            title="打开另一个文件夹"
          />
          <!-- <button 
            @click="refreshFolder"
            v-if="rootHandle"
            class="p-1 text-gray-400 hover:text-primary hover:bg-white/50 rounded transition-colors"
            title="刷新"
          >
            <i class="ph ph-arrow-clockwise text-sm"></i>
          </button> -->
          <img 
            :src="refreshIcon" 
            alt="刷新" 
            class="w-5 h-5 mr-0.5 drop-shadow-sm opacity-90 cursor-pointer hover:scale-105 transition-transform" 
            @click="refreshFolder"
            title="刷新"
          />
        </div>
      </div>
      
      <!-- 当前文件夹名称 -->
      <div 
        v-if="rootHandle"
        class="flex items-center gap-2 text-gray-700 bg-gray-100/50 px-2.5 py-1.5 rounded border border-gray-200/30 text-sm"
      >
        <i class="ph ph-folder-open text-primary"></i>
        <span class="font-medium truncate">{{ rootHandle.name }}</span>
      </div>
      
      <!-- 未打开文件夹时的提示 -->
      <div 
        v-else
        class="flex flex-col items-center justify-center py-6 text-center"
      >
        <img 
          :src="Folder" 
          alt="Empty Folder" 
          class="w-16 h-16 mb-2 drop-shadow-sm opacity-90 cursor-pointer hover:scale-105 transition-transform" 
          @click="openFolder"
          title="点击打开文件夹"
        />
        <p class="text-xs font-bold bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600 bg-clip-text text-transparent border border-indigo-100/50 rounded-lg hover:bg-indigo-50/50 transition-colors mb-3">尚未打开文件夹</p>
        <!-- 之前样式
        <button 
          @click="openFolder"
          class="px-3 py-1.5 text-xs bg-primary/10 text-primary rounded-lg hover:bg-primary/20 transition-colors"
        >
          打开文件夹
        </button>
        -->
        <!-- <button 
          @click="openFolder"
          class="px-3 py-1.5 text-xs font-bold bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600 bg-clip-text text-transparent border border-indigo-100/50 rounded-lg hover:bg-indigo-50/50 transition-colors"
        >
          打开文件夹
        </button> -->
      </div>
    </div>

    <!-- 文件树 -->
    <div class="flex-1 overflow-y-auto py-1 px-1" v-if="rootHandle">
      <FileTreeNode 
        v-for="node in fileTree" 
        :key="node.path"
        :node="node"
        :activeFile="activeFile"
        :depth="0"
        @select-file="handleFileSelect"
        @toggle-folder="toggleFolder"
      />
    </div>

    <!-- 不支持提示 -->
    <div v-if="!isSupported" class="p-3 bg-yellow-50 border-t border-yellow-200">
      <p class="text-xs text-yellow-700">
        <i class="ph ph-warning mr-1"></i>
        您的浏览器不支持文件系统访问。请使用 Chrome 或 Edge 浏览器。
      </p>
    </div>
  </aside>
</template>

<script setup>
import { ref, computed, defineAsyncComponent, shallowRef, markRaw } from 'vue'
import FileTreeNode from './FileTreeNode.vue'
import Folder from '@/assets/Vue-icons/file.webp'
import smallFolder from '@/assets/Vue-icons/folder.webp'
import refreshIcon from '@/assets/Vue-icons/refresh.webp'

const props = defineProps({
  activeFile: {
    type: String,
    default: ''
  },
  collapsed: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['select-file', 'open-file'])

// 检查浏览器是否支持 File System Access API
const isSupported = computed(() => 'showDirectoryPicker' in window)

// 根目录句柄
const rootHandle = shallowRef(null)

// 文件树数据
const fileTree = ref([])

// 打开文件夹
const openFolder = async () => {
  if (!isSupported.value) {
    alert('您的浏览器不支持文件系统访问功能，请使用 Chrome 或 Edge 浏览器。')
    return
  }

  try {
    const handle = await window.showDirectoryPicker({
      mode: 'readwrite'
    })
    rootHandle.value = handle
    await loadDirectoryContents(handle)
  } catch (err) {
    if (err.name !== 'AbortError') {
      console.error('打开文件夹失败:', err)
    }
  }
}

// 刷新文件夹
const refreshFolder = async () => {
  if (rootHandle.value) {
    await loadDirectoryContents(rootHandle.value)
  }
}

// 加载目录内容
const loadDirectoryContents = async (dirHandle, parentPath = '') => {
  const entries = []
  
  for await (const entry of dirHandle.values()) {
    const path = parentPath ? `${parentPath}/${entry.name}` : entry.name
    
    // 跳过隐藏文件和常见的忽略目录
    if (entry.name.startsWith('.') || 
        entry.name === 'node_modules' || 
        entry.name === '__pycache__' ||
        entry.name === 'vendor') {
      continue
    }

    const node = {
      name: entry.name,
      path: path,
      type: entry.kind,
      handle: markRaw(entry),
      expanded: false,
      children: []
    }

    if (entry.kind === 'directory') {
      // 目录：延迟加载子项
      node.loaded = false
    } else {
      // 文件：提取扩展名
      node.extension = entry.name.split('.').pop().toLowerCase()
    }

    entries.push(node)
  }

  // 排序：目录在前，文件在后，按名称字母排序
  entries.sort((a, b) => {
    if (a.type !== b.type) {
      return a.type === 'directory' ? -1 : 1
    }
    return a.name.localeCompare(b.name)
  })

  if (!parentPath) {
    fileTree.value = entries
  }
  
  return entries
}

// 切换文件夹展开/折叠
const toggleFolder = async (node) => {
  if (node.type !== 'directory') return

  node.expanded = !node.expanded

  // 如果展开且未加载过，则加载子目录
  if (node.expanded && !node.loaded) {
    try {
      node.children = await loadDirectoryContents(node.handle, node.path)
      node.loaded = true
    } catch (err) {
      console.error('加载目录失败:', err)
      node.expanded = false
    }
  }
}

// 处理文件选择
const handleFileSelect = async (node) => {
  if (node.type === 'directory') {
    await toggleFolder(node)
  } else {
    // 读取文件内容
    try {
      const file = await node.handle.getFile()
      const content = await file.text()
      
      emit('open-file', {
        name: node.name,
        path: node.path,
        handle: node.handle,
        content: content,
        extension: node.extension
      })
    } catch (err) {
      console.error('读取文件失败:', err)
    }
  }
}
</script>

<style scoped>
/* 响应式侧边栏：小屏幕隐藏 */
@media (max-width: 800px) {
  .sidebar-responsive {
    display: none !important;
    width: 0 !important;
  }
}
</style>
