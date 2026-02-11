<template>
  <div 
    class="h-screen w-full bg-cover bg-center bg-no-repeat flex flex-col font-sans selection:bg-primary/20 overflow-hidden relative transition-all duration-300"
    :style="{ backgroundImage: `url(${backgroundImage})` }"
  >
    <!-- 顶部导航栏 -->
    <AppHeader 
      v-model:currentLanguage="currentLanguage"
      :loading="loading"
      :sidebarOpen="!sidebarCollapsed"
      @run="runCode"
      @save="saveToLocal"
      @toggle-history="showHistory = !showHistory"
      @toggle-sidebar="sidebarCollapsed = !sidebarCollapsed"
      @toggle-analysis="showAnalysis = !showAnalysis"
      class="flex-none z-10 relative"
    />

    <!-- 主工作区 -->
    <main class="flex-1 flex overflow-hidden min-h-0">
      <!-- 侧边栏切换按钮 -->
      <button
        @click="sidebarCollapsed = !sidebarCollapsed"
        class="absolute left-0 top-1/2 -translate-y-1/2 z-20 p-1.5 bg-white/80 hover:bg-white border border-gray-200 rounded-r-lg shadow-md transition-all hover:scale-105"
        :class="sidebarCollapsed ? 'md:hidden' : 'hidden'"
        title="展开侧边栏"
      >
        <i class="ph ph-caret-right text-gray-600"></i>
      </button>

      <!-- 侧边栏 -->
      <Sidebar 
        :activeFile="currentFilePath"
        :collapsed="sidebarCollapsed"
        @open-file="handleOpenFile"
      />

      <!-- 编辑器区域 -->
      <section class="flex-1 flex flex-col min-w-0">
        <!-- 标签页 -->
        <!-- <EditorTabs 
          :files="files"
          :activeFile="activeFileName"
          @select-file="handleFileSelect"
        /> -->

        <!-- 代码编辑器 -->
        <div ref="editorContainer" class="relative bg-[#0c0c0e]/70 transition-all duration-300" :style="{ height: terminalCollapsed ? (editorHeight + terminalHeight - 32) + 'px' : editorHeight + 'px' }">
          <CodeEditor v-model="code" :language="currentLanguage" :fontSize="editorFontSize" />
        </div>

        <!-- 可拖动分割线 -->
        <div 
          @mousedown="startDragging"
          class="h-0.1 bg-gray-300/20 hover:bg-primary cursor-row-resize transition-colors relative group"
          :class="{ 'bg-primary': isDragging }"
        >
          <div class="absolute inset-0 flex items-center justify-center">
            <div class="w-12 h-0.5 bg-gray-400 group-hover:bg-primary rounded-full transition-colors"></div>
          </div>
        </div>

        <!-- 终端面板 -->
        <div 
          ref="terminalContainer"
          class="flex flex-col bg-gray-900/70 transition-all duration-300"
          :style="{ height: terminalCollapsed ? '32px' : terminalHeight + 'px' }"
        >
          <Terminal 
            ref="terminalRef" 
            :fontSize="terminalFontSize"
            :collapsed="terminalCollapsed"
            :code="code"
            :language="currentLanguage"
            @toggle-collapse="terminalCollapsed = !terminalCollapsed"
            @send-input="handleTerminalInput"
          />
        </div>
      </section>
    </main>

    <!-- 状态栏 -->
    <StatusBar 
      :line="cursorLine"
      :column="cursorColumn"
      :language="languageDisplayName"
      :errorCount="0"
    />

    <!-- 历史记录面板 -->
    <HistoryPanel 
      v-model:show="showHistory" 
      @load-code="loadFromHistory"
    />

    <!-- AI 分析面板 -->
    <Transition name="slide-right">
      <div 
        v-if="showAnalysis" 
        class="fixed right-0 top-14 bottom-6 w-96 max-w-full z-40 shadow-2xl border-l border-gray-200"
      >
        <AnalysisPanel 
          :code="code" 
          :language="currentLanguage"
          @close="showAnalysis = false"
        />
      </div>
    </Transition>

    <!-- 保存模态框 -->
    <SaveModal 
      v-model:show="showSaveModal"
      :defaultName="`aurora_code_${Date.now()}`"
      :extension="fileExtensions[currentLanguage] || 'txt'"
      @confirm="handleSaveConfirm"
      @cancel="showSaveModal = false"
    />

    <!-- 保存成功提示 -->
    <Transition name="toast">
      <div v-if="showSaveToast" class="fixed bottom-8 right-6 bg-success/90 text-white px-5 py-2.5 rounded-lg shadow-2xl flex items-center gap-2 z-50 text-sm">
        <i class="ph ph-check-circle text-lg"></i>
        <span class="font-medium">代码已保存到本地</span>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, shallowRef, watch, computed, onMounted, onBeforeUnmount, defineAsyncComponent } from 'vue'
