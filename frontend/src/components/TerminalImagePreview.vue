<template>
  <!-- 图片预览区域 -->
  <div v-if="images.length" class="absolute bottom-4 right-4 flex flex-col gap-2 max-h-[60%] overflow-y-auto p-2 bg-gray-900/80 backdrop-blur rounded border border-gray-700">
    <div class="text-xs text-gray-400 font-bold px-1">生成结果:</div>
    <div v-for="(img, idx) in images" :key="idx" class="relative group">
      <img 
        :src="img" 
        class="w-40 h-auto object-contain border border-gray-600 rounded bg-white cursor-zoom-in hover:border-cyan-500 transition-colors"
        @click="previewImage(img)"
        title="点击放大"
      />
      <button @click.stop="removeImage(idx)" class="absolute -top-2 -right-2 bg-red-500 text-white rounded-full p-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
        <i class="ph ph-x text-xs"></i>
      </button>
    </div>
  </div>

  <!-- 全屏图片预览模态框 -->
  <Teleport to="body">
    <div v-if="previewUrl" class="fixed inset-0 z-50 flex items-center justify-center bg-black/90 p-10" @click="closePreview">
      <img :src="previewUrl" class="max-w-full max-h-full object-contain rounded shadow-2xl" />
      <button class="absolute top-5 right-5 text-white/50 hover:text-white text-4xl">&times;</button>
    </div>
  </Teleport>
</template>

<script setup>
import { ref } from 'vue';

const props = defineProps({
  images: {
    type: Array,
    required: true
  }
});

const emit = defineEmits(['remove-image']);

const previewUrl = ref(null);

const previewImage = (url) => {
  previewUrl.value = url;
};

const closePreview = () => {
  previewUrl.value = null;
};

const removeImage = (idx) => {
  emit('remove-image', idx);
};
</script>
