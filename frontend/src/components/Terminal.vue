<template>
  <div class="flex flex-col bg-gray-900 border-t border-gray-700 h-full">
    <!-- 终端头部 -->
    <div class="h-8 border-b border-gray-700 flex items-center justify-between px-1 flex-shrink-0 bg-gray-800">
      <div class="flex items-center gap-1">
        <div class="pl-0 pr-2.5 py-1 text-xs font-medium rounded text-white border-gray-700 flex items-center gap-3">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 mb-0.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="4 17 10 11 4 5"></polyline>
            <line x1="12" y1="19" x2="20" y2="19"></line>
          </svg>
          终端
        </div>
      </div>
      <div class="flex items-center gap-2">
        <!-- 输入按钮 -->
        <button 
          @click="toggleInput"
          class="px-2 py-1 text-xs font-medium rounded transition-colors flex items-center gap-1"
          :class="inputEnabled 
            ? 'text-cyan-400 bg-cyan-500/20 hover:bg-cyan-500/30' 
            : 'text-gray-400 hover:text-gray-200 hover:bg-gray-700'"
          title="显示/隐藏输入框 (用于需要标准输入的程序)"
        >
          <i class="ph ph-keyboard"></i>
          <span>输入</span>
        </button>
        <!-- 分析报错按钮 -->
        <button 
          @click="analyzeTerminalError"
          :disabled="analyzing || !hasContent"
          class="px-2 py-1 text-xs font-medium rounded transition-colors flex items-center gap-1"
          :class="analyzing 
            ? 'text-purple-400 bg-purple-500/20 cursor-wait' 
            : hasContent 
              ? 'text-orange-400 hover:text-orange-300 hover:bg-orange-500/20' 
              : 'text-gray-600 cursor-not-allowed'"
          :title="hasContent ? '使用 AI 分析终端报错' : '终端无内容可分析'"
        >
          <i :class="analyzing ? 'ph ph-spinner animate-spin' : 'ph ph-bug'"></i>
          <span>{{ analyzing ? '分析中...' : '分析报错' }}</span>
        </button>
        <button 
          @click="clear"
          class="p-1 text-gray-400 hover:text-gray-200 hover:bg-gray-700 rounded transition-colors"
          title="Clear"
        >
          <i class="ph ph-trash text-sm"></i>
        </button>
        <button 
          @click="$emit('toggle-collapse')"
          class="p-1 text-gray-400 hover:text-gray-200 hover:bg-gray-700 rounded transition-colors"
          :title="collapsed ? '展开' : '折叠'"
        >
          <i :class="collapsed ? 'ph ph-caret-up' : 'ph ph-caret-down'" class="text-sm"></i>
        </button>
      </div>
    </div>
    
    <!-- 终端内容 -->
    <div class="relative flex-1 overflow-hidden">
      <div v-show="!collapsed" ref="terminalRef" class="h-full w-full"></div>
      
      <!-- 图片预览子组件 -->
      <TerminalImagePreview 
        :images="images" 
        @remove-image="images.splice($event, 1)" 
      />
    </div>

    <!-- AI 分析弹窗子组件 -->
    <TerminalAnalysis 
      :visible="showAnalysisModal"
      :result="analysisResult"
      :error="analysisError"
      @close="showAnalysisModal = false"
    />

    <!-- 输入区域 -->
    <div v-show="!collapsed && inputEnabled" class="flex-shrink-0 border-t border-gray-700 bg-gray-800/80 p-2">
      <div class="flex items-center gap-2">
        <div class="flex-1 relative">
          <input
            ref="inputRef"
            v-model="inputValue"
            @keydown.enter="submitInput"
            type="text"
            placeholder="输入数据后按 Enter 或点击提交..."
            class="w-full bg-gray-900 border border-gray-600 rounded-lg px-3 py-2 text-sm text-gray-200 placeholder-gray-500 focus:outline-none focus:border-cyan-500 focus:ring-2 focus:ring-cyan-500/20 font-mono"
          />
        </div>
        <button
          @click="submitInput"
          :disabled="!inputValue.trim()"
          class="px-4 py-2 bg-gradient-to-r from-cyan-500 to-blue-500 text-white text-sm font-medium rounded-lg hover:brightness-110 active:scale-95 transition-all disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:brightness-100"
        >
          <i class="ph ph-paper-plane-right mr-1"></i>
          提交
        </button>
      </div>
      <p class="text-xs text-gray-500 mt-1.5 flex items-center gap-1">
        <i class="ph ph-info"></i>
        程序正在等待输入，输入后点击提交或按 Enter 键
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref, shallowRef, computed, onMounted, onUnmounted, watch } from 'vue';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { analyzeError } from '../api';
import TerminalImagePreview from './TerminalImagePreview.vue';
import TerminalAnalysis from './TerminalAnalysis.vue';
import 'xterm/css/xterm.css';

const props = defineProps({
  height: { type: String, default: '200px' },
  fontSize: { type: Number, default: 13 },
  collapsed: { type: Boolean, default: false },
  code: { type: String, default: '' },
  language: { type: String, default: 'python' }
});

const emit = defineEmits(['send-input', 'toggle-collapse']);

const terminalRef = shallowRef(null);
const inputRef = shallowRef(null);
const inputValue = ref('');
const inputEnabled = ref(false);
const images = ref([]);

// 报错分析状态
const analyzing = ref(false);
const analysisResult = ref('');
const analysisError = ref('');
const showAnalysisModal = ref(false);
const terminalContent = ref('');

const hasContent = computed(() => terminalContent.value.trim().length > 0);

let term = null;
let fitAddon = null;
let wheelHandler = null;
let keyHandler = null;
let resizeObserver = null;

