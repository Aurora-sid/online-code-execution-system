<template>
  <div class="app-container">
    <header class="app-header">
      <h1>在线编程系统</h1>
      <div class="language-selector">
        <el-select v-model="selectedLanguage" placeholder="选择编程语言" size="large">
          <el-option label="Python" value="python" />
          <el-option label="Go" value="go" />
          <el-option label="Java" value="java" />
        </el-select>
      </div>
    </header>
    
    <main class="app-main">
      <div class="editor-section">
        <CodeEditor 
          v-model="code" 
          :language="selectedLanguage"
        />
      </div>
      
      <div class="terminal-section">
        <div class="control-panel">
          <el-button type="primary" @click="runCode" :loading="isRunning">运行</el-button>
          <el-button @click="stopCode" :disabled="!isRunning">停止</el-button>
          <el-button @click="resetCode">重置</el-button>
        </div>
        <WebTerminal 
          ref="terminalRef"
          :is-connecting="isConnecting"
        />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import CodeEditor from './components/CodeEditor.vue'
import WebTerminal from './components/WebTerminal.vue'
import { useCodeStore } from './stores/code'
import { useTerminalStore } from './stores/terminal'

const codeStore = useCodeStore()
const terminalStore = useTerminalStore()

const selectedLanguage = ref<string>('python')
const code = ref<string>('print("Hello, World!")')
const isRunning = ref<boolean>(false)
const isConnecting = ref<boolean>(false)
const terminalRef = ref<any>(null)

// 运行代码
const runCode = async () => {
  if (!code.value.trim()) {
    terminalStore.writeLine('请输入代码后再运行')
    return
  }
  
  isRunning.value = true
  isConnecting.value = true
  
  try {
    await terminalStore.connect()
    await terminalStore.runCode(code.value, selectedLanguage.value)
  } catch (error) {
    terminalStore.writeLine(`运行出错: ${error}`)
    isRunning.value = false
    isConnecting.value = false
  }
}

// 停止代码
const stopCode = () => {
  terminalStore.disconnect()
  isRunning.value = false
  isConnecting.value = false
}

// 重置代码
const resetCode = () => {
  code.value = ''
  terminalStore.clear()
}

// 监听语言变化，切换默认代码
const languageDefaultCode: Record<string, string> = {
  python: 'print("Hello, World!")',
  go: 'package main\n\nimport "fmt"\n\nfunc main() {\n    fmt.Println("Hello, World!")\n}',
  java: 'public class Main {\n    public static void main(String[] args) {\n        System.out.println("Hello, World!");\n    }\n}'
}

// 监听语言变化
const handleLanguageChange = () => {
  code.value = languageDefaultCode[selectedLanguage.value] || ''
}

// 组件挂载时
onMounted(() => {
  // 可以在这里加载用户历史代码等
})

// 组件卸载时
onUnmounted(() => {
  terminalStore.disconnect()
})
</script>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}

.app-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  height: 60px;
  background-color: #1e1e1e;
  color: white;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.app-header h1 {
  margin: 0;
  font-size: 20px;
  font-weight: 500;
}

.language-selector {
  width: 200px;
}

.app-main {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.editor-section {
  flex: 1;
  padding: 10px;
  overflow: hidden;
  background-color: #1e1e1e;
}

.terminal-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 10px;
  overflow: hidden;
  background-color: #1e1e1e;
  border-left: 1px solid #333;
}

.control-panel {
  display: flex;
  gap: 10px;
  padding: 10px 0;
}

@media (max-width: 768px) {
  .app-main {
    flex-direction: column;
  }
  
  .terminal-section {
    border-left: none;
    border-top: 1px solid #333;
  }
}
</style>