import AppHeader from '../components/AppHeader.vue'
import CodeEditor from '../components/CodeEditor.vue'
import Terminal from '../components/Terminal.vue'
import Sidebar from '../components/Sidebar.vue'

import StatusBar from '../components/StatusBar.vue'
import backgroundImage from '@/assets/background.webp'
import { runCode as apiRunCode, getWebSocketURL } from '../api'
import { useEditorStore } from '../stores/editor'

const HistoryPanel = defineAsyncComponent(() => import('../components/HistoryPanel.vue'))
const SaveModal = defineAsyncComponent(() => import('../components/SaveModal.vue'))
const AnalysisPanel = defineAsyncComponent(() => import('../components/AnalysisPanel.vue'))



const editorStore = useEditorStore()
const currentLanguage = ref(editorStore.currentLanguage)
const code = ref(editorStore.codeMap[editorStore.currentLanguage] || editorStore.snippets[editorStore.currentLanguage])
const loading = ref(false)
const terminalRef = shallowRef(null)
const showHistory = ref(false)
const showAnalysis = ref(false)
const showSaveModal = ref(false)
const showSaveToast = ref(false)
const lastOutput = ref('')
const sidebarCollapsed = ref(true) // 默认隐藏侧边栏，点击"打开文件夹"后显示
const isMobile = ref(window.innerWidth < 768) // 小屏幕检测
const terminalCollapsed = ref(false)
const terminalHeight = ref(200)
const editorHeight = ref(400)

// 拖动状态
const isDragging = ref(false)
const editorContainer = shallowRef(null)
const terminalContainer = shallowRef(null)

const editorFontSize = ref(14)
const terminalFontSize = ref(13)
const cursorLine = ref(14)
const cursorColumn = ref(32)

// 当前打开的文件信息
const currentFilePath = ref('')
const currentFileHandle = shallowRef(null)
const isFileModified = ref(false)

// WebSocket 连接（用于交互式输入）
let currentWebSocket = null

// 侧边栏和标签页的文件列表
const files = computed(() => {
  const ext = fileExtensions[currentLanguage.value] || 'txt'
  return [
    { name: currentFile.value, type: ext, unsaved: true },
    { name: 'utils.' + ext, type: ext, unsaved: false },
    { name: ext === 'go' ? 'go.mod' : 'config.json', type: 'mod', unsaved: false },
    { name: 'README.md', type: 'md', unsaved: false },
  ]
})

const activeFileName = computed(() => currentFile.value)

const currentFile = computed(() => {
  const map = {
    cpp: 'main.cpp',
    c: 'main.c',
    java: 'Main.java',
    python: 'main.py',

    go: 'main.go',
    javascript: 'main.js',
    rust: 'main.rs',
    csharp: 'Program.cs',
    typescript: 'main.ts',
  }
  return map[currentLanguage.value] || 'script'
})

const languageDisplayName = computed(() => {
  const map = {
    cpp: 'C++',
    c: 'C',
    java: 'Java',
    python: 'Python',

    go: 'Go',
    javascript: 'JavaScript',
    rust: 'Rust',
    csharp: 'C#',
    typescript: 'TypeScript',
  }
  return map[currentLanguage.value] || 'Text'
})

const fileExtensions = {
  cpp: 'cpp',
  c: 'c',
  java: 'java',
  python: 'py',

  go: 'go',
  javascript: 'js',
  rust: 'rs',
  csharp: 'cs',
  typescript: 'ts'
}

const mimeTypes = {
  cpp: 'text/x-c++src',
  c: 'text/x-csrc',
  java: 'text/x-java-source',
  python: 'text/x-python',

  go: 'text/x-go',
  javascript: 'text/javascript',
  rust: 'text/x-rustsrc',
  csharp: 'text/x-csharp',
  typescript: 'text/typescript'
}

