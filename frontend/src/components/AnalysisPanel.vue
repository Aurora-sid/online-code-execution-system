<template>
  <div class="analysis-panel">
    <!-- 头部 -->
    <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200">
      <div class="flex items-center gap-2">
        <i class="ph ph-sparkle text-purple-600 text-lg"></i>
        <span class="font-semibold text-gray-700">智能分析</span>
      </div>
      <button 
        @click="$emit('close')" 
        class="p-1.5 rounded-lg hover:bg-gray-100 transition-colors"
        title="关闭"
      >
        <i class="ph ph-x text-gray-500 hover:text-gray-700"></i>
      </button>
    </div>

    <!-- 分析类型选择 -->
    <div class="flex gap-2 p-3 border-b border-gray-200">
      <button 
        v-for="option in analysisTypes" 
        :key="option.value"
        @click="selectedType = option.value"
        :class="[
          'px-3 py-1.5 rounded-lg text-sm font-medium transition-all',
          selectedType === option.value 
            ? 'bg-purple-600 text-white' 
            : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
        ]"
      >
        <i :class="option.icon" class="mr-1"></i>
        {{ option.label }}
      </button>
    </div>

    <!-- 分析按钮 -->
    <div class="p-3 border-b border-gray-200">
      <button 
        @click="runAnalysis"
        :disabled="loading || !code"
        class="w-full py-2.5 px-4 rounded-lg bg-gradient-to-r from-purple-600 to-indigo-600 text-white font-semibold shadow-lg hover:shadow-xl hover:brightness-110 transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
      >
        <i v-if="loading" class="ph ph-spinner animate-spin text-lg"></i>
        <i v-else class="ph ph-magic-wand text-lg"></i>
        {{ loading ? '分析中...' : '开始分析' }}
      </button>
    </div>

    <!-- 结果展示区 -->
    <div class="flex-1 overflow-y-auto p-4 custom-scrollbar">
      <!-- 错误提示 -->
      <div v-if="error" class="p-4 rounded-lg bg-red-50 border border-red-200 text-red-700">
        <div class="flex items-center gap-2 mb-2">
          <i class="ph ph-warning-circle text-lg"></i>
          <span class="font-semibold">分析失败</span>
        </div>
        <p class="text-sm">{{ error }}</p>
      </div>

      <!-- 加载动画 - 优化后的等待体验 -->
      <div v-else-if="loading" class="space-y-4">
        <!-- 进度条动画 -->
        <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100">
          <div class="flex items-center justify-between mb-3">
            <span class="text-sm font-medium text-gray-700">{{ loadingStage }}</span>
            <span class="text-xs text-purple-600">{{ loadingTip }}</span>
          </div>
          <div class="h-2 bg-gray-200 rounded-full overflow-hidden">
            <div 
              class="h-full bg-gradient-to-r from-purple-500 to-indigo-500 rounded-full transition-all duration-300"
              :style="{ width: loadingProgress + '%' }"
            ></div>
          </div>
        </div>

        <!-- 骨架屏模拟分析结果 -->
        <div class="bg-white rounded-lg p-4 shadow-sm border border-gray-100 space-y-3">
          <div class="flex items-center gap-2">
            <div class="w-5 h-5 rounded bg-purple-200 animate-pulse"></div>
            <div class="h-4 w-32 bg-gray-200 rounded animate-pulse"></div>
          </div>
          <div class="space-y-2 pl-7">
            <div class="h-3 w-full bg-gray-100 rounded animate-pulse"></div>
            <div class="h-3 w-5/6 bg-gray-100 rounded animate-pulse"></div>
            <div class="h-3 w-4/6 bg-gray-100 rounded animate-pulse"></div>
          </div>
          <div class="flex items-center gap-2 mt-4">
            <div class="w-5 h-5 rounded bg-indigo-200 animate-pulse"></div>
            <div class="h-4 w-40 bg-gray-200 rounded animate-pulse"></div>
          </div>
          <div class="space-y-2 pl-7">
            <div class="h-3 w-full bg-gray-100 rounded animate-pulse"></div>
            <div class="h-3 w-3/4 bg-gray-100 rounded animate-pulse"></div>
          </div>
        </div>

        <!-- 等待提示 -->
        <div class="text-center text-sm text-gray-500 animate-pulse">
          <i class="ph ph-robot text-lg mr-1"></i>
          {{ funTip }}
        </div>
      </div>

      <!-- 分析结果 -->
      <div v-else-if="result" class="prose prose-sm prose-invert max-w-none">
        <div class="markdown-body" v-html="renderedResult"></div>
      </div>

      <!-- 空状态 -->
      <div v-else class="text-center py-12 text-gray-400">
        <i class="ph ph-robot text-5xl mb-4"></i>
        <p>选择分析类型并点击"开始分析"</p>
        <p class="text-sm mt-2">AI 将帮助你检测问题、优化代码</p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onUnmounted } from 'vue'
