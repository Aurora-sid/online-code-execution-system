<template>
  <div ref="terminalContainer" class="w-full h-full bg-black p-2 rounded-lg overflow-hidden"></div>
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
    fontFamily: 'Consolas, monospace',
    fontSize: 14,
    theme: {
      background: '#1e1e1e'
    },
    convertEol: true,
  })
  
  fitAddon = new FitAddon()
  term.loadAddon(fitAddon)
  
  term.open(terminalContainer.value)
  
  // Use setTimeout to ensure container has proper dimensions
  setTimeout(() => {
    fitAddon.fit()
  }, 100)
  
  term.writeln('Welcome to Online Code Executor')
  term.writeln('Waiting for output...')
  
  // Refit on window resize
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
