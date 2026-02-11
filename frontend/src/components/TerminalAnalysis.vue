<template>
  <!-- 报错分析结果弹窗 -->
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm p-6" @click.self="close">
      <div class="bg-gray-900 border border-gray-700 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[80vh] flex flex-col overflow-hidden">
        <!-- 弹窗头部 -->
        <div class="flex items-center justify-between px-5 py-4 border-b border-gray-700 bg-gradient-to-r from-orange-500/10 to-purple-500/10">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-orange-500 to-purple-600 flex items-center justify-center">
              <i class="ph ph-bug text-white text-xl"></i>
            </div>
            <div>
              <h3 class="text-lg font-bold text-white">AI 报错分析</h3>
              <p class="text-xs text-gray-400">智能诊断运行错误</p>
            </div>
          </div>
          <button @click="close" class="p-2 rounded-lg hover:bg-gray-700 transition-colors">
            <i class="ph ph-x text-gray-400 hover:text-white text-lg"></i>
          </button>
        </div>
        <!-- 弹窗内容 -->
        <div class="flex-1 overflow-y-auto p-5">
          <div v-if="error" class="p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-300">
            <div class="flex items-center gap-2 mb-2">
              <i class="ph ph-warning-circle text-lg"></i>
              <span class="font-semibold">分析失败</span>
            </div>
            <p class="text-sm">{{ error }}</p>
          </div>
          <div v-else-if="result" class="prose prose-invert prose-sm max-w-none markdown-body" v-html="renderedResult"></div>
          <div v-else class="space-y-6 py-8 px-4">
            <!-- 进度条动画 -->
            <div class="bg-gray-800 rounded-lg p-4 shadow-inner border border-gray-700">
              <div class="flex items-center justify-between mb-3">
                <span class="text-sm font-medium text-gray-300">{{ loadingStage }}</span>
                <span class="text-xs text-purple-400">{{ loadingTip }}</span>
              </div>
              <div class="h-2 bg-gray-700 rounded-full overflow-hidden">
                <div 
                  class="h-full bg-gradient-to-r from-orange-500 to-purple-600 rounded-full transition-all duration-300 relative overflow-hidden"
                  :style="{ width: loadingProgress + '%' }"
                >
                  <div class="absolute inset-0 bg-white/20 animate-[shimmer_2s_infinite]"></div>
                </div>
              </div>
            </div>

            <!-- 骨架屏 -->
            <div class="space-y-4 opacity-50">
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded bg-gray-700 animate-pulse"></div>
                <div class="h-5 w-48 bg-gray-700 rounded animate-pulse"></div>
              </div>
              <div class="space-y-2 pl-11">
                <div class="h-3 w-full bg-gray-700 rounded animate-pulse"></div>
                <div class="h-3 w-5/6 bg-gray-700 rounded animate-pulse"></div>
                <div class="h-3 w-4/6 bg-gray-700 rounded animate-pulse"></div>
              </div>
            </div>

            <!-- 等待提示 -->
            <div class="text-center">
              <p class="text-sm text-gray-400 animate-pulse flex items-center justify-center gap-2">
                <i class="ph ph-robot text-lg"></i>
                {{ funTip }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, onUnmounted } from 'vue';
import MarkdownIt from 'markdown-it';

const md = new MarkdownIt();

const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  result: {
    type: String,
    default: ''
  },
  error: {
    type: String,
    default: ''
  }
});



const emit = defineEmits(['close']);

const loadingProgress = ref(0);
const loadingStage = ref('');
const loadingTip = ref('');
const funTip = ref('');
let progressTimer = null;

// 加载阶段提示
const loadingStages = [
  { progress: 15, stage: '正在读取报错信息...', tip: '解析日志' },
  { progress: 35, stage: '分析错误堆栈...', tip: '定位问题' },
  { progress: 55, stage: 'AI 正在思考解决方案...', tip: '生成建议' },
  { progress: 75, stage: '整理修复方案...', tip: '即将完成' },
  { progress: 90, stage: '优化输出格式...', tip: '最后一步' }
];

// 有趣的等待提示
const funTips = [
  'AI 正在侦探模式下工作 🕵️',
  '正在解开代码的谜团...',
  'Bug 哪里逃！AI 正在追捕中...',
  '正在施展修复魔法 ✨',
  '稍等片刻，答案即将揭晓...'
];

const startProgressAnimation = () => {
  loadingProgress.value = 0;
  let stageIndex = 0;
  funTip.value = funTips[Math.floor(Math.random() * funTips.length)];
  
  if (progressTimer) clearInterval(progressTimer);
  
  progressTimer = setInterval(() => {
    if (stageIndex < loadingStages.length) {
      const target = loadingStages[stageIndex];
      if (loadingProgress.value < target.progress) {
        loadingProgress.value += 2;
        loadingStage.value = target.stage;
        loadingTip.value = target.tip;
      } else {
        stageIndex++;
      }
    } else {
      if (loadingProgress.value < 95) {
        loadingProgress.value += 0.5;
      }
    }
  }, 100);
};

const stopProgressAnimation = () => {
  if (progressTimer) {
    clearInterval(progressTimer);
    progressTimer = null;
  }
};

watch(() => props.visible, (newVal) => {
  if (newVal) {
    if (!props.result && !props.error) {
      startProgressAnimation();
    }
  } else {
    stopProgressAnimation();
  }
});

watch(() => props.result, (newVal) => {
  if (newVal) stopProgressAnimation();
});

watch(() => props.error, (newVal) => {
  if (newVal) stopProgressAnimation();
});

onUnmounted(() => {
  stopProgressAnimation();
});

const renderedResult = computed(() => {
  if (!props.result) return '';
  return md.render(props.result);
});

const close = () => {
  emit('close');
};
</script>

<style scoped>
/* Markdown 样式（暗色主题） */
.markdown-body {
  font-size: 14px;
  line-height: 1.6;
  color: #e5e7eb;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3) {
  margin-top: 16px;
  margin-bottom: 8px;
  font-weight: 600;
  color: #f9fafb;
}

.markdown-body :deep(code) {
  background: #374151;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
  color: #fbbf24;
}

.markdown-body :deep(pre) {
  background: #1f2937;
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  border: 1px solid #374151;
}

.markdown-body :deep(pre code) {
  background: transparent;
  padding: 0;
  color: #e5e7eb;
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

.markdown-body :deep(strong) {
  color: #f9fafb;
}
</style>
