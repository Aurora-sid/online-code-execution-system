<template>
  <div class="h-screen w-full bg-zinc-900 text-zinc-200 flex flex-col font-sans selection:bg-cyan-500/30 overflow-hidden relative">
    <!-- Aurora Background Effects -->
    <div class="absolute top-0 left-0 w-full h-full overflow-hidden pointer-events-none z-0">
      <!-- 静态背景图片（性能优化） -->
      <img src="../assets/login.png" alt="Background" class="w-full h-full object-cover opacity-20" />
      
      <!-- 网格覆盖层 -->
      <div class="absolute inset-0 bg-[linear-gradient(rgba(255,255,255,0.02)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,0.02)_1px,transparent_1px)] bg-[size:50px_50px] opacity-20"></div>
    </div>

    <!-- 头部：固定高度 -->
    <AppHeader 
      v-model:currentLanguage="currentLanguage"
      :loading="loading"
      @run="runCode"
      @save="saveToLocal"
      @toggle-history="showHistory = !showHistory"
      class="flex-none z-10 relative"
    />

    <!-- 主要内容：填充剩余空间 -->
    <main class="flex-1 flex flex-col min-h-0 p-6 z-10 relative">
      <!-- 网格容器：填充主要内容 -->
      <div class="flex-1 w-full mx-auto grid grid-rows-[minmax(0,1fr)_300px] gap-6">
        
        <!-- 编辑器部分 -->
        <div class="bg-zinc-900/60 backdrop-blur-md rounded-2xl border border-white/10 shadow-2xl overflow-hidden flex flex-col group hover:border-white/20 transition-all duration-300 min-h-0 relative ring-1 ring-white/5">
          <!-- 编辑器标题栏 -->
          <div class="bg-black/40 px-4 py-3 border-b border-white/5 flex items-center justify-between flex-none">
            <div class="flex items-center gap-3">
              <div class="flex gap-2">
                <div class="w-3 h-3 rounded-full bg-red-500/90 shadow-sm"></div>
                <div class="w-3 h-3 rounded-full bg-yellow-500/90 shadow-sm"></div>
                <div class="w-3 h-3 rounded-full bg-green-500/90 shadow-sm"></div>
              </div>
              <div class="h-4 w-[1px] bg-white/10 mx-1"></div>
              <span class="text-sm font-medium text-zinc-400 font-mono flex items-center gap-2">
                <svg class="w-4 h-4 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                   <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
                </svg>
                {{ currentFile }}
              </span>
            </div>
            <div class="flex items-center gap-2">
               <span class="text-xs text-zinc-500 font-mono px-2 py-1 rounded bg-white/5">online-code-system</span>
            </div>
          </div>
          <!-- 编辑器组件 -->
          <div class="flex-1 min-h-0 relative">
             <CodeEditor v-model="code" :language="currentLanguage" />
          </div>
        </div>

        <!-- 终端部分 -->
        <div class="bg-zinc-900/60 backdrop-blur-md rounded-2xl border border-white/10 shadow-2xl overflow-hidden flex flex-col group hover:border-white/20 transition-all duration-300 min-h-0 relative ring-1 ring-white/5">
          <div class="bg-black/40 px-4 py-2.5 border-b border-white/5 flex items-center justify-between flex-none">
            <div class="flex items-center gap-2">
              <svg class="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
              </svg>
              <span class="text-xs font-bold text-zinc-300 tracking-wider">终端</span>
              <!-- <span class="text-xs font-bold text-zinc-300 tracking-wider">终端 | TERMINAL</span> -->
            </div>
            <div class="flex gap-1.5">
               <div class="w-1.5 h-1.5 rounded-full bg-emerald-500/50 animate-pulse"></div>
            </div>
          </div>
          <Terminal ref="terminalRef" class="flex-1 min-h-0" />
        </div>

      </div>
    </main>

    <!-- History Panel -->
    <HistoryPanel 
      v-model:show="showHistory" 
      @load-code="loadFromHistory"
    />

    <!-- Save Modal -->
    <SaveModal 
      v-model:show="showSaveModal"
      :defaultName="`aurora_code_${Date.now()}`"
      :extension="fileExtensions[currentLanguage] || 'txt'"
      @confirm="handleSaveConfirm"
      @cancel="showSaveModal = false"
    />

    <!-- Save Success Toast -->
    <Transition name="toast">
      <div v-if="showSaveToast" class="fixed bottom-6 right-6 bg-green-500/90 backdrop-blur text-white px-6 py-3 rounded-lg shadow-2xl flex items-center gap-3 z-50">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
        </svg>
        <span class="font-medium">代码已保存到本地</span>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'
