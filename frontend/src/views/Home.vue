<template>
  <div class="min-h-screen bg-base-100">
    <!-- Mini Hero with Quote -->
    <div class="bg-base-200/50 border-b border-base-300">
      <div class="max-w-4xl mx-auto px-4 py-6">
        <div class="text-center">
          <h1 class="text-2xl font-light text-base-content/80 mb-2">oinote</h1>
          <div class="flex items-center justify-center gap-2">
            <p class="text-sm text-base-content/60 italic">"{{ currentQuote }}"</p>
            <button
              @click="refreshQuote"
              class="btn btn-ghost btn-xs btn-circle opacity-50 hover:opacity-100 transition-opacity"
              title="换一句"
            >
              <RefreshCw class="w-3 h-3" />
            </button>
          </div>
          <router-link to="/register" v-if="!authStore.isAuthenticated" class="btn btn-neutral btn-sm mt-3">开始使用</router-link>
        </div>
      </div>
    </div>

    <!-- Main Content -->
    <div class="max-w-4xl mx-auto px-4 py-6 space-y-8 pb-20 lg:pb-6">
      <!-- Notes Section -->
      <section id="notes">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-xl font-semibold text-base-content">笔记</h2>
          <router-link v-if="authStore.isAuthenticated" to="/note" class="btn btn-outline btn-sm">新建笔记</router-link>
        </div>
        <div v-if="publicNotes.length === 0" class="text-center py-8 text-base-content/40">
          <div class="text-sm">暂无公开笔记</div>
        </div>
        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <router-link 
            v-for="note in publicNotes" 
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
              <div class="text-xs text-base-content/70 mb-2" v-if="note.content" style="white-space: pre-wrap; word-break: break-word; display: -webkit-box; -webkit-line-clamp: 4; -webkit-box-orient: vertical; overflow: hidden;">
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
      </section>

      <!-- Channels Section -->
      <section id="channels">
        <div class="flex items-center justify-between mb-4">
                  <h2 class="text-xl font-semibold text-base-content">频道</h2>
                  <button
                    v-if="authStore.isAuthenticated"
                    @click="$emit('create-channel')"
                    class="btn btn-outline btn-sm"
                  >
                    新建频道
                  </button>
                </div>        <div v-if="publicChannels.length === 0" class="text-center py-8 text-base-content/40">
          <div class="text-sm">暂无公开频道</div>
        </div>
        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div
            v-for="ch in publicChannels"
            :key="ch.id"
            class="card bg-base-100 border border-base-300 shadow-sm hover:shadow-md transition-shadow hover:border-base-400"
          >
            <div class="card-body p-4">
              <div class="flex items-start gap-2 mb-2">
                <div class="w-8 h-8 bg-neutral text-neutral-content rounded-full flex items-center justify-center text-xs font-bold shrink-0">
                  {{ ch.name.charAt(0).toUpperCase() }}
                </div>
                <div class="flex-1 min-w-0">
                  <h3 class="font-medium text-base-content mb-1 truncate text-sm">{{ ch.name }}</h3>
                  <div class="flex items-center gap-2 text-xs text-base-content/50">
                    <span>{{ ch.owner?.nickname || ch.owner?.username || '未知' }}</span>
                    <span class="shrink-0">{{ formatDate(ch.created_at) }}</span>
                  </div>
                </div>
              </div>
              <p class="text-sm text-base-content/70 mb-2 line-clamp-2">{{ ch.description || '暂无描述' }}</p>
              <div v-if="ch.tags" class="flex flex-wrap gap-1 mb-2">
                <span
                  v-for="tag in parseTags(ch.tags)"
                  :key="tag"
                  class="badge badge-ghost badge-xs text-xs"
                >
                  {{ tag }}
                </span>
              </div>
              <div class="flex gap-2">
                <router-link
                  v-if="ch.is_member || !authStore.isAuthenticated"
                  :to="{ name: 'channel', params: { id: ch.id } }"
                  class="btn btn-sm btn-neutral flex-1"
                >
                  进入频道
                </router-link>
                <button
                  v-else-if="ch.is_pending"
                  class="btn btn-sm btn-ghost flex-1"
                  disabled
                >
                  待审核
                </button>
                <button
                  v-else-if="ch.is_invited"
                  @click="acceptInvitation(ch.id)"
                  class="btn btn-sm btn-primary flex-1"
                >
                  接受邀请
                </button>
                <button
                  v-else
                  @click="joinChannel(ch.id)"
                  class="btn btn-sm btn-neutral flex-1"
                >
                  申请加入
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue';
import api from '../api/axios';
import { useAuthStore } from '../stores/auth';
import wsClient from '../utils/websocket';
import { RefreshCw } from 'lucide-vue-next';

// Define emits
const emit = defineEmits(['create-channel']);

