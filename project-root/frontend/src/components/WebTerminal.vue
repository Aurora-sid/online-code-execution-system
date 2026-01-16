<template>
  <div class="web-terminal">
    <div ref="terminalContainer" class="terminal-container"></div>
    <div v-if="isConnecting" class="connecting">连接中...</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, defineProps } from 'vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import { useTerminalStore } from '../stores/terminal'

// 定义组件属性
const props = defineProps<{
  isConnecting: boolean
}>()

// 终端相关
const terminalContainer = ref<HTMLElement | null>(null)
let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null

// 状态管理
const terminalStore = useTerminalStore()

// 初始化终端
const initTerminal = () => {
  if (!terminalContainer.value) return

  // 创建终端实例
  terminal = new Terminal({
    fontFamily: 'Monaco, Menlo, "Ubuntu Mono", Consolas, source-code-pro, monospace',
    fontSize: 14,
    lineHeight: 1.4,
    theme: {
      background: '#1e1e1e',
      foreground: '#d4d4d4',
      cursor: '#ffffff',
      selection: '#264f78',
      black: '#000000',
      blue: '#0000ff',
      cyan: '#00ffff',
      green: '#00ff00',
      magenta: '#ff00ff',
      red: '#ff0000',
      white: '#ffffff',
      yellow: '#ffff00',
      brightBlack: '#808080',
      brightBlue: '#0000ff',
      brightCyan: '#00ffff',
      brightGreen: '#00ff00',
      brightMagenta: '#ff00ff',
      brightRed: '#ff0000',
      brightWhite: '#ffffff',
      brightYellow: '#ffff00'
    },
    cursorBlink: true,
    scrollback: 1000
  })

  // 创建并加载fit插件
  fitAddon = new FitAddon()
  terminal.loadAddon(fitAddon)

  // 打开终端
  terminal.open(terminalContainer.value)
  fitAddon.fit()

  // 监听终端数据输入
  terminal.onData((data) => {
    terminalStore.sendData(data)
  })

  // 监听终端调整大小
  window.addEventListener('resize', handleResize)

  // 绑定终端到store
  terminalStore.setTerminal(terminal)
}

// 处理窗口大小变化
const handleResize = () => {
  if (terminal && fitAddon) {
    fitAddon.fit()
  }
}

// 组件挂载时初始化终端
onMounted(() => {
  initTerminal()
})

// 组件卸载前清理资源
onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  
  if (terminal) {
    terminal.dispose()
    terminal = null
  }
  
  if (fitAddon) {
    fitAddon.dispose()
    fitAddon = null
  }
})
</script>

<style scoped>
.web-terminal {
  position: relative;
  width: 100%;
  height: 100%;
  border-radius: 4px;
  overflow: hidden;
  background-color: #1e1e1e;
}

.terminal-container {
  width: 100%;
  height: 100%;
}

.connecting {
  position: absolute;
  top: 10px;
  right: 10px;
  padding: 4px 12px;
  background-color: rgba(0, 0, 0, 0.5);
  color: #ffffff;
  border-radius: 4px;
  font-size: 12px;
  z-index: 10;
}
</style>