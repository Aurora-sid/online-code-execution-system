<template>
  <div ref="editorContainer" class="code-editor"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, onBeforeUnmount, defineProps, defineEmits } from 'vue'
import * as monaco from 'monaco-editor'

// 定义组件属性
const props = defineProps<{
  modelValue: string
  language: string
  theme?: string
}>()

// 定义组件事件
const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

// 编辑器容器和实例
const editorContainer = ref<HTMLElement | null>(null)
let editor: monaco.editor.IStandaloneCodeEditor | null = null

// 初始化编辑器
const initEditor = () => {
  if (!editorContainer.value) return

  // 创建编辑器实例
  editor = monaco.editor.create(editorContainer.value, {
    value: props.modelValue,
    language: props.language,
    theme: props.theme || 'vs-dark',
    minimap: {
      enabled: true
    },
    scrollBeyondLastLine: false,
    automaticLayout: true,
    tabSize: 2,
    wordWrap: 'on'
  })

  // 监听编辑器内容变化
  editor.onDidChangeModelContent(() => {
    const value = editor?.getValue() || ''
    emit('update:modelValue', value)
  })
}

// 监听语言变化
watch(() => props.language, (newLanguage) => {
  if (!editor) return
  
  const model = editor.getModel()
  if (model) {
    monaco.editor.setModelLanguage(model, newLanguage)
  }
})

// 监听内容变化（外部更新）
watch(() => props.modelValue, (newValue) => {
  if (!editor) return
  
  const currentValue = editor.getValue()
  if (newValue !== currentValue) {
    editor.setValue(newValue)
  }
})

// 组件挂载时初始化编辑器
onMounted(() => {
  initEditor()
})

// 组件卸载前销毁编辑器
onBeforeUnmount(() => {
  if (editor) {
    editor.dispose()
    editor = null
  }
})
</script>

<style scoped>
.code-editor {
  width: 100%;
  height: 100%;
  border-radius: 4px;
  overflow: hidden;
}
</style>