import MarkdownIt from 'markdown-it'
import { analyzeCode } from '../api'

const md = new MarkdownIt()

const props = defineProps({
  code: { type: String, default: '' },
  language: { type: String, default: 'python' }
})

const emit = defineEmits(['close'])

const loading = ref(false)
const result = ref('')
const error = ref('')
const selectedType = ref('bug')
const loadingProgress = ref(0)
const loadingStage = ref('')
const loadingTip = ref('')
const funTip = ref('')

// 进度动画定时器
let progressTimer = null

const analysisTypes = [
  { value: 'bug', label: 'Bug 检测', icon: 'ph ph-bug' },
  { value: 'optimize', label: '优化建议', icon: 'ph ph-lightning' },
  { value: 'explain', label: '代码解释', icon: 'ph ph-book-open' },
  { value: 'security', label: '安全检查', icon: 'ph ph-shield-check' }
]

// 加载阶段提示
const loadingStages = [
  { progress: 15, stage: '正在解析代码结构...', tip: '理解语法' },
  { progress: 35, stage: '分析代码逻辑...', tip: '深度分析' },
  { progress: 55, stage: 'AI 正在思考...', tip: '生成建议' },
  { progress: 75, stage: '整理分析结果...', tip: '即将完成' },
  { progress: 90, stage: '优化输出格式...', tip: '最后一步' }
]

// 有趣的等待提示
const funTips = [
  'AI 正在认真阅读你的代码...',
  '让 AI 帮你找出隐藏的 Bug~',
  '好代码值得仔细分析...',
  '正在施展代码魔法 ✨',
  'AI 正在努力工作中...'
]

// 使用 markdown-it 渲染 Markdown
const renderedResult = computed(() => {
  if (!result.value) return ''
  return md.render(result.value)
})

// 代码变化时清空结果
watch(() => props.code, () => {
  result.value = ''
  error.value = ''
})

// 启动模拟进度动画
const startProgressAnimation = () => {
  loadingProgress.value = 0
  let stageIndex = 0
  funTip.value = funTips[Math.floor(Math.random() * funTips.length)]
  
  progressTimer = setInterval(() => {
    if (stageIndex < loadingStages.length) {
      const target = loadingStages[stageIndex]
      if (loadingProgress.value < target.progress) {
        loadingProgress.value += 2
        loadingStage.value = target.stage
        loadingTip.value = target.tip
      } else {
        stageIndex++
      }
    } else {
      // 最后阶段缓慢增加
      if (loadingProgress.value < 95) {
        loadingProgress.value += 0.5
      }
    }
  }, 100)
}

// 停止进度动画
const stopProgressAnimation = () => {
  if (progressTimer) {
    clearInterval(progressTimer)
    progressTimer = null
  }
  loadingProgress.value = 100
}

const runAnalysis = async () => {
  if (!props.code || loading.value) return

  loading.value = true
  error.value = ''
  result.value = ''
  startProgressAnimation()

  try {
    const response = await analyzeCode(props.code, props.language, selectedType.value)
    stopProgressAnimation()
    
    if (response.status === 'success') {
      result.value = response.result
    } else {
      error.value = response.error || '分析失败'
    }
  } catch (err) {
    stopProgressAnimation()
    error.value = err.response?.data?.error || err.message || '请求失败'
  } finally {
    loading.value = false
  }
}

onUnmounted(() => {
  stopProgressAnimation()
})
</script>

<style scoped>
.analysis-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #f9fafb;
}

.markdown-body {
  font-size: 14px;
  line-height: 1.6;
  color: #374151;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  margin-top: 16px;
  margin-bottom: 8px;
  font-weight: 600;
  color: #111827;
}

.markdown-body :deep(code) {
  background: #f3f4f6;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
  color: #1f2937;
}

.markdown-body :deep(pre) {
  background: #1f2937;
  color: #f3f4f6;
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
}

.markdown-body :deep(pre code) {
  background: transparent;
  padding: 0;
  color: inherit;
}

/* 滚动条样式 */
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background-color: #d1d5db;
  border-radius: 3px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background-color: #9ca3af;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  padding-left: 1.5em;
  margin: 8px 0;
}

.markdown-body :deep(li) {
  margin: 4px 0;
}

.markdown-body :deep(p) {
  margin: 8px 0;
}
</style>