// 默认代码片段
const snippets = {
  cpp: '#include <iostream>\n\nint main() {\n    std::cout << "Hello Aurora Code from C++17!" << std::endl;\n    return 0;\n}',
  c: '#include <stdio.h>\n\nint main() {\n    printf("Hello Aurora Code from C (gcc8.3.0)!\\n");\n    return 0;\n}',
  java: 'public class Main {\n    public static void main(String[] args) {\n        System.out.println("Hello Aurora Code from Java 11!");\n    }\n}',
  python: 'print("Hello Aurora Code from Python 3.7.3!")',

  go: 'package main\n\nimport "fmt"\n\nfunc main() {\n    fmt.Println("Hello Aurora Code from Go 1.19.5")\n}',
  javascript: 'console.log("Hello Aurora Code from JavaScript (Node.js)!");',
  rust: 'fn main() {\n    println!("Hello Aurora Code from Rust!");\n}',
  csharp: 'using System;\n\nclass Program {\n    static void Main() {\n        Console.WriteLine("Hello Aurora Code from C# (.NET 8)!");\n    }\n}',
  typescript: 'const msg: string = "Hello Aurora Code from TypeScript!";\nconsole.log(msg);'
}

watch(currentLanguage, (newLang) => {
  editorStore.currentLanguage = newLang
  code.value = editorStore.codeMap[newLang] || editorStore.snippets[newLang] || ''
})

// 处理打开文件夹按钮点击
const handleOpenFolder = async () => {
  // 显示侧边栏
  sidebarCollapsed.value = false
  // 这里可以扩展为调用 File System Access API 打开文件夹
}

// 处理从侧边栏打开文件
const handleOpenFile = (fileInfo) => {
  // 设置文件内容
  code.value = fileInfo.content
  currentFilePath.value = fileInfo.path
  currentFileHandle.value = fileInfo.handle
  isFileModified.value = false
  
  // 根据扩展名设置语言
  const extToLang = {
    go: 'go',
    py: 'python',
    js: 'javascript',
    ts: 'typescript',
    java: 'java',
    cpp: 'cpp',
    c: 'c',
    rs: 'rust',
    rb: 'ruby',
    php: 'php',
    html: 'html',
    css: 'css',
    json: 'json',
    md: 'markdown',
    vue: 'vue',
  }
  
  const lang = extToLang[fileInfo.extension] || 'plaintext'
  currentLanguage.value = lang
  
  terminalRef.value?.clear()
  terminalRef.value?.write(`\x1b[32m> 已打开文件: ${fileInfo.name}\x1b[0m\r\n`)
}

// 监听代码变化，标记修改状态
watch(code, (newCode) => {
  editorStore.updateCode(currentLanguage.value, newCode)
  if (currentFileHandle.value) {
    isFileModified.value = true
  }
})

const handleFileSelect = (file) => {
  // 目前仅记录日志 - 仅主文件可编辑
  console.log('Selected file:', file.name)
}

