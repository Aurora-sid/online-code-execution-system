<template>
  <div ref="editorContainer" class="absolute inset-0 w-full h-full">
    <!-- 编辑器加载骨架屏 -->
    <div v-if="!isEditorReady" class="absolute inset-0 bg-[#1e1e1e] flex flex-col">
      <!-- 模拟行号区域 -->
      <div class="flex h-full">
        <div class="w-12 bg-[#1e1e1e] border-r border-gray-700/50 pt-2">
          <div v-for="i in 15" :key="i" class="h-5 px-2 flex justify-end">
            <div class="w-4 h-3 bg-gray-600/30 rounded animate-pulse"></div>
          </div>
        </div>
        <!-- 模拟代码区域 -->
        <div class="flex-1 pt-2 pl-4 space-y-1">
          <div class="h-4 w-32 bg-gray-600/30 rounded animate-pulse"></div>
          <div class="h-4 w-48 bg-gray-600/30 rounded animate-pulse"></div>
          <div class="h-4 w-24 bg-gray-600/30 rounded animate-pulse"></div>
          <div class="h-4 w-64 bg-gray-600/30 rounded animate-pulse"></div>
          <div class="h-4 w-40 bg-gray-600/30 rounded animate-pulse"></div>
          <div class="h-4 w-56 bg-gray-600/30 rounded animate-pulse"></div>
          <div class="h-4 w-36 bg-gray-600/30 rounded animate-pulse"></div>
        </div>
      </div>
      <!-- 加载提示 -->
      <div class="absolute bottom-4 left-1/2 -translate-x-1/2 flex items-center gap-2 text-gray-400 text-sm">
        <div class="w-4 h-4 border-2 border-gray-500 border-t-cyan-400 rounded-full animate-spin"></div>
        <span>正在加载编辑器...</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, shallowRef, onMounted, onBeforeUnmount, watch } from 'vue'

const props = defineProps({
  language: { type: String, default: 'python' },
  modelValue: { type: String, default: '' },
  fontSize: { type: Number, default: 14 }
})

const emit = defineEmits(['update:modelValue'])
const editorContainer = ref(null)
const isEditorReady = ref(false)

// 使用 shallowRef 存储 monaco 模块和编辑器实例，避免响应式深度追踪
let monacoModule = null
let editorInstance = null
let resizeObserver = null

// 防抖函数
function debounce(func, wait) {
  let timeout
  return function (...args) {
    const context = this
    clearTimeout(timeout)
    timeout = setTimeout(() => func.apply(context, args), wait)
  }
}

// 懒加载 Monaco Editor
async function loadMonacoEditor() {
  if (monacoModule) return monacoModule
  
  // 并行加载 monaco 和 worker
  const [monaco, editorWorkerModule] = await Promise.all([
    import('monaco-editor'),
    import('monaco-editor/esm/vs/editor/editor.worker?worker')
  ])
  
  // 配置 Worker - 只使用基础 Editor Worker
  self.MonacoEnvironment = {
    getWorker() {
      return new editorWorkerModule.default()
    }
  }
  
  monacoModule = monaco
  return monaco
}

// 初始化编辑器
async function initEditor() {
  const monaco = await loadMonacoEditor()
  
  if (!editorContainer.value) return
  
  editorInstance = monaco.editor.create(editorContainer.value, {
    value: props.modelValue,
    language: props.language,
    theme: 'vs-dark',
    automaticLayout: false,
    minimap: { enabled: false },
    fontSize: props.fontSize,
    scrollBeyondLastLine: false,
    wordWrap: 'on',
    lineNumbers: 'on',
    lineNumbersMinChars: 3,
    glyphMargin: false,
    folding: true,
    lineDecorationsWidth: 10,
  })

  // 使用防抖处理内容更新
  const debouncedUpdate = debounce(() => {
    emit('update:modelValue', editorInstance.getValue())
  }, 300)

  editorInstance.onDidChangeModelContent(() => {
    debouncedUpdate()
  })

  // 监听容器大小变化
  resizeObserver = new ResizeObserver(() => {
    editorInstance?.layout()
  })
  resizeObserver.observe(editorContainer.value)

  // Ctrl + 滚轮缩放
  const handleWheel = (e) => {
    if (e.ctrlKey) {
      e.preventDefault()
      const delta = e.deltaY > 0 ? -1 : 1
      const currentSize = editorInstance.getOption(monaco.editor.EditorOption.fontSize)
      const newSize = Math.max(10, Math.min(30, currentSize + delta))
      editorInstance.updateOptions({ fontSize: newSize })
    }
  }
  
  editorContainer.value.addEventListener('wheel', handleWheel, { passive: false })
  
  // 缩放快捷键
  const handleKeyDown = (e) => {
    if (e.ctrlKey) {
      let shouldPrevent = false
      let newSize = null
      const currentSize = editorInstance.getOption(monaco.editor.EditorOption.fontSize)
      
      if (e.key === '=' || e.key === '+') {
        shouldPrevent = true
        newSize = Math.min(30, currentSize + 1)
      } else if (e.key === '-' || e.key === '_') {
        shouldPrevent = true
        newSize = Math.max(10, currentSize - 1)
      } else if (e.key === '0') {
        shouldPrevent = true
        newSize = props.fontSize || 14
      }
      
      if (shouldPrevent) {
        e.preventDefault()
        if (newSize !== null) {
          editorInstance.updateOptions({ fontSize: newSize })
        }
      }
    }
  }
  
  editorContainer.value.addEventListener('keydown', handleKeyDown)
  
  // 存储清理函数
  editorInstance._wheelCleanup = () => {
    editorContainer.value?.removeEventListener('wheel', handleWheel)
    editorContainer.value?.removeEventListener('keydown', handleKeyDown)
  }
  
  isEditorReady.value = true
}

onMounted(() => {
  initEditor()
})

watch(() => props.language, (newLang) => {
  if (editorInstance && monacoModule) {
    monacoModule.editor.setModelLanguage(editorInstance.getModel(), newLang)
  }
})

watch(() => props.fontSize, (newSize) => {
  if (editorInstance) {
    editorInstance.updateOptions({ fontSize: newSize })
  }
})

watch(() => props.modelValue, (newValue) => {
  if (editorInstance && newValue !== editorInstance.getValue()) {
    editorInstance.setValue(newValue)
  }
})

onBeforeUnmount(() => {
  if (resizeObserver) {
    resizeObserver.disconnect()
  }
  if (editorInstance) {
    editorInstance._wheelCleanup?.()
    editorInstance.dispose()
  }
})
</script>
