<template>
  <Transition name="fade">
    <div v-if="show" class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <div 
        class="w-full max-w-md bg-surface/90 backdrop-blur-xl border border-white/10 rounded-2xl shadow-2xl overflow-hidden transform transition-all"
        @click.stop
      >
        <!-- Header -->
        <div class="px-6 py-4 border-b border-white/10 bg-white/5 flex items-center justify-between">
          <h3 class="text-lg font-bold text-white flex items-center gap-2">
            <svg class="w-5 h-5 text-accent-cyan" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4"/>
            </svg>
            保存文件
          </h3>
          <button @click="$emit('cancel')" class="text-gray-400 hover:text-white transition-colors">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <!-- Body -->
        <div class="p-6 space-y-4">
          <div class="space-y-2">
            <label class="text-sm font-medium text-gray-300">文件名</label>
            <div class="relative group">
              <input 
                v-model="localFilename"
                type="text" 
                ref="inputRef"
                placeholder="请输入文件名"
                class="w-full bg-background/50 border border-white/10 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-accent-cyan/50 focus:ring-2 focus:ring-accent-cyan/20 transition-all"
                @keyup.enter="confirm"
              />
              <div class="absolute right-3 top-1/2 -translate-y-1/2 text-xs font-mono text-gray-500">
                .{{ extension }}
              </div>
            </div>
            <p class="text-xs text-gray-500">提示：如果浏览器支持，将直接弹出系统保存对话框。</p>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 bg-white/5 flex justify-end gap-3">
          <button 
            @click="$emit('cancel')"
            class="px-4 py-2 rounded-lg text-sm font-medium text-gray-400 hover:text-white hover:bg-white/5 transition-all"
          >
            取消
          </button>
          <button 
            @click="confirm"
            class="px-6 py-2 rounded-lg text-sm font-bold text-white bg-gradient-to-r from-accent-cyan to-primary hover:from-primary hover:to-accent-cyan shadow-lg shadow-cyan-500/20 transition-all transform active:scale-95"
          >
            确定保存
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'

const props = defineProps({
  show: Boolean,
  defaultName: String,
  extension: String
})

const emit = defineEmits(['confirm', 'cancel'])

const localFilename = ref('')
const inputRef = ref(null)

watch(() => props.show, (newVal) => {
  if (newVal) {
    localFilename.value = props.defaultName
    nextTick(() => {
      inputRef.value?.focus()
      inputRef.value?.select()
    })
  }
})

const confirm = () => {
  if (!localFilename.value.trim()) return
  emit('confirm', localFilename.value.trim())
}
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