const runCode = async () => {
  loading.value = true
  lastOutput.value = ''
  terminalCollapsed.value = false
  terminalRef.value.clear()
  terminalRef.value.disableInput() // 先禁用输入
  editorStore.clearOutput() // 清除持久化的输出
  terminalRef.value.write('\r\n\x1b[34m> Compiling and running...\x1b[0m\r\n')
  editorStore.appendOutput('\r\n\x1b[34m> Compiling and running...\x1b[0m\r\n')

  // 关闭之前的 WebSocket 连接
  if (currentWebSocket) {
    currentWebSocket.close()
    currentWebSocket = null
  }

  try {
    const data = await apiRunCode(currentLanguage.value, code.value)
    
    const taskId = data.taskId
    const queuedMsg = `\x1b[32m> Task Queued: ${taskId}\x1b[0m\r\n> Waiting for execution...\r\n`
    terminalRef.value.write(queuedMsg)
    editorStore.appendOutput(queuedMsg)

    // 连接 WebSocket（双向通信）
    const ws = new WebSocket(getWebSocketURL(taskId))
    currentWebSocket = ws

    ws.onopen = () => {
      terminalRef.value.write('> Connected to output stream\r\n')
      // 不再自动启用输入，用户可以通过点击"输入"按钮手动启用
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'stdout' && msg.data) {
          // 检查是否是"等待输入"信号
          if (msg.data.includes('__WAITING_INPUT__')) {
            // 启用输入框
            terminalRef.value.enableInput()
            terminalRef.value.write('\r\n\x1b[33m> 程序正在等待输入...\x1b[0m\r\n')
            return
          }

          lastOutput.value += msg.data
          // 直接写入原始输出（不用 writeln，因为输出可能是部分行）
          const outText = msg.data.replace(/\n/g, '\r\n')
          terminalRef.value.write(outText)
          editorStore.appendOutput(outText)
          
          // 检查是否执行完成
          if (msg.data.includes('[执行完成]')) {
            terminalRef.value.disableInput()
            loading.value = false
          }
        }
      } catch (e) {
        // 如果不是 JSON，直接写入
        lastOutput.value += event.data
        const outText = event.data.replace(/\n/g, '\r\n')
        terminalRef.value.write(outText)
        editorStore.appendOutput(outText)
      }
    }

    ws.onerror = (error) => {
      terminalRef.value.write(`\r\n\x1b[31m> Connection Error\x1b[0m\r\n`)
      terminalRef.value.disableInput()
    }

    ws.onclose = () => {
      terminalRef.value.write('\r\n\x1b[90m> Connection closed\x1b[0m\r\n')
      terminalRef.value.disableInput()
      currentWebSocket = null
      loading.value = false
      
      // 保存运行历史到 localStorage
      saveRunHistory({
        filePath: currentFilePath.value || '未命名',
        code: code.value,
        language: currentLanguage.value,
        output: lastOutput.value,
        timestamp: Date.now()
      })
    }
    
  } catch (error) {
    terminalRef.value.write(`\r\n\x1b[31mError: ${error.message}\x1b[0m`)
    terminalRef.value.disableInput()
    loading.value = false
  }
}

// 处理终端输入
const handleTerminalInput = (input) => {
  if (currentWebSocket && currentWebSocket.readyState === WebSocket.OPEN) {
    // 发送 stdin 消息到后端
    currentWebSocket.send(JSON.stringify({
      type: 'stdin',
      data: input
    }))
  }
}

const saveToLocal = async () => {
  // 始终弹出文件保存对话框，让用户选择路径和文件名
  const ext = fileExtensions[currentLanguage.value] || 'txt'
  // 使用当前文件名（如果有）或生成默认名称
  const suggestedName = currentFilePath.value 
    ? currentFilePath.value.split(/[/\\]/).pop() 
    : `aurora_code_${Date.now()}.${ext}`

  if ('showSaveFilePicker' in window) {
    try {
      const options = {
        suggestedName: suggestedName,
        types: [{
          description: `${currentLanguage.value.toUpperCase()} Source File`,
          accept: {
            'text/plain': [`.${ext}`]
          }
        }],
      }
      
      const handle = await window.showSaveFilePicker(options)
      const writable = await handle.createWritable()
      await writable.write(code.value)
      await writable.close()
      
      // 更新当前文件句柄和路径
      currentFileHandle.value = handle
      currentFilePath.value = handle.name
      isFileModified.value = false
      
      showSaveToast.value = true
      setTimeout(() => showSaveToast.value = false, 3000)
      terminalRef.value?.write(`\x1b[32m> 文件已保存: ${handle.name}\x1b[0m\r\n`)
      return
    } catch (err) {
      if (err.name === 'AbortError') return // 用户取消
      console.warn('File Picker API failed, falling back to legacy save:', err)
      // 回退到旧版保存模态框
    }
  }

  // 浏览器不支持 File Picker API，使用旧版模态框
  showSaveModal.value = true
}

