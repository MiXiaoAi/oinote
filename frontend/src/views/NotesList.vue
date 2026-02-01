<template>
  <div class="bg-base-100">
    <div class="max-w-4xl mx-auto px-4 py-6 pb-20 lg:pb-6">
      <div class="flex items-center justify-between mb-6">
        <h1 class="text-2xl font-bold">笔记</h1>
        <router-link v-if="authStore.isAuthenticated" to="/note" class="btn btn-neutral btn-sm">
          新建笔记
        </router-link>
      </div>

      <div v-if="loading" class="text-center py-12 text-base-content/50">
        <div class="text-sm">加载中...</div>
      </div>

      <div v-else-if="notes.length === 0" class="text-center py-12 text-base-content/40">
        <div class="text-sm">{{ authStore.isAuthenticated ? '暂无笔记' : '暂无公开笔记' }}</div>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <router-link
          v-for="note in notes"
          :key="note.id"
          :to="{ name: 'note-editor', params: { id: note.id } }"
          class="card bg-base-100 border border-base-300 shadow-sm hover:shadow-md transition-shadow cursor-pointer hover:border-base-400"
        >
          <div class="card-body p-4">
            <div class="flex items-start justify-between mb-2">
              <div class="flex-1 min-w-0">
                <h3 class="font-medium text-base-content mb-1 truncate text-sm">{{ note.title || '无标题' }}</h3>
                <div class="flex items-center gap-2 text-xs text-base-content/50">
                  <span class="truncate">{{ note.owner?.nickname || note.owner?.username || '未知' }}</span>
                  <span class="shrink-0">{{ formatDate(note.created_at) }}</span>
                </div>
              </div>
            </div>
            <div class="text-xs text-base-content/70 line-clamp-2 mb-2" v-if="note.content">
              {{ stripHtml(note.content) }}
            </div>
            <div v-if="note.tags" class="flex flex-wrap gap-1">
              <span
                v-for="tag in parseTags(note.tags)"
                :key="tag"
                class="badge badge-ghost badge-xs text-xs"
              >
                {{ tag }}
              </span>
            </div>
          </div>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue';
import api from '../api/axios';
import { useAuthStore } from '../stores/auth';
import wsClient from '../utils/websocket';

const authStore = useAuthStore();
const notes = ref([]);
const loading = ref(true);

const loadNotes = async () => {
  loading.value = true;
  try {
    if (authStore.isAuthenticated) {
      const res = await api.get('/notes', { params: { channel_id: 0 } });
      notes.value = res.data || [];
    } else {
      const res = await api.get('/public/notes');
      notes.value = res.data || [];
    }
  } catch (error) {
    console.error('Failed to fetch notes:', error);
    notes.value = [];
  } finally {
    loading.value = false;
  }
};

const parseTags = (value) =>
  String(value || '')
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0);

const stripHtml = (html) => {
  const div = document.createElement('div');
  div.innerHTML = html;
  return div.textContent || div.innerText || '';
};

const formatDate = (dateString) => {
  if (!dateString) return '';
  const date = new Date(dateString);
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  });
};

const handleWsMessage = async (message) => {
  if (message.type === 'note') {
    if (message.action === 'create' || message.action === 'update') {
      await loadNotes();
    } else if (message.action === 'delete') {
      notes.value = notes.value.filter(n => {
        const noteId = String(n.id);
        const dataId = String(message.data.id);
        return noteId !== dataId;
      });
    }
  }
};

onMounted(async () => {
  await loadNotes();

  const setupWebSocketListeners = () => {
    if (localStorage.getItem('userId')) {
      wsClient.on('note_create', handleWsMessage);
      wsClient.on('note_update', handleWsMessage);
      wsClient.on('note_delete', handleWsMessage);
    }
  };

  setupWebSocketListeners();
  wsClient.on('connected', setupWebSocketListeners);
});

onBeforeUnmount(() => {
  wsClient.off('note_create', handleWsMessage);
  wsClient.off('note_update', handleWsMessage);
  wsClient.off('note_delete', handleWsMessage);
  wsClient.off('connected', () => {});
});
</script>