const authStore = useAuthStore();
const publicChannels = ref([]);
const publicNotes = ref([]);
const quotes = ref([]);
const currentQuote = ref('');

// Load quotes from the text file
const loadQuotes = async () => {
  try {
    const response = await fetch('/src/data/quotes.txt');
    const text = await response.text();
    quotes.value = text.split('\n').filter(line => line.trim());
    // Select a random quote
    if (quotes.value.length > 0) {
      const randomIndex = Math.floor(Math.random() * quotes.value.length);
      currentQuote.value = quotes.value[randomIndex];
    }
  } catch (error) {
    console.error('Failed to load quotes:', error);
    // Fallback quote
    currentQuote.value = '记录思想，连接灵感';
  }
};

// Refresh quote function
const refreshQuote = () => {
  if (quotes.value.length > 0) {
    const randomIndex = Math.floor(Math.random() * quotes.value.length);
    currentQuote.value = quotes.value[randomIndex];
  }
};

// Load public data
const loadPublicData = async () => {
  try {
    const [chRes, noteRes] = await Promise.all([
      api.get('/public/channels'),
      api.get('/public/notes')
    ]);
    publicChannels.value = chRes.data || [];
    publicNotes.value = noteRes.data || [];
  } catch (error) {
    console.error('Failed to fetch public data:', error);
    publicChannels.value = [];
    publicNotes.value = [];
  }
};

// Utility functions
const parseTags = (value) =>
  String(value || '')
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0);

const stripHtml = (html) => {
  const div = document.createElement('div');
  div.innerHTML = html;
  
  // 将块级元素转换为换行符
  const blockElements = div.querySelectorAll('p, div, li, h1, h2, h3, h4, h5, h6, br');
  blockElements.forEach(el => {
    el.innerHTML = '\n' + el.innerHTML + '\n';
  });
  
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

// 申请加入频道
const joinChannel = async (channelId) => {
  try {
    await api.post(`/channels/${channelId}/join`);
    alert('申请已提交，请等待管理员审核');
    await loadPublicData();
  } catch (error) {
    console.error('Failed to join channel:', error);
    alert(error.response?.data?.error || '申请失败');
  }
};

// 接受邀请
const acceptInvitation = async (channelId) => {
  try {
    // 查找邀请记录
    const res = await api.get('/channels/approvals/pending');
    const invitation = res.data.find(item => item.channel?.id === channelId && item.status === 'invited');
    if (invitation) {
      await api.post(`/channels/approvals/${invitation.id}/accept`);
      alert('已成功加入频道');
      await loadPublicData();
    }
  } catch (error) {
    console.error('Failed to accept invitation:', error);
    alert(error.response?.data?.error || '接受邀请失败');
  }
};

// WebSocket 消息处理
const handleWsMessage = async (message) => {
  if (message.type === 'channel') {
    if (message.action === 'create' || message.action === 'update') {
      // 重新加载公开频道列表
      await loadPublicData();
    } else if (message.action === 'delete') {
      // 删除已删除的频道
      publicChannels.value = publicChannels.value.filter(c => {
        const channelId = String(c.id);
        const dataId = String(message.data.id);
        return channelId !== dataId;
      });
    }
  } else if (message.type === 'note') {
    if (message.action === 'create' || message.action === 'update') {
      // 重新加载公开笔记列表
      await loadPublicData();
    } else if (message.action === 'delete') {
      // 删除已删除的笔记
      publicNotes.value = publicNotes.value.filter(n => {
        const noteId = String(n.id);
        const dataId = String(message.data.id);
        return noteId !== dataId;
      });
    }
  }
};

onMounted(async () => {
  // Load quotes first
  await loadQuotes();

  // Load public data
  await loadPublicData();

  // 设置 WebSocket 监听器的函数
  const setupWebSocketListeners = () => {
    if (localStorage.getItem('userId')) {
      wsClient.on('channel_create', handleWsMessage);
      wsClient.on('channel_update', handleWsMessage);
      wsClient.on('channel_delete', handleWsMessage);
      wsClient.on('note_create', handleWsMessage);
      wsClient.on('note_update', handleWsMessage);
      wsClient.on('note_delete', handleWsMessage);
    }
  };

  // 如果已经有 userId，立即设置监听器
  setupWebSocketListeners();

  // 监听 WebSocket 连接成功事件
  wsClient.on('connected', setupWebSocketListeners);
});

onBeforeUnmount(() => {
  // 清除 WebSocket 监听器
  wsClient.off('channel_create', handleWsMessage);
  wsClient.off('channel_update', handleWsMessage);
  wsClient.off('channel_delete', handleWsMessage);
  wsClient.off('note_create', handleWsMessage);
  wsClient.off('note_update', handleWsMessage);
  wsClient.off('note_delete', handleWsMessage);
  wsClient.off('connected', () => {}); // 移除 connected 事件监听器
});
</script>