const handleSaveConfirm = (filename) => {
  const ext = fileExtensions[currentLanguage.value] || 'txt'
  const fullFilename = filename.endsWith(`.${ext}`) ? filename : `${filename}.${ext}`
  
  const blob = new Blob([code.value], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  
  const a = document.createElement('a')
  a.href = url
  a.download = fullFilename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
  
  showSaveModal.value = false
  showSaveToast.value = true
  setTimeout(() => showSaveToast.value = false, 3000)
}

const loadFromHistory = (record) => {
  currentLanguage.value = record.language
  code.value = record.code
  showHistory.value = false
}

// 保存运行历史到 localStorage
const saveRunHistory = (record) => {
  try {
    const historyKey = 'aurora_run_history'
    const existingHistory = JSON.parse(localStorage.getItem(historyKey) || '[]')
    
    // 添加唯一 ID
    record.id = `run_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    
    // 添加到历史记录开头
    existingHistory.unshift(record)
    
    // 限制最多保存 50 条记录
    if (existingHistory.length > 50) {
      existingHistory.pop()
    }
    
    localStorage.setItem(historyKey, JSON.stringify(existingHistory))
  } catch (err) {
    console.error('保存运行历史失败:', err)
  }
}

// 可拖动分割线逻辑
const startDragging = (e) => {
  isDragging.value = true
  document.addEventListener('mousemove', handleDrag)
  document.addEventListener('mouseup', stopDragging)
  e.preventDefault()
}

const handleDrag = (e) => {
  if (!isDragging.value) return
  
  const container = editorContainer.value?.parentElement
  if (!container) return
  
  const containerRect = container.getBoundingClientRect()
  const newEditorHeight = e.clientY - containerRect.top
  const totalHeight = containerRect.height
  
  // 最小高度
  const minEditorHeight = 150
  const minTerminalHeight = 100
  
  if (newEditorHeight >= minEditorHeight && (totalHeight - newEditorHeight - 4) >= minTerminalHeight) {
    editorHeight.value = newEditorHeight
    terminalHeight.value = totalHeight - newEditorHeight - 4 // 4px for divider
    
    // 根据新高度更新字体大小
    updateFontSizes()
  }
}

const stopDragging = () => {
  isDragging.value = false
  document.removeEventListener('mousemove', handleDrag)
  document.removeEventListener('mouseup', stopDragging)
}

// 根据容器高度计算字体大小
const calculateFontSize = (height, min = 10, max = 20) => {
  // 公式：fontSize = height / 30，限制在 min 和 max 之间
  const calculated = Math.floor(height / 30)
  return Math.max(min, Math.min(max, calculated))
}

const updateFontSizes = () => {
  editorFontSize.value = calculateFontSize(editorHeight.value, 12, 20)
  terminalFontSize.value = calculateFontSize(terminalHeight.value, 11, 16)
}

// 窗口大小调整处理函数（需要存储引用以便清理）
const handleWindowResize = () => {
  // 更新移动端检测
  isMobile.value = window.innerWidth < 768
  
  const container = editorContainer.value?.parentElement
  if (container) {
    const totalHeight = container.clientHeight
    const ratio = editorHeight.value / (editorHeight.value + terminalHeight.value + 4)
    editorHeight.value = Math.floor(totalHeight * ratio)
    terminalHeight.value = totalHeight - editorHeight.value - 4
    updateFontSizes()
  }
}

// 挂载时初始化高度
onMounted(() => {
  // 根据可用空间计算初始高度
  const container = editorContainer.value?.parentElement
  if (container) {
    const totalHeight = container.clientHeight
    editorHeight.value = Math.floor(totalHeight * 0.65) // 65% for editor
    terminalHeight.value = Math.floor(totalHeight * 0.35) - 4 // 35% for terminal, minus divider
    updateFontSizes()
  }
  
  // 窗口调整大小时更新
  window.addEventListener('resize', handleWindowResize)
  
  // 恢复之前运行的输出
  if (editorStore.terminalOutput) {
      setTimeout(() => {
          if (terminalRef.value) {
              terminalRef.value.write(editorStore.terminalOutput)
              // Auto-expand terminal if there is content
              if (terminalCollapsed.value) {
                  terminalCollapsed.value = false
              }
          }
      }, 100)
  }
})

// 组件销毁时清理资源
onBeforeUnmount(() => {
  // 移除 resize 监听器
  window.removeEventListener('resize', handleWindowResize)
  
  // 关闭 WebSocket 连接
  if (currentWebSocket) {
    currentWebSocket.close(1000, 'Component unmounted')
    currentWebSocket = null
  }
})
</script>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(20px);
}

/* AI 分析面板滑入动画 */
.slide-right-enter-active,
.slide-right-leave-active {
  transition: all 0.3s ease;
}
.slide-right-enter-from,
.slide-right-leave-to {
  opacity: 0;
  transform: translateX(100%);
}
</style>