onMounted(() => {
  term = new Terminal({
    fontSize: props.fontSize,
    fontFamily: '"JetBrains Mono", monospace',
    theme: {
      background: '#111827',
      foreground: '#f1f5f9',
      cursor: '#06b6d4',
      cursorAccent: '#111827',
      selectionBackground: 'rgba(6, 182, 212, 0.3)',
      black: '#374151',
      red: '#f87171',
      green: '#4ade80',
      yellow: '#fbbf24',
      blue: '#60a5fa',
      magenta: '#c084fc',
      cyan: '#22d3ee',
      white: '#f9fafb',
    },
    cursorBlink: true,
    cursorStyle: 'block',
    cursorInactiveStyle: 'outline',
    allowTransparency: true,
  });

  fitAddon = new FitAddon();
  term.loadAddon(fitAddon);

  if (terminalRef.value) {
    term.open(terminalRef.value);
    
    resizeObserver = new ResizeObserver(() => {
      if (fitAddon && !props.collapsed) {
        requestAnimationFrame(() => fitAddon.fit());
      }
    });
    
    resizeObserver.observe(terminalRef.value);
    setTimeout(() => fitAddon.fit(), 100);
    setTimeout(() => fitAddon.fit(), 300);
  }

  // Ctrl + 滚轮缩放
  wheelHandler = (e) => {
    if (e.ctrlKey) {
      e.preventDefault();
      const delta = e.deltaY > 0 ? -1 : 1;
      const currentSize = term.options.fontSize;
      const newSize = Math.max(8, Math.min(24, currentSize + delta));
      term.options.fontSize = newSize;
      setTimeout(() => fitAddon?.fit(), 50);
    }
  };
  terminalRef.value.addEventListener('wheel', wheelHandler, { passive: false });

  // 快捷键缩放
  keyHandler = (e) => {
    if (e.ctrlKey) {
      let shouldPrevent = false;
      let newSize = null;
      const currentSize = term.options.fontSize;
      
      if (e.key === '=' || e.key === '+') {
        shouldPrevent = true;
        newSize = Math.min(24, currentSize + 1);
      } else if (e.key === '-' || e.key === '_') {
        shouldPrevent = true;
        newSize = Math.max(8, currentSize - 1);
      } else if (e.key === '0') {
        shouldPrevent = true;
        newSize = props.fontSize || 13;
      }
      
      if (shouldPrevent) {
        e.preventDefault();
        if (newSize !== null) {
          term.options.fontSize = newSize;
          setTimeout(() => fitAddon?.fit(), 50);
        }
      }
    }
  };
  terminalRef.value.addEventListener('keydown', keyHandler);
});

watch(() => props.fontSize, (newSize) => {
  if (term) {
    term.options.fontSize = newSize;
    setTimeout(() => fitAddon?.fit(), 50);
  }
});

onUnmounted(() => {
  resizeObserver?.disconnect();
  if (wheelHandler && terminalRef.value) {
    terminalRef.value.removeEventListener('wheel', wheelHandler);
  }
  if (keyHandler && terminalRef.value) {
    terminalRef.value.removeEventListener('keydown', keyHandler);
  }
  term?.dispose();
});

// 写入终端
const write = (text) => {
  const imgRegex = /<<<IMAGE_START>>>>(.*?)<<<<IMAGE_END>>>>/s;
  const match = text.match(imgRegex);

  if (match) {
    const base64Str = match[1];
    const imgSrc = `data:image/png;base64,${base64Str}`;
    images.value.push(imgSrc);
    text = text.replace(match[0], '\n\x1b[36m✨ [System] 成功捕获图片文件: /app/output.png\n   已将其渲染至右下角"生成结果"区域，点击可查看大图。\x1b[0m\n');
  }

  terminalContent.value += text + '\n';
  term?.writeln(text);
  term?.scrollToBottom();
};

const clear = () => {
  term?.clear();
  terminalContent.value = '';
};

const enableInput = () => {
  inputEnabled.value = true;
  setTimeout(() => inputRef.value?.focus(), 100);
};

const disableInput = () => {
  inputEnabled.value = false;
  inputValue.value = '';
};

const submitInput = () => {
  if (!inputValue.value.trim()) return;
  const value = inputValue.value;
  term?.writeln(`\x1b[33m> ${value}\x1b[0m`);
  emit('send-input', value + '\n');
  inputValue.value = '';
};

const toggleInput = () => {
  inputEnabled.value = !inputEnabled.value;
  if (inputEnabled.value) {
    setTimeout(() => inputRef.value?.focus(), 100);
  }
};

const analyzeTerminalError = async () => {
  if (analyzing.value || !terminalContent.value.trim()) return;
  
  analyzing.value = true;
  analysisResult.value = '';
  analysisError.value = '';
  showAnalysisModal.value = true;
  
  try {
    const code = props.code || '';
    const language = props.language || 'unknown';
    const errorOutput = terminalContent.value;
    
    const response = await analyzeError(code, language, errorOutput);
    
    if (response.status === 'success') {
      analysisResult.value = response.result;
    } else {
      analysisError.value = response.error || '分析失败';
    }
  } catch (err) {
    analysisError.value = err.response?.data?.error || err.message || '请求失败，请检查网络连接';
  } finally {
    analyzing.value = false;
  }
};

defineExpose({ write, clear, enableInput, disableInput, toggleInput });
</script>

<style scoped>
:deep(.xterm) {
  padding: 8px 12px;
  height: 100% !important;
  width: 100% !important;
}

:deep(.xterm-viewport) {
  background-color: transparent !important;
  height: 100% !important;
  overflow-y: auto !important;
}

:deep(.xterm-screen) {
  height: 100% !important;
}

:deep(.xterm-rows) {
  height: 100% !important;
}
</style>