import axios from 'axios'
import AppHeader from '../components/AppHeader.vue'
import CodeEditor from '../components/CodeEditor.vue'
import Terminal from '../components/Terminal.vue'
import HistoryPanel from '../components/HistoryPanel.vue'
import SaveModal from '../components/SaveModal.vue'
import { useAuth } from '../stores/auth'

const { getToken } = useAuth()

const currentLanguage = ref('python')
const code = ref('print("Hello World")')
const loading = ref(false)
const terminalRef = ref(null)
const showHistory = ref(false)
const showSaveModal = ref(false)
const showSaveToast = ref(false)
const lastOutput = ref('')

const currentFile = computed(() => {
  const map = {
    cpp: 'main.cpp',
    c: 'main.c',
    java: 'Main.java',
    python: 'main.py',
    pypy: 'main.py',
    go: 'main.go',
    javascript: 'main.js',
  }
  return map[currentLanguage.value] || 'script'
})

const fileExtensions = {
  cpp: 'cpp',
  c: 'c',
  java: 'java',
  python: 'py',
  pypy: 'py',
  go: 'go',
  javascript: 'js'
}

const mimeTypes = {
  cpp: 'text/x-c++src',
  c: 'text/x-csrc',
  java: 'text/x-java-source',
  python: 'text/x-python',
  pypy: 'text/x-python',
  go: 'text/x-go',
  javascript: 'text/javascript'
}

// 默认代码片段
const snippets = {
  cpp: '#include <iostream>\n\nint main() {\n    std::cout << "Hello Aurora Code from C++17!" << std::endl;\n    return 0;\n}',
  c: '#include <stdio.h>\n\nint main() {\n    printf("Hello Aurora Code from C (gcc8.3.0)!\\n");\n    return 0;\n}',
  java: 'public class Main {\n    public static void main(String[] args) {\n        System.out.println("Hello Aurora Code from Java 11!");\n    }\n}',
  python: 'print("Hello Aurora Code from Python 3.7.3!")',
  pypy: 'print("Hello Aurora Code from PyPy3 (7.3.8)!")',
  go: 'package main\n\nimport "fmt"\n\nfunc main() {\n    fmt.Println("Hello Aurora Code from Go 1.19.5")\n}',
  javascript: 'console.log("Hello Aurora Code from JavaScript (Node.js)!");'
}

watch(currentLanguage, (newLang) => {
  code.value = snippets[newLang] || ''
})

const runCode = async () => {
  loading.value = true
  lastOutput.value = ''
  terminalRef.value.clear()
  terminalRef.value.write('\r\n\x1b[34m> Compiling and running...\x1b[0m\r\n')

  try {
    const response = await axios.post('http://localhost:8080/api/run', {
      language: currentLanguage.value,
      code: code.value
    }, {
      headers: {
        'Authorization': `Bearer ${getToken()}`
      }
    })
    
    const taskId = response.data.taskId
    terminalRef.value.write(`\x1b[32m> Task Queued: ${taskId}\x1b[0m\r\n`)
    terminalRef.value.write(`> Waiting for execution...\r\n`)

    // Connect WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${protocol}//localhost:8080/api/ws?taskId=${taskId}`)

    ws.onopen = () => {
      terminalRef.value.write('> Connected to output stream\r\n')
    }

    ws.onmessage = (event) => {
      lastOutput.value += event.data
      terminalRef.value.write(event.data)
    }

    ws.onerror = (error) => {
      terminalRef.value.write(`\r\n\x1b[31m> Connection Error: ${error}\x1b[0m\r\n`)
    }

    ws.onclose = () => {
      terminalRef.value.write('\r\n\x1b[90m> Connection closed\x1b[0m\r\n')
    }
    
  } catch (error) {
    terminalRef.value.write(`\r\n\x1b[31mError: ${error.message}\x1b[0m`)
  } finally {
    setTimeout(() => { loading.value = false }, 1000)
  }
}

const saveToLocal = async () => {
  const ext = fileExtensions[currentLanguage.value] || 'txt'
  const defaultName = `aurora_code_${Date.now()}`

  // 1. Try modern File System Access API
  if ('showSaveFilePicker' in window) {
    try {
      const handle = await window.showSaveFilePicker({
        suggestedName: `${defaultName}.${ext}`,
        types: [{
          description: `${currentLanguage.value.toUpperCase()} Code`,
          accept: { [mimeTypes[currentLanguage.value] || 'text/plain']: [`.${ext}`] },
        }],
      })
      const writable = await handle.createWritable()
      await writable.write(code.value)
      await writable.close()
      
      showSaveToast.value = true
      setTimeout(() => showSaveToast.value = false, 3000)
      return
    } catch (err) {
      // User cancelled or error, don't show modal if user cancelled
      if (err.name === 'AbortError') return
      console.error('File Picker Error:', err)
    }
  }

  // 2. Fallback to custom modal + traditional download
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
</style>
