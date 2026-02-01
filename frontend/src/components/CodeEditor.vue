<template>
  <div ref="editorContainer" class="absolute inset-0 w-full h-full"></div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import * as monaco from 'monaco-editor'

// 工作线程配置
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker'
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker'
import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker'
import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker'

self.MonacoEnvironment = {
  getWorker(_, label) {
    if (label === 'json') return new jsonWorker()
    if (label === 'css' || label === 'scss' || label === 'less') return new cssWorker()
    if (label === 'html' || label === 'handlebars' || label === 'razor') return new htmlWorker()
    if (label === 'typescript' || label === 'javascript') return new tsWorker()
    return new editorWorker()
  }
}

const props = defineProps({
  language: { type: String, default: 'python' },
  modelValue: { type: String, default: '' },
  fontSize: { type: Number, default: 14 }
})

const emit = defineEmits(['update:modelValue'])
const editorContainer = ref(null)
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

onMounted(() => {
  editorInstance = monaco.editor.create(editorContainer.value, {
    value: props.modelValue,
    language: props.language,
    theme: 'vs-dark',
    automaticLayout: false, // 禁用自动布局，改用 ResizeObserver
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

  // 使用防抖处理内容更新，减少 v-model 同步频率
  const debouncedUpdate = debounce(() => {
    emit('update:modelValue', editorInstance.getValue())
  }, 300) // 300ms 延迟

  editorInstance.onDidChangeModelContent(() => {
    debouncedUpdate()
  })

  // 手动监听容器大小变化
  resizeObserver = new ResizeObserver(() => {
    editorInstance.layout()
  })
  resizeObserver.observe(editorContainer.value)

  // 添加 Ctrl + 滚轮缩放功能
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
  
  // 添加缩放快捷键（Ctrl+/-, Ctrl+=, Ctrl+0）
  const handleKeyDown = (e) => {
    if (e.ctrlKey) {
      let shouldPrevent = false
      let newSize = null
      const currentSize = editorInstance.getOption(monaco.editor.EditorOption.fontSize)
      
      if (e.key === '=' || e.key === '+') {
        // 放大
        shouldPrevent = true
        newSize = Math.min(30, currentSize + 1)
      } else if (e.key === '-' || e.key === '_') {
        // 缩小
        shouldPrevent = true
        newSize = Math.max(10, currentSize - 1)
      } else if (e.key === '0') {
        // 重置为默认
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
})

watch(() => props.language, (newLang) => {
  if (editorInstance) {
    monaco.editor.setModelLanguage(editorInstance.getModel(), newLang)
  }
})

watch(() => props.fontSize, (newSize) => {
  if (editorInstance) {
    editorInstance.updateOptions({ fontSize: newSize })
  }
})

watch(() => props.modelValue, (newValue) => {
  // 只有当传入的值和编辑器当前值不同步时才设置，避免光标跳动
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
