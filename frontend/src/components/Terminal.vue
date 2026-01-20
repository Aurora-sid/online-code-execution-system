<template>
  <div ref="terminalContainer" class="w-full h-full pl-2 pb-2"></div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'

const terminalContainer = ref(null)
let term = null
let fitAddon = null

onMounted(() => {
  term = new Terminal({
    cursorBlink: true,
    cursorStyle: 'bar',
    cursorInactiveStyle: 'bar',
    fontFamily: '"JetBrains Mono", Consolas, monospace',
    fontSize: 14,
    theme: {
      background: '#00000000', // Transparent
      foreground: '#F1F5F9'
    },
    convertEol: true,
  })
  
  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  
  term.open(terminalContainer.value)
  
  // 使用setTimeout确保容器具有正确的尺寸
  setTimeout(() => {
    fitAddon.fit()
  }, 100)
  
  term.writeln('欢迎使用在线代码执行器')
  term.writeln('等待输出...')
  
  // 在窗口调整大小时重新适应
  window.addEventListener('resize', () => {
    if (fitAddon) fitAddon.fit()
  })
})

const write = (text) => {
  if (term) term.write(text)
}

const clear = () => {
  if (term) term.clear()
}

defineExpose({ write, clear })

onBeforeUnmount(() => {
  if (term) term.dispose()
})
</script>

<style scoped>
/* Force cursor to blink even when not focused */
:deep(.xterm-cursor) {
  animation: xterm-cursor-blink 1s step-end infinite !important;
}

@keyframes xterm-cursor-blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}
</style>
