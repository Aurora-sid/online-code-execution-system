<template>
  <div class="h-screen bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 text-white flex flex-col overflow-hidden">
    <!-- Header -->
    <header class="bg-gray-900/80 backdrop-blur-sm border-b border-gray-700 shadow-lg">
      <div class="px-6 py-4">
        <div class="flex justify-between items-center">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"/>
              </svg>
            </div>
            <h1 class="text-2xl font-bold bg-gradient-to-r from-blue-400 to-purple-500 bg-clip-text text-transparent">
              在线代码执行平台
            </h1>
          </div>
          
          <div class="flex items-center gap-4">
            <select 
              v-model="currentLanguage" 
              class="bg-gray-800/80 border border-gray-600 rounded-lg px-4 py-2 text-white hover:border-blue-500 transition-colors focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="python">🐍 Python</option>
              <option value="java">☕ Java</option>
              <option value="cpp">⚡ C++</option>
              <option value="go">🔷 Go</option>
            </select>
            
            <button 
              @click="runCode" 
              :disabled="loading"
              class="bg-gradient-to-r from-green-500 to-emerald-600 hover:from-green-600 hover:to-emerald-700 px-6 py-2 rounded-lg font-semibold shadow-lg transition-all transform hover:scale-105 disabled:opacity-50 disabled:cursor-not-allowed disabled:transform-none"
            >
              <span v-if="loading" class="flex items-center gap-2">
                <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                运行中...
              </span>
              <span v-else class="flex items-center gap-2">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                </svg>
                运行代码
              </span>
            </button>
          </div>
        </div>
      </div>
    </header>

    <!-- Main Content -->
    <div class="flex-1 px-6 py-6 flex flex-col gap-4 overflow-hidden">
      <!-- Editor Section -->
      <div class="flex-1 flex flex-col bg-gray-800/50 backdrop-blur-sm rounded-xl shadow-2xl border border-gray-700 overflow-hidden">
        <div class="bg-gray-800/80 px-4 py-3 border-b border-gray-700 flex items-center gap-2">
          <div class="flex gap-1.5">
            <div class="w-3 h-3 rounded-full bg-red-500"></div>
            <div class="w-3 h-3 rounded-full bg-yellow-500"></div>
            <div class="w-3 h-3 rounded-full bg-green-500"></div>
          </div>
          <span class="ml-3 text-sm font-medium text-gray-300">代码编辑器</span>
        </div>
        <CodeEditor v-model="code" :language="currentLanguage" class="flex-1" />
      </div>

      <!-- Terminal Section -->
      <div class="h-64 flex flex-col bg-gray-900/80 backdrop-blur-sm rounded-xl shadow-2xl border border-gray-700 overflow-hidden">
        <div class="bg-gray-800/80 px-4 py-3 border-b border-gray-700 flex items-center gap-2">
          <svg class="w-4 h-4 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
          </svg>
          <span class="text-sm font-medium text-gray-300">终端输出</span>
        </div>
        <Terminal ref="terminalRef" class="flex-1 bg-black/90" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import axios from 'axios'
import CodeEditor from './components/CodeEditor.vue'
import Terminal from './components/Terminal.vue'

const currentLanguage = ref('python')
const code = ref('print("Hello World")')
const loading = ref(false)
const terminalRef = ref(null)

// Default snippets
const snippets = {
  python: 'print("Hello World from Python")',
  java: 'public class Main {\n    public static void main(String[] args) {\n        System.out.println("Hello World from Java");\n    }\n}',
  cpp: '#include <iostream>\n\nint main() {\n    std::cout << "Hello World from C++" << std::endl;\n    return 0;\n}',
  go: 'package main\n\nimport "fmt"\n\nfunc main() {\n    fmt.Println("Hello World from Go")\n}'
}

watch(currentLanguage, (newLang) => {
  code.value = snippets[newLang] || ''
})

const runCode = async () => {
  loading.value = true
  terminalRef.value.clear()
  terminalRef.value.write('\r\n> 编译并运行中...\r\n')

  try {
    const response = await axios.post('http://localhost:8080/api/run', {
      language: currentLanguage.value,
      code: code.value
    })
    
    const taskId = response.data.taskId
    terminalRef.value.write(`> 任务已入队: ${taskId}\r\n`)
    terminalRef.value.write(`> 等待执行...\r\n`)

    // Connect WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const ws = new WebSocket(`${protocol}//localhost:8080/api/ws?taskId=${taskId}`)

    ws.onopen = () => {
      terminalRef.value.write('> 已连接到输出流\r\n')
    }

    ws.onmessage = (event) => {
      terminalRef.value.write(event.data)
    }

    ws.onerror = (error) => {
      terminalRef.value.write(`\r\n> 连接错误: ${error}\r\n`)
    }

    ws.onclose = () => {
      terminalRef.value.write('\r\n> 连接已关闭\r\n')
    }
    
  } catch (error) {
    terminalRef.value.write(`\r\n错误: ${error.message}`)
  } finally {
    setTimeout(() => { loading.value = false }, 1000)
  }
}
</script>

<style>
/* Custom scrollbar */
::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

::-webkit-scrollbar-track {
  background: #1f2937;
}

::-webkit-scrollbar-thumb {
  background: #4b5563;
  border-radius: 4px;
}

::-webkit-scrollbar-thumb:hover {
  background: #6b7280;
}
</style>
