<template>
  <div v-if="isOpen" class="modal modal-open z-50" @click.self="close">
    <div class="modal-box max-w-5xl p-4" :class="mediaType === 'image' ? 'max-h-[90vh]' : ''">
      <div class="flex justify-between items-center mb-4">
        <h3 class="font-bold text-lg">
          {{ mediaType === 'image' ? '图片查看器' : mediaType === 'video' ? '视频播放器' : '音频播放器' }}
        </h3>
        <button @click="close" class="btn btn-sm btn-circle btn-ghost">✕</button>
      </div>

      <div class="relative" :class="mediaType === 'image' ? 'flex-1 min-h-0' : ''">
        <!-- 图片查看器 -->
        <div v-if="mediaType === 'image'" class="flex items-center justify-center min-h-[60vh]">
          <button @click="prevImage" class="absolute left-2 btn btn-circle btn-ghost z-10">❮</button>

          <img
            :src="currentUrl"
            :alt="currentFileName"
            class="max-w-full max-h-[70vh] object-contain"
            @wheel="handleWheel"
            :style="{ transform: `scale(${zoom})` }"
          />

          <button @click="nextImage" class="absolute right-2 btn btn-circle btn-ghost z-10">❯</button>
        </div>

        <!-- 视频播放器 -->
        <div v-else-if="mediaType === 'video'" class="flex flex-col gap-4">
          <div class="relative bg-black rounded-lg overflow-hidden">
            <video
              :src="currentUrl"
              class="w-full max-h-[70vh]"
              ref="videoPlayer"
              controls
              @loadedmetadata="onVideoLoaded"
            ></video>
          </div>

          <!-- 文件信息 -->
          <div class="text-center text-sm text-base-content/60">
            <p class="font-semibold text-base-content">{{ currentFileName || '视频文件' }}</p>
          </div>
        </div>

        <!-- 音频播放器 -->
        <div v-else-if="mediaType === 'audio'" class="flex flex-col items-center gap-4 py-8">
          <div class="w-32 h-32 rounded-full bg-base-300 flex items-center justify-center">
            <span class="text-4xl">🎵</span>
          </div>

          <div class="text-center">
            <p class="font-semibold">{{ currentFileName || '音频文件' }}</p>
          </div>

          <audio
            :src="currentUrl"
            controls
            preload="metadata"
            class="w-full max-w-lg"
          ></audio>
        </div>
      </div>

      <!-- 图片控制栏 -->
      <div v-if="mediaType === 'image' && imageList.length > 1" class="flex justify-center items-center gap-4 mt-4">
        <span class="text-sm text-base-content/60">
          {{ currentIndex + 1 }} / {{ imageList.length }}
        </span>
        <div class="flex gap-2">
          <button @click="zoomOut" class="btn btn-xs btn-ghost">缩小</button>
          <button @click="resetZoom" class="btn btn-xs btn-ghost">重置</button>
          <button @click="zoomIn" class="btn btn-xs btn-ghost">放大</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import { getFileUrl } from '../utils/urlHelper';

const props = defineProps({
  isOpen: Boolean,
  mediaList: Array,
  startIndex: Number,
  currentMedia: Object
});

const emit = defineEmits(['close']);

const zoom = ref(1);
const currentIndex = ref(0);
const videoPlayer = ref(null);
const duration = ref(0);

// 键盘事件处理
const handleKeyDown = (e) => {
  if (!props.isOpen || !videoPlayer.value || mediaType.value !== 'video') return;

  const video = videoPlayer.value;
  const step = 5; // 进度步长（秒）
  const volumeStep = 0.1; // 音量步长

  switch(e.key) {
    case 'ArrowLeft':
      e.preventDefault();
      video.currentTime = Math.max(0, video.currentTime - step);
      break;
    case 'ArrowRight':
      e.preventDefault();
      video.currentTime = Math.min(video.duration, video.currentTime + step);
      break;
    case 'ArrowUp':
      e.preventDefault();
      video.volume = Math.min(1, video.volume + volumeStep);
      break;
    case 'ArrowDown':
      e.preventDefault();
      video.volume = Math.max(0, video.volume - volumeStep);
      break;
  }
};

onMounted(() => {
  window.addEventListener('keydown', handleKeyDown);
});

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeyDown);
});

// 视频事件处理
const onVideoLoaded = () => {
  if (videoPlayer.value) {
    duration.value = videoPlayer.value.duration;
  }
};

// 根据当前媒体确定类型
const mediaType = computed(() => {
  if (!props.currentMedia) return null;
  const filePath = props.currentMedia.file_path || '';
  if (/\.(png|jpe?g|gif|webp|bmp)$/i.test(filePath)) return 'image';
  if (/\.(mp4|webm|ogg|mov)$/i.test(filePath)) return 'video';
  if (/\.(mp3|wav|ogg|m4a|flac)$/i.test(filePath)) return 'audio';
  return null;
});

// 当前显示的URL
const currentUrl = computed(() => {
  if (!props.currentMedia) return '';
  return getFileUrl(props.currentMedia.file_path);
});

// 当前文件名
const currentFileName = computed(() => {
  return props.currentMedia?.file_name || '';
});

// 图片列表
const imageList = computed(() => {
  if (!props.mediaList) return [];
  return props.mediaList.filter(m => {
    const path = m.file_path || '';
    return /\.(png|jpe?g|gif|webp|bmp)$/i.test(path);
  });
});

// 重置缩放
const resetZoom = () => {
  zoom.value = 1;
};

// 放大
const zoomIn = () => {
  zoom.value = Math.min(zoom.value + 0.25, 5);
};

// 缩小
const zoomOut = () => {
  zoom.value = Math.max(zoom.value - 0.25, 0.25);
};

// 鼠标滚轮缩放
const handleWheel = (e) => {
  e.preventDefault();
  if (e.deltaY < 0) {
    zoomIn();
  } else {
    zoomOut();
  }
};

// 上一张图片
const prevImage = () => {
  currentIndex.value = (currentIndex.value - 1 + imageList.value.length) % imageList.value.length;
};

// 下一张图片
const nextImage = () => {
  currentIndex.value = (currentIndex.value + 1) % imageList.value.length;
};

// 关闭
const close = () => {
  resetZoom();
  if (videoPlayer.value) {
    videoPlayer.value.pause();
  }
  emit('close');
};

// 监听打开状态，重置缩放
watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    resetZoom();
    currentIndex.value = props.startIndex || 0;
  }
});
</script>

<style scoped>

img {

  transition: transform 0.2s ease;

}

</